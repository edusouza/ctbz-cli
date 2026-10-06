package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

var reCodigoSimples = regexp.MustCompile(`^\d{12}$`)

func newEmpresaCredenciaisCmd() *cobra.Command {
	var mostrar bool
	cmd := &cobra.Command{
		Use:   "credenciais",
		Short: "Mostra os dados de acesso a órgãos públicos (código do Simples, prefeitura, Dataprev)",
		Long: `Mostra as credenciais que a Contabilizei usa nos órgãos públicos: código de acesso do
Simples Nacional, usuário e senha da prefeitura e do Dataprev, e se o painel pede para
redefinir a senha da prefeitura.

O código e as senhas aparecem só como "preenchido"; --mostrar exibe os valores, apenas num
terminal e na tabela. Para usuário administrador a Contabilizei recusa (403).

Para trocar: ctbz empresa credenciais atualizar; para responder ao alerta da prefeitura:
ctbz empresa credenciais confirmar-prefeitura.`,
		Example: `  ctbz empresa credenciais
  ctbz empresa credenciais --mostrar`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			if mostrar && (!s.interactive || f != output.FormatTable) {
				return usageError{errors.New("--mostrar só funciona num terminal e na tabela")}
			}
			d, err := api.BuscarDadosAcesso(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			fo := d.Formulario
			segredo := func(v *string) any {
				switch {
				case v == nil || *v == "":
					return nil
				case mostrar:
					return semControle(*v)
				}
				return "preenchido"
			}
			texto := func(v *string) any {
				if v == nil {
					return nil
				}
				return nilIfEmpty(semControle(*v))
			}
			rec := &output.Record{}
			rec.Add("codigo_simples", "Código de acesso do Simples", segredo(fo.ChaveAcessoSimples)).
				Add("usuario_prefeitura", "Usuário da prefeitura", texto(fo.UsuarioPrefeitura)).
				Add("senha_prefeitura", "Senha da prefeitura", segredo(fo.SenhaPrefeitura)).
				Add("usuario_dataprev", "Usuário do Dataprev", texto(fo.UsuarioDataprev)).
				Add("senha_dataprev", "Senha do Dataprev", segredo(fo.SenhaDataprev)).
				Add("redefinir_senha_prefeitura", "Redefinir a senha da prefeitura", d.AlertaPendenciaCredencial)
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().BoolVar(&mostrar, "mostrar", false, "mostra o código e as senhas (só num terminal)")
	cmd.AddCommand(newCredenciaisAtualizarCmd(), newCredenciaisConfirmarPrefeituraCmd())
	return cmd
}

func newCredenciaisAtualizarCmd() *cobra.Command {
	var simples, prefeitura, dataprev bool
	cmd := &cobra.Command{
		Use:   "atualizar",
		Short: "Troca o código do Simples ou o usuário e a senha da prefeitura ou do Dataprev",
		Long: `Troca as credenciais escolhidas, pedindo os valores sem eco no terminal (nunca por
argumento): --codigo-simples (12 dígitos), --prefeitura e --dataprev (usuário e senha, com
pelo menos 3 caracteres; Enter no usuário mantém o atual). As demais credenciais seguem como
estão. Fora de um terminal, os valores são lidos da entrada padrão, uma linha por pergunta.

Depois de salvar, a Contabilizei verifica o acesso e avisa por e-mail se for negado.

Risco alto: credencial errada faz falhar a entrega de declarações (multas e juros). Aceita
--yes e --dry-run (que mostra *** no lugar dos segredos).`,
		Example: `  ctbz empresa credenciais atualizar --prefeitura
  ctbz empresa credenciais atualizar --codigo-simples --dataprev`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !simples && !prefeitura && !dataprev {
				return usageError{errors.New("escolha o que atualizar: --codigo-simples, --prefeitura ou --dataprev")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			d, err := api.BuscarDadosAcesso(cmd.Context(), g)
			if err != nil {
				return err
			}
			mudancas := map[string]string{}
			var nomes []string
			if simples {
				v, err := lerSegredo(s, "Código de acesso do Simples Nacional (12 dígitos): ")
				if err != nil {
					return err
				}
				if v = strings.TrimSpace(v); !reCodigoSimples.MatchString(v) {
					return usageError{errors.New("código de acesso deve possuir 12 caracteres apenas numéricos")}
				}
				mudancas[api.CampoChaveAcessoSimples] = v
				nomes = append(nomes, "código do Simples")
			}
			for _, c := range []struct {
				marcado                  bool
				orgao                    string // com o artigo: "da prefeitura"
				campoUsuario, campoSenha string
				atual                    *string
			}{
				{prefeitura, "da prefeitura", api.CampoUsuarioPrefeitura, api.CampoSenhaPrefeitura, d.Formulario.UsuarioPrefeitura},
				{dataprev, "do Dataprev", api.CampoUsuarioDataprev, api.CampoSenhaDataprev, d.Formulario.UsuarioDataprev},
			} {
				if !c.marcado {
					continue
				}
				atual, rotulo := "", "Usuário "+c.orgao+": "
				if c.atual != nil && *c.atual != "" {
					atual = *c.atual
					rotulo = fmt.Sprintf("Usuário %s [Enter mantém %s]: ", c.orgao, semControle(atual))
				}
				usuario := lerTexto(s, rotulo, atual)
				senha, err := lerSegredo(s, "Senha "+c.orgao+": ")
				if err != nil {
					return err
				}
				if utf8.RuneCountInString(usuario) < 3 || utf8.RuneCountInString(senha) < 3 {
					return usageError{fmt.Errorf("usuário e senha %s precisam de pelo menos 3 caracteres", c.orgao)}
				}
				mudancas[c.campoUsuario], mudancas[c.campoSenha] = usuario, senha
				nomes = append(nomes, "usuário e senha "+c.orgao)
			}
			var idEmpresa any
			if e, err := empresaDaSessao(); err == nil {
				idEmpresa = e.ID
			}
			corpo, err := api.NovoFormularioDadosAcesso(d.FormularioBruto, idEmpresa, mudancas)
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoAlto, Resumo: "Atualizar os dados de acesso: " + strings.Join(nomes, ", "),
				Consequencia: "Credencial errada faz falhar a entrega de declarações (multas e juros). As demais credenciais seguem como estão."}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroAdmin(api.AtualizarDadosAcesso(cmd.Context(), snd, corpo))
			})
			if err != nil || !enviado {
				return err
			}
			fmt.Fprintln(s.err, "Verificação de acesso iniciada: se o acesso for negado, a Contabilizei avisa por e-mail.")
			return output.Write(s.out, f, resultadoEscrita("atualizar credenciais", "enviado", nil).Add("campos", "Campos", strings.Join(nomes, ", ")))
		},
	}
	cmd.Flags().BoolVar(&simples, "codigo-simples", false, "troca o código de acesso do Simples Nacional")
	cmd.Flags().BoolVar(&prefeitura, "prefeitura", false, "troca o usuário e a senha da prefeitura")
	cmd.Flags().BoolVar(&dataprev, "dataprev", false, "troca o usuário e a senha do Dataprev")
	addWriteFlags(cmd)
	return cmd
}

func newCredenciaisConfirmarPrefeituraCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirmar-prefeitura",
		Short: `Responde "Sim, já atualizei a senha" ao alerta da senha da prefeitura`,
		Long: `Declara que a senha da prefeitura cadastrada é válida e limpa o alerta "Redefina a senha
da prefeitura e atualize os dados de acesso". Se a senha mudou, use antes ctbz empresa
credenciais atualizar --prefeitura.

Risco médio: se a declaração for falsa, as obrigações municipais falham. Aceita --yes e
--dry-run.`,
		Example: `  ctbz empresa credenciais confirmar-prefeitura`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			d, err := api.BuscarDadosAcesso(cmd.Context(), g)
			if err != nil {
				return err
			}
			if !d.AlertaPendenciaCredencial {
				fmt.Fprintln(s.err, "O painel não pede para redefinir a senha da prefeitura; nada foi enviado.")
				return output.Write(s.out, f, resultadoEscrita("confirmar-prefeitura", "sem mudança", nil))
			}
			op := operacao{Risco: riscoMedio, Resumo: "Declarar que a senha da prefeitura cadastrada está correta"}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroAdmin(api.ConfirmarCredencialPrefeitura(cmd.Context(), snd))
			})
			if err != nil || !enviado {
				return err
			}
			relido, err := api.BuscarDadosAcesso(cmd.Context(), g)
			situacao := "enviado"
			if err == nil && !relido.AlertaPendenciaCredencial {
				situacao = "confirmada"
			}
			return output.Write(s.out, f, resultadoEscrita("confirmar-prefeitura", situacao, nil))
		},
	}
	addWriteFlags(cmd)
	return cmd
}
