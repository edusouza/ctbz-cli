package cli

import (
	"strings"
	"testing"
)

const pathCentralInit = "/api/plataforma/central-rotinas/init"

func TestPendenciasTermos(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{pathCentralInit: fixture(t, "central_rotinas_init")}))
	out, stderr, code := execCLI(t, "", "pendencias", "termos", "-o", "csv")
	want := "chave,titulo,pendente,prazo_aceite_tacito_dias\n" +
		"carta-responsabilidade,Carta de Responsabilidade da Administração,true,\n" +
		"termo-debitos,Termo de Ciência e Responsabilidade (retiradas de lucros),true,15\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s\nquero:\n%s", code, stderr, out, want)
	}
	out, _, _ = execCLI(t, "", "pendencias", "termos", "--todos", "-o", "csv")
	if !strings.Contains(out, "termo-totalpass,Termo de adesão ao TotalPass,false,") {
		t.Errorf("--todos:\n%s", out)
	}
}

func TestPendenciasTermo(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{pathCentralInit: fixture(t, "central_rotinas_init")}))
	out, _, code := execCLI(t, "", "pendencias", "termo", "carta-responsabilidade")
	want := "Carta de Responsabilidade da Administração\n\nCarta de Responsabilidade da Administração\nDeclaramos que TEXTO EXEMPLO.\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d\n%q\nquero\n%q", code, out, want)
	}
	out, _, _ = execCLI(t, "", "pendencias", "termo", "termo-debitos", "-o", "json")
	if !strings.Contains(out, `"texto": "Termo de Ciência e Responsabilidade: TEXTO EXEMPLO.\n"`) || !strings.Contains(out, `"prazo_aceite_tacito_dias": 15`) {
		t.Errorf("json:\n%s", out)
	}
	if _, _, code := execCLI(t, "", "pendencias", "termo", "outro"); code != ExitUsage {
		t.Errorf("chave desconhecida: código %d", code)
	}
}
