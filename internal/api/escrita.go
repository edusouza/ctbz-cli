package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"
)

// Sender envia uma escrita (POST, PUT, PATCH ou DELETE) autenticada e decodifica a resposta
// em v. A implementação confere a sessão antes e envia uma única vez, sem retentativa
// (ADR-0018). v pode ser nil (resposta ignorada) ou *string (resposta em texto puro).
type Sender interface {
	// Send envia body como JSON. body pode ser nil (sem corpo), []byte ou json.RawMessage
	// (enviados como estão), JSONString (string JSON) ou qualquer valor serializável.
	Send(ctx context.Context, method, path string, body any, v any) error
	// SendMultipart envia um formulário multipart/form-data (upload de arquivos).
	SendMultipart(ctx context.Context, method, path string, campos []Campo, v any) error
}

// Campo é um campo de um formulário multipart: texto (Valor) ou arquivo (Arquivo e Conteudo).
type Campo struct {
	Nome  string
	Valor string
	// Arquivo é o nome do arquivo enviado; vazio indica um campo de texto.
	Arquivo  string
	Conteudo []byte
	// Tipo é o Content-Type do arquivo; vazio deduz pela extensão do nome.
	Tipo string
}

// JSONString envia V serializado como uma string JSON, como o painel faz em algumas
// escritas (JSON.stringify do objeto e o resultado como corpo JSON).
type JSONString struct{ V any }

// MarshalJSON serializa V e devolve o resultado como string JSON.
func (s JSONString) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(s.V)
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(b))
}

// EncodeBody serializa o corpo de Sender.Send; nil devolve nil (requisição sem corpo).
func EncodeBody(body any) ([]byte, error) {
	switch b := body.(type) {
	case nil:
		return nil, nil
	case []byte:
		return b, nil
	case json.RawMessage:
		return b, nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("montando o corpo da requisição: %w", err)
	}
	return data, nil
}

// EncodeMultipart monta o corpo de Sender.SendMultipart e devolve o Content-Type com o boundary.
func EncodeMultipart(campos []Campo) ([]byte, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for _, c := range campos {
		if err := writeCampo(w, c); err != nil {
			return nil, "", fmt.Errorf("montando o campo %q: %w", c.Nome, err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}

func writeCampo(w *multipart.Writer, c Campo) error {
	if c.Arquivo == "" {
		return w.WriteField(c.Nome, c.Valor)
	}
	tipo := c.Tipo
	if tipo == "" {
		tipo = mime.TypeByExtension(strings.ToLower(filepath.Ext(c.Arquivo)))
	}
	if tipo == "" {
		tipo = "application/octet-stream"
	}
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, c.Nome, c.Arquivo))
	h.Set("Content-Type", tipo)
	part, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = part.Write(c.Conteudo)
	return err
}

// DecodeResponse decodifica a resposta de uma escrita em v: nil ignora, *string recebe o
// texto puro e corpo vazio deixa v como está.
func DecodeResponse(data []byte, v any) error {
	switch p := v.(type) {
	case nil:
		return nil
	case *string:
		*p = string(data)
		return nil
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("resposta inesperada: %w", err)
	}
	return nil
}

// Escrita liga uma escrita a uma chamada de exemplo, com valores fictícios, para os testes
// de requisição: a requisição gerada é comparada com testdata/requisicoes/<Name>.json.
type Escrita struct {
	Name    string
	Exemplo func(ctx context.Context, s Sender) error
}

// Escritas lista todas as escritas tipadas, na ordem do roadmap. Toda escrita nova entra
// aqui; o teste de cobertura falha se faltar o golden.
func Escritas() []Escrita {
	return []Escrita{
		{Name: "caixa_salvar_lancamento", Exemplo: func(ctx context.Context, s Sender) error {
			return SalvarLancamento(ctx, s, SalvarLancamentoCaixa{Ano: "2026", Mes: 9, LancamentoUsuario: LancamentoCaixaUsuario{
				Data: DataISOBrasilia(time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)), Descricao: "Guia do Simples",
				IDContaUsuario: 1000000000000013, IDVinculo: json.RawMessage(`"1000000000000021"`), Valor: -150.25,
			}})
		}},
		{Name: "caixa_editar_lancamento", Exemplo: func(ctx context.Context, s Sender) error {
			id := int64(1000000000000002)
			return SalvarLancamento(ctx, s, SalvarLancamentoCaixa{Ano: "2026", Mes: 9, LancamentoUsuario: LancamentoCaixaUsuario{
				Data: DataISOBrasilia(time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)), Descricao: "Venda à vista", ID: &id,
				IDContaUsuario: 1000000000000011, Valor: 900,
			}})
		}},
	}
}
