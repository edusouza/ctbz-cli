package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

// hostContado é um servidor que só conta as requisições recebidas.
func hostContado(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		io.WriteString(w, `{"vazou":true}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

// ctbz api não envia a sessão para outro host: a URL é recusada antes de qualquer requisição,
// inclusive a conferência da sessão das escritas (#941).
func TestAPIRecusaURLDeOutroHost(t *testing.T) {
	painel, nPainel := hostContado(t)
	withSession(t, painel)
	outro, nOutro := hostContado(t)
	hostPainel := strings.TrimPrefix(painel.URL, "http://")
	hostOutro := strings.TrimPrefix(outro.URL, "http://")
	for _, args := range [][]string{
		{outro.URL + "/x"},
		{"HTTP://" + strings.ToUpper(hostOutro) + "/x"},
		{"http://" + hostPainel + "@" + hostOutro + "/x"},
		{"https://" + hostPainel + "/x"},
		{"ftp://" + hostOutro + "/x"},
		{"-X", "POST", "-d", `{"a":1}`, outro.URL + "/x"},
		{"-F", "a=1", outro.URL + "/x"},
		{"--verbose", outro.URL + "/x"},
	} {
		_, stderr, code := execCLI(t, "", append([]string{"api"}, args...)...)
		if code != ExitUsage || !strings.Contains(stderr, "caminho fora de "+hostPainel) {
			t.Errorf("%v: código %d, quero %d\n%s", args, code, ExitUsage, stderr)
		}
	}
	if nPainel.Load() != 0 || nOutro.Load() != 0 {
		t.Errorf("requisições enviadas: painel %d, outro host %d; quero nenhuma", nPainel.Load(), nOutro.Load())
	}
}

// Sem sessão salva, a URL de outro host também é recusada como uso incorreto, sem pedir login.
func TestAPIRecusaURLDeOutroHostSemSessao(t *testing.T) {
	t.Setenv("CTBZ_HOME", t.TempDir())
	t.Setenv("CTBZ_BASE_URL", "")
	outro, n := hostContado(t)
	_, stderr, code := execCLI(t, "", "api", outro.URL+"/x")
	if code != ExitUsage || !strings.Contains(stderr, "caminho fora de app.contabilizei.com.br") || n.Load() != 0 {
		t.Errorf("código %d, %d requisição(ões)\n%s", code, n.Load(), stderr)
	}
}

func TestAPIAceitaURLDoProprioPainel(t *testing.T) {
	var mu sync.Mutex
	var caminhos []string
	painel := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "SESSION=COOKIE-DA-SESSAO" {
			t.Errorf("%s sem o cookie da sessão: %q", r.URL, r.Header.Get("Cookie"))
		}
		if r.URL.Path != "/api/plataforma/"+api.PathAppBar {
			mu.Lock()
			caminhos = append(caminhos, r.Method+" "+r.URL.RequestURI())
			mu.Unlock()
		}
		io.WriteString(w, `{"razaoSocial":"X"}`)
	}))
	t.Cleanup(painel.Close)
	withSession(t, painel)
	store, _ := ctbz.DefaultStore()
	sess, _ := store.LoadSession()
	sess.Cookies = map[string]*http.Cookie{"SESSION": {Name: "SESSION", Value: "COOKIE-DA-SESSAO"}}
	if err := store.SaveSession(sess); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{painel.URL + "/api/legado/empresa/cnpj?cnpj=1"},
		{"HTTP://" + strings.ToUpper(strings.TrimPrefix(painel.URL, "http://")) + "/api/legado/empresa/cnpj?cnpj=1"},
		{"-X", "POST", "-d", `{"a":1}`, painel.URL + "/api/legado/x"},
	} {
		out, stderr, code := execCLI(t, "", append([]string{"api"}, args...)...)
		if code != ExitOK || !strings.Contains(out, `"razaoSocial": "X"`) {
			t.Errorf("%v: código %d\n%s", args, code, stderr)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	want := []string{"GET /api/legado/empresa/cnpj?cnpj=1", "GET /api/legado/empresa/cnpj?cnpj=1", "POST /api/legado/x"}
	if strings.Join(caminhos, "\n") != strings.Join(want, "\n") {
		t.Errorf("requisições:\n%s", strings.Join(caminhos, "\n"))
	}
}

func TestAPIAjudaExplicaHost(t *testing.T) {
	out, _, code := execCLI(t, "", "api", "--help")
	if code != ExitOK || !strings.Contains(out, "só aceita caminhos relativos ou URLs do próprio painel") {
		t.Errorf("código %d, ajuda:\n%s", code, out)
	}
}
