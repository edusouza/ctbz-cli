package cli

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newImpostosFaturamentoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "faturamento",
		Short: "Mostra o faturamento, o pró-labore e os impostos pagos nos últimos 12 meses",
		Long: `Mostra mês a mês o faturamento e o pró-labore considerados na apuração (últimos 12 meses,
do mais antigo ao mais recente) e o total pago em impostos em cada mês.

Na tabela, o resumo vai para stderr: faturamento acumulado em 12 meses (RBT12, que define
a alíquota do Simples Nacional), pró-labore acumulado e o percentual do Fator R.`,
		Example: `  ctbz impostos faturamento
  ctbz impostos faturamento -o csv > faturamento.csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			c, err := api.BuscarCalculoImposto(cmd.Context(), g)
			if err != nil {
				return err
			}
			meses := mesesDoHistorico(c)
			pagos := map[int]*api.DadosGrafico{}
			for _, m := range meses {
				if _, ok := pagos[m.ano]; !ok {
					if pagos[m.ano], err = api.BuscarImpostosPagosNoAno(cmd.Context(), g, m.ano); err != nil {
						if !anoAntesDaAbertura(err) {
							return err
						}
						fmt.Fprintf(s.err, "Impostos pagos indisponíveis em %d: a Contabilizei recusou o ano por ser anterior à abertura ou à ativação da empresa.\n", m.ano)
					}
				}
			}
			if err := output.Write(s.out, f, faturamentoList(meses, pagos)); err != nil {
				return err
			}
			if f == output.FormatTable {
				fmt.Fprintf(s.err, "Faturamento 12 meses (RBT12): %s · Pró-labore 12 meses: %s · Fator R: %s\n",
					brlOrDash(c.ValorFaturamentoUltimos12Meses), brlOrDash(c.ValorProlaboreUltimos12Meses), percentOrDash(c.PercentualFatorR))
			}
			return nil
		},
	}
}

type mesFaturamento struct {
	mes, ano               int
	faturamento, prolabore *float64
}

// mesesDoHistorico converte o histórico ("set./26") em meses em ordem cronológica.
func mesesDoHistorico(c *api.CalculoImposto) []mesFaturamento {
	var out []mesFaturamento
	for _, h := range c.HistoricoFaturamento {
		comp, ok := competenciaDeTexto(h.Mes).(string)
		if !ok || len(comp) != 7 {
			continue
		}
		mes, _ := strconv.Atoi(comp[:2])
		ano, _ := strconv.Atoi(comp[3:])
		out = append(out, mesFaturamento{mes, ano, h.ValorFaturamento, h.ValorProlabore})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ano*100+out[i].mes < out[j].ano*100+out[j].mes })
	return out
}

// anoAntesDaAbertura diz se a Contabilizei recusou o ano por ser anterior à abertura (ou à
// ativação) da empresa: os 12 meses do histórico podem começar antes dela, e nesses meses não
// há impostos pagos a mostrar.
func anoAntesDaAbertura(err error) bool {
	var he *ctbz.HTTPError
	return errors.As(err, &he) && he.Status == http.StatusBadRequest &&
		bytes.Contains(he.Body, []byte("impostos/historico-impostos/ano-invalido"))
}

func faturamentoList(meses []mesFaturamento, pagos map[int]*api.DadosGrafico) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "competencia", Header: "Competência"},
		{Key: "faturamento", Header: "Faturamento"},
		{Key: "prolabore", Header: "Pró-labore"},
		{Key: "impostos_pagos", Header: "Impostos pagos"},
	}}
	for _, m := range meses {
		var pago any
		if d := pagos[m.ano]; d != nil {
			pago = moneyOrNil(d.Meses[strconv.Itoa(m.mes)].TotalPago)
		}
		l.Append(competencia(m.mes, m.ano), moneyOrNil(m.faturamento), moneyOrNil(m.prolabore), pago)
	}
	return l
}

func brlOrDash(v *float64) string {
	if v == nil {
		return "—"
	}
	return output.FormatBRL(*v)
}

func percentOrDash(v *float64) string {
	if v == nil {
		return "—"
	}
	return strconv.FormatFloat(*v, 'f', -1, 64) + "%"
}
