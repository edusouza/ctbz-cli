package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newProlaboreZerarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zerar-sem-faturamento on|off",
		Short: "Liga ou desliga o pró-labore zerado nos meses sem faturamento",
		Long: `Liga (on) ou desliga (off) a preferência da empresa "Não quero ter pró-labore cadastrado
em meses sem faturamento". Vale para todos os sócios.

Com a preferência ligada, nos meses sem faturamento os sócios ficam sem pró-labore e não
contribuem para o INSS. Numa empresa na gestão inteligente, ligar vale a partir deste mês e
não altera os meses passados.

on é de risco alto (deixar de contribuir para o INSS) e off, de risco médio. Os dois são
reversíveis. Aceita --yes e --dry-run.`,
		Example: `  ctbz prolabore zerar-sem-faturamento on
  ctbz prolabore zerar-sem-faturamento off --yes`,
		Args:      exactArgs(1, "on ou off"),
		ValidArgs: []string{"on", "off"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "on" && args[0] != "off" {
				return usageError{fmt.Errorf("use on ou off: %q", args[0])}
			}
			ligar := args[0] == "on"
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			c, err := api.BuscarProlaboreCentral(cmd.Context(), g)
			if err != nil {
				return err
			}
			if c.ZerarProlabore == ligar {
				fmt.Fprintf(s.err, "A preferência já está %s; nada foi enviado.\n", estadoZerar(ligar))
				return output.Write(s.out, f, resultadoEscrita("zerar-sem-faturamento", "sem mudança", nil).Add("zerar", "Zerar sem faturamento", ligar))
			}
			gi := ligar && c.TipoGerenciamento == api.GerenciamentoInteligente
			op := operacao{Risco: riscoMedio, Resumo: "Desligar o pró-labore zerado nos meses sem faturamento: os sócios voltam a ter pró-labore e INSS todo mês"}
			if ligar {
				linhas := []string{"Nos meses sem faturamento os sócios ficam sem pró-labore e não contribuem para o INSS. Dá para desfazer com off."}
				if gi {
					linhas = append(linhas, "A empresa está na gestão inteligente: vale a partir deste mês e não altera os meses passados.")
				}
				op = operacao{Risco: riscoAlto, Resumo: "Ligar o pró-labore zerado nos meses sem faturamento", Consequencia: strings.Join(linhas, "\n")}
			}
			if n := len(c.Socios); n > 1 {
				op.Resumo += fmt.Sprintf(" (vale para os %d sócios)", n)
			}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				if gi {
					return api.ZerarProlaboreGestaoInteligente(cmd.Context(), snd)
				}
				return api.DefinirZerarProlabore(cmd.Context(), snd, ligar)
			})
			if err != nil || !enviado {
				return err
			}
			relida, err := api.BuscarProlaboreCentral(cmd.Context(), g)
			if err != nil {
				return err
			}
			situacao := "enviado"
			if relida.ZerarProlabore == ligar {
				situacao = estadoZerar(ligar)
			}
			return output.Write(s.out, f, resultadoEscrita("zerar-sem-faturamento", situacao, nil).Add("zerar", "Zerar sem faturamento", relida.ZerarProlabore))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

func estadoZerar(ligado bool) string {
	if ligado {
		return "ligada"
	}
	return "desligada"
}
