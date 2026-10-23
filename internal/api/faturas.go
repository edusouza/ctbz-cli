package api

import (
	"context"
	"net/url"
)

// PathFaturasPagas é o histórico da tela legada "Consultar histórico" (rota pagto-historico do
// front sistema): as faturas da Contabilizei já pagas. Fica na API de pagamentos, fora do BFF,
// e pede o CNPJ da empresa da sessão (#856).
func PathFaturasPagas(cnpj string) string {
	return "/api/pagamentos/faturas?cnpj=" + url.QueryEscape(cnpj) + "&confirmadas=true"
}

// FaturaPaga é uma fatura da Contabilizei já paga. A resposta real traz também dados da empresa
// e do responsável, linha digitável e Pix, que a CLI não lê.
type FaturaPaga struct {
	Competencia struct {
		Mes int `json:"mes"`
		Ano int `json:"ano"`
	} `json:"competencia"`
	Status         string   `json:"status"`                          // ex.: CONFIRMADA
	StatusLabel    string   `json:"statusLabel" contract:"optional"` // ex.: Paga
	DataVencimento *int64   `json:"dataVencimento"`                  // epoch em ms
	DataPagamento  *int64   `json:"dataPagamento"`                   // epoch em ms; null na fatura zerada
	Total          *float64 `json:"total"`
	Pagamento      *struct {
		FormaPagamento string `json:"formaPagamento"` // ex.: BOLETO, PIX
	} `json:"pagamento" contract:"optional"` // null na fatura zerada
	Itens []ItemFatura `json:"itens" contract:"optional"`
}

// ItemFatura é um item da fatura.
type ItemFatura struct {
	Tipo  string  `json:"tipo"` // MENSALIDADE, MULTA, JUROS, SERVICO_AVULSO, CREDITO_ABERTURA…
	Valor float64 `json:"valor"`
}

// MultaEJuros soma os itens de multa e juros (atraso da fatura anterior).
func (f FaturaPaga) MultaEJuros() float64 {
	var t float64
	for _, i := range f.Itens {
		if i.Tipo == "MULTA" || i.Tipo == "JUROS" {
			t += i.Valor
		}
	}
	return t
}

// BuscarFaturasPagas lê as faturas já pagas da empresa.
func BuscarFaturasPagas(ctx context.Context, g Getter, cnpj string) ([]FaturaPaga, error) {
	f, err := get[[]FaturaPaga](ctx, g, PathFaturasPagas(cnpj))
	if err != nil {
		return nil, err
	}
	return *f, nil
}
