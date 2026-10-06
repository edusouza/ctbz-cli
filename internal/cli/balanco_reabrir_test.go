package cli

import (
	"strings"
	"testing"
)

func reabrirFake(t *testing.T, restricoes string) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{pathRestricoes: restricoes})
	f.onWrite = func(f *escritaFake) {
		f.gets[pathRestricoes] = strings.Replace(f.gets[pathRestricoes], `"status":"NENHUM"`, `"status":"EM_ANDAMENTO"`, 1)
	}
	return f
}

func restricoesReabertura(fluxo, status string) string {
	return `{"restricoes":{"pendenciaDocumental":{"possuiPendencia":true,"fluxoRegularizacao":"` + fluxo + `","valorServicoAdicional":450}},"processoReabertura":{"status":"` + status + `"}}`
}

func TestBalancoReabrir(t *testing.T) {
	f := reabrirFake(t, restricoesReabertura("REABERTURA_BALANCO", "NENHUM"))
	out, stderr, code := execCLI(t, "", "balanco", "reabrir", "--ano", "2025", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "POST /api/plataforma/informerendimento/reabrir-balanco/2025 " {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "(risco alto)") || !strings.Contains(stderr, "fica indisponível até a análise") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if out != "acao,situacao,id\nreabrir,EM_ANDAMENTO,2025\n" {
		t.Errorf("saída:\n%s", out)
	}
}

func TestBalancoReabrirRecusas(t *testing.T) {
	for _, tc := range []struct {
		nome, restricoes string
		args             []string
		code             int
		want             string
	}{
		{"sem ano", restricoesReabertura("REABERTURA_BALANCO", "NENHUM"), nil, ExitUsage, "informe --ano"},
		{"em andamento", restricoesReabertura("REABERTURA_BALANCO", "ANALISANDO"), []string{"--ano", "2025"}, ExitError, "já há uma reabertura do balanço de 2025 em andamento (ANALISANDO)"},
		{"sem pendência", fixture(t, "restricoes_informe"), []string{"--ano", "2025"}, ExitError, "não tem pendência documental"},
		{"serviço pago", restricoesReabertura("CONTRATAR_SERVICO_ADICIONAL", "NENHUM"), []string{"--ano", "2025"}, ExitError, "serviço adicional que custa R$ 450,00"},
		{"outro fluxo", restricoesReabertura("ENVIAR_DOCUMENTOS", "NENHUM"), []string{"--ano", "2025"}, ExitError, "não oferece reabrir o balanço de 2025 (regularização: ENVIAR_DOCUMENTOS)"},
	} {
		f := reabrirFake(t, tc.restricoes)
		_, stderr, code := execCLI(t, "", append([]string{"balanco", "reabrir", "--yes"}, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%s: código %d, %d escritas, %q (quero %q)", tc.nome, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestBalancoReabrirAdministrador(t *testing.T) {
	f := reabrirFake(t, restricoesReabertura("REABERTURA_BALANCO", "NENHUM"))
	f.status, f.resposta = 403, `{"mensagem":"Usuário admin não pode fazer a reabertura do exercício contábil para um cliente."}`
	_, stderr, code := execCLI(t, "", "balanco", "reabrir", "--ano", "2025", "--yes")
	if code != ExitError || !strings.Contains(stderr, "sessão é de administrador") {
		t.Errorf("código %d: %s", code, stderr)
	}
}
