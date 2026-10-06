package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Envio de documentos para pendências da Central de Rotinas. Ver
// docs/api/escrita/contabilidade-e-documentos.md, seções 4.1 e 4.2.

// PathEnvioDocumentoInit descreve os documentos pedidos pelas pendências (GET).
func PathEnvioDocumentoInit(idsPendencia []string) string {
	q := make([]string, len(idsPendencia))
	for i, id := range idsPendencia {
		q[i] = "id=" + url.QueryEscape(id)
	}
	return "documentos/envio-documento/init?" + strings.Join(q, "&")
}

// CompetenciaDocumento é a competência de um documento pedido.
type CompetenciaDocumento struct {
	Mes     int             `json:"mes"`
	Ano     int             `json:"ano"`
	Periodo json.RawMessage `json:"periodo" contract:"optional"`
}

// PropriedadesDocumento são os dados de uma pendência de documento, reenviados como metadados.
// Ids e valores ficam como vieram (ADR-0021).
type PropriedadesDocumento struct {
	IDPendencia      json.RawMessage       `json:"idPendencia"`
	IDContaBancaria  json.RawMessage       `json:"idContaBancaria" contract:"optional"`
	TipoInvestimento json.RawMessage       `json:"tipoInvestimento" contract:"optional"`
	Valor            json.RawMessage       `json:"valor" contract:"optional"`
	Competencia      *CompetenciaDocumento `json:"competencia" contract:"optional"`
	Banco            string                `json:"banco" contract:"optional"`
	Agencia          string                `json:"agencia" contract:"optional"`
	ContaCorrente    string                `json:"contaCorrente" contract:"optional"`
}

// IDTexto devolve o id da pendência sem aspas.
func (p PropriedadesDocumento) IDTexto() string { return strings.Trim(string(p.IDPendencia), `"`) }

// DocumentoPendente é um documento pedido por uma pendência.
type DocumentoPendente struct {
	Tipo         string                `json:"tipo"`
	Propriedades PropriedadesDocumento `json:"propriedades"`
}

// EnvioDocumentoInit é a resposta de PathEnvioDocumentoInit (ADR-0021).
type EnvioDocumentoInit struct {
	TiposPermitidos []string            `json:"tiposPermitidos" contract:"optional"`
	Documentos      []DocumentoPendente `json:"documentos"`
}

// BuscarEnvioDocumento lê os documentos pedidos pelas pendências.
func BuscarEnvioDocumento(ctx context.Context, g Getter, idsPendencia []string) (*EnvioDocumentoInit, error) {
	return get[EnvioDocumentoInit](ctx, g, PathEnvioDocumentoInit(idsPendencia))
}

// TipoExtratoMovimentacoes é enviado por outro fluxo (importação de extrato).
const TipoExtratoMovimentacoes = "EXTRATO_BANCARIO_MOVIMENTACOES"

// metadadosDocumento é o JSON "metadados", na ordem do front; nulos quando não se aplicam.
type metadadosDocumento struct {
	Valor            json.RawMessage `json:"valor,omitempty"`
	IDPendencia      json.RawMessage `json:"idPendencia"`
	Periodo          json.RawMessage `json:"periodo"`
	IDContaBancaria  json.RawMessage `json:"idContaBancaria"`
	TipoInvestimento json.RawMessage `json:"tipoInvestimento"`
}

type competenciaJSON struct {
	Mes int `json:"mes"`
	Ano int `json:"ano"`
}

func nulo(r json.RawMessage) json.RawMessage {
	if len(r) == 0 {
		return json.RawMessage("null")
	}
	return r
}

func metadados(d DocumentoPendente, comValor bool) metadadosDocumento {
	p := d.Propriedades
	m := metadadosDocumento{IDPendencia: nulo(p.IDPendencia), Periodo: json.RawMessage("null"),
		IDContaBancaria: nulo(p.IDContaBancaria), TipoInvestimento: nulo(p.TipoInvestimento)}
	if p.Competencia != nil {
		m.Periodo = nulo(p.Competencia.Periodo)
	}
	if comValor {
		m.Valor = nulo(p.Valor)
	}
	return m
}

func competencia(d DocumentoPendente) competenciaJSON {
	if c := d.Propriedades.Competencia; c != nil {
		return competenciaJSON{Mes: c.Mes, Ano: c.Ano}
	}
	return competenciaJSON{}
}

