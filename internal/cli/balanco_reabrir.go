package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newBalancoReabrirCmd() *cobra.Command {
	var ano int
	cmd := &cobra.Command{
		Use:   "reabrir",
		Short: "Reabre o balanço de um ano para regularizar a pendência documental do informe",
		Long: `Reabre o exercício contábil (balanço) de um ano, o "Regularizar pendências" do informe de
rendimentos. Só é enviado quando o informe do ano tem pendência documental, o caminho de
regularização é a reabertura do balanço e não há outra reabertura em andamento.

Quando a regularização é um serviço pago (contratar serviço adicional), a CLI recusa e mostra
o custo: contrate pelo painel. Veja a situação em ctbz lucros informe restricoes.

Risco alto: afeta o balanço, o informe de rendimentos (indisponível até a análise) e a
distribuição de lucros, sem desfazer pela API. Aceita --yes e --dry-run.`,
		Example: `  ctbz balanco reabrir --ano 2025`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if ano == 0 {
				return usageError{errors.New("informe --ano (o ano do exercício a reabrir)")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			r, err := api.BuscarRestricoesInforme(cmd.Context(), g, ano)
			if err != nil {
				return err
			}
			if err := podeReabrir(ano, r); err != nil {
				return err
			}
			op := operacao{Risco: riscoAlto, ID: fmt.Sprint(ano), Resumo: fmt.Sprintf("Reabrir o balanço de %d para regularizar a pendência documental", ano),
				Consequencia: "O informe de rendimentos fica indisponível até a análise da Contabilizei, e a reabertura não dá para desfazer pela CLI."}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroAdmin(api.ReabrirBalanco(cmd.Context(), snd, ano))
			})
			if err != nil || !enviado {
				return err
			}
			relida, err := api.BuscarRestricoesInforme(cmd.Context(), g, ano)
			if err != nil {
				return err
			}
			status := relida.ProcessoReabertura.Status
			situacao := "enviado"
			if status != "" && status != "NENHUM" {
				situacao = status
			}
			fmt.Fprintln(s.err, "Acompanhe na Central de Rotinas: ctbz pendencias.")
			return output.Write(s.out, f, resultadoEscrita("reabrir", situacao, fmt.Sprint(ano)))
		},
	}
	cmd.Flags().IntVar(&ano, "ano", 0, "ano do exercício a reabrir")
	addWriteFlags(cmd)
	return cmd
}

// podeReabrir confere o que o painel confere antes de oferecer a reabertura.
func podeReabrir(ano int, r *api.RestricoesInforme) error {
	pd := r.Restricoes.PendenciaDocumental
	switch status := r.ProcessoReabertura.Status; {
	case status != "" && status != "NENHUM":
		return fmt.Errorf("já há uma reabertura do balanço de %d em andamento (%s)", ano, status)
	case !pd.PossuiPendencia:
		return fmt.Errorf("o informe de %d não tem pendência documental: não há o que regularizar", ano)
	case pd.FluxoRegularizacao == api.FluxoContratarServicoAdicional:
		custo := "com cobrança"
		if pd.ValorServicoAdicional != nil {
			custo = "que custa " + output.FormatBRL(*pd.ValorServicoAdicional)
		}
		return fmt.Errorf("o período contábil de %d já foi fechado: a reabertura é um serviço adicional %s; contrate pelo painel", ano, custo)
	case pd.FluxoRegularizacao != api.FluxoReaberturaBalanco:
		return fmt.Errorf("o painel não oferece reabrir o balanço de %d (regularização: %s)", ano, firstNonEmpty(pd.FluxoRegularizacao, "não informada"))
	}
	return nil
}
