package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Escritório virtual (correspondências). Ver docs/api/escrita/pendencias-e-conta.md, seção
// 3.4. A conta de desenvolvimento não contratou o serviço (HTTP 560): os formatos vêm do
// front e são sintéticos (ADR-0021).
const (
	PathCorrespondencias       = "escritorio-virtual/recuperar-mensagens"
	PathEnderecoEntrega        = "escritorio-virtual/endereco-entrega"
	PathAutorizacaoRecebimento = "escritorio-virtual/verifica-autorizacao-recebimento"
	PathSalvarEnderecoEntrega  = "escritorio-virtual/salvar-endereco-correspondencia"
	// StatusServicoNaoContratado é a resposta das leituras sem o serviço contratado.
	StatusServicoNaoContratado = 560
)

// Correspondencia é uma correspondência recebida no escritório virtual.
type Correspondencia struct {
	ID        any      `json:"id" contract:"optional"`
	Remetente string   `json:"remetente" contract:"optional"`
	Descricao string   `json:"descricao" contract:"optional"`
	Data      any      `json:"dataRecebimento" contract:"optional"`
	Taxa      *float64 `json:"taxa" contract:"optional"` // "Valor do envio"
	Status    string   `json:"status" contract:"optional"`
}

// PaginaCorrespondencias é uma página de correspondências (paginação por meta.cursor).
type PaginaCorrespondencias struct {
	Data []Correspondencia `json:"data" contract:"optional"`
	Meta struct {
		Cursor *string `json:"cursor" contract:"optional"`
		Count  int     `json:"count" contract:"optional"`
	} `json:"meta" contract:"optional"`
}

// BuscarCorrespondencias lê todas as páginas de correspondências.
func BuscarCorrespondencias(ctx context.Context, g Getter) ([]Correspondencia, error) {
	var out []Correspondencia
	cursor := ""
	for range maxPaginas {
		p, err := get[PaginaCorrespondencias](ctx, g, PathCorrespondencias+"?cursor="+url.QueryEscape(cursor))
		if err != nil {
			return nil, err
		}
		out = append(out, p.Data...)
		if len(p.Data) == 0 || p.Meta.Cursor == nil || *p.Meta.Cursor == "" || *p.Meta.Cursor == cursor {
			break
		}
		cursor = *p.Meta.Cursor
	}
	return out, nil
}

// LinkCorrespondencia é o link assinado para baixar uma correspondência.
func LinkCorrespondencia(ctx context.Context, g Getter, id string) (*LinkDownload, error) {
	return get[LinkDownload](ctx, g, "escritorio-virtual/gerar-link-download-correspondencia/"+url.PathEscape(id))
}

// EnderecoEntrega é o endereço para onde as correspondências são enviadas.
type EnderecoEntrega struct {
	CEP         string `json:"cep" contract:"optional"`
	Logradouro  string `json:"logradouro" contract:"optional"`
	Numero      string `json:"numero" contract:"optional"`
	Complemento string `json:"complemento" contract:"optional"`
	Bairro      string `json:"bairro" contract:"optional"`
	Cidade      string `json:"cidade" contract:"optional"`
	UF          string `json:"uf" contract:"optional"`
}

// BuscarEnderecoEntrega lê o endereço de envio das correspondências.
func BuscarEnderecoEntrega(ctx context.Context, g Getter) (*EnderecoEntrega, error) {
	return get[EnderecoEntrega](ctx, g, PathEnderecoEntrega)
}

// PesquisarEnderecoCorrespondencia consulta um CEP (8 dígitos) para o endereço de envio.
func PesquisarEnderecoCorrespondencia(ctx context.Context, g Getter, cep string) (*EnderecoEntrega, error) {
	return get[EnderecoEntrega](ctx, g, "escritorio-virtual/pesquisarEndereco/"+url.PathEscape(cep))
}

// SalvarEnderecoCorrespondencia troca o endereço de envio; logradouro e cidade vêm do CEP.
func SalvarEnderecoCorrespondencia(ctx context.Context, s Sender, cep, numero, complemento string) (*EnderecoEntrega, error) {
	var e EnderecoEntrega
	err := s.Send(ctx, "POST", PathSalvarEnderecoEntrega, struct {
		CEP         string `json:"cep"`
		Numero      string `json:"numero"`
		Complemento string `json:"complemento"`
	}{cep, numero, complemento}, &e)
	return &e, err
}

// AutorizacaoRecebimento diz se a empresa autoriza receber correspondências no endereço.
type AutorizacaoRecebimento struct {
	Autoriza bool `json:"autoriza" contract:"optional"`
}

// BuscarAutorizacaoRecebimento lê a autorização de recebimento.
func BuscarAutorizacaoRecebimento(ctx context.Context, g Getter) (*AutorizacaoRecebimento, error) {
	return get[AutorizacaoRecebimento](ctx, g, PathAutorizacaoRecebimento)
}

// AutorizarRecebimento liga ou desliga a autorização de recebimento (reversível).
func AutorizarRecebimento(ctx context.Context, s Sender, autoriza bool) error {
	return s.Send(ctx, "PUT", fmt.Sprintf("escritorio-virtual/autorizar-recebimento/%s", strconv.FormatBool(autoriza)), nil, nil)
}