func jsonTexto(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// Caminhos de envio com arquivo (multipart).
const (
	PathEnviarDocumento            = "documentos/envio-documento/enviar"
	PathEnviarDocumentoConsolidado = "documentos/envio-documento/enviar/consolidado"
)

// EnviarDocumento envia um arquivo para uma pendência (tipo e metadados vêm do init).
func EnviarDocumento(ctx context.Context, s Sender, d DocumentoPendente, nomeArquivo string, conteudo []byte) error {
	return s.SendMultipart(ctx, "POST", PathEnviarDocumento, []Campo{
		{Nome: "arquivo", Arquivo: nomeArquivo, Conteudo: conteudo, Tipo: TipoArquivo(nomeArquivo)},
		{Nome: "tipoDocumento", Valor: d.Tipo},
		{Nome: "competencia", Valor: jsonTexto(competencia(d))},
		{Nome: "metadados", Valor: jsonTexto(metadados(d, true))},
		{Nome: "ordemArquivoCompetencia", Valor: "null"},
	}, nil)
}

// EnviarDocumentoConsolidado envia um arquivo único para várias pendências do mesmo tipo.
func EnviarDocumentoConsolidado(ctx context.Context, s Sender, ds []DocumentoPendente, nomeArquivo string, conteudo []byte) error {
	type item struct {
		Competencia competenciaJSON    `json:"competencia"`
		Metadados   metadadosDocumento `json:"metadados"`
	}
	itens := make([]item, len(ds))
	for i, d := range ds {
		itens[i] = item{competencia(d), metadados(d, false)}
	}
	tipo := ""
	if len(ds) > 0 {
		tipo = ds[0].Tipo
	}
	return s.SendMultipart(ctx, "POST", PathEnviarDocumentoConsolidado, []Campo{
		{Nome: "arquivo", Arquivo: nomeArquivo, Conteudo: conteudo, Tipo: TipoArquivo(nomeArquivo)},
		{Nome: "tipoDocumento", Valor: tipo},
		{Nome: "documentos", Valor: jsonTexto(itens)},
	}, nil)
}

// TiposSemArquivo são os documentos que aceitam a declaração "não tenho" (sem arquivo).
var TiposSemArquivo = map[string]bool{"ESTOQUE": true, "CONTROLE_DE_INTERMEDIACOES": true, "CONTRATO_DE_AFAC": true}

// PathEnviarSemArquivo fecha pendências declarando que o documento não existe.
const PathEnviarSemArquivo = "documentos/envio-documento/enviar/sem-arquivo"

// PendenciaSemArquivo é um item do corpo de PathEnviarSemArquivo.
type PendenciaSemArquivo struct {
	IDPendencia json.RawMessage `json:"idPendencia"`
	Tipo        string          `json:"tipo"`
}

// EnviarSemArquivo declara a ausência do documento (fecha a pendência, sem desfazer).
func EnviarSemArquivo(ctx context.Context, s Sender, itens []PendenciaSemArquivo) error {
	return s.Send(ctx, "POST", PathEnviarSemArquivo, itens, nil)
}

// TipoExtratoAplicacao é o extrato de aplicação financeira, que tem declaração própria.
const TipoExtratoAplicacao = "EXTRATO_APLICACAO_FINANCEIRA"

// PathSemAplicacao declara "não tenho aplicação nesta conta" para as competências pendentes.
const PathSemAplicacao = "upload-documentos/extrato-aplicacao-financeira/enviar/sem-aplicacao-financeira"

// CompetenciaPendente é uma competência da declaração sem aplicação.
type CompetenciaPendente struct {
	ID  json.RawMessage `json:"id"` // id da pendência
	Mes int             `json:"mes"`
	Ano int             `json:"ano"`
}

// DeclaracaoSemAplicacao é o corpo de PathSemAplicacao.
type DeclaracaoSemAplicacao struct {
	TipoDocumento         string                `json:"tipoDocumento"`
	CompetenciasPendentes []CompetenciaPendente `json:"competenciasPendentes"`
	IDContaBancaria       json.RawMessage       `json:"idContaBancaria"`
	TipoPendencia         string                `json:"tipoPendencia"`
}

// ResultadoSemAplicacao traz as competências que falharam (sucesso parcial).
type ResultadoSemAplicacao struct {
	CompetenciasComErro []competenciaJSON `json:"competenciasComErro"`
}

// CompetenciasComErroTexto formata as competências que falharam como MM/AAAA.
func (r ResultadoSemAplicacao) CompetenciasComErroTexto() []string {
	var out []string
	for _, c := range r.CompetenciasComErro {
		out = append(out, fmt.Sprintf("%02d/%d", c.Mes, c.Ano))
	}
	return out
}

// DeclararSemAplicacao declara que não houve aplicação na conta nas competências.
func DeclararSemAplicacao(ctx context.Context, s Sender, d DeclaracaoSemAplicacao) (*ResultadoSemAplicacao, error) {
	var r ResultadoSemAplicacao
	if err := s.Send(ctx, "POST", PathSemAplicacao, d, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
