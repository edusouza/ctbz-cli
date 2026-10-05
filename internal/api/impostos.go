package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const PathGuiasAPagar = "impostos/v5/impostos-a-pagar/guias"

// Rotulo é o formato de exibição usado pelas rotas v5: {"label": …, "descricao": …}.
type Rotulo struct {
	Label string `json:"label"`
}

// ValorRotulo é um valor monetário no formato de exibição v5; nulo enquanto calcula.
type ValorRotulo struct {
	Label *float64 `json:"label"`
}

// GuiaResumo é uma guia (ou parcela) na lista de impostos a pagar.
type GuiaResumo struct {
	ID                   int64       `json:"id"`
	Nome                 Rotulo      `json:"nome"`
	Tipo                 string      `json:"tipo"` // GUIA ou PARCELA
	IdentificadorImposto string      `json:"identificadorImposto"`
	VencimentoOriginal   string      `json:"vencimentoOriginal"` // AAAA-MM-DD
	Valor                ValorRotulo `json:"valor"`
	Competencia          string      `json:"competencia"` // ex.: "Jul de 2026"
	Status               Rotulo      `json:"status"`
}

// GuiasAPagar é a resposta de impostos/v5/impostos-a-pagar/guias.
type GuiasAPagar struct {
	EmAtraso   []GuiaResumo `json:"emAtraso"`
	EsteMes    []GuiaResumo `json:"esteMes"`
	ProximoMes []GuiaResumo `json:"proximoMes"`
}

// BuscarGuiasAPagar lê as guias em atraso, do mês e do próximo mês.
func BuscarGuiasAPagar(ctx context.Context, g Getter) (*GuiasAPagar, error) {
	return get[GuiasAPagar](ctx, g, PathGuiasAPagar)
}

const (
	PathCalculoImposto = "impostos/como-imposto-foi-calculado/init"
	PathTabelaIRRF     = "impostos/como-imposto-foi-calculado/tabela-irrf"
)

// PathGuia é o detalhe de uma guia (o ID vem de GuiasAPagar).
func PathGuia(id int64) string {
	return fmt.Sprintf("impostos/v5/impostos-a-pagar/guia/%d", id)
}

// Montante é um valor no formato {"valor": …, "descricao": …}; nulo quando não se aplica.
type Montante struct {
	Valor *float64 `json:"valor"`
}

// GuiaDetalhe é a resposta de impostos/v5/impostos-a-pagar/guia/{id}.
type GuiaDetalhe struct {
	ID                   int64  `json:"id"`
	Tipo                 string `json:"tipo"`
	Origem               string `json:"origem" contract:"optional"` // usada nas escritas; padrão GUIAS
	IdentificadorImposto string `json:"identificadorImposto"`
	Competencia          string `json:"competencia"`
	Vencimento           string `json:"vencimento"` // dd/mm/aaaa
	Oraculo              struct {
		Nome       string `json:"nome"`
		Descricao  string `json:"descricao"`
		Frequencia string `json:"frequencia"`
		Impacto    string `json:"impacto"`
	} `json:"oraculo"`
	Status           []string  `json:"status"`
	ValorTotal       *Montante `json:"valorTotal"`
	ValorOriginal    *Montante `json:"valorOriginal"`
	ValorJurosEMulta *Montante `json:"valorJurosEMulta"`
	ValorEstimado    *Montante `json:"valorEstimado"`
	ValorEmAtraso    *Montante `json:"valorEmAtraso"`
	AcoesBotoes      []string  `json:"acoesBotoes"`
}

// BuscarGuia lê o detalhe de uma guia.
func BuscarGuia(ctx context.Context, g Getter, id int64) (*GuiaDetalhe, error) {
	return get[GuiaDetalhe](ctx, g, PathGuia(id))
}

