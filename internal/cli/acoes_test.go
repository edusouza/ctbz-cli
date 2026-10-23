package cli

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

func TestAcoesCmd(t *testing.T) {
	store := &ctbz.Store{Dir: t.TempDir()}
	t.Setenv("CTBZ_HOME", store.Dir)
	t.Setenv("CTBZ_OUTPUT", "")
	out, _, code := execCLI(t, "", "acoes", "-o", "json")
	if code != ExitOK || strings.TrimSpace(out) != "[]" {
		t.Fatalf("sem registro: código %d, %q", code, out)
	}
	base := time.Date(2026, 9, 30, 12, 0, 0, 0, time.Local)
	for i := 0; i < 3; i++ {
		store.AppendAcao(ctbz.Acao{Data: base.AddDate(0, 0, i), CNPJ: "22222222000122", Comando: "ctbz caixa remover",
			Metodo: "DELETE", Caminho: "/api/plataforma/x/" + string(rune('1'+i)), Status: 200, Resultado: "enviada", ID: string(rune('1' + i))})
	}
	store.AppendAcao(ctbz.Acao{Data: base.AddDate(0, 0, 3), Comando: "ctbz api", Metodo: "POST", Caminho: "/api/plataforma/y", Resultado: "sem resposta"})

	out, _, code = execCLI(t, "", "acoes", "--desde", "2026-10-01", "-o", "json")
	var v []map[string]any
	if code != ExitOK || json.Unmarshal([]byte(out), &v) != nil || len(v) != 3 {
		t.Fatalf("--desde: código %d\n%s", code, out)
	}
	if v[0]["id"] != "2" || v[0]["status"] != float64(200) || v[0]["cnpj"] != "22222222000122" ||
		v[2]["status"] != nil || v[2]["id"] != nil || v[2]["resultado"] != "sem resposta" {
		t.Errorf("JSON:\n%s", out)
	}
	out, _, _ = execCLI(t, "", "acoes", "--limite", "1", "-o", "csv")
	if lines := strings.Split(strings.TrimSpace(out), "\n"); len(lines) != 2 ||
		lines[0] != "data,cnpj,comando,metodo,caminho,status,resultado,id,arquivos" || !strings.Contains(lines[1], "ctbz api") {
		t.Errorf("--limite 1 em CSV:\n%s", out)
	}
	out, _, _ = execCLI(t, "", "acoes")
	if !strings.Contains(out, "22.222.222/0001-22") || !strings.Contains(out, "ctbz caixa remover") {
		t.Errorf("tabela:\n%s", out)
	}
	for _, args := range [][]string{{"acoes", "--desde", "01/10/2026"}, {"acoes", "--limite", "-1"}} {
		if _, _, code := execCLI(t, "", args...); code != ExitUsage {
			t.Errorf("%v: código %d, quero %d", args, code, ExitUsage)
		}
	}
}
