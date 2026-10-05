package cli

import (
	"strings"
	"testing"
)

const pathChecklist = "/api/plataforma/checklist-onboarding/aside/init"

func procuracaoFake(t *testing.T) *escritaFake {
	t.Helper()
	return newEscritaFake(t, map[string]string{
		pathCentralInit: fixture(t, "central_rotinas_init"),
		pathChecklist:   fixture(t, "checklist_onboarding"),
	})
}

const semProcuracaoNaCentral = `{"pendencias":{"pendenciasCriticas":{"pendenciaProcuracaoEcac":{"possuiPendencia":false}},"outrasPendencias":{}}}`

func TestPendenciasProcuracaoLeitura(t *testing.T) {
	f := procuracaoFake(t)
	out, stderr, code := execCLI(t, "", "pendencias", "procuracao", "-o", "json")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, `"origem": "central-de-rotinas"`) ||
		!strings.Contains(out, `"situacao": "SEM_PROCURACAO"`) || !strings.Contains(out, `"cnpj_outorgado": "00000000000191"`) ||
		!strings.Contains(stderr, "ao CNPJ 00.000.000/0001-91") {
		t.Errorf("código %d\n%s\n%s", code, out, stderr)
	}
}

func TestPendenciasProcuracaoJaCriei(t *testing.T) {
	for _, tc := range []struct {
		name, central, escrita, origem string
	}{
		{"central de rotinas", "", "POST /api/plataforma/central-rotinas/resolver-pendencia-procuracao-ecac ", "central-de-rotinas"},
		{"checklist", semProcuracaoNaCentral, "POST /api/plataforma/checklist-onboarding/procuracoes/compartilhamentos ", "primeiros-passos"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := procuracaoFake(t)
			if tc.central != "" {
				f.gets[pathCentralInit] = tc.central
			}
			f.onWrite = func(f *escritaFake) {
				f.gets[pathCentralInit] = semProcuracaoNaCentral
				f.gets[pathChecklist] = `{"etapas":[{"nome":"GERAR_PROCURACAO_VIRTUAL","situacao":"CONCLUIDA"}]}`
			}
			out, stderr, code := execCLI(t, "", "pendencias", "procuracao", "--ja-criei", "--yes", "-o", "json")
			if code != ExitOK || len(f.writes) != 1 || f.writes[0] != tc.escrita {
				t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
			}
			if !strings.Contains(out, `"situacao": "declarado"`) || !strings.Contains(out, `"id": "`+tc.origem+`"`) {
				t.Errorf("saída:\n%s", out)
			}
		})
	}
}

func TestPendenciasProcuracaoNadaADeclarar(t *testing.T) {
	f := procuracaoFake(t)
	f.gets[pathCentralInit] = semProcuracaoNaCentral
	f.gets[pathChecklist] = `{"etapas":[]}`
	out, stderr, code := execCLI(t, "", "pendencias", "procuracao", "--ja-criei", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, "nada_a_declarar") || !strings.Contains(stderr, "Nada a declarar") {
		t.Errorf("código %d\n%s\n%s", code, out, stderr)
	}
}
