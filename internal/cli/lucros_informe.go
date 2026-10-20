package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newLucrosInformeCmd() *cobra.Command {
	var ano int
	cmd := &cobra.Command{
		Use:   "informe",
		Short: "Mostra os valores do informe de rendimentos dos sócios (para o IR)",
		Long: `Mostra, para cada sócio, os valores do comprovante de rendimentos do ano: rendimentos
tributáveis (pró-labore), contribuição previdenciária (INSS), IRRF retido, 13º salário e
seu IRRF, e lucros e dividendos isentos. São os números que vão para a declaração de IR.

O painel monta o PDF do comprovante no navegador a partir desses mesmos valores; a API não
oferece o PDF. O padrão de --ano é o ano anterior (o da declaração).

O que bloqueia o informe e os aceites: ctbz lucros informe restricoes, carta, aceitar e
decidir.`,
		Example: `  ctbz lucros informe
  ctbz lucros informe --ano 2025 -o csv > informe-2025.csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			if ano == 0 {
				ano = now().Year() - 1
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			socios, err := api.BuscarSociosInforme(cmd.Context(), g, ano)
			if err != nil {
				return err
			}
			var cs []*api.ComprovanteRendimentos
			for _, so := range socios {
				c, err := api.BuscarComprovanteRendimentos(cmd.Context(), g, idTexto(so.ID), ano)
				if err != nil {
					return err
				}
				if c.Nome == "" {
					c.Nome = so.Nome
				}
				cs = append(cs, c)
			}
			if len(cs) == 0 && f == output.FormatTable {
				fmt.Fprintf(s.err, "Nenhum sócio com informe de rendimentos em %d.\n", ano)
			}
			return output.Write(s.out, f, informeList(ano, cs))
		},
	}
	cmd.Flags().IntVar(&ano, "ano", 0, "ano-calendário (padrão: o ano anterior)")
	cmd.AddCommand(newInformeRestricoesCmd(), newInformeCartaCmd(), newInformeAceitarCmd(), newInformeDecidirCmd())
	return cmd
}

func informeList(ano int, cs []*api.ComprovanteRendimentos) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "ano", Header: "Ano"},
		{Key: "socio", Header: "Sócio"},
		{Key: "cpf", Header: "CPF"},
		{Key: "rendimentos", Header: "Rendimentos tributáveis"},
		{Key: "previdencia", Header: "Previdência (INSS)"},
		{Key: "irrf_retido", Header: "IRRF retido"},
		{Key: "decimo_terceiro", Header: "13º salário"},
		{Key: "irrf_decimo_terceiro", Header: "IRRF sobre o 13º"},
		{Key: "lucros_isentos", Header: "Lucros isentos"},
	}}
	for _, c := range cs {
		l.Append(ano, c.Nome, output.NewCPF(c.CPF), valorDeValor(c.Rendimentos), valorDeValor(c.Previdencia),
			valorDeValor(c.IRRFRetido), valorDeValor(c.DecimoTerceiro), valorDeValor(c.IRRFDecimoTerceiro), valorDeValor(c.Lucro))
	}
	return l
}

// valorDeValor interpreta um valor em reais de formato não verificado: número ou texto
// ("R$ 1.234,56" ou "1234.56").
func valorDeValor(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case float64:
		return output.Money(x)
	case string:
		if strings.Contains(x, ",") {
			if m, ok := output.ParseBRL(x); ok {
				return m
			}
		} else if f, err := strconv.ParseFloat(strings.TrimSpace(x), 64); err == nil {
			return output.Money(f)
		}
		return nilIfEmpty(x)
	}
	return fmt.Sprint(v)
}
