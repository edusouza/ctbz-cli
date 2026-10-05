package cli

import (
	"strings"
	"testing"
)

const (
	pathGuia1     = "/api/plataforma/impostos/v5/impostos-a-pagar/guia/1000000000000001"
	pathRollout   = "/api/plataforma/impostos/rollout"
	pathCentralRt = "/api/plataforma/dashboard/v2/central-rotinas"
	centralLimpa  = `{"pendencias":{"pendenciasCriticas":{"pendenciaAceiteTermoDebitos":{"possuiPendencia":false}},"outrasPendencias":{}},"rotinas":[],"rotinasContabilizei":[]}`
)

func impostosFake(t *testing.T, rollout string) *escritaFake {
	t.Helper()
	return newEscritaFake(t, map[string]string{
		pathGuia1:     fixture(t, "guia_detalhe"),
		pathRollout:   `{"versao":"` + rollout + `"}`,
		pathCentralRt: centralLimpa,
	})
}

func TestImpostosConfirmar(t *testing.T) {
	for _, tc := range []struct {
		name, rollout string
		args          []string
		escrita       string
		situacao      string
	}{
		{"v5 já paguei", "v5", []string{"confirmar"},
			`PUT /api/plataforma/impostos/v5/impostos-a-pagar/guia/1000000000000001/confirmar-pagamento {"tipo":"GUIA","origem":"GUIAS","pagamentoConfirmado":true}`, "pago"},
		{"v3 não paguei", "v4", []string{"confirmar", "--nao-paguei"},
			`PUT /api/plataforma/impostos/v3/impostos-a-pagar/guia/1000000000000001/confirmar-pagamento {"tipo":"GUIA","origem":"GUIAS","pagamentoConfirmado":false}`, "nao_pago"},
		{"desmarcar", "v5", []string{"desmarcar"},
			`"pagamentoConfirmado":false}`, "nao_pago"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := impostosFake(t, tc.rollout)
			f.resposta = `{"movidaPara":"ESTE_MES"}`
			args := append([]string{"impostos"}, tc.args...)
			out, stderr, code := execCLI(t, "", append(args, "1000000000000001", "--yes", "-o", "json")...)
			if code != ExitOK || len(f.writes) != 1 || !strings.HasSuffix(f.writes[0], tc.escrita) {
				t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
			}
			if !strings.Contains(out, `"situacao": "`+tc.situacao+`"`) || !strings.Contains(out, `"movida_para": "ESTE_MES"`) || !strings.Contains(out, `"RECALCULADA"`) {
				t.Errorf("saída:\n%s", out)
			}
			if !strings.Contains(stderr, "não serve como comprovante oficial") || !strings.Contains(stderr, "FULANO DE TAL de Jul de 2026 (R$ 1.234,56), situação atual RECALCULADA") {
				t.Errorf("stderr: %s", stderr)
			}
		})
	}
}

func TestImpostosConfirmarBloqueadoPorPendenciaCritica(t *testing.T) {
	f := impostosFake(t, "v5")
	f.gets[pathCentralRt] = strings.Replace(centralLimpa, `"possuiPendencia":false`, `"possuiPendencia":true`, 1)
	_, stderr, code := execCLI(t, "", "impostos", "confirmar", "1000000000000001", "--yes")
	if code != ExitError || len(f.writes) != 0 || !strings.Contains(stderr, "pendências críticas na Central de Rotinas (pendenciaAceiteTermoDebitos)") {
		t.Errorf("código %d, escritas %d, %s", code, len(f.writes), stderr)
	}
}

func TestImpostosConfirmarDryRun(t *testing.T) {
	f := impostosFake(t, "v5")
	out, _, code := execCLI(t, "", "impostos", "confirmar", "1000000000000001", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, "PUT /api/plataforma/impostos/v5/impostos-a-pagar/guia/1000000000000001/confirmar-pagamento") {
		t.Errorf("código %d:\n%s", code, out)
	}
	if _, _, code := execCLI(t, "", "impostos", "confirmar", "abc", "--yes"); code != ExitUsage {
		t.Errorf("id inválido: código %d", code)
	}
}
