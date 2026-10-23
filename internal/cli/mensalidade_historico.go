package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newMensalidadeHistoricoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "historico",
		Short: "Lista as faturas pagas da mensalidade",
		Long: `Lista as faturas da Contabilizei já pagas, como a tela "Consultar histórico" do painel: por
competência, o vencimento, a data do pagamento, o valor, a multa e os juros cobrados pelo
atraso da fatura anterior, a forma de pagamento e o status. A mais recente vem primeiro.

A baixa da fatura pode demorar depois do pagamento: para saber se a mensalidade está em dia,
use ctbz mensalidade situacao.

Os campos debito_automatico_ativo, debito_automatico_habilitado, status, competencia e
proximo_pagamento são do débito automático de impostos (o mesmo de ctbz impostos recorrente),
não da mensalidade. Ficam por compatibilidade e estão obsoletos.

Só campos conhecidos são copiados para a saída: dados de cartão, chaves do gateway de
pagamento e os códigos de pagamento das faturas nunca são mostrados.`,
		Example: `  ctbz mensalidade historico
  ctbz mensalidade historico -o json | jq '.pagamentos[] | {competencia, valor, multa_juros}'`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			cnpj, err := cnpjDaSessao()
			if err != nil {
				return err
			}
			var ini json.RawMessage
			if err := g.GetJSON(cmd.Context(), api.PathRecorrenciaInit, &ini); err != nil {
				return err
			}
			faturas, err := api.BuscarFaturasPagas(cmd.Context(), g, cnpj)
			if err != nil {
				return err
			}
			rec, err := buildHistorico(ini, faturas)
			if err != nil {
				return err
			}
			return output.Write(s.out, f, rec)
		},
	}
}

// buildHistorico monta o registro de saída a partir das faturas pagas e de
// payments/recorrencia/init (débito automático de impostos, mantido por compatibilidade).
// Só campos conhecidos são copiados: a resposta da recorrência traz chaves do gateway de
// pagamento (adyenClientKey) e a das faturas, Pix e linha digitável, que nunca devem chegar
// à saída.
func buildHistorico(initBody []byte, faturas []api.FaturaPaga) (*output.Record, error) {
	var ini map[string]any
	if err := decodeJSON(initBody, &ini); err != nil {
		return nil, fmt.Errorf("resposta inesperada de %s: %w", api.PathRecorrenciaInit, err)
	}

	sort.SliceStable(faturas, func(i, j int) bool {
		a, b := faturas[i].Competencia, faturas[j].Competencia
		return a.Ano > b.Ano || (a.Ano == b.Ano && a.Mes > b.Mes)
	})
	pagamentos := make([]output.Record, 0, len(faturas))
	for _, x := range faturas {
		status := x.StatusLabel
		if status == "" {
			status = x.Status
		}
		var forma any
		if x.Pagamento != nil {
			forma = nilIfEmpty(x.Pagamento.FormaPagamento)
		}
		r := output.Record{}
		r.Add("data", "Data", dataMillis(x.DataPagamento)).
			Add("competencia", "Competência", fmt.Sprintf("%02d/%d", x.Competencia.Mes, x.Competencia.Ano)).
			Add("vencimento", "Vencimento", dataMillis(x.DataVencimento)).
			Add("valor", "Valor", moneyOrNil(x.Total)).
			Add("multa_juros", "Multa e juros", output.Money(x.MultaEJuros())).
			Add("forma_pagamento", "Forma", forma).
			Add("status", "Status", nilIfEmpty(status))
		pagamentos = append(pagamentos, r)
	}

	status, _ := pick(ini, "status")
	competencia, _ := pick(ini, "competencia")
	proximo, _ := pick(ini, "dataProximoPagamento")
	habilitado, _ := pick(ini, "habilitado")
	ativado, _ := pick(ini, "ativado")

	rec := &output.Record{}
	rec.Add("debito_automatico_ativo", "Débito automático de impostos ativo", toBool(ativado)).
		Add("debito_automatico_habilitado", "Débito automático de impostos habilitado", toBool(habilitado)).
		Add("status", "Status do débito automático de impostos", toText(status)).
		Add("competencia", "Competência do débito automático de impostos", toText(competencia)).
		Add("proximo_pagamento", "Próximo débito de impostos", toDate(proximo)).
		Add("pagamentos", "Faturas pagas", pagamentos)
	return rec, nil
}

// dataMillis converte um epoch em ms opcional numa data (nil quando ausente).
func dataMillis(ms *int64) any {
	if ms == nil {
		return nil
	}
	return output.DateFromMillis(*ms)
}

// cnpjDaSessao é o CNPJ da empresa escolhida na sessão salva.
func cnpjDaSessao() (string, error) {
	store, err := ctbz.DefaultStore()
	if err != nil {
		return "", err
	}
	sess, err := store.LoadSession()
	if err != nil {
		return "", err
	}
	if cnpj := currentCNPJ(sess); cnpj != "" {
		return cnpj, nil
	}
	if sess.CNPJ != "" {
		return sess.CNPJ, nil
	}
	return "", errors.New("a sessão não tem a empresa escolhida; rode `ctbz login` ou `ctbz empresa usar`")
}

func decodeJSON(b []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	return dec.Decode(v)
}

// pick devolve o primeiro valor não nulo entre as chaves candidatas.
func pick(m map[string]any, keys ...string) (any, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v, true
		}
	}
	return nil, false
}

func toText(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		return nilIfEmpty(x)
	case json.Number:
		return x.String()
	case bool:
		return strconv.FormatBool(x)
	default:
		return nil
	}
}

func toBool(v any) any {
	b, ok := v.(bool)
	if !ok {
		return nil
	}
	return b
}

// toDate aceita texto (dd/mm/aaaa ou ISO) ou epoch em milissegundos.
func toDate(v any) output.Date {
	switch x := v.(type) {
	case string:
		d, _ := output.ParseDate(x)
		return d
	case json.Number:
		if ms, err := x.Int64(); err == nil {
			return output.DateFromMillis(ms)
		}
	}
	return output.Date{}
}
