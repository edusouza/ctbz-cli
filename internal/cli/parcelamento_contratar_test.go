package cli

import (
	"net/http"
	"strings"
	"testing"
)

const (
	pathSimPGFN = "/api/plataforma/impostos/parcelamento/negociacao-automatica/pgfn-previdenciario/init"
	pathSimSN   = "/api/plataforma/impostos/parcelamento/negociacao-automatica/simples-nacional/init"
	pathSimEsp  = "/api/plataforma/impostos/parcelamento/negociacao-especializada/init?tipoNegociacao=VENCIDOS"
)

func parcelamentoFake(t *testing.T) *escritaFake {
	t.Helper()
	sim := fixture(t, "simulacao_parcelamento")
	return newEscritaFake(t, map[string]string{
		pathSimPGFN: sim, pathSimSN: sim, pathSimEsp: sim,
		"/api/plataforma/impostos/parcelamento/negociacao-especializada/detalhes/init/T-77": `{"status":"EM_ANALISE"}`,
	})
}

func TestParcelamentoContratar(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		escrita  string
		resposta string
		saida    string
	}{
		{"pgfn com parcelas", []string{"pgfn-previdenciario", "--parcelas", "24"},
			`POST /api/plataforma/impostos/parcelamento/negociacao-automatica/pgfn-previdenciario/contratar {"quantidadeParcelas":24}`, "", `"parcelas": 24`},
		{"pgfn padrão = primeira opção", []string{"pgfn-previdenciario"},
			`{"quantidadeParcelas":12}`, "", `"entrada": 500.00`},
		{"simples sem corpo", []string{"simples-nacional"},
			`POST /api/plataforma/impostos/parcelamento/negociacao-automatica/simples-nacional/contratar `, "", `"tipo": "simples-nacional"`},
		{"especializado", []string{"especializado", "--negociacao", "VENCIDOS"},
			`POST /api/plataforma/impostos/parcelamento/negociacao-especializada/contratar {"tipoNegociacao":"VENCIDOS"}`, `{"idTicket":"T-77"}`, `"status": "EM_ANALISE"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := parcelamentoFake(t)
			f.resposta = tc.resposta
			out, stderr, code := execCLI(t, "", append(append([]string{"impostos", "parcelamento", "contratar"}, tc.args...), "--yes", "-o", "json")...)
			if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], tc.escrita) {
				t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
			}
			if !strings.Contains(out, `"situacao": "solicitado"`) || !strings.Contains(out, tc.saida) {
				t.Errorf("saída:\n%s", out)
			}
			for _, w := range []string{"(risco alto)", "confissão de dívida", "cancelado automaticamente", "Selic", "serviço adicional R$ 99,90; emissão de guia R$ 15,00"} {
				if !strings.Contains(stderr, w) {
					t.Errorf("stderr sem %q: %s", w, stderr)
				}
			}
		})
	}
}

func TestParcelamentoContratarValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"pgfn-previdenciario", "--parcelas", "36"}, ExitUsage, "36 parcelas não é uma opção; use 12, 24"},
		{[]string{"simples-nacional", "--parcelas", "12"}, ExitUsage, "só vale para os tipos pgfn"},
		{[]string{"especializado"}, ExitUsage, "exige --negociacao"},
	} {
		f := parcelamentoFake(t)
		_, stderr, code := execCLI(t, "", append(append([]string{"impostos", "parcelamento", "contratar"}, tc.args...), "--yes")...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, %q", tc.args, code, len(f.writes), stderr)
		}
	}
	// Sem débitos: nada é contratado.
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("escrita sem débitos: %s %s", r.Method, r.URL)
		}
		w.WriteHeader(560)
	})
	if _, stderr, code := execCLI(t, "", "impostos", "parcelamento", "contratar", "simples-nacional", "--yes"); code != ExitError || !strings.Contains(stderr, "nenhum débito") {
		t.Errorf("sem débitos: código %d, %s", code, stderr)
	}
}
