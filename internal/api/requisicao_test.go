package api

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

var update = flag.Bool("update", false, "regrava os goldens de testdata/requisicoes")

// requisicao é o que o golden guarda de cada requisição de escrita.
type requisicao struct {
	Metodo      string      `json:"metodo"`
	Caminho     string      `json:"caminho"`
	Query       string      `json:"query,omitempty"`
	ContentType string      `json:"content_type,omitempty"`
	Corpo       any         `json:"corpo,omitempty"`
	Campos      []campoGrav `json:"campos,omitempty"`
}

// campoGrav é um campo multipart; de arquivos, só nome, tipo e tamanho.
type campoGrav struct {
	Nome    string `json:"nome"`
	Valor   string `json:"valor,omitempty"`
	Arquivo string `json:"arquivo,omitempty"`
	Tipo    string `json:"tipo,omitempty"`
	Tamanho int    `json:"tamanho,omitempty"`
}

// httpSender envia pela mesma pilha da CLI (codificação de api e ctbz.Client), sem sessão.
type httpSender struct{ c *ctbz.Client }

func (h httpSender) Send(ctx context.Context, method, path string, body any, v any) error {
	data, err := EncodeBody(body)
	if err != nil {
		return err
	}
	ct := ""
	if data != nil {
		ct = "application/json"
	}
	return h.do(ctx, method, path, data, ct, v)
}

func (h httpSender) SendMultipart(ctx context.Context, method, path string, campos []Campo, v any) error {
	data, ct, err := EncodeMultipart(campos)
	if err != nil {
		return err
	}
	return h.do(ctx, method, path, data, ct, v)
}

func (h httpSender) do(ctx context.Context, method, path string, body []byte, ct string, v any) error {
	resp, err := h.c.Send(ctx, method, path, bytes.NewReader(body), ct)
	if err != nil {
		return err
	}
	return DecodeResponse(resp.Body, v)
}

// gravar roda exemplo contra um servidor que grava as requisições e responde resposta
// (padrão "{}").
func gravar(t *testing.T, exemplo func(context.Context, Sender) error, resposta string) []requisicao {
	t.Helper()
	var mu sync.Mutex
	var reqs []requisicao
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := lerRequisicao(r)
		if err != nil {
			t.Errorf("lendo a requisição: %v", err)
		}
		mu.Lock()
		reqs = append(reqs, req)
		mu.Unlock()
		if resposta == "" {
			resposta = "{}"
		}
		io.WriteString(w, resposta)
	}))
	defer srv.Close()
	if err := exemplo(context.Background(), httpSender{ctbz.NewClient(srv.URL)}); err != nil {
		t.Fatalf("exemplo: %v", err)
	}
	return reqs
}

func lerRequisicao(r *http.Request) (requisicao, error) {
	req := requisicao{Metodo: r.Method, Caminho: r.URL.Path, Query: r.URL.RawQuery}
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return req, nil
	}
	mt, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return req, err
	}
	req.ContentType = mt
	if mt == "multipart/form-data" {
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return req, nil
			}
			if err != nil {
				return req, err
			}
			data, _ := io.ReadAll(p)
			c := campoGrav{Nome: p.FormName()}
			if p.FileName() == "" {
				c.Valor = string(data)
			} else {
				c.Arquivo, c.Tipo, c.Tamanho = p.FileName(), p.Header.Get("Content-Type"), len(data)
			}
			req.Campos = append(req.Campos, c)
		}
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return req, err
	}
	var v any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if dec.Decode(&v) == nil {
		req.Corpo = v
	} else {
		req.Corpo = string(data)
	}
	return req, nil
}

func goldenRequisicoes(name string) string {
	return filepath.Join("testdata", "requisicoes", name+".json")
}

// TestRequisicoes compara a requisição de cada escrita com o golden.
// Para regravar: go test ./internal/api -run Requisicoes -update
func TestRequisicoes(t *testing.T) {
	for _, e := range Escritas() {
		t.Run(e.Name, func(t *testing.T) {
			compararGolden(t, goldenRequisicoes(e.Name), gravar(t, e.Exemplo, e.Resposta))
		})
	}
}

func compararGolden(t *testing.T, path string, reqs []requisicao) {
	t.Helper()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(reqs); err != nil {
		t.Fatal(err)
	}
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden ausente: rode `go test ./internal/api -run Requisicoes -update` e revise (%v)", err)
	}
	if !bytes.Equal(want, buf.Bytes()) {
		t.Errorf("requisição diferente do golden %s:\n--- golden\n%s\n--- gerada\n%s", path, want, buf.Bytes())
	}
}

// TestRequisicoesCobertura falha com escrita sem golden, nome repetido ou golden órfão.
func TestRequisicoesCobertura(t *testing.T) {
	names := map[string]bool{}
	for _, e := range Escritas() {
		if names[e.Name] {
			t.Errorf("escrita duplicada: %s", e.Name)
		}
		names[e.Name] = true
		if _, err := os.Stat(goldenRequisicoes(e.Name)); err != nil && !*update {
			t.Errorf("escrita sem golden: %s", e.Name)
		}
	}
	files, _ := filepath.Glob(goldenRequisicoes("*"))
	for _, f := range files {
		if name := strings.TrimSuffix(filepath.Base(f), ".json"); !names[name] {
			t.Errorf("golden sem escrita: %s", f)
		}
	}
}

// TestGravacao verifica o próprio gravador: JSON, string JSON, query, DELETE sem corpo e
// multipart comparado por campo, sem o conteúdo do arquivo.
func TestGravacao(t *testing.T) {
	reqs := gravar(t, func(ctx context.Context, s Sender) error {
		if err := s.Send(ctx, "POST", "caixa/lancamentousuario/novo/?x=1", map[string]any{"valor": -150.25}, nil); err != nil {
			return err
		}
		if err := s.Send(ctx, "PUT", "movimentacao-financeira/classificar", JSONString{map[string]int{"id": 7}}, nil); err != nil {
			return err
		}
		if err := s.Send(ctx, "DELETE", "/api/emissor/x/1", nil, nil); err != nil {
			return err
		}
		return s.SendMultipart(ctx, "POST", "upload", []Campo{
			{Nome: "competencia", Valor: "2026-09"},
			{Nome: "arquivo", Arquivo: "a.pdf", Conteudo: []byte("%PDF-1.4")},
		}, nil)
	}, "")
	got, _ := json.Marshal(reqs)
	want := `[{"metodo":"POST","caminho":"/api/plataforma/caixa/lancamentousuario/novo/","query":"x=1","content_type":"application/json","corpo":{"valor":-150.25}},` +
		`{"metodo":"PUT","caminho":"/api/plataforma/movimentacao-financeira/classificar","content_type":"application/json","corpo":"{\"id\":7}"},` +
		`{"metodo":"DELETE","caminho":"/api/emissor/x/1"},` +
		`{"metodo":"POST","caminho":"/api/plataforma/upload","content_type":"multipart/form-data","campos":[{"nome":"competencia","valor":"2026-09"},{"nome":"arquivo","arquivo":"a.pdf","tipo":"application/pdf","tamanho":8}]}]`
	if string(got) != want {
		t.Errorf("gravação:\n got %s\nwant %s", got, want)
	}
}
