package cli

import (
	"strings"
	"testing"
)

func TestLucros(t *testing.T) {
	fixNow(t, "2026-10-03")
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/informerendimento/recuperardadosdistribuicaocliente": fixture(t, "distribuicao_lucros"),
		"/api/plataforma/informerendimento/v2/2025/restricoes":                fixture(t, "restricoes_informe"),
	}))
	out, stderr, code := execCLI(t, "", "lucros", "-o", "json")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	for _, want := range []string{`"exercicio": 2025`, `"saldo": 1000.00`, `"data_limite": null`, `"pendencia_documental": false`,
		`"reabertura_balanco": "NENHUM"`, `"socios": []`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON sem %s:\n%s", want, out)
		}
	}
}

func TestLucrosComExercicio(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/informerendimento/recuperardadosdistribuicaocliente": `{"ano":2025,"saldo":5000,"totalDistribuido":3000,
"totalAdiantamentos":0,"limitePermitidoDistribuicaoLucros":8000,"exercicioFechado":true,"podeAlterar":false,
"motivoNaoPodeAlterar":"PRAZO_ENCERRADO","dataLimite":"31/03/2026","lucrosSocios":[{"id":1000000000000001,"socio":"A","porcentagem":66.67,"valor":2000},{"socio":"B","valor":1000}]}`,
		"/api/plataforma/informerendimento/v2/2025/restricoes": `{"restricoes":{"pendenciaDocumental":{"possuiPendencia":true,
"fluxoRegularizacao":"X"},"debitosFederais":{"possuiPendencia":false,"divergenciaContabilFiscal":false}},"processoReabertura":{"status":"EM_ANDAMENTO"}}`,
	}))
	out, _, code := execCLI(t, "", "lucros", "-o", "json")
	for _, want := range []string{`"exercicio": 2025`, `"data_limite": "2026-03-31"`, `"motivo": "PRAZO_ENCERRADO"`,
		`"pendencia_documental": true`, `"socio": "B"`, `"valor": 1000.00`,
		`"porcentagem": 66.67`, `"id": "1000000000000001"`, `"porcentagem": null`, `"id": null`} {
		if code != ExitOK || !strings.Contains(out, want) {
			t.Errorf("JSON sem %s (código %d):\n%s", want, code, out)
		}
	}
}