// CalculoImposto é a resposta de "como meu imposto foi calculado": a memória de cálculo
// do mês mais recente (DAS do Simples e DARF de INSS/IRRF sobre o pró-labore).
type CalculoImposto struct {
	NomeMesCompetencia string   `json:"nomeMesCompetencia"`
	FaturamentoTotal   *float64 `json:"faturamentoTotal"`
	DasSimples         struct {
		ImpostoBruto    *float64 `json:"impostoBruto"`
		DeducaoRetencao *float64 `json:"deducaoRetencao"`
		ImpostoTotal    *float64 `json:"impostoTotal"`
		Inconsistente   bool     `json:"inconsistente"`
	} `json:"dasSimples"`
	Darf struct {
		INSS struct {
			Prolabore    *float64 `json:"prolabore"`
			Aliquota     *float64 `json:"aliquota"`
			Teto         *float64 `json:"teto"`
			TotalImposto *float64 `json:"totalImposto"`
		} `json:"inss"`
		IRRF struct {
			Prolabore       *float64 `json:"prolabore"`
			BaseCalculoIRRF *float64 `json:"baseCalculoIrrf"`
			Aliquota        *float64 `json:"aliquota"`
			DeducaoIRRF     *float64 `json:"deducaoIrrf"`
			TotalImposto    *float64 `json:"totalImposto"`
		} `json:"irrf"`
		Total *float64 `json:"total"`
	} `json:"darf"`
	ValorFaturamentoUltimos12Meses *float64 `json:"valorFaturamentoUltimos12Meses"`
	ValorProlaboreUltimos12Meses   *float64 `json:"valorProlaboreUltimos12Meses"`
	PercentualFatorR               *float64 `json:"percentualFatorR"`
	HistoricoFaturamento           []struct {
		Mes              string   `json:"mes"` // ex.: "set./26"
		ValorFaturamento *float64 `json:"valorFaturamento"`
		ValorProlabore   *float64 `json:"valorProlabore"`
	} `json:"historicoFaturamento"`
}

// BuscarCalculoImposto lê a memória de cálculo do mês mais recente.
func BuscarCalculoImposto(ctx context.Context, g Getter) (*CalculoImposto, error) {
	return get[CalculoImposto](ctx, g, PathCalculoImposto)
}

// FaixaIRRF é uma faixa da tabela progressiva do IRRF, em texto como a API devolve.
type FaixaIRRF struct {
	BaseCalculo string `json:"baseCalculo"`
	Aliquota    string `json:"aliquota"`
	Deducao     string `json:"deducao"`
}

// BuscarTabelaIRRF lê a tabela progressiva do IRRF vigente.
func BuscarTabelaIRRF(ctx context.Context, g Getter) ([]FaixaIRRF, error) {
	t, err := get[[]FaixaIRRF](ctx, g, PathTabelaIRRF)
	if err != nil {
		return nil, err
	}
	return *t, nil
}

// PathLinkGuia é o link de download do PDF de uma guia ou parcela. A rota é a v3 mesmo
// na tela v5 do painel.
func PathLinkGuia(id int64, tipo string) string {
	if strings.EqualFold(tipo, "PARCELA") {
		return fmt.Sprintf("impostos/v3/impostos-a-pagar/parcela/%d/baixar-parcela", id)
	}
	return fmt.Sprintf("impostos/v3/impostos-a-pagar/guia/%d/baixar-guia", id)
}

// LinkDownload é um link temporário (URL assinada) para baixar um arquivo.
type LinkDownload struct {
	URL string `json:"url"`
}

// BuscarLinkGuia lê o link de download do PDF de uma guia ou parcela.
func BuscarLinkGuia(ctx context.Context, g Getter, id int64, tipo string) (*LinkDownload, error) {
	return get[LinkDownload](ctx, g, PathLinkGuia(id, tipo))
}

const (
	PathHistoricoResumo = "impostos/v2/historico-impostos/init"
	PathHistoricoGuias  = "impostos/v2/historico-impostos/guias"
)

// HistoricoResumo é a resposta de impostos/v2/historico-impostos/init.
type HistoricoResumo struct {
	EmDia                   bool `json:"emDia"`
	QuantidadeGuiasVencidas int  `json:"quantidadeGuiasVencidas"`
}

// BuscarHistoricoResumo lê se a empresa está em dia e quantas guias venceram.
func BuscarHistoricoResumo(ctx context.Context, g Getter) (*HistoricoResumo, error) {
	return get[HistoricoResumo](ctx, g, PathHistoricoResumo)
}

// GuiaHistorico é uma guia do histórico de impostos.
type GuiaHistorico struct {
	ID               int64  `json:"id"`
	Imposto          string `json:"imposto"`
	ImpostoDescricao string `json:"impostoDescricao"`
	Competencia      struct {
		Mes int `json:"mes"`
		Ano int `json:"ano"`
	} `json:"competencia"`
	DataVencimento string   `json:"dataVencimento"` // dd/mm/aaaa
	ValorPrincipal *float64 `json:"valorPrincipal"`
	ValorPago      *float64 `json:"valorPago"`
	Status         string   `json:"status"`
	Tipo           string   `json:"tipo"`
}

