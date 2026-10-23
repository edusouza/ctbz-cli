package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

func TestOutputPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name, env string
		args      []string
		wantStart string
	}{
		{"padrão é tabela", "", []string{"version"}, "Versão:"},
		{"CTBZ_OUTPUT", "csv", []string{"version"}, "versao,commit,data"},
		{"-o vence CTBZ_OUTPUT", "csv", []string{"version", "-o", "table"}, "Versão:"},
		{"-o antes do comando", "", []string{"-o", "json", "version"}, "{"},
		{"--output=", "", []string{"--output=csv", "version"}, "versao,"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CTBZ_OUTPUT", tc.env)
			out, stderr, code := execCLI(t, "", tc.args...)
			if code != ExitOK || !strings.HasPrefix(out, tc.wantStart) {
				t.Errorf("código %d, stderr %q, saída:\n%s", code, stderr, out)
			}
		})
	}
}

// withSession cria uma sessão apontando para srv num CTBZ_HOME temporário.
func withSession(t *testing.T, srv *httptest.Server) {
	t.Helper()
	store := &ctbz.Store{Dir: t.TempDir()}
	t.Setenv("CTBZ_HOME", store.Dir)
	t.Setenv("CTBZ_OUTPUT", "")
	t.Setenv("CTBZ_OTP_CMD", "")
	empresa, err := os.ReadFile(filepath.Join("..", "api", "testdata", "sessao_empresa.json"))
	if err != nil {
		t.Fatal(err)
	}
	sess := &ctbz.Session{BaseURL: srv.URL, CreatedAt: time.Now(), Storage: map[string]json.RawMessage{"e": empresa}}
	if err := store.SaveSession(sess); err != nil {
		t.Fatal(err)
	}
}

// fakeAPI responde caminhos fixos com corpos fixos.
func fakeAPI(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := routes[r.URL.RequestURI()]
		if !ok {
			http.NotFound(w, r)
			return
		}
		io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

const dadosEmpresaJSON = `{"empresaAtual":{"cnpj":"22.222.222/0001-22","razaoSocial":"FULANO TECNOLOGIA LTDA",
	"regimeTributario":"Simples Nacional","statusEmpresa":"ATIVO","ramosAtividade":[],"plano":"Básico",
	"certificado":{"status":{"descricao":"Ativo"},"dataValidade":"22/07/2027"}},
	"empresas":[{"cnpj":"22222222000122","razaoSocial":"FULANO TECNOLOGIA LTDA","statusEmpresa":"ATIVO"},
	            {"cnpj":"11111111000111","razaoSocial":"FULANO MEI","statusEmpresa":"INATIVO"}]}`

func TestEmpresaFormats(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{"/api/plataforma/dadosempresa/get": dadosEmpresaJSON}))

	// --json é atalho para -o json.
	for _, args := range [][]string{{"empresa", "--json"}, {"empresa", "-o", "json"}, {"--output=json", "empresa"}} {
		out, stderr, code := execCLI(t, "", args...)
		if code != ExitOK {
			t.Fatalf("%v: código %d: %s", args, code, stderr)
		}
		var v struct {
			CNPJ                string `json:"cnpj"`
			CertificadoValidade string `json:"certificado_validade"`
			OutrasEmpresas      []struct {
				CNPJ string `json:"cnpj"`
			} `json:"outras_empresas"`
		}
		if err := json.Unmarshal([]byte(out), &v); err != nil {
			t.Fatalf("%v: stdout não é JSON: %v\n%s", args, err, out)
		}
		if v.CNPJ != "22222222000122" || v.CertificadoValidade != "2027-07-22" || len(v.OutrasEmpresas) != 1 {
			t.Errorf("%v: JSON inesperado: %s", args, out)
		}
		if stderr != "" {
			t.Errorf("%v: stderr deveria estar vazio, veio %q", args, stderr)
		}
	}

	out, _, _ := execCLI(t, "", "empresa", "-o", "csv")
	if lines := strings.Split(strings.TrimSpace(out), "\n"); len(lines) != 2 || !strings.HasPrefix(lines[0], "razao_social,cnpj,") {
		t.Errorf("CSV inesperado:\n%s", out)
	}

	out, _, _ = execCLI(t, "", "empresa")
	for _, want := range []string{"CNPJ:", "22.222.222/0001-22", "22/07/2027", "Outras empresas:", "FULANO MEI",
		"Sociedade Empresária Limitada (206-2)", "07/2026", "CEP 00000-000"} {
		if !strings.Contains(out, want) {
			t.Errorf("tabela sem %q:\n%s", want, out)
		}
	}
}

func TestAPICommand(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/menu/get":        `[{"id":"home","label":"Home"},{"id":"rotinas","label":"Minhas Rotinas"}]`,
		"/api/plataforma/texto":           `OK`,
		"/api/legado/empresa/cnpj?cnpj=1": `{"razaoSocial":"X"}`,
	}))
	t.Setenv("CTBZ_OUTPUT", "csv") // api ignora CTBZ_OUTPUT

	out, _, code := execCLI(t, "", "api", "menu/get")
	if code != ExitOK || !strings.Contains(out, `"label": "Minhas Rotinas"`) {
		t.Errorf("JSON: código %d\n%s", code, out)
	}
	out, _, _ = execCLI(t, "", "api", "-o", "csv", "menu/get")
	if out != "id,label\nhome,Home\nrotinas,Minhas Rotinas\n" {
		t.Errorf("CSV:\n%s", out)
	}
	out, _, _ = execCLI(t, "", "api", "texto")
	if out != "OK\n" {
		t.Errorf("texto = %q", out)
	}
	out, _, _ = execCLI(t, "", "api", "--raw", "texto")
	if out != "OK" {
		t.Errorf("--raw = %q", out)
	}
	if _, _, code := execCLI(t, "", "api", "/api/legado/empresa/cnpj?cnpj=1"); code != ExitOK {
		t.Errorf("caminho absoluto: código %d", code)
	}
	if _, _, code := execCLI(t, "", "api", "nao/existe"); code != ExitError {
		t.Errorf("404 deveria sair com %d, veio %d", ExitError, code)
	}
}

func TestAPICaminhoGitBash(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{"/api/legado/empresa/cnpj?cnpj=1": `{"razaoSocial":"X"}`}))
	// "//" é a forma que o Git Bash não converte; em outros shells chega assim e vale "/".
	for _, p := range []string{"//api/legado/empresa/cnpj?cnpj=1", "///api/legado/empresa/cnpj?cnpj=1"} {
		if out, stderr, code := execCLI(t, "", "api", p); code != ExitOK || !strings.Contains(out, `"razaoSocial": "X"`) {
			t.Errorf("%s: código %d, %s", p, code, stderr)
		}
	}
	// O caminho que o MSYS converteu nem chega a ser enviado.
	for _, p := range []string{`C:/Program Files/Git/api/legado/empresa/cnpj?cnpj=1`, `D:\api\legado`} {
		_, stderr, code := execCLI(t, "", "api", p)
		if code != ExitUsage || !strings.Contains(stderr, "convertido pelo Git Bash do Windows") || !strings.Contains(stderr, "MSYS_NO_PATHCONV=1") {
			t.Errorf("%s: código %d, %s", p, code, stderr)
		}
	}
}
