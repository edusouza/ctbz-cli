package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// PathNotasEmitidas é a listagem v2 do emissor de NFS-e (a que o painel usa quando
// novo-emissor/v2/feature-flag é verdadeiro).
const PathNotasEmitidas = "novo-emissor/v2/listagem/notas/filtro"

// Campos de filtro aceitos pela listagem (um por vez).
const (
	FiltroNomeTomador = "nomeTomador"
	FiltroDocumento   = "documento"
	FiltroNumeroNota  = "numeroNota"
)

// NotaEmitida é uma NFS-e emitida. Os campos são os lidos pelo front; todos são opcionais e
// os de formato incerto ficam como any (dataEmissao veio como array [ano, mês, dia, h, min, s]
// em 02/11/2026, #1052).
type NotaEmitida struct {
	ID                   any        `json:"id" contract:"optional"`
	Numero               any        `json:"numero" contract:"optional"`
	CPFCNPJTomador       string     `json:"cpfCnpjTomador" contract:"optional"`
	NomeRazaoTomador     string     `json:"nomeRazaoTomador" contract:"optional"`
	ValorServico         *float64   `json:"valorServico" contract:"optional"`
	DataEmissao          any        `json:"dataEmissao" contract:"optional"`
	DataEmissaoFormatada string     `json:"dataEmissaoFormatada" contract:"optional"`
	StatusNotaFiscal     StatusNota `json:"statusNotaFiscal" contract:"optional"`
	SituacaoNota         string     `json:"situacaoNota" contract:"optional"`
	TomadorExterior      bool       `json:"tomadorExterior" contract:"optional"`
}

// StatusNota é o rótulo de status que o servidor monta para a listagem. Vem como objeto
// {descricao, text, badgeType} desde 11/2026 (#1052); antes era texto, que ainda é aceito
// (vai para Text) para não quebrar numa nova troca de formato.
type StatusNota struct {
	Descricao string `json:"descricao"` // ex.: "NFS-e processada com sucesso"
	Text      string `json:"text"`      // ex.: "Emitida", o rótulo que o painel mostra
	BadgeType string `json:"badgeType"` // ex.: "success"
}

// UnmarshalJSON aceita o objeto, um texto ou null.
func (s *StatusNota) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		*s = StatusNota{}
		return json.Unmarshal(b, &s.Text)
	}
	type objeto StatusNota
	var o objeto
	if err := json.Unmarshal(b, &o); err != nil {
		return fmt.Errorf("statusNotaFiscal inesperado %s", b)
	}
	*s = StatusNota(o)
	return nil
}

// Rotulo devolve o texto do status para exibir: Text, ou Descricao quando Text vem vazio.
func (s StatusNota) Rotulo() string {
	if s.Text != "" {
		return s.Text
	}
	return s.Descricao
}

// ListaNotas é uma página da listagem de notas.
type ListaNotas struct {
	List  []NotaEmitida `json:"list"`
	Total int           `json:"total"`
}

// FiltroNotas seleciona as notas de um mês (1–12), opcionalmente filtradas por um campo.
type FiltroNotas struct {
	Ano, Mes     int
	Campo, Termo string // Campo é um dos Filtro*; vazio, sem filtro
}

// tamanhoPaginaNotas é o tamanho de página padrão do painel.
const tamanhoPaginaNotas = 10

// BuscarNotasEmitidas lê todas as páginas de notas emitidas de um mês.
func BuscarNotasEmitidas(ctx context.Context, g Getter, f FiltroNotas) ([]NotaEmitida, error) {
	var out []NotaEmitida
	for pagina := 1; pagina <= maxPaginas; pagina++ {
		path := fmt.Sprintf("%s?pagina=%d&limite=%d&ano=%d&mes=%d", PathNotasEmitidas, pagina, tamanhoPaginaNotas, f.Ano, f.Mes)
		if f.Campo != "" {
			path += "&" + f.Campo + "=" + url.QueryEscape(f.Termo)
		}
		l, err := get[ListaNotas](ctx, g, path)
		if err != nil {
			return nil, err
		}
		out = append(out, l.List...)
		if len(l.List) == 0 || len(out) >= l.Total {
			break
		}
	}
	return out, nil
}
