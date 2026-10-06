package api

import (
	"context"
	"encoding/json"
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