// HistoricoGuias é uma página de impostos/v2/historico-impostos/guias.
type HistoricoGuias struct {
	PaginaAtual  int `json:"paginaAtual"`
	TotalPaginas int `json:"totalPaginas"`
	Competencias []struct {
		Competencia string          `json:"competencia"`
		Guias       []GuiaHistorico `json:"guias"`
	} `json:"competencias"`
}

// FiltroHistorico restringe o histórico; zero ou vazio não filtra.
type FiltroHistorico struct {
	Ano, Mes int
	Status   string
}

// BuscarHistoricoGuias lê todas as páginas do histórico de guias.
func BuscarHistoricoGuias(ctx context.Context, g Getter, f FiltroHistorico) ([]GuiaHistorico, error) {
	var out []GuiaHistorico
	for pagina := 1; ; pagina++ {
		q := url.Values{"pagina": {strconv.Itoa(pagina)}}
		if f.Ano > 0 {
			q.Set("ano", strconv.Itoa(f.Ano))
		}
		if f.Mes > 0 {
			q.Set("mes", strconv.Itoa(f.Mes))
		}
		if f.Status != "" {
			q.Set("status", f.Status)
		}
		h, err := get[HistoricoGuias](ctx, g, PathHistoricoGuias+"?"+q.Encode())
		if err != nil {
			return nil, err
		}
		for _, c := range h.Competencias {
			out = append(out, c.Guias...)
		}
		if pagina >= h.TotalPaginas || pagina >= maxPaginas {
			return out, nil
		}
	}
}

// maxPaginas protege contra paginação que nunca termina.
const maxPaginas = 100

const PathDadosGrafico = "impostos/v5/historico-impostos/dados-grafico"

// DadosGrafico é a resposta de dados-grafico: total pago em impostos por mês ("1" a "12").
type DadosGrafico struct {
	Meses map[string]struct {
		TotalPago *float64 `json:"totalPago"`
	} `json:"meses"`
}

// BuscarImpostosPagosNoAno lê o total pago em impostos em cada mês do ano.
func BuscarImpostosPagosNoAno(ctx context.Context, g Getter, ano int) (*DadosGrafico, error) {
	return get[DadosGrafico](ctx, g, fmt.Sprintf("%s?ano=%d", PathDadosGrafico, ano))
}

const (
	// PathImpostosV3 é o init da tela de impostos v3/v4. Mesmo com a tela v5 ativa ele
	// responde, e é a única fonte da lista de parcelamentos (abaParcelamentos).
	PathImpostosV3      = "impostos/v3/impostos-a-pagar/init"
	PathDebitosFederais = "dashboard/informerendimento/debitos-federais"
)

// PathParcelamento é o detalhe de um parcelamento (o ID vem de ParcelamentosV3).
func PathParcelamento(id int64) string {
	return fmt.Sprintf("impostos/parcelamento/detalhes/%d", id)
}

// Parcelamento é um item da aba de parcelamentos. Campos vistos no front; a conta usada
// nos testes não tinha parcelamentos, por isso são todos opcionais no contrato.
type Parcelamento struct {
	IDParcelamento   int64  `json:"idParcelamento" contract:"optional"`
	Titulo           string `json:"titulo" contract:"optional"`
	TipoParcelamento string `json:"tipoParcelamento" contract:"optional"`
	Status           string `json:"status" contract:"optional"`
	ParcelaAtual     *struct {
		NumeroParcela int `json:"numeroParcela" contract:"optional"`
	} `json:"parcelaAtual" contract:"optional"`
}

// ParcelamentosV3 é a parte de impostos/v3/impostos-a-pagar/init usada pela CLI.
type ParcelamentosV3 struct {
	AbaParcelamentos struct {
		EmAndamento []Parcelamento `json:"emAndamento"`
		Ativos      []Parcelamento `json:"ativos"`
		Historico   []Parcelamento `json:"historico"`
	} `json:"abaParcelamentos"`
}

// BuscarParcelamentos lê os parcelamentos em andamento, ativos e encerrados.
func BuscarParcelamentos(ctx context.Context, g Getter) (*ParcelamentosV3, error) {
	return get[ParcelamentosV3](ctx, g, PathImpostosV3)
}

// DebitosFederais é a resposta de dashboard/informerendimento/debitos-federais.
type DebitosFederais struct {
	PossuiDebitosFederais bool `json:"possuiDebitosFederais"`
}

// BuscarDebitosFederais lê se há débitos federais em aberto.
func BuscarDebitosFederais(ctx context.Context, g Getter) (*DebitosFederais, error) {
	return get[DebitosFederais](ctx, g, PathDebitosFederais)
}
