package api

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// PathReclassificarInit lê os lançamentos e as opções de classificação das pendências de
// reclassificação da Central de Rotinas. Ver docs/api/escrita/contabilidade-e-documentos.md, 3.2.
func PathReclassificarInit(idsPendencia []string) string {
	q := make([]string, len(idsPendencia))
	for i, id := range idsPendencia {
		q[i] = "id=" + url.QueryEscape(id)
	}
	return "documentos/reclassificar/init?" + strings.Join(q, "&")
}

// SocioReclassificacao é um sócio oferecido numa opção de classificação.
type SocioReclassificacao struct {
	IDSocio int64  `json:"idSocio"`
	Nome    string `json:"nome"`
}

// OpcaoReclassificacao é uma classificação aceita para o lançamento da pendência.
type OpcaoReclassificacao struct {
	IDClassificacao int64                  `json:"idClassificacao"`
	Nome            string                 `json:"nome"`
	Socios          []SocioReclassificacao `json:"socios" contract:"optional"`
}

// PendenciaReclassificacao é um item de PathReclassificarInit (ADR-0021).
type PendenciaReclassificacao struct {
	// IDPendencia é reenviado como veio (o painel o trata como texto).
	IDPendencia        json.RawMessage        `json:"idPendencia"`
	IDLancamento       int64                  `json:"idLancamento"`
	ClassificacaoAtual any                    `json:"classificacaoAtual" contract:"optional"`
	Data               Data                   `json:"data" contract:"optional"`
	Valor              *float64               `json:"valor"`
	Descricao          string                 `json:"descricao"`
	Banco              string                 `json:"banco" contract:"optional"`
	Classificacoes     []OpcaoReclassificacao `json:"classificacoes"`
}

// IDTexto devolve o id da pendência sem aspas.
func (p PendenciaReclassificacao) IDTexto() string { return strings.Trim(string(p.IDPendencia), `"`) }

// BuscarReclassificacoes lê as pendências de reclassificação pedidas.
func BuscarReclassificacoes(ctx context.Context, g Getter, idsPendencia []string) ([]PendenciaReclassificacao, error) {
	ps, err := get[[]PendenciaReclassificacao](ctx, g, PathReclassificarInit(idsPendencia))
	if err != nil {
		return nil, err
	}
	return *ps, nil
}

// PathReclassificar grava a nova classificação e conclui as pendências.
const PathReclassificar = "documentos/reclassificar"

// Reclassificacao é um item do corpo de PathReclassificar.
type Reclassificacao struct {
	IDPendencia     json.RawMessage `json:"idPendencia"`
	IDClassificacao int64           `json:"idClassificacao"`
	IDSocio         *int64          `json:"idSocio"`
	IDLancamento    int64           `json:"idLancamento"`
}

// Reclassificar reclassifica os lançamentos e conclui as pendências (a rotina não reabre).
func Reclassificar(ctx context.Context, s Sender, itens []Reclassificacao) error {
	return s.Send(ctx, "POST", PathReclassificar, itens, nil)
}
