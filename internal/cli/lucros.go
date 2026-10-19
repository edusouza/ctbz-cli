package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newLucrosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lucros",
		Short: "Mostra a distribuição de lucros do exercício e o que a impede",
		Long: `Mostra a distribuição de lucros como a tela de informe de rendimentos do painel: exercício,
saldo disponível na empresa, total distribuído aos sócios, adiantamentos, limite permitido,
se o exercício está fechado, se a distribuição ainda pode ser alterada (e até quando) e o
valor por sócio, com o id que ctbz lucros distribuir pede em --socio.

Também mostra as restrições do exercício: pendências documentais, débitos federais e
reabertura do balanço. A API não recebe ano: o exercício é o que a Contabilizei tem aberto
(sem exercício aberto, as restrições são as do ano anterior).`,
		Example: `  ctbz lucros
  ctbz lucros -o json | jq .saldo`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			d, err := api.BuscarDistribuicaoLucros(cmd.Context(), g)
			if err != nil {
				return err
			}
			ano := d.Ano
			if ano == 0 {
				ano = now().Year() - 1 // o mesmo padrão do painel
			}
			r, err := api.BuscarRestricoesInforme(cmd.Context(), g, ano)
			if err != nil {
				return err
			}
			return output.Write(s.out, f, lucrosRecord(d, ano, r))
		},
	}
	cmd.AddCommand(newLucrosInformeCmd(), newLucrosDistribuirCmd())
	return cmd
}

func lucrosRecord(d *api.DistribuicaoLucros, ano int, r *api.RestricoesInforme) *output.Record {
	socios := make([]output.Record, 0, len(d.LucrosSocios))
	for _, s := range d.LucrosSocios {
		rec := output.Record{}
		rec.Add("socio", "Sócio", s.Socio)
		rec.Add("valor", "Valor", moneyOrNil(s.Valor))
		rec.Add("porcentagem", "Percentual", numeroDeValor(s.Porcentagem))
		rec.Add("id", "ID", nilIfEmpty(idTexto(s.ID)))
		socios = append(socios, rec)
	}
	rec := &output.Record{}
	rec.Add("exercicio", "Exercício", ano)
	rec.Add("saldo", "Saldo na empresa", moneyOrNil(d.Saldo))
	rec.Add("total_distribuido", "Total distribuído", moneyOrNil(d.TotalDistribuido))
	rec.Add("total_adiantamentos", "Adiantamentos", moneyOrNil(d.TotalAdiantamentos))
	rec.Add("limite_distribuicao", "Limite para distribuir", moneyOrNil(d.LimitePermitidoDistribuicaoLucros))
	rec.Add("exercicio_fechado", "Exercício fechado", d.ExercicioFechado)
	rec.Add("pode_alterar", "Pode alterar", d.PodeAlterar)
	rec.Add("data_limite", "Data limite", dataDeValor(d.DataLimite))
	rec.Add("motivo", "Motivo", textoDeValor(d.MotivoNaoPodeAlterar))
	rec.Add("pendencia_documental", "Pendência documental", r.Restricoes.PendenciaDocumental.PossuiPendencia)
	rec.Add("debitos_federais", "Débitos federais", r.Restricoes.DebitosFederais.PossuiPendencia)
	rec.Add("reabertura_balanco", "Reabertura do balanço", nilIfEmpty(r.ProcessoReabertura.Status))
	rec.Add("socios", "Por sócio", socios)
	return rec
}

// numeroDeValor mostra um número de formato não verificado: número fica número; o resto, texto.
func numeroDeValor(v any) any {
	if f, ok := v.(float64); ok {
		return f
	}
	return textoDeValor(v)
}

// textoDeValor mostra um campo de formato não verificado como texto (nulo continua nulo).
func textoDeValor(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		return nilIfEmpty(x)
	case map[string]any:
		return nomeDeValor(x)
	}
	return fmt.Sprint(v)
}
