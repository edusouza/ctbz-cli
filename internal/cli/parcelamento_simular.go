package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// statusSemDebitos é a resposta da simulação quando não há dívidas para parcelar.
const statusSemDebitos = 560

// simularParcelamento lê a simulação; sem dívidas (HTTP 560) devolve nil, sem erro.
func simularParcelamento(cmd *cobra.Command, g api.Getter, tipo, negociacao string) (*api.SimulacaoParcelamento, error) {
	sim, err := api.SimularParcelamento(cmd.Context(), g, tipo, negociacao)
	var he *ctbz.HTTPError
	if errors.As(err, &he) && he.Status == statusSemDebitos {
		return nil, nil
	}
	return sim, err
}

func simulacaoList(tipo string, s *api.SimulacaoParcelamento) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "tipo", Header: "Tipo"},
		{Key: "parcelas", Header: "Parcelas"},
		{Key: "entrada", Header: "Primeira parcela"},
		{Key: "demais_parcelas", Header: "Demais parcelas"},
		{Key: "servico_adicional", Header: "Serviço adicional"},
		{Key: "emissao_guia", Header: "Emissão de guia"},
	}}
	if s == nil {
		return l
	}
	var servico any
	if s.CobrarServicoAdicional {
		servico = moneyOrNil(s.ValorServicoAdicional)
	}
	for _, p := range s.Parcelas {
		l.Append(tipo, p.Quantidade, moneyOrNil(p.ValorEntrada), moneyOrNil(p.ValorDemaisParcelas), servico, moneyOrNil(s.ValorEmissaoGuia))
	}
	return l
}

func newParcelamentoSimularCmd() *cobra.Command {
	var negociacao string
	cmd := &cobra.Command{
		Use:   "simular TIPO",
		Short: "Simula um parcelamento de débitos (opções de parcelas e custos)",
		Long: `Mostra as opções de parcelamento de um tipo de débito: quantidade de parcelas, valor da
primeira e das demais, e os custos adicionais cobrados na mensalidade (serviço adicional e
emissão de guia). Só leitura: nada é contratado.

TIPO: ` + strings.Join(api.TiposParcelamento(), ", ") + `. O tipo especializado (negociação feita
por um especialista) exige --negociacao ` + strings.Join(api.NegociacoesEspecializadas, "|") + `.

Sem débitos para parcelar, a CLI avisa e termina com código 0.`,
		Example: `  ctbz impostos parcelamento simular simples-nacional
  ctbz impostos parcelamento simular especializado --negociacao DIVIDA_ATIVA -o json`,
		Args: exactArgs(1, "o TIPO de parcelamento"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := api.ValidarParcelamento(args[0], negociacao); err != nil {
				return usageError{err}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			sim, err := simularParcelamento(cmd, sessionGetter{s}, args[0], negociacao)
			if err != nil {
				return err
			}
			if sim == nil {
				fmt.Fprintln(s.err, "Nenhum débito para parcelar.")
			} else if len(sim.Parcelas) == 0 {
				fmt.Fprintln(s.err, "A simulação não trouxe opções de parcelas; veja a resposta completa com ctbz api", api.PathSimulacaoParcelamento(args[0], negociacao))
			}
			return output.Write(s.out, f, simulacaoList(args[0], sim))
		},
	}
	cmd.Flags().StringVar(&negociacao, "negociacao", "", "no tipo especializado: "+strings.Join(api.NegociacoesEspecializadas, ", "))
	return cmd
}
