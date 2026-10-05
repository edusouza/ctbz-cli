package ctbz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// WriteError descreve uma escrita recusada pelo servidor, com a mensagem que o painel
// mostraria. Não inclui cabeçalhos nem cookies, e o corpo só aparece resumido.
type WriteError struct {
	Method  string
	Path    string
	Status  int
	Message string
	// Identificador é o código do erro de negócio (detalhes[0].identificador, ex.:
	// "exception/movimentacao-financeira-901"), quando a resposta traz.
	Identificador string
}

func (e *WriteError) Error() string {
	return fmt.Sprintf("%s %s: %s (HTTP %d)", e.Method, e.Path, e.Message, e.Status)
}

// NewWriteError traduz uma resposta de erro de escrita numa mensagem em português.
func NewWriteError(method, path string, resp *Response) *WriteError {
	msg := statusText(resp.Status)
	if detalhe := errorDetail(resp.Status, resp.Body); detalhe != "" {
		msg += ": " + detalhe
	}
	return &WriteError{Method: method, Path: path, Status: resp.Status, Message: msg, Identificador: identificador(resp.Body)}
}

func identificador(body []byte) string {
	var v struct {
		Detalhes []struct {
			Identificador string `json:"identificador"`
		} `json:"detalhes"`
	}
	if json.Unmarshal(body, &v) != nil || len(v.Detalhes) == 0 {
		return ""
	}
	return v.Detalhes[0].Identificador
}

func statusText(status int) string {
	switch {
	case status == http.StatusBadRequest:
		return "requisição recusada"
	case status == http.StatusUnauthorized:
		return "sessão expirada; rode `ctbz login`"
	case status == http.StatusForbidden:
		return "sem permissão"
	case status == http.StatusNotFound:
		return "não encontrado"
	case status == http.StatusNotAcceptable || status == http.StatusUnprocessableEntity:
		return "dados recusados pelo servidor"
	case status >= 300 && status < 400:
		return "redirecionado pelo servidor (sessão inválida?)"
	case status == 560:
		return "recusado pela regra de negócio da Contabilizei"
	case status >= 500:
		return "erro no servidor da Contabilizei"
	default:
		return "resposta inesperada"
	}
}

// maxDetail limita o trecho do corpo mostrado ao usuário.
const maxDetail = 300

// errorDetail extrai a mensagem do corpo de erro, nos formatos usados pela API:
// {"detalhes":[{"detalhe":…}]}, {"message":…}, string JSON, texto puro ou HTML.
func errorDetail(status int, body []byte) string {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return ""
	}
	if body[0] == '<' {
		if status == http.StatusNotFound {
			return "o servidor respondeu uma página HTML (caminho incompleto ou inexistente)"
		}
		return "o servidor respondeu uma página HTML"
	}
	if msg, ok := jsonMessage(body); ok {
		return truncate(msg)
	}
	return truncate(string(body))
}

func jsonMessage(body []byte) (string, bool) {
	var str string
	if json.Unmarshal(body, &str) == nil {
		return str, str != ""
	}
	var v struct {
		Detalhes []struct {
			Detalhe string `json:"detalhe"`
		} `json:"detalhes"`
		Message  string `json:"message"`
		Mensagem string `json:"mensagem"`
		Error    any    `json:"error"`
		Erro     string `json:"erro"`
	}
	if json.Unmarshal(body, &v) != nil {
		return "", false
	}
	for _, d := range v.Detalhes {
		if d.Detalhe != "" {
			return d.Detalhe, true
		}
	}
	for _, m := range []string{v.Message, v.Mensagem, v.Erro} {
		if m != "" {
			return m, true
		}
	}
	if e, ok := v.Error.(string); ok && e != "" {
		return e, true
	}
	return "", false
}

func truncate(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > maxDetail {
		return string(r[:maxDetail]) + "…"
	}
	return s
}
