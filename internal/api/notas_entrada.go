package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Listagens do front de notas de entrada (/nota-entrada/), que usa a base /api/emissor/.
const (
	PathNotasEntradaManifestacao  = "/api/emissor/notasentrada/listar/"
	PathNotasEntradaClassificacao = "/api/emissor/classificacaonotas/listar/"
)

// Listas de notas de entrada, como as abas da tela.
const (
	ListaAManifestar   = "a-manifestar"
	ListaManifestadas  = "manifestadas"
	ListaAClassificar  = "a-classificar"
	ListaClassificadas = "classificadas"
)

// NotaEntrada é uma NF-e recebida pela empresa (nota de compra). Os campos são os do
// exemplo do tutorial do front; a conta verificada não tinha notas de entrada.
type NotaEntrada struct {
	ID           any      `json:"id" contract:"optional"`
	Chave        string   `json:"chave" contract:"optional"`
	CNPJEmitente string   `json:"cnpjEmitente" contract:"optional"`
	RazaoSocial  string   `json:"razaoSocial" contract:"optional"` // do emitente
	DataEmissao  any      `json:"dataEmissao" contract:"optional"` // epoch em ms no exemplo
	Valor        *float64 `json:"valor" contract:"optional"`
	Situacao     *struct {
		ID        string `json:"id" contract:"optional"`
		Descricao string `json:"descricao" contract:"optional"`
	} `json:"situacao" contract:"optional"`
	// TipoClassificacao, nas listas de classificação: ESTOQUE, INSUMO… ou PROCESSANDO
	// enquanto uma reclassificação é aplicada (citado no front).
	TipoClassificacao string `json:"tipoClassificacao" contract:"optional"`
}

// ListaNotasEntrada é uma página das listagens de notas de entrada.
type ListaNotasEntrada struct {
	List   []NotaEntrada `json:"list"`
	Total  int           `json:"total"`
	Cursor *string       `json:"cursor"`
}

// FiltroNotasEntrada seleciona as notas de entrada de um mês (1–12).
type FiltroNotasEntrada struct {
	Lista    string // um dos Lista*
	Ano, Mes int
	Emitente string // filtro pela razão social do emitente (campo "empresa" da API)
}

// tamanhoPaginaEntrada é o tamanho de página usado nas listagens de entrada.
const tamanhoPaginaEntrada = 10

// BuscarNotasEntrada lê todas as páginas de uma lista de notas de entrada.
func BuscarNotasEntrada(ctx context.Context, g Getter, f FiltroNotasEntrada) ([]NotaEntrada, error) {
	brutas, err := BuscarNotasEntradaBrutas(ctx, g, f)
	if err != nil {
		return nil, err
	}
	out := make([]NotaEntrada, len(brutas))
	for i, b := range brutas {
		if err := json.Unmarshal(b, &out[i]); err != nil {
			return nil, fmt.Errorf("nota de entrada: %w", err)
		}
	}
	return out, nil
}

// BuscarNotasEntradaBrutas é BuscarNotasEntrada com cada nota como veio do servidor: a
// classificação em lote reenvia os objetos inteiros (ADR-0022).
func BuscarNotasEntradaBrutas(ctx context.Context, g Getter, f FiltroNotasEntrada) ([]json.RawMessage, error) {
	type pagina struct {
		List   []json.RawMessage `json:"list"`
		Total  int               `json:"total"`
		Cursor *string           `json:"cursor"`
	}
	var out []json.RawMessage
	cursor := ""
	for n := 0; n < maxPaginas; n++ {
		path, err := pathNotasEntrada(f, cursor, len(out))
		if err != nil {
			return nil, err
		}
		l, err := get[pagina](ctx, g, path)
		if err != nil {
			return nil, err
		}
		out = append(out, l.List...)
		if len(l.List) == 0 || len(out) >= l.Total {
			break
		}
		if l.Cursor != nil {
			cursor = *l.Cursor
		}
	}
	return out, nil
}

// pathNotasEntrada monta a URL de uma página: a manifestação pagina por cursor e a
// classificação por offset, como o front.
func pathNotasEntrada(f FiltroNotasEntrada, cursor string, offset int) (string, error) {
	q := fmt.Sprintf("?mes=%d&ano=%d&empresa=%s", f.Mes, f.Ano, url.QueryEscape(f.Emitente))
	switch f.Lista {
	case ListaAManifestar, ListaManifestadas:
		situacao := 0
		if f.Lista == ListaManifestadas {
			situacao = 1
		}
		return fmt.Sprintf("%s%d%s&qtdPagina=%d&cursor=%s", PathNotasEntradaManifestacao, situacao, q,
			tamanhoPaginaEntrada, url.QueryEscape(cursor)), nil
	case ListaAClassificar, ListaClassificadas:
		tipo := 0
		if f.Lista == ListaClassificadas {
			tipo = 1
		}
		return fmt.Sprintf("%s%s&tipo=%d&limite=%d&cursor=&offset=%d", PathNotasEntradaClassificacao, q, tipo,
			tamanhoPaginaEntrada, offset), nil
	}
	return "", fmt.Errorf("lista de notas de entrada desconhecida: %q", f.Lista)
}
