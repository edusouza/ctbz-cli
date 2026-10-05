package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newImpostosCmd() *cobra.Command {
	var soAtrasadas, failOnAtraso bool
	cmd := &cobra.Command{
		Use:   "impostos",
		Short: "Lista as guias de impostos a pagar (em atraso, do mês e do próximo mês)",
		Long: `Lista as guias de impostos a pagar, agrupadas em atraso, deste mês e do próximo mês,
com competência, vencimento, valor e situação. Na tabela, os totais por grupo vão para stderr.

Com --fail-on-atraso, o comando termina com código 4 quando há guias em atraso (útil em
scripts e alertas).`,
		Example: `  ctbz impostos
  ctbz impostos --atrasadas
  ctbz impostos -o csv > guias.csv
  ctbz impostos --fail-on-atraso || notify-send "Há impostos em atraso"`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g, err := api.BuscarGuiasAPagar(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			grupos := gruposDeGuias(g, soAtrasadas)
			if err := output.Write(s.out, f, guiasList(grupos)); err != nil {
				return err
			}
			if f == output.FormatTable {
				fmt.Fprintln(s.err, totaisDeGuias(grupos))
			}
			if failOnAtraso && len(g.EmAtraso) > 0 {
				return errAttention
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&soAtrasadas, "atrasadas", false, "mostra só as guias em atraso")
	cmd.Flags().BoolVar(&failOnAtraso, "fail-on-atraso", false, "termina com código 4 se houver guias em atraso")
	cmd.AddCommand(newImpostosConfirmarCmd(), newImpostosDesmarcarCmd(), newImpostosRecalcularCmd(), newImpostosGuiaCmd(), newImpostosCalculoCmd(), newImpostosTabelaIRRFCmd(), newImpostosBaixarCmd(), newImpostosHistoricoCmd(), newImpostosFaturamentoCmd(),
		newImpostosParcelamentosCmd(), newImpostosParcelamentoCmd(), newImpostosDebitosCmd(), newImpostosRecorrenteCmd())
	return cmd
}

// grupoGuias é um grupo da tela de impostos a pagar.
type grupoGuias struct {
	chave, nome string
	guias       []api.GuiaResumo
}

func gruposDeGuias(g *api.GuiasAPagar, soAtrasadas bool) []grupoGuias {
	grupos := []grupoGuias{{"em_atraso", "Em atraso", g.EmAtraso}}
	if !soAtrasadas {
		grupos = append(grupos, grupoGuias{"este_mes", "Este mês", g.EsteMes}, grupoGuias{"proximo_mes", "Próximo mês", g.ProximoMes})
	}
	return grupos
}

func guiasList(grupos []grupoGuias) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "grupo", Header: "Grupo"},
		{Key: "id", Header: "ID"},
		{Key: "imposto", Header: "Imposto"},
		{Key: "competencia", Header: "Competência"},
		{Key: "vencimento", Header: "Vencimento"},
		{Key: "valor", Header: "Valor"},
		{Key: "situacao", Header: "Situação"},
		{Key: "tipo", Header: "Tipo"},
	}}
	for _, gr := range grupos {
		for _, g := range gr.guias {
			venc, _ := output.ParseDate(g.VencimentoOriginal)
			l.Append(gr.chave, g.ID, g.Nome.Label, competenciaDeTexto(g.Competencia), venc,
				moneyOrNil(g.Valor.Label), g.Status.Label, strings.ToLower(g.Tipo))
		}
	}
	return l
}

func totaisDeGuias(grupos []grupoGuias) string {
	var parts []string
	for _, gr := range grupos {
		var total float64
		for _, g := range gr.guias {
			if g.Valor.Label != nil {
				total += *g.Valor.Label
			}
		}
		parts = append(parts, fmt.Sprintf("%s: %s (%d)", gr.nome, output.FormatBRL(total), len(gr.guias)))
	}
	return "Totais — " + strings.Join(parts, " · ")
}

// moneyOrNil converte um valor opcional; nulo continua nulo (ex.: guia em cálculo).
func moneyOrNil(v *float64) any {
	if v == nil {
		return nil
	}
	return output.Money(*v)
}
