package cli

import (
	"strings"
	"testing"
)

const pathCartaInforme = "/api/plataforma/informerendimento/carta-responsabilidade/2025"

const restricoesComTudo = `{"restricoes":{"debitosFederais":{"possuiPendencia":true,"debitos":[{"descricao":"DARF 2025"}]},
"pendenciaDocumental":{"possuiPendencia":true,"fluxoRegularizacao":"REABERTURA_BALANCO","pendencias":[{}],"valorServicoAdicional":300}},"processoReabertura":{"status":"NENHUM"}}`

func informeFake(t *testing.T, restricoes string) *escritaFake {
	t.Helper()
	fixNow(t, "2026-10-04")
	f := newEscritaFake(t, map[string]string{
		pathRestricoes:   restricoes,
		pathCartaInforme: fixture(t, "carta_responsabilidade_informe"),
	})
	f.onWrite = func(f *escritaFake) {
		f.gets[pathCartaInforme] = `{"deveAssinarCartaResponsabilidade":false,"html":""}`
	}
	return f
}

func TestInformeRestricoesECarta(t *testing.T) {
	informeFake(t, restricoesComTudo)
	out, stderr, code := execCLI(t, "", "lucros", "informe", "restricoes", "-o", "json")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	for _, s := range []string{`"ano": 2025`, `"carta_pendente": true`, `"debitos": 1`, `"fluxo_regularizacao": "REABERTURA_BALANCO"`, `"valor_servico_adicional": 300.00`, `"pendencias": 1`} {
		if !strings.Contains(out, s) {
			t.Errorf("restrições sem %s:\n%s", s, out)
		}
	}
	out, _, code = execCLI(t, "", "lucros", "informe", "carta")
	if code != ExitOK || out != "Declaro, para os devidos fins, que as informações prestadas à contabilidade são verdadeiras.\n" {
		t.Errorf("carta: código %d, %q", code, out)
	}
}

func TestInformeAceitarCarta(t *testing.T) {
	f := informeFake(t, restricoesComTudo)
	out, stderr, code := execCLI(t, "", "lucros", "informe", "aceitar", "carta-responsabilidade", "--ano", "2025", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "POST /api/plataforma/informerendimento/aceitar-carta-responsabilidade/2025 " {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "Declaro, para os devidos fins") || !strings.Contains(stderr, "(risco alto)") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if out != "acao,situacao,id\naceitar carta-responsabilidade,aceita,2025\n" {
		t.Errorf("saída:\n%s", out)
	}
	// Já aceita: nada é enviado.
	out, stderr, code = execCLI(t, "", "lucros", "informe", "aceitar", "carta-responsabilidade", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(stderr, "não precisa ser aceita") || !strings.Contains(out, "sem mudança") {
		t.Errorf("já aceita: código %d, %d escritas, %s", code, len(f.writes), stderr)
	}
}

func TestInformeAceitarTermoDebitos(t *testing.T) {
	f := informeFake(t, restricoesComTudo)
	_, stderr, code := execCLI(t, "", "lucros", "informe", "aceitar", "termo-debitos", "--yes")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `POST /api/plataforma/informerendimento/aceitar-termo-debitos/2025 {"aceite":"PRIMEIRA_VEZ"}` {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, `Débitos federais apontados no informe de 2025: 1`) || !strings.Contains(stderr, `{"descricao":"DARF 2025"}`) {
		t.Errorf("stderr:\n%s", stderr)
	}
	f = informeFake(t, fixture(t, "restricoes_informe"))
	_, stderr, code = execCLI(t, "", "lucros", "informe", "aceitar", "termo-debitos", "--yes")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(stderr, "não há termo a aceitar") {
		t.Errorf("sem débitos: código %d, %d escritas, %s", code, len(f.writes), stderr)
	}
	if _, _, code := execCLI(t, "", "lucros", "informe", "aceitar", "outro"); code != ExitUsage {
		t.Errorf("aceite inválido: código %d", code)
	}
}

func TestInformeDecidir(t *testing.T) {
	f := informeFake(t, restricoesComTudo)
	out, stderr, code := execCLI(t, "", "lucros", "informe", "decidir", "--regularizar-pendencia", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `POST /api/plataforma/informerendimento/aceite/2025 {"tipoAceite":"REGULARIZAR_PENDENCIA_DOCUMENTAL"}` {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "ticket_form_id=360000562139") || out != "acao,situacao,id,decisao\ndecidir,enviado,2025,regularizar-pendencia\n" {
		t.Errorf("stderr:\n%s\nsaída:\n%s", stderr, out)
	}
	_, stderr, code = execCLI(t, "", "lucros", "informe", "decidir", "--nao-distribuir-lucros", "--yes")
	if code != ExitOK || len(f.writes) != 2 || !strings.Contains(f.writes[1], `"IR_NAO_DISTRIBUIR_LUCROS"`) || !strings.Contains(stderr, "sem distribuição de lucros isenta") {
		t.Errorf("não distribuir: código %d, %q\n%s", code, f.writes, stderr)
	}
}

func TestInformeDecidirRecusas(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{nil, ExitUsage, "escolha uma decisão"},
		{[]string{"--regularizar-pendencia", "--nao-distribuir-lucros"}, ExitUsage, "escolha uma decisão"},
		{[]string{"--nao-regularizar-pendencia"}, ExitError, "não tem pendência documental"},
		{[]string{"--nao-distribuir-lucros"}, ExitError, "não aponta débitos federais"},
	} {
		f := informeFake(t, fixture(t, "restricoes_informe"))
		_, stderr, code := execCLI(t, "", append([]string{"lucros", "informe", "decidir", "--yes"}, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestInformeAceiteAdministrador(t *testing.T) {
	f := informeFake(t, restricoesComTudo)
	f.status, f.resposta = 403, `{"mensagem":"Acesso negado"}`
	_, stderr, code := execCLI(t, "", "lucros", "informe", "decidir", "--nao-regularizar-pendencia", "--yes")
	if code != ExitError || !strings.Contains(stderr, "sessão é de administrador") {
		t.Errorf("código %d: %s", code, stderr)
	}
}
