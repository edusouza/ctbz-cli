package ctbz

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// servidorContado conta as requisições recebidas e guarda o último cabeçalho Cookie.
func servidorContado(t *testing.T) (*httptest.Server, *atomic.Int32, *atomic.Value) {
	t.Helper()
	var n atomic.Int32
	var cookie atomic.Value
	cookie.Store("")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		cookie.Store(r.Header.Get("Cookie"))
		w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &n, &cookie
}

func clientComSessao(base string) *Client {
	c := NewClient(base)
	c.Cookies["SESSION"] = &http.Cookie{Name: "SESSION", Value: "COOKIE-SECRETO"}
	return c
}

func TestDoRecusaURLDeOutroHost(t *testing.T) {
	painel, nPainel, _ := servidorContado(t)
	outro, nOutro, _ := servidorContado(t)
	hostPainel := strings.TrimPrefix(painel.URL, "http://")
	hostOutro := strings.TrimPrefix(outro.URL, "http://")
	portaOutra := hostPainel[:strings.LastIndex(hostPainel, ":")] + hostOutro[strings.LastIndex(hostOutro, ":"):]
	c := clientComSessao(painel.URL)
	for _, u := range []string{
		outro.URL + "/x",
		"http://" + hostPainel + "@" + hostOutro + "/x",     // o host do painel como usuário
		"http://usuario:SENHA-SECRETA@" + hostPainel + "/x", // usuário e senha no próprio host
		"https://" + hostPainel + "/x",                      // outro esquema
		"HTTP://" + strings.ToUpper(hostOutro) + "/x",
		"http://" + portaOutra + "/x", // mesmo host, outra porta
	} {
		for _, metodo := range []string{http.MethodGet, http.MethodPost} {
			_, err := c.Do(context.Background(), metodo, u, strings.NewReader("{}"), nil)
			var fora *ForaDoHostError
			if !errors.As(err, &fora) {
				t.Errorf("%s %s: erro %v, quero ForaDoHostError", metodo, u, err)
				continue
			}
			if msg := err.Error(); !strings.Contains(msg, "caminho fora de "+hostPainel) || strings.Contains(msg, "SENHA-SECRETA") {
				t.Errorf("%s: mensagem %q", u, msg)
			}
		}
	}
	if nPainel.Load() != 0 || nOutro.Load() != 0 {
		t.Errorf("requisições enviadas: painel %d, outro host %d; quero nenhuma", nPainel.Load(), nOutro.Load())
	}
	// Send (a camada usada por leituras e escritas) passa pela mesma guarda.
	if _, err := c.Send(context.Background(), http.MethodPut, outro.URL+"/x", strings.NewReader("{}"), "application/json"); err == nil || nOutro.Load() != 0 {
		t.Errorf("Send para outro host: erro %v, %d requisição(ões)", err, nOutro.Load())
	}
}

func TestDoAceitaURLDoProprioHost(t *testing.T) {
	painel, n, cookie := servidorContado(t)
	c := clientComSessao(painel.URL)
	host := strings.TrimPrefix(painel.URL, "http://")
	for _, u := range []string{painel.URL + "/api/x", "HTTP://" + strings.ToUpper(host) + "/api/x", "/api/x", "dadosempresa/get"} {
		resp, err := c.API(context.Background(), http.MethodGet, u, nil)
		if err != nil || resp.Status != http.StatusOK {
			t.Errorf("%s: %v", u, err)
		}
		if got := cookie.Load().(string); got != "SESSION=COOKIE-SECRETO" {
			t.Errorf("%s: cookie %q", u, got)
		}
	}
	if n.Load() != 4 {
		t.Errorf("%d requisições, quero 4", n.Load())
	}
}

// Um redirecionamento do login para outro host não leva os cookies da sessão.
func TestFollowRedirectsNaoSaiDoHost(t *testing.T) {
	painel, _, _ := servidorContado(t)
	outro, n, _ := servidorContado(t)
	c := clientComSessao(painel.URL)
	_, err := c.followRedirects(context.Background(), &Response{Status: http.StatusFound, Location: outro.URL + "/painel"})
	var fora *ForaDoHostError
	if !errors.As(err, &fora) || n.Load() != 0 {
		t.Errorf("erro %v, %d requisição(ões) ao outro host", err, n.Load())
	}
}

func TestResolvePath(t *testing.T) {
	for in, want := range map[string]string{
		"dadosempresa/get":    "/api/plataforma/dadosempresa/get",
		"/api/legado/x":       "/api/legado/x",
		"https://h.exemplo/x": "https://h.exemplo/x",
		"HTTPS://h.exemplo/x": "HTTPS://h.exemplo/x",
		"httpstatus/get":      "/api/plataforma/httpstatus/get",
		"http:h.exemplo/x":    "/api/plataforma/http:h.exemplo/x",
		"":                    "/api/plataforma/",
	} {
		if got := ResolvePath(in); got != want {
			t.Errorf("ResolvePath(%q) = %q, quero %q", in, got, want)
		}
	}
}
