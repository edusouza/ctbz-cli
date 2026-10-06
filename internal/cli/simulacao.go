package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// requisicaoSimulada é uma escrita registrada pelo --dry-run em vez de enviada.
type requisicaoSimulada struct {
	Metodo      string
	Caminho     string
	ContentType string
	Corpo       any         // JSON decodificado (com segredos mascarados) ou nil
	Campos      []api.Campo // multipart (o conteúdo dos arquivos não é mostrado)
}

// dryRunSender implementa api.Sender sem enviar nada: guarda cada requisição, com o corpo
// montado pelos mesmos auxiliares do envio real, e responde como se a resposta fosse vazia.
type dryRunSender struct{ reqs *[]requisicaoSimulada }

func (d dryRunSender) Send(_ context.Context, method, path string, body any, _ any) error {
	data, err := api.EncodeBody(body)
	if err != nil {
		return err
	}
	r := requisicaoSimulada{Metodo: method, Caminho: ctbz.ResolvePath(path)}
	if data != nil {
		r.ContentType = "application/json"
		r.Corpo = mascarar(decodificar(data))
	}
	*d.reqs = append(*d.reqs, r)
	return nil
}

func (d dryRunSender) SendMultipart(_ context.Context, method, path string, campos []api.Campo, _ any) error {
	if _, _, err := api.EncodeMultipart(campos); err != nil {
		return err
	}
	*d.reqs = append(*d.reqs, requisicaoSimulada{Metodo: method, Caminho: ctbz.ResolvePath(path),
		ContentType: "multipart/form-data", Campos: campos})
	return nil
}

func decodificar(data []byte) any {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if dec.Decode(&v) != nil {
		return string(data)
	}
	return v
}

// chavesSecretas identificam campos mostrados como *** no --dry-run (ADR-0018).
var chavesSecretas = []string{"senha", "password", "secret", "token", "otp", "chaveacesso"}

func secreta(chave string) bool {
	k := strings.ToLower(chave)
	if k == "codigo" || k == "code" {
		return true
	}
	for _, s := range chavesSecretas {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}

func mascarar(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			if secreta(k) && e != nil {
				x[k] = "***"
			} else {
				x[k] = mascarar(e)
			}
		}
	case []any:
		for i, e := range x {
			x[i] = mascarar(e)
		}
	}
	return v
}

func descreverCampo(c api.Campo) string {
	if c.Arquivo == "" {
		if secreta(c.Nome) {
			return "***"
		}
		return c.Valor
	}
	return fmt.Sprintf("arquivo %s (%d bytes)", c.Arquivo, len(c.Conteudo))
}

// escreverSimulacao imprime as requisições: em texto na tabela, como lista em JSON e CSV.
func escreverSimulacao(w io.Writer, f output.Format, reqs []requisicaoSimulada) error {
	if f != output.FormatTable {
		l := &output.List{Columns: []output.Column{
			{Key: "metodo", Header: "Método"},
			{Key: "caminho", Header: "Caminho"},
			{Key: "content_type", Header: "Content-Type"},
			{Key: "corpo", Header: "Corpo"},
			{Key: "campos", Header: "Campos"},
		}}
		for _, r := range reqs {
			var campos any
			if r.Campos != nil {
				cs := make([]map[string]string, len(r.Campos))
				for i, c := range r.Campos {
					cs[i] = map[string]string{"nome": c.Nome, "valor": descreverCampo(c)}
				}
				campos = cs
			}
			l.Append(r.Metodo, r.Caminho, nilIfEmpty(r.ContentType), r.Corpo, campos)
		}
		return output.Write(w, f, l)
	}
	for i, r := range reqs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s %s\n", r.Metodo, r.Caminho)
		if r.ContentType != "" {
			fmt.Fprintf(w, "Content-Type: %s\n", r.ContentType)
		}
		if r.Corpo != nil {
			b, err := json.MarshalIndent(r.Corpo, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "\n%s\n", b)
		}
		if len(r.Campos) > 0 {
			fmt.Fprintln(w)
			for _, c := range r.Campos {
				fmt.Fprintf(w, "%s: %s\n", c.Nome, descreverCampo(c))
			}
		}
	}
	return nil
}
