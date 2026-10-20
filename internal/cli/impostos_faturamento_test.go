package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImpostosFaturamento(t *testing.T) {
	calculo := `{"nomeMesCompetencia":"Setembro","faturamentoTotal":10,"dasSimples":{"inconsistente":false},"darf":{"inss":{},"irrf":{}},
"valorFaturamentoUltimos12Meses":36000,"valorProlaboreUltimos12Meses":10000,"percentualFatorR":27.78,
"historicoFaturamento":[{"mes":"jan./27","valorFaturamento":3000,"valorProlabore":1000},{"mes":"dez./26","valorFaturamento":2000,"valorProlabore":null}]}`
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/impostos/como-imposto-foi-calculado/init":              calculo,
		"/api/plataforma/impostos/v5/historico-impostos/dados-grafico?ano=2026": `{"meses":{"12":{"totalPago":150.5}}}`,
		"/api/plataforma/impostos/v5/historico-impostos/dados-grafico?ano=2027": fixture(t, "impostos_pagos_no_ano"),
	}))
	out, stderr, code := execCLI(t, "", "impostos", "faturamento", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	want := "competencia,faturamento,prolabore,impostos_pagos\n" +
		"12/2026,2000.00,,150.50\n" +
		"01/2027,3000.00,1000.00,1000.00\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}
	_, stderr, _ = execCLI(t, "", "impostos", "faturamento")
	if !strings.Contains(stderr, "RBT12): R$ 36.000,00") || !strings.Contains(stderr, "Fator R: 27.78%") {
		t.Errorf("resumo: %q", stderr)
	}
}

// Os 12 meses do histórico podem começar antes da abertura da empresa; a Contabilizei recusa
// esses anos com 400 e, neles, não há impostos pagos.
func TestImpostosFaturamentoAnoAntesDaAbertura(t *testing.T) {
	calculo := `{"nomeMesCompetencia":"Setembro","faturamentoTotal":0,"dasSimples":{"inconsistente":false},"darf":{"inss":{},"irrf":{}},
"historicoFaturamento":[{"mes":"dez./25","valorFaturamento":0,"valorProlabore":0},{"mes":"jul./26","valorFaturamento":0,"valorProlabore":100}]}`
	anoInvalido := `{"titulo":"Ano inválido para consulta do histórico de impostos","detalhes":[{"identificador":"impostos/historico-impostos/ano-invalido","detalhe":"O parâmetro 'ano' é obrigatório e deve ser maior ou igual ao ano de abertura ou ativação da empresa, recebido: 2025"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.RequestURI() {
		case "/api/plataforma/impostos/como-imposto-foi-calculado/init":
			io.WriteString(w, calculo)
		case "/api/plataforma/impostos/v5/historico-impostos/dados-grafico?ano=2025":
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, anoInvalido)
		case "/api/plataforma/impostos/v5/historico-impostos/dados-grafico?ano=2026":
			io.WriteString(w, `{"meses":{"7":{"totalPago":12.92}}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	withSession(t, srv)

	out, stderr, code := execCLI(t, "", "impostos", "faturamento", "-o", "csv")
	want := "competencia,faturamento,prolabore,impostos_pagos\n" +
		"12/2025,0.00,0.00,\n" +
		"07/2026,0.00,100.00,12.92\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\nCSV:\n%s\nesperado:\n%s", code, stderr, out, want)
	}
	if !strings.Contains(stderr, "Impostos pagos indisponíveis em 2025") {
		t.Errorf("sem aviso do ano recusado: %q", stderr)
	}
}
