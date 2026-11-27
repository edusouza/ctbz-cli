package cli

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/edusouza/ctbz-cli/internal/api"
)

// avisoIncertaTeste executa uma escrita cujo POST recebe o status dado e devolve a saída.
func avisoIncertaTeste(t *testing.T, status int, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	op := operacao{Risco: riscoBaixo, Resumo: "Criar lançamento"}
	enviar := func(s api.Sender) error {
		return s.Send(t.Context(), "POST", "lancamentos", map[string]any{"a": 1}, nil)
	}
	return execWrite(t, false, "", op, enviar, append([]string{"--yes"}, args...)...)
}

func TestFalhaIncertaAvisaPossivelmenteAplicada(t *testing.T) {
	for _, status := range []int{302, 500, 502, 504} {
		_, stderr, code := avisoIncertaTeste(t, status)
		if code != ExitError {
			t.Errorf("HTTP %d: código %d", status, code)
		}
		for _, quero := range []string{"pode ter sido aplicada", "ctbz acoes"} {
			if !strings.Contains(stderr, quero) {
				t.Errorf("HTTP %d: stderr sem %q: %q", status, quero, stderr)
			}
		}
	}
}

func TestFalhaRecusadaNaoAvisaPossivelmenteAplicada(t *testing.T) {
	for _, status := range []int{400, 422, 560} {
		stdout, stderr, _ := avisoIncertaTeste(t, status, "-o", "json")
		if strings.Contains(stderr, "pode ter sido aplicada") || stdout != "" {
			t.Errorf("HTTP %d: não devia avisar; stdout=%q stderr=%q", status, stdout, stderr)
		}
	}
}

func TestFalhaIncertaNoJSON(t *testing.T) {
	stdout, _, code := avisoIncertaTeste(t, 504, "-o", "json")
	if code != ExitError {
		t.Fatalf("código %d", code)
	}
	var v struct {
		Resultado             string `json:"resultado"`
		PossivelmenteAplicada bool   `json:"possivelmente_aplicada"`
		Status                int    `json:"status"`
		Aviso                 string `json:"aviso"`
	}
	if err := json.Unmarshal([]byte(stdout), &v); err != nil {
		t.Fatalf("stdout não é JSON: %v: %q", err, stdout)
	}
	if v.Resultado != "incerta" || !v.PossivelmenteAplicada || v.Status != 504 || !strings.Contains(v.Aviso, "ctbz acoes") {
		t.Errorf("JSON inesperado: %+v", v)
	}
}

func TestFalhaIncertaTabelaNaoPoluiStdout(t *testing.T) {
	stdout, _, _ := avisoIncertaTeste(t, 504)
	if stdout != "" {
		t.Errorf("stdout deveria ficar vazio fora do JSON: %q", stdout)
	}
}
