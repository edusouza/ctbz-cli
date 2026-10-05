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

func TestPendenciasAceitar(t *testing.T) {
	f := newEscritaFake(t, map[string]string{pathCentralInit: fixture(t, "central_rotinas_init")})
	f.onWrite = func(f *escritaFake) {
		f.gets[pathCentralInit] = strings.Replace(f.gets[pathCentralInit], `"possuiPendencia": true,
        "conteudoAceiteTermoDebito"`, `"possuiPendencia": false,
        "conteudoAceiteTermoDebito"`, 1)
	}
	out, stderr, code := execCLI(t, "", "pendencias", "aceitar", "termo-debitos", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "POST /api/plataforma/central-rotinas/aceitar-termo-debitos " {
		t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	for _, w := range []string{"Termo de Ciência e Responsabilidade: TEXTO EXEMPLO.", "(risco alto)", "aceite tácito", "15 dias"} {
		if !strings.Contains(stderr, w) {
			t.Errorf("stderr sem %q:\n%s", w, stderr)
		}
	}
	if !strings.Contains(out, `"situacao": "aceito"`) || !strings.Contains(out, `"id": "termo-debitos"`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestPendenciasAceitarNadaAAceitar(t *testing.T) {
	f := newEscritaFake(t, map[string]string{pathCentralInit: fixture(t, "central_rotinas_init")})
	out, stderr, code := execCLI(t, "", "pendencias", "aceitar", "termo-totalpass", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(stderr, "Nada a aceitar") || !strings.Contains(out, `"situacao": "nada_a_aceitar"`) {
		t.Errorf("código %d, escritas %d\n%s\n%s", code, len(f.writes), out, stderr)
	}
}

func TestPendenciasAceitarAdmin(t *testing.T) {
	f := newEscritaFake(t, map[string]string{pathCentralInit: fixture(t, "central_rotinas_init")})
	f.status = 403
	f.resposta = `{"message":"Não é possível realizar a assinatura como admin!"}`
	_, stderr, code := execCLI(t, "", "pendencias", "aceitar", "carta-responsabilidade", "--yes")
	if code != ExitError || !strings.Contains(stderr, "sessão é de administrador") || !strings.Contains(stderr, "como admin!") {
		t.Errorf("código %d, %s", code, stderr)
	}
}
