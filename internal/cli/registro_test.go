package cli

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

func TestAcoesRegistraCadaEnvio(t *testing.T) {
	status := http.StatusOK
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	store, _ := ctbz.DefaultStore()
	sess, _ := store.LoadSession()
	sess.Cookies = map[string]*http.Cookie{"oauth-token": {Name: "oauth-token", Value: "SEGREDO-COOKIE"}}
	if err := store.SaveSession(sess); err != nil {
		t.Fatal(err)
	}
	snd := sessionSender{s: testSender().s, origem: origemEscrita{comando: "ctbz caixa remover", id: "123"}}
	body := map[string]string{"senha": "SEGREDO-SENHA", "descricao": "SEGREDO-CORPO"}
	ctx := context.Background()
	if err := snd.Send(ctx, "POST", "caixa/lancamentousuario/novo/?cpf=SEGREDO-QUERY", body, nil); err != nil {
		t.Fatal(err)
	}
	status = http.StatusBadRequest
	if err := snd.Send(ctx, "DELETE", "caixa/lancamentousuario/remover/2026/9/123", nil, nil); err == nil {
		t.Fatal("400 deveria falhar")
	}

	raw, err := os.ReadFile(filepath.Join(store.Dir, "acoes.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "SEGREDO") {
		t.Errorf("registro com corpo, query ou cookie:\n%s", raw)
	}
	acoes, _, _ := store.LoadAcoes()
	if len(acoes) != 2 {
		t.Fatalf("%d ações, quero 2:\n%s", len(acoes), raw)
	}
	want := []ctbz.Acao{
		{Comando: "ctbz caixa remover", Metodo: "POST", Caminho: "/api/plataforma/caixa/lancamentousuario/novo/", Status: 200, Resultado: "enviada", ID: "123"},
		{Comando: "ctbz caixa remover", Metodo: "DELETE", Caminho: "/api/plataforma/caixa/lancamentousuario/remover/2026/9/123", Status: 400, Resultado: "recusada", ID: "123"},
	}
	for i, a := range acoes {
		if a.CNPJ == "" || a.Data.IsZero() {
			t.Errorf("ação %d sem CNPJ ou data: %+v", i, a)
		}
		a.CNPJ, a.Data = "", time.Time{}
		if a != want[i] {
			t.Errorf("ação %d = %+v, quero %+v", i, a, want[i])
		}
	}
}

// Só 4xx e o 560 de regra de negócio são "recusada" (nada foi alterado). 3xx e 5xx são
// resultado incerto: um 502/504 de gateway pode ter vindo depois de a escrita ser aplicada.
func TestAcoesClassificaStatus(t *testing.T) {
	status := http.StatusOK
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	snd := sessionSender{s: testSender().s, origem: origemEscrita{comando: "ctbz caixa remover"}}
	want := map[int]string{
		200: ctbz.ResultadoEnviada,
		400: ctbz.ResultadoRecusada,
		404: ctbz.ResultadoRecusada,
		560: ctbz.ResultadoRecusada,
		302: ctbz.ResultadoIncerta,
		500: ctbz.ResultadoIncerta,
		502: ctbz.ResultadoIncerta,
		504: ctbz.ResultadoIncerta,
	}
	for st, res := range want {
		status = st
		snd.Send(context.Background(), "DELETE", "caixa/x", nil, nil)
		store, _ := ctbz.DefaultStore()
		acoes, _, _ := store.LoadAcoes()
		if got := acoes[len(acoes)-1]; got.Status != st || got.Resultado != res {
			t.Errorf("HTTP %d registrado como %q, quero %q", st, got.Resultado, res)
		}
	}
}

func TestAcoesSemRespostaEDryRun(t *testing.T) {
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, _ *http.Request) {
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close()
	})
	execWrite(t, false, "", opMedio, removerLancamento, "--yes")
	execWrite(t, false, "", opMedio, removerLancamento, "--dry-run")
	store, _ := ctbz.DefaultStore()
	acoes, _, _ := store.LoadAcoes()
	if len(acoes) != 1 || acoes[0].Resultado != ctbz.ResultadoSemResposta || acoes[0].Status != 0 || acoes[0].Comando != "ctbz escrita-teste" {
		t.Errorf("ações = %+v", acoes)
	}
}

func TestAcoesFalhaAoRegistrarNaoEscondeResultado(t *testing.T) {
	writeServer(t, http.StatusOK, func(http.ResponseWriter, *http.Request) {})
	store, _ := ctbz.DefaultStore()
	if err := os.Mkdir(filepath.Join(store.Dir, "acoes.jsonl"), 0o700); err != nil {
		t.Fatal(err)
	}
	out, stderr, code := execWrite(t, false, "", opMedio, removerLancamento, "--yes")
	if code != ExitOK || !strings.Contains(out, `"acao": "remover"`) || !strings.Contains(stderr, "aviso: não foi possível registrar a ação") {
		t.Errorf("código %d, stdout %q, stderr %q", code, out, stderr)
	}
}
