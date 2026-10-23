// Package ctbz implementa um cliente HTTP para a plataforma web da Contabilizei
// (app.contabilizei.com.br), reproduzindo o fluxo de login do navegador
// (usuário/senha → OTP por e-mail → seleção de empresa) e as chamadas às APIs
// internas usadas pelo painel.
package ctbz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultBaseURL   = "https://app.contabilizei.com.br"
	defaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0 Safari/537.36"
)

// ErrUnauthorized indica que a sessão expirou ou é inválida.
var ErrUnauthorized = errors.New("sessão inválida ou expirada (HTTP 401); rode `ctbz login`")

// Client mantém os cookies da sessão manualmente (em vez de net/http/cookiejar)
// para que possam ser serializados em disco entre execuções da CLI.
type Client struct {
	BaseURL   string
	UserAgent string
	HTTP      *http.Client
	Cookies   map[string]*http.Cookie
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		UserAgent: defaultUserAgent,
		HTTP: &http.Client{
			Timeout: 60 * time.Second,
			// Redirecionamentos são tratados manualmente: o login sinaliza erros
			// via fragmento na URL (ex.: /login#incorreto).
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		Cookies: map[string]*http.Cookie{},
	}
}

// Response é uma resposta HTTP já lida.
type Response struct {
	Status   int
	Header   http.Header
	Body     []byte
	Location string
}

func (c *Client) URL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.BaseURL + path
}

// Do executa uma requisição enviando os cookies da sessão e absorvendo os
// cookies retornados pelo servidor. Com um Tracer no contexto, cada troca é passada a ele.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*Response, error) {
	trace := TracerFrom(ctx)
	var reqBody []byte
	if trace != nil && body != nil {
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		reqBody, body = b, bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.URL(path), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Origin", c.BaseURL)
	req.Header.Set("Referer", c.BaseURL+"/login")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	now := time.Now()
	for _, ck := range c.Cookies {
		if !ck.Expires.IsZero() && ck.Expires.Before(now) {
			continue
		}
		req.AddCookie(&http.Cookie{Name: ck.Name, Value: ck.Value})
	}
	inicio := time.Now()
	resp, err := c.HTTP.Do(req)
	if err != nil {
		if trace != nil {
			trace(Exchange{Method: method, URL: req.URL.String(), Header: req.Header, Body: reqBody, Elapsed: time.Since(inicio), Err: err})
		}
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if trace != nil {
		trace(Exchange{Method: method, URL: req.URL.String(), Header: req.Header, Body: reqBody, Status: resp.StatusCode,
			RespHeader: resp.Header, RespBody: data, Elapsed: time.Since(inicio), Err: err})
	}
	if err != nil {
		return nil, err
	}
	for _, ck := range resp.Cookies() {
		if ck.MaxAge < 0 || (ck.Value == "" && !ck.Expires.IsZero() && ck.Expires.Before(now)) {
			delete(c.Cookies, ck.Name)
			continue
		}
		if ck.MaxAge > 0 {
			ck.Expires = now.Add(time.Duration(ck.MaxAge) * time.Second)
		}
		c.Cookies[ck.Name] = ck
	}
	return &Response{Status: resp.StatusCode, Header: resp.Header, Body: data, Location: resp.Header.Get("Location")}, nil
}

func (c *Client) postForm(ctx context.Context, path string, form url.Values) (*Response, error) {
	return c.Do(ctx, http.MethodPost, path, strings.NewReader(form.Encode()),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
}

// API chama um endpoint JSON autenticado. path pode ser absoluto ("/api/...")
// ou relativo ao BFF da plataforma ("dadosempresa/get" → "/api/plataforma/dadosempresa/get").
func (c *Client) API(ctx context.Context, method, path string, body io.Reader) (*Response, error) {
	contentType := ""
	if body != nil {
		contentType = "application/json"
	}
	return c.Send(ctx, method, path, body, contentType)
}

// Send é como API, com o Content-Type do corpo escolhido por quem chama (ex.: multipart).
func (c *Client) Send(ctx context.Context, method, path string, body io.Reader, contentType string) (*Response, error) {
	path = ResolvePath(path)
	headers := map[string]string{
		"Accept":  "application/json, text/plain, */*",
		"Referer": c.BaseURL + "/painel-de-controle/",
	}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	resp, err := c.Do(ctx, method, path, body, headers)
	if err != nil {
		return nil, err
	}
	if resp.Status == http.StatusUnauthorized {
		return resp, ErrUnauthorized
	}
	return resp, nil
}

// ResolvePath completa um caminho relativo com o BFF da plataforma
// ("dadosempresa/get" → "/api/plataforma/dadosempresa/get"); absolutos ficam como estão.
func ResolvePath(path string) string {
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "http") {
		return "/api/plataforma/" + path
	}
	return path
}

// HTTPError descreve uma resposta inesperada do servidor.
type HTTPError struct {
	Step   string
	Status int
	Body   []byte
}

func (e *HTTPError) Error() string {
	snippet := strings.TrimSpace(string(e.Body))
	if len(snippet) > 200 {
		snippet = snippet[:200] + "…"
	}
	return fmt.Sprintf("%s: resposta inesperada HTTP %d: %s", e.Step, e.Status, snippet)
}
