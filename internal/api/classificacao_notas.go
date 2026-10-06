package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Classificação das notas de entrada (base /api/emissor/, ADR-0022). Ver
// docs/api/escrita/notas-e-prolabore.md, seção 3.2.
const (
	PathClassificarLote   = "/api/emissor/classificacaonotas/salvarloteclassificacao/"
	PathProdutosNota      = "/api/emissor/classificacaonotas/listarprodutos/"
	PathClassificarNota   = "/api/emissor/classificacaonotas/salvarclassificacao/"
	PathReclassificarNota = "/api/emissor/classificacaonotas/reclassificar"
)

// Classificações de uma nota inteira, como o lote as recebe no caminho.
const (
	ClassificacaoEstoque          = "ESTOQUE"
	ClassificacaoInsumo           = "INSUMO"
	ClassificacaoUsoConsumo       = "USO_CONSUMO"
	ClassificacaoAtivoImobilizado = "ATIVO_IMOBILIZADO"
	ClassificacaoPrestacaoServico = "PRESTACAO_SERVICO"
)

// ClassificarNotasLote classifica notas inteiras numa das Classificacao*. O corpo são as notas
// da lista "a classificar" como vieram (BuscarNotasEntradaBrutas).
func ClassificarNotasLote(ctx context.Context, s Sender, classificacao string, notas []json.RawMessage) error {
	return s.Send(ctx, "POST", PathClassificarLote+url.PathEscape(classificacao), notas, nil)
}

// ProdutoNota é um item de uma nota de entrada, com a distribuição das quantidades entre as
// classificações. Campos citados no front; a conta verificada não tinha notas de entrada.
type ProdutoNota struct {
	ID                  any      `json:"id" contract:"optional"`
	Descricao           string   `json:"descricao" contract:"optional"`
	NCM                 string   `json:"ncm" contract:"optional"`
	Valor               *float64 `json:"valor" contract:"optional"`
	QuantidadeTotal     float64  `json:"quantidadeTotal" contract:"optional"`
	QuantidadeEstoque   float64  `json:"quantidadeEstoque" contract:"optional"`
	QuantidadeInsumo    float64  `json:"quantidadeInsumo" contract:"optional"`
	QuantidadeAtivo     float64  `json:"quantidadeAtivo" contract:"optional"`
	QuantidadeConsumo   float64  `json:"quantidadeConsumo" contract:"optional"`
	QuantidadePrestacao float64  `json:"quantidadePrestacao" contract:"optional"`
}

// ProdutoBruto é um produto como veio do servidor, com a leitura tipada ao lado.
type ProdutoBruto struct {
	ProdutoNota
	Bruto json.RawMessage
}

// ListarProdutosNota lê os itens de uma nota de entrada.
func ListarProdutosNota(ctx context.Context, g Getter, idNota string) ([]ProdutoBruto, error) {
	brutos, err := get[[]json.RawMessage](ctx, g, PathProdutosNota+url.PathEscape(idNota))
	if err != nil {
		return nil, err
	}
	out := make([]ProdutoBruto, len(*brutos))
	for i, b := range *brutos {
		out[i].Bruto = b
		if err := json.Unmarshal(b, &out[i].ProdutoNota); err != nil {
			return nil, fmt.Errorf("produto da nota %s: %w", idNota, err)
		}
	}
	return out, nil
}
