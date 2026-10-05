package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// avisosParcelamento são os avisos do painel na contratação.
const avisosParcelamento = "Contratar é uma confissão de dívida perante a Receita ou a PGFN. Se a primeira parcela não for paga, " +
	"o parcelamento é cancelado automaticamente. As parcelas são corrigidas todo mês pela Selic + 1% de juros. " +
	"Não dá para desfazer pela CLI nem pelo painel."

// opcaoEscolhida acha a opção de parcelas: --parcelas (validado) ou a primeira, como o painel.
func opcaoEscolhida(sim *api.SimulacaoParcelamento, parcelas int) (*api.OpcaoParcelas, error) {
	if len(sim.Parcelas) == 0 {
		return nil, nil
	}
	if parcelas == 0 {
		return &sim.Parcelas[0], nil
	}
	var aceitas []string
	for i, p := range sim.Parcelas {
		if p.Quantidade == parcelas {
			return &sim.Parcelas[i], nil
		}
		aceitas = append(aceitas, fmt.Sprint(p.Quantidade))
	}
	return nil, usageError{fmt.Errorf("%d parcelas não é uma opção; use %s (ver ctbz impostos parcelamento simular)", parcelas, strings.Join(aceitas, ", "))}
}

func newParcelamentoContratarCmd() *cobra.Command {
	var negociacao string
	var parcelas int
	cmd := &cobra.Command{
		Use:   "contratar TIPO",
		Short: "Contrata um parcelamento de débitos",
		Long: `Contrata um parcelamento de débitos (risco alto). A CLI roda a simulação antes e mostra a
opção escolhida e os custos adicionais cobrados na mensalidade. ` + avisosParcelamento + `

TIPO: ` + strings.Join(api.TiposParcelamento(), ", ") + `.
- pgfn-*: --parcelas escolhe a quantidade entre as opções da simulação (padrão: a primeira,
  como no painel).
- simples-nacional: a quantidade não é escolhida aqui (o pedido vai sem corpo).
- especializado: exige --negociacao; vira um pedido de atendimento com idTicket.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz impostos parcelamento contratar pgfn-previdenciario --parcelas 24
  ctbz impostos parcelamento contratar especializado --negociacao VENCIDOS --dry-run`,
		Args: exactArgs(1, "o TIPO de parcelamento"),
		RunE: func(cmd *cobra.Command, args []string) error {
			tipo := args[0]
			if err := api.ValidarParcelamento(tipo, negociacao); err != nil {
				return usageError{err}
			}
			if parcelas < 0 || (parcelas > 0 && !strings.HasPrefix(tipo, "pgfn-")) {
				return usageError{errors.New("--parcelas só vale para os tipos pgfn-* (e deve ser positivo)")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			sim, err := simularParcelamento(cmd, g, tipo, negociacao)
			if err != nil {
				return err
			}
			if sim == nil {
				return errors.New("nenhum débito para parcelar")
			}
			opcao, err := opcaoEscolhida(sim, parcelas)
			if err != nil {
				return err
			}
			if strings.HasPrefix(tipo, "pgfn-") && opcao == nil {
				return errors.New("a simulação não trouxe opções de parcelas; contrate pelo painel")
			}
			resumo := "Contratar parcelamento " + tipo
			if negociacao != "" {
				resumo += " (" + negociacao + ")"
			}
			var qtd int
			if opcao != nil {
				qtd = opcao.Quantidade
				resumo += fmt.Sprintf(": %d parcelas, primeira de %s e demais de %s", opcao.Quantidade,
					formatarCentavos(centavosPtr(opcao.ValorEntrada)), formatarCentavos(centavosPtr(opcao.ValorDemaisParcelas)))
			}
			consequencia := avisosParcelamento
			if sim.CobrarServicoAdicional || sim.ValorEmissaoGuia != nil {
				consequencia += fmt.Sprintf(" Cobrados na próxima mensalidade: serviço adicional %s; emissão de guia %s.",
					formatarCentavos(centavosPtr(sim.ValorServicoAdicional)), formatarCentavos(centavosPtr(sim.ValorEmissaoGuia)))
			}
			if !strings.HasPrefix(tipo, "pgfn-") {
				qtd = 0 // só a PGFN envia a quantidade
			}
			op := operacao{Risco: riscoAlto, Resumo: resumo, Consequencia: consequencia}
			var resp *api.RespostaContratacao
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				var err error
				resp, err = api.ContratarParcelamento(ctx, snd, tipo, qtd, negociacao)
				return err
			})
			if err != nil || !enviado {
				return err
			}
			var id any
			var detalhes json.RawMessage
			if resp != nil && len(resp.IDTicket) > 0 && string(resp.IDTicket) != "null" {
				ticket := strings.Trim(string(resp.IDTicket), `"`)
				id = ticket
				if r, err := authedAPI(ctx, s, "GET", api.PathDetalheNegociacaoEspecializada(ticket), nil); err == nil && r.Status == 200 && json.Valid(r.Body) {
					detalhes = r.Body
				} else {
					fmt.Fprintln(s.err, "aviso: pedido feito, mas não foi possível ler o detalhe; veja ctbz impostos parcelamentos")
				}
			}
			fmt.Fprintln(s.err, "Recebemos seu pedido de parcelamento. Acompanhe com ctbz impostos parcelamentos.")
			rec := resultadoEscrita("contratar", "solicitado", id).
				Add("tipo", "Tipo", tipo).
				Add("negociacao", "Negociação", nilIfEmpty(negociacao))
			if opcao != nil {
				rec.Add("parcelas", "Parcelas", opcao.Quantidade).
					Add("entrada", "Primeira parcela", moneyOrNil(opcao.ValorEntrada)).
					Add("demais_parcelas", "Demais parcelas", moneyOrNil(opcao.ValorDemaisParcelas))
			}
			if detalhes != nil {
				rec.Add("detalhes", "Detalhes", detalhes)
			}
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().StringVar(&negociacao, "negociacao", "", "no tipo especializado: "+strings.Join(api.NegociacoesEspecializadas, ", "))
	cmd.Flags().IntVar(&parcelas, "parcelas", 0, "quantidade de parcelas (tipos pgfn-*; padrão: a primeira opção da simulação)")
	addWriteFlags(cmd)
	return cmd
}

func centavosPtr(v *float64) int64 {
	if v == nil {
		return 0
	}
	return centavos(*v)
}
