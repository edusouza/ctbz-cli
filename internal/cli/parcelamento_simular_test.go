package cli

import (
	"net/http"
	"strings"
	"testing"
)

func TestParcelamentoSimular(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/impostos/parcelamento/negociacao-automatica/pgfn-previdenciario/init": fixture(t, "simulacao_parcelamento"),
	}))
	out, stderr, code := execCLI(t, "", "impostos", "parcelamento", "simular", "pgfn-previdenciario", "-o", "csv")
	want := "tipo,parcelas,entrada,demais_parcelas,servico_adicional,emissao_guia\n" +
		"pgfn-previdenciario,12,500.00,250.50,99.90,15.00\n" +
		"pgfn-previdenciario,24,300.00,140.25,99.90,15.00\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s\nquero:\n%s", code, stderr, out, want)
	}
}

func TestParcelamentoSimularSemDebitos(t *testing.T) {
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(560)
		w.Write([]byte(`{"detalhes":[{"detalhe":"Empresa não possui guias para simulação"}]}`))
	})
	out, stderr, code := execCLI(t, "", "impostos", "parcelamento", "simular", "especializado", "--negociacao", "VENCIDOS", "-o", "json")
	if code != ExitOK || strings.TrimSpace(out) != "[]" || !strings.Contains(stderr, "Nenhum débito para parcelar") {
		t.Errorf("código %d, %q, %s", code, out, stderr)
	}
}

func TestParcelamentoSimularUso(t *testing.T) {
	t.Setenv("CTBZ_HOME", t.TempDir())
	for _, args := range [][]string{
		{"impostos", "parcelamento", "simular", "mei"},
		{"impostos", "parcelamento", "simular", "especializado"},
		{"impostos", "parcelamento", "simular"},
	} {
		if _, _, code := execCLI(t, "", args...); code != ExitUsage {
			t.Errorf("%v: código %d", args, code)
		}
	}
}
