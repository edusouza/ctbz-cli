package api

import (
	"context"
	"strings"
)

const (
	PathFatura      = "dashboard/fatura"
	PathMensalidade = "dashboard/v1/mensalidade"
)

// Fatura é a fatura atual da Contabilizei (card "Mensalidade" do painel).
type Fatura struct {
	Total                 *float64 `json:"total"`
	Competencia           string   `json:"competencia"` // ex.: "Outubro, 2026"
	Status                string   `json:"status"`      // ex.: "Em aberto"
	Vencimento            any      `json:"vencimento"`  // formato não verificado (vinha null)
	MensagemAdicional     string   `json:"mensagemAdicional"`
	PossuiCartaoPrincipal bool     `json:"possuiCartaoPrincipal"`
	// Ações dos botões do card ("<tipo>:<valor>", ver ParseAcaoCard). O adicional ("Gerar
	// segunda via s/ juros", "request:dashboard/nova-fatura/AAAAMM") vem null com a fatura em dia.
	BotaoTitulo          string `json:"botaoTitulo"`
	BotaoAcao            string `json:"botaoAcao"`
	BotaoAdicionalTitulo string `json:"botaoAdicionalTitulo"`
	BotaoAdicionalAcao   string `json:"botaoAdicionalAcao"`
}

// BuscarFatura lê dashboard/fatura.
func BuscarFatura(ctx context.Context, g Getter) (*Fatura, error) {
	return get[Fatura](ctx, g, PathFatura)
}

// Mensalidade é a resposta de dashboard/v1/mensalidade. Na conta verificada só
// competenciaAnteriorAtrasada vinha preenchido; os demais campos têm formato não verificado.
type Mensalidade struct {
	Status                      any      `json:"status"`
	Valor                       *float64 `json:"valor"`
	DataVencimento              any      `json:"dataVencimento"`
	CompetenciaAnteriorAtrasada bool     `json:"competenciaAnteriorAtrasada"`
}

// BuscarMensalidade lê dashboard/v1/mensalidade.
func BuscarMensalidade(ctx context.Context, g Getter) (*Mensalidade, error) {
	return get[Mensalidade](ctx, g, PathMensalidade)
}

// PathSituacaoMensalidade responde texto puro, não JSON.
const PathSituacaoMensalidade = "inadimplencia/consultasituacaomensalidadeempresa"

// SituacaoEmDia é a resposta de PathSituacaoMensalidade para uma empresa em dia. Outros
// valores não foram vistos; qualquer coisa diferente é tratada como "não está em dia".
const SituacaoEmDia = "OK"

// BuscarSituacaoMensalidade lê a situação da empresa com a Contabilizei (ex.: "OK"),
// sem espaços nem aspas em volta.
func BuscarSituacaoMensalidade(ctx context.Context, g TextGetter) (string, error) {
	s, err := g.GetText(ctx, PathSituacaoMensalidade)
	if err != nil {
		return "", err
	}
	return strings.Trim(strings.TrimSpace(s), `"`), nil
}
