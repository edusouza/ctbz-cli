package cli

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

func TestAPIFormEnviaMultipart(t *testing.T) {
	dir := t.TempDir()
	csv := filepath.Join(dir, "extrato.csv")
	os.WriteFile(csv, []byte("data;valor\nCONTEUDO-DO-ARQUIVO\n"), 0o600)
	pdf := filepath.Join(dir, "comprovante.pdf")
	os.WriteFile(pdf, []byte("%PDF-1.4"), 0o600)
	writes := writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/plataforma/upload-documentos/extrato/enviar/bucket" ||
			!strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Errorf("requisição inesperada: %s %s %q", r.Method, r.URL.Path, r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if got := r.MultipartForm.Value; got["ano"][0] != "2026" || got["idConta"][0] != "7" || got["texto"][0] != "a=b" {
			t.Errorf("campos de texto %v", got)
		}
		doc := r.MultipartForm.File["documento"][0]
		f, _ := doc.Open()
		conteudo, _ := io.ReadAll(f)
		if doc.Filename != "extrato.csv" || doc.Header.Get("Content-Type") != "text/plain" || string(conteudo) != "data;valor\nCONTEUDO-DO-ARQUIVO\n" {
			t.Errorf("arquivo %q %q %q", doc.Filename, doc.Header.Get("Content-Type"), conteudo)
		}
		if c := r.MultipartForm.File["comprovante"][0]; c.Filename != "comprovante.pdf" || c.Header.Get("Content-Type") != "application/pdf" {
			t.Errorf("segundo arquivo %q %q", c.Filename, c.Header.Get("Content-Type"))
		}
		io.WriteString(w, `{"ok":true}`)
	})
	out, stderr, code := execCLI(t, "", "api", "-F", "ano=2026", "-F", "idConta=7", "-F", "texto=a=b",
		"-F", "documento=@"+csv+";type=text/plain", "-F", "comprovante=@"+pdf, "upload-documentos/extrato/enviar/bucket")
	if code != ExitOK || writes.Load() != 1 || !strings.Contains(out, `"ok": true`) {
		t.Fatalf("código %d, %d escrita(s)\nsaída %s\nstderr %s", code, writes.Load(), out, stderr)
	}

	store, _ := ctbz.DefaultStore()
	raw, _ := os.ReadFile(filepath.Join(store.Dir, "acoes.jsonl"))
	if strings.Contains(string(raw), "CONTEUDO-DO-ARQUIVO") {
		t.Errorf("o registro tem o conteúdo do arquivo:\n%s", raw)
	}
	acoes, _, _ := store.LoadAcoes()
	if len(acoes) != 1 || acoes[0].Comando != "ctbz api" || acoes[0].Metodo != "POST" ||
		acoes[0].Arquivos != "documento: extrato.csv (31 bytes); comprovante: comprovante.pdf (8 bytes)" {
		t.Errorf("ações %+v", acoes)
	}
	out, _, _ = execCLI(t, "", "acoes", "-o", "csv")
	if !strings.Contains(out, ",arquivos\n") || !strings.Contains(out, "documento: extrato.csv (31 bytes)") {
		t.Errorf("ctbz acoes sem os arquivos:\n%s", out)
	}
}

func TestAPIFormComPUT(t *testing.T) {
	writes := writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("requisição %s %q", r.Method, r.Header.Get("Content-Type"))
		}
	})
	if _, stderr, code := execCLI(t, "", "api", "-X", "PUT", "-F", "a=1", "x/y"); code != ExitOK || writes.Load() != 1 {
		t.Errorf("código %d, %d escrita(s)\n%s", code, writes.Load(), stderr)
	}
}

func TestAPIFormUsoIncorreto(t *testing.T) {
	dir := t.TempDir()
	grande := filepath.Join(dir, "grande.csv")
	f, _ := os.Create(grande)
	f.Truncate(api.TamanhoMaximoExtrato + 1)
	f.Close()
	writes := writeServer(t, http.StatusOK, func(http.ResponseWriter, *http.Request) {})
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"-F", "a=1", "-d", `{"a":1}`}, "-F e -d não podem ser usados juntos"},
		{[]string{"-X", "GET", "-F", "a=1"}, "-F envia um corpo"},
		{[]string{"-F", "sem-igual"}, `-F "sem-igual": use campo=valor ou campo=@arquivo`},
		{[]string{"-F", "=1"}, `-F "=1": use campo=valor ou campo=@arquivo`},
		{[]string{"-F", "doc=@" + filepath.Join(dir, "nao-existe.csv")}, "lendo o arquivo de -F doc"},
		{[]string{"-F", "doc=@" + grande}, "o arquivo de -F doc passa do limite de 30 MB"},
	} {
		_, stderr, code := execCLI(t, "", append(append([]string{"api"}, c.args...), "x/y")...)
		if code != ExitUsage || !strings.Contains(stderr, c.want) {
			t.Errorf("%v: código %d, quero %d com %q\n%s", c.args, code, ExitUsage, c.want, stderr)
		}
	}
	if writes.Load() != 0 {
		t.Errorf("%d escrita(s) enviada(s) com uso incorreto", writes.Load())
	}
}
