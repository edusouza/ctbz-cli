package api

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Segunda via da mensalidade (#851). O card "Fatura" do dashboard traz a ação do botão
// adicional ("request:dashboard/nova-fatura/AAAAMM"); o GET desse caminho registra uma
// segunda via com vencimento novo e devolve a URL do checkout com o UUID da fatura. O
// checkout lê a fatura em endpoints públicos (/api/fintech/public/…), em que o UUID é a
// credencial: ele não vai para logs nem para o registro de ações. Ver
// docs/api/escrita/impostos-e-pagamentos.md, seção 2.3.

// AcaoCard é a ação de um botão do card: "<tipo>:<valor>".
type AcaoCard struct {
	Tipo  string // "redirect" (navegação do painel) ou "request" (GET em /api/plataforma/<valor>)
	Valor string
}

// ParseAcaoCard separa "request:dashboard/nova-fatura/202610" em tipo e valor.
func ParseAcaoCard(s string) (AcaoCard, bool) {
	tipo, valor, ok := strings.Cut(strings.TrimSpace(s), ":")
	if !ok || tipo == "" || valor == "" {
		return AcaoCard{}, false
	}
	return AcaoCard{Tipo: tipo, Valor: valor}, true
}

// caminhoSegundaVia aceita só um caminho relativo do BFF, sem esquema, query nem "..": é o
// servidor que o escolhe, e a CLI não deve seguir para outro lugar.
var caminhoSegundaVia = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9/_-]*$`)

// CaminhoSegundaVia valida o valor de uma ação "request:" do card.
func CaminhoSegundaVia(valor string) (string, error) {
	if !caminhoSegundaVia.MatchString(valor) || strings.Contains(valor, "..") || strings.Contains(valor, "//") {
		return "", fmt.Errorf("o card da mensalidade pediu um caminho inesperado: %q", valor)
	}
	return valor, nil
}

// competenciaNoCaminho reconhece o AAAAMM no fim de "dashboard/nova-fatura/202610".
var competenciaNoCaminho = regexp.MustCompile(`/(\d{4})(\d{2})$`)

// CompetenciaSegundaVia devolve ano e mês do caminho da segunda via, quando ele os traz.
func CompetenciaSegundaVia(caminho string) (ano, mes int, ok bool) {
	m := competenciaNoCaminho.FindStringSubmatch(caminho)
	if m == nil {
		return 0, 0, false
	}
	fmt.Sscan(m[1], &ano)
	fmt.Sscan(m[2], &mes)
	return ano, mes, mes >= 1 && mes <= 12
}

// NovaFatura é a resposta do GET de segunda via.
type NovaFatura struct {
	URL string `json:"url"`
}

// GerarSegundaVia faz o GET com efeito colateral que registra a segunda via (é uma escrita:
// vai pelo Sender, com confirmação e --dry-run).
func GerarSegundaVia(ctx context.Context, s Sender, caminho string) (*NovaFatura, error) {
	var r NovaFatura
	if err := s.Send(ctx, "GET", caminho, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// uuidFatura é o formato do idFatura do checkout.
var uuidFatura = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IDFaturaDaURL tira o idFatura da URL do checkout
// ("…/checkout/#/fatura-contabilizei?idFatura=<UUID>"); a query fica no fragmento.
func IDFaturaDaURL(raw string) (string, bool) {
	_, depois, ok := strings.Cut(raw, "idFatura=")
	if !ok {
		return "", false
	}
	id, _, _ := strings.Cut(depois, "&")
	id, err := url.QueryUnescape(id)
	if err != nil || !uuidFatura.MatchString(id) {
		return "", false
	}
	return id, true
}

// PathFaturaPublica é a fatura com os meios de pagamento (v2 do checkout).
func PathFaturaPublica(id string) string { return "/api/fintech/public/v2/fatura-contabilizei/" + id }

// PathFaturaPublicaResumo é a fatura com os itens e o status (v1 do checkout).
func PathFaturaPublicaResumo(id string) string {
	return "/api/fintech/public/fatura-contabilizei/" + id
}

// MeioPagamento é um meio de pagamento da fatura pública. No boleto, textoPagamento é a linha
// digitável e urlPagamento, apesar do nome, o código de barras. No Pix, textoPagamento é o
// copia e cola (urlPagamento é o QR em data-URI, que a CLI não usa).
type MeioPagamento struct {
	TextoPagamento string `json:"textoPagamento"`
	URLPagamento   string `json:"urlPagamento" contract:"optional"`
	Exibir         bool   `json:"exibir"`
}

// FaturaPublica é a resposta de PathFaturaPublica (ADR-0021). O campo token da resposta não
// é lido nem mostrado.
type FaturaPublica struct {
	DataVencimento  any      `json:"dataVencimento"` // formato não verificado
	Total           *float64 `json:"total"`
	MeioDePagamento struct {
		Pix    *MeioPagamento `json:"pix" contract:"optional"`
		Boleto *MeioPagamento `json:"boleto" contract:"optional"`
	} `json:"meioDePagamento"`
}

// ItemFaturaPublica é uma linha do resumo da fatura (mensalidade, multa e juros, serviços…).
type ItemFaturaPublica struct {
	Descricao string   `json:"descricao"`
	Valor     *float64 `json:"valor"`
}

// FaturaPublicaResumo é a resposta de PathFaturaPublicaResumo (ADR-0021).
type FaturaPublicaResumo struct {
	Resumo struct {
		Itens []ItemFaturaPublica `json:"itens"`
	} `json:"resumo"`
	Status *struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao" contract:"optional"`
	} `json:"status" contract:"optional"`
}

// BuscarFaturaPublica lê a fatura pública com os meios de pagamento.
func BuscarFaturaPublica(ctx context.Context, g Getter, id string) (*FaturaPublica, error) {
	return get[FaturaPublica](ctx, g, PathFaturaPublica(id))
}

// BuscarFaturaPublicaResumo lê os itens e o status da fatura pública.
func BuscarFaturaPublicaResumo(ctx context.Context, g Getter, id string) (*FaturaPublicaResumo, error) {
	return get[FaturaPublicaResumo](ctx, g, PathFaturaPublicaResumo(id))
}
