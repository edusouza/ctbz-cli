// Package cli define os comandos do ctbz (árvore Cobra). O pacote main só chama Execute;
// o gerador de documentação (tools/gendocs) usa NewRootCmd para ler a mesma árvore.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/edusouza/ctbz-cli/internal/output"
)

// Códigos de saída do processo.
const (
	ExitOK      = 0
	ExitError   = 1
	ExitUsage   = 2
	ExitPending = 3
	// ExitAttention indica que o comando funcionou, mas há algo que exige atenção
	// (ex.: guias em atraso com --fail-on-atraso).
	ExitAttention = 4
)

// errAttention sinaliza, sem mensagem de erro, que há algo que exige atenção.
var errAttention = errors.New("atenção")

// errPending sinaliza que o login foi salvo aguardando uma entrada do usuário.
var errPending = errors.New("login pendente")

// usageError marca erros de uso (flag inválida, argumentos faltando).
type usageError struct{ error }

// Execute roda a CLI e devolve o código de saída.
func Execute(ctx context.Context, version string, args []string) int {
	root := NewRootCmd(version)
	root.SetArgs(args)
	return run(ctx, root)
}

// run executa a árvore já configurada e traduz o erro em código de saída.
func run(ctx context.Context, root *cobra.Command) int {
	err := root.ExecuteContext(ctx)
	switch {
	case err == nil:
		return ExitOK
	case errors.Is(err, errPending):
		return ExitPending
	case errors.Is(err, errAttention):
		return ExitAttention
	}
	fmt.Fprintln(root.ErrOrStderr(), "erro:", err)
	var ue usageError
	if errors.As(err, &ue) || strings.HasPrefix(err.Error(), "unknown command") {
		return ExitUsage
	}
	return ExitError
}

// NewRootCmd monta a árvore de comandos.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "ctbz",
		Short: "CLI para a Contabilizei",
		Long: `ctbz opera a Contabilizei pela linha de comando: faz o mesmo login do site
(usuário, senha e código enviado por e-mail) e consulta os dados da empresa.

Dados vão para stdout no formato escolhido com -o; mensagens e progresso vão para stderr.`,
		Version:       resolveVersion(version).Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringP("output", "o", "", "formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)")
	root.PersistentFlags().BoolP("help", "h", false, "mostra a ajuda do comando")
	root.Flags().BoolP("version", "v", false, "mostra a versão do ctbz")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError{err} })
	root.CompletionOptions.HiddenDefaultCmd = true
	root.SetUsageTemplate(usageTemplate)
	root.SetVersionTemplate("ctbz {{.Version}}\n")
	root.SetHelpCommand(&cobra.Command{
		Use:    "help [comando]",
		Short:  "Mostra a ajuda de um comando",
		Hidden: true,
		Run: func(c *cobra.Command, args []string) {
			cmd, _, err := c.Root().Find(args)
			if cmd == nil || err != nil {
				cmd = c.Root()
			}
			cmd.Help()
		},
	})

	root.AddCommand(
		newLoginCmd(),
		newStatusCmd(),
		newEmpresaCmd(),
		newEmpresasCmd(),
		newImpostosCmd(),
		newPendenciasCmd(),
		newRotinasCmd(),
		newPrimeirosPassosCmd(),
		newChamadosCmd(),
		newResumoCmd(),
		newMensalidadeCmd(),
		newPlanoCmd(),
		newNotasCmd(),
		newProlaboreCmd(),
		newLucrosCmd(),
		newBalanceteCmd(),
		newBalancoCmd(),
		newRazaoCmd(),
		newCaixaCmd(),
		newExtratosCmd(),
		newContasBancariasCmd(),
		newContasCmd(),
		newDocumentosCmd(),
		newCertificadoCmd(),
		newUsuariosCmd(),
		newContaCmd(),
		newCorrespondenciasCmd(),
		newAcoesCmd(),
		newAPICmd(),
		newLogoutCmd(),
		newVersionCmd(version),
	)
	return root
}

// exactArgs valida o número de argumentos como erro de uso.
func exactArgs(n int, name string) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != n {
			return usageError{fmt.Errorf("informe %s", name)}
		}
		return nil
	}
}

// outputFormat resolve o formato: -o explícito > fallback do comando > CTBZ_OUTPUT > table.
func outputFormat(cmd *cobra.Command, fallback string) (output.Format, error) {
	v := ""
	if f := cmd.Flag("output"); f != nil && f.Changed {
		v = f.Value.String()
	} else if fallback != "" {
		v = fallback
	} else if env := os.Getenv("CTBZ_OUTPUT"); env != "" {
		v = env
	} else {
		v = string(output.FormatTable)
	}
	f, err := output.ParseFormat(v)
	if err != nil {
		return "", usageError{err}
	}
	return f, nil
}

// streams agrupa as entradas e saídas de um comando (trocadas nos testes).
type streams struct {
	in          io.Reader
	out         io.Writer
	err         io.Writer
	interactive bool
}

func streamsOf(cmd *cobra.Command) streams {
	in := cmd.InOrStdin()
	return streams{
		in:          in,
		out:         cmd.OutOrStdout(),
		err:         cmd.ErrOrStderr(),
		interactive: isTerminal(in),
	}
}

// isTerminal diz se a entrada é um terminal (trocada nos testes de confirmação).
var isTerminal = func(in io.Reader) bool {
	f, ok := in.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

const usageTemplate = `Uso:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [comando]{{end}}{{if gt (len .Aliases) 0}}

Apelidos:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Exemplos:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Comandos:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Flags globais:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [comando] --help" para a ajuda de um comando.{{end}}
`
