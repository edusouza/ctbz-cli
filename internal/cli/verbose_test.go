package cli

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

// segredos são valores que nunca podem aparecer na saída verbose.
var segredos = []string{"COOKIE-SECRETO", "SETCOOKIE-SECRETO", "TOKEN-SECRETO", "SENHA-SECRETA", "AUTH-SECRETO",
	"0f8fad5b-d9cb-469f-a165-70867728950e", "00000000000100"}

func semSegredos(t *testing.T, saida string) {
	t.Helper()
	for _, s := range segredos {
		if strings.Contains(saida, s) {
			t.Errorf("a saída verbose mostra %q:\n%s", s, saida)
		}
	}
}

// verboseServer responde JSON com segredos no corpo e num cookie novo.
func verboseServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "oauth-token", Value: "SETCOOKIE-SECRETO", Path: "/"})
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"nome":"Empresa Exemplo","token":"TOKEN-SECRETO",`+
			`"url":"https://app.example.com/checkout/#/fatura?idFatura=0f8fad5b-d9cb-469f-a165-70867728950e"}`)
	}))
	t.Cleanup(srv.Close)
	withSession(t, srv)
	store, _ := ctbz.DefaultStore()
	sess, _ := store.LoadSession()
	sess.Cookies = map[string]*http.Cookie{"__C": {Name: "__C", Value: "COOKIE-SECRETO"}}
	if err := store.SaveSession(sess); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CTBZ_DEBUG", "")
	return srv
}

func TestVerboseMostraRequisicaoSemSegredos(t *testing.T) {
	verboseServer(t)
	out, stderr, code := execCLI(t, "", "api", "--verbose", "empresa/cnpj?cnpj=00000000000100")
	if code != ExitOK {
		t.Fatalf("código %d\n%s", code, stderr)
	}
	for _, want := range []string{
		"> GET http://127.0.0.1",
		"/api/plataforma/empresa/cnpj?cnpj=***",
		"> Cookie: __C=***",
		"> User-Agent: Mozilla/5.0",
		"> Referer: http://127.0.0.1",
		"< HTTP 200 (",
		"< Set-Cookie: oauth-token=***",
		"< Content-Type: application/json",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr sem %q:\n%s", want, stderr)
		}
	}
	if strings.Contains(stderr, "Empresa Exemplo") {
		t.Errorf("no nível 1 o corpo não aparece:\n%s", stderr)
	}
	if strings.Contains(out, "> GET") {
		t.Errorf("a saída verbose foi para o stdout:\n%s", out)
	}
	semSegredos(t, stderr)
}

func TestVerboseNivel2MostraCorposMascarados(t *testing.T) {
	verboseServer(t)
	for _, args := range [][]string{
		{"api", "--verbose", "--verbose", "-X", "POST", "-d", `{"senha":"SENHA-SECRETA","valor":12}`, "x/y"},
		{"api", "-X", "POST", "-d", `{"senha":"SENHA-SECRETA","valor":12}`, "x/y"},
	} {
		if args[1] != "--verbose" {
			t.Setenv("CTBZ_DEBUG", "2")
		}
		_, stderr, code := execCLI(t, "", args...)
		if code != ExitOK {
			t.Fatalf("%v: código %d\n%s", args, code, stderr)
		}
		for _, want := range []string{
			"> POST http://127.0.0.1",
			`> {"senha":"***","valor":12}`,
			`< {"nome":"Empresa Exemplo","token":"***","url":"https://app.example.com/checkout/#/fatura?idFatura=***"}`,
		} {
			if !strings.Contains(stderr, want) {
				t.Errorf("%v: stderr sem %q:\n%s", args, want, stderr)
			}
		}
		semSegredos(t, stderr)
	}
}

func TestVerboseDesligado(t *testing.T) {
	verboseServer(t)
	_, stderr, code := execCLI(t, "", "api", "x/y")
	if code != ExitOK || strings.Contains(stderr, "> GET") {
		t.Errorf("código %d, sem --verbose a saída não tem requisições:\n%s", code, stderr)
	}
	t.Setenv("CTBZ_DEBUG", "1")
	_, stderr, _ = execCLI(t, "", "api", "x/y")
	if !strings.Contains(stderr, "> GET") || strings.Contains(stderr, "Empresa Exemplo") {
		t.Errorf("CTBZ_DEBUG=1 liga o nível 1:\n%s", stderr)
	}
}

func TestVerboseValorInvalidoDeCTBZDebug(t *testing.T) {
	verboseServer(t)
	t.Setenv("CTBZ_DEBUG", "talvez")
	_, stderr, code := execCLI(t, "", "api", "x/y")
	if code != ExitUsage || !strings.Contains(stderr, "CTBZ_DEBUG") {
		t.Errorf("código %d\n%s", code, stderr)
	}
}

func TestVerboseFormularioEMultipart(t *testing.T) {
	var buf bytes.Buffer
	escreverTroca(&buf, 2, ctbz.Exchange{
		Method: "POST", URL: "https://app.example.com/login",
		Header: http.Header{"Content-Type": {"application/x-www-form-urlencoded"}, "Authorization": {"Bearer AUTH-SECRETO"}},
		Body:   []byte("user=pessoa%40example.com&password=SENHA-SECRETA&token=TOKEN-SECRETO&v=2"),
		Status: 302, RespHeader: http.Header{"Location": {"/painel?token=TOKEN-SECRETO#incorreto"}},
		Elapsed: 15 * time.Millisecond,
	})
	saida := buf.String()
	for _, want := range []string{
		"> Authorization: ***",
		"> password=***&token=***&user=pessoa%40example.com&v=2",
		"< HTTP 302 (15 ms, 0 bytes)",
		"< Location: /painel?token=***#incorreto",
	} {
		if !strings.Contains(saida, want) {
			t.Errorf("saída sem %q:\n%s", want, saida)
		}
	}
	semSegredos(t, saida)

	var corpo bytes.Buffer
	mw := multipart.NewWriter(&corpo)
	mw.WriteField("ano", "2026")
	mw.WriteField("senha", "SENHA-SECRETA")
	fw, _ := mw.CreateFormFile("documento", "extrato.csv")
	fw.Write([]byte("data;valor\nCONTEUDO-DO-ARQUIVO\n"))
	mw.Close()
	buf.Reset()
	escreverTroca(&buf, 2, ctbz.Exchange{Method: "POST", URL: "https://app.example.com/api/plataforma/up",
		Header: http.Header{"Content-Type": {mw.FormDataContentType()}}, Body: corpo.Bytes(), Status: 200})
	saida = buf.String()
	for _, want := range []string{"> ano=2026", "> senha=***", "> documento: arquivo extrato.csv (31 bytes)"} {
		if !strings.Contains(saida, want) {
			t.Errorf("saída sem %q:\n%s", want, saida)
		}
	}
	if strings.Contains(saida, "CONTEUDO-DO-ARQUIVO") {
		t.Errorf("o conteúdo do arquivo apareceu:\n%s", saida)
	}
	semSegredos(t, saida)
}

func TestVerboseErroDeConexaoECorpoGrande(t *testing.T) {
	var buf bytes.Buffer
	escreverTroca(&buf, 2, ctbz.Exchange{Method: "GET", URL: "https://app.example.com/x", Err: io.ErrUnexpectedEOF, Elapsed: time.Second})
	if !strings.Contains(buf.String(), "< sem resposta (1000 ms): unexpected EOF") {
		t.Errorf("saída:\n%s", buf.String())
	}
	buf.Reset()
	grande := `{"lista":"` + strings.Repeat("a", 5000) + `"}`
	escreverTroca(&buf, 2, ctbz.Exchange{Method: "GET", URL: "https://app.example.com/x", Status: 200,
		RespHeader: http.Header{"Content-Type": {"application/json"}}, RespBody: []byte(grande)})
	if !strings.Contains(buf.String(), "… (truncado; 5012 bytes)") || len(buf.String()) > 3000 {
		t.Errorf("corpo não truncado (%d bytes):\n%.300s", len(buf.String()), buf.String())
	}
	buf.Reset()
	escreverTroca(&buf, 2, ctbz.Exchange{Method: "GET", URL: "https://app.example.com/login", Status: 200,
		RespHeader: http.Header{"Content-Type": {"text/html"}}, RespBody: []byte("<html>token=TOKEN-SECRETO</html>")})
	if !strings.Contains(buf.String(), "< (corpo text/html de 32 bytes omitido)") {
		t.Errorf("saída:\n%s", buf.String())
	}
	semSegredos(t, buf.String())
}
