package cli

import (
	"strings"
	"testing"
)

const (
	pathGuias      = "/api/plataforma/impostos/v5/impostos-a-pagar/guias"
	pathPendencias = "/api/plataforma/home/pendencia/pendenciasEmpresa"
	pathCentral    = "/api/plataforma/dashboard/v2/central-rotinas"
)

func TestResumo(t *testing.T) {
	fixNow(t, "2026-10-03")
	central := strings.Replace(centralRotinasJSON, `"pendencias":{"pendenciasCriticas":{},"outrasPendencias":{}}`,
		`"pendencias":{"pendenciasCriticas":{"pendenciaNovaCoisa":{"possuiPendencia":true},"pendenciaCertificadoDigital":{"possuiPendencia":true},
"pendenciaImposto":{"possuiPendencia":false}},"outrasPendencias":{"pendenciaImportacaoExtrato":{"possuiPendencia":true}}}`, 1)
	withSession(t, fakeAPI(t, map[string]string{
		pathGuias:      fixture(t, "guias_a_pagar"),
		pathPendencias: fixture(t, "pendencias_empresa"),
		pathCentral:    central,
	}))

	out, stderr, code := execCLI(t, "", "resumo", "--fail-on-atencao", "-o", "csv")
	if code != ExitAttention {
		t.Errorf("com itens vencidos e críticos, código %d, quero %d: %s", code, ExitAttention, stderr)
	}
	want := "secao,item,prazo,valor,situacao,alerta\n" +
		"impostos,DARF Unificado 07/2026,2026-10-06,1000.00,RECALCULADA,próxima\n" +
		"pendencias,Cadastre o PIS para ativar o pró-labore automático,2026-07-21,,Pendente,vencida\n" +
		"rotinas,Importar extrato bancário de setembro,2026-10-05,,EM_ABERTO,próxima\n" +
		"mensalidade,Mensalidade da Contabilizei,2026-10-15,15.90,EM_ABERTO,\n" +
		"painel,Certificado digital,,,crítica,crítica\n" +
		"painel,Nova coisa,,,crítica,crítica\n" +
		"painel,Importar extrato bancário,,,aviso,\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}
}

func TestResumoImpostoEmAtraso(t *testing.T) {
	fixNow(t, "2026-10-03")
	withSession(t, fakeAPI(t, map[string]string{
		pathGuias:      guiasEmAtrasoJSON,
		pathPendencias: `[]`,
		pathCentral:    `{"pendencias":{"pendenciasCriticas":{},"outrasPendencias":{}},"rotinas":[],"rotinasContabilizei":[]}`,
	}))
	out, _, code := execCLI(t, "", "resumo", "--fail-on-atencao", "-o", "csv")
	if code != ExitAttention || !strings.Contains(out, "impostos,DAS 07/2026,2026-08-20,150.50,VENCIDA,vencida\n") {
		t.Errorf("imposto em atraso (código %d):\n%s", code, out)
	}
}

// Uma guia já paga continua entre as do mês, mas não pede ação, mesmo com o vencimento passado.
func TestResumoIgnoraGuiaPaga(t *testing.T) {
	fixNow(t, "2026-10-20")
	withSession(t, fakeAPI(t, map[string]string{
		pathGuias: `{"emAtraso":[],"esteMes":[{"id":9,"nome":{"label":"DARF Unificado"},"tipo":"GUIA","identificadorImposto":"DARF_UNIFICADO",
"vencimentoOriginal":"2026-10-06","valor":{"label":12.92},"competencia":"Jul de 2026","status":{"label":"PAGA"}}],"proximoMes":[]}`,
		pathPendencias: `[]`,
		pathCentral:    `{"pendencias":{"pendenciasCriticas":{},"outrasPendencias":{}},"rotinas":[],"rotinasContabilizei":[]}`,
	}))
	out, stderr, code := execCLI(t, "", "resumo", "--fail-on-atencao")
	if code != ExitOK || !strings.Contains(stderr, "Nada precisa de atenção.") {
		t.Errorf("guia paga: código %d, stdout %q, stderr %q", code, out, stderr)
	}
}

func TestGuiaPaga(t *testing.T) {
	for situacao, want := range map[string]bool{"PAGA": true, "Paga": true, "PAGO": true, "PAGAR": false, "RECALCULADA": false, "": false} {
		if got := guiaPaga(situacao); got != want {
			t.Errorf("guiaPaga(%q) = %v, quero %v", situacao, got, want)
		}
	}
}

func TestResumoNadaAFazer(t *testing.T) {
	fixNow(t, "2026-08-15") // a fixture não tem rotinas em agosto
	withSession(t, fakeAPI(t, map[string]string{
		pathGuias:      `{"emAtraso":[],"esteMes":[],"proximoMes":[]}`,
		pathPendencias: `[]`,
		pathCentral:    fixture(t, "central_rotinas"),
	}))
	out, stderr, code := execCLI(t, "", "resumo", "--fail-on-atencao")
	if code != ExitOK || !strings.Contains(stderr, "Nada precisa de atenção.") {
		t.Errorf("código %d, stdout %q, stderr %q", code, out, stderr)
	}
}

func TestResumoFalhaSeUmaConsultaFalhar(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{pathGuias: `{"emAtraso":[],"esteMes":[],"proximoMes":[]}`, pathPendencias: `[]`}))
	out, stderr, code := execCLI(t, "", "resumo")
	if code != ExitError || out != "" || !strings.Contains(stderr, "central-rotinas") {
		t.Errorf("código %d, stdout %q, stderr %q", code, out, stderr)
	}
}

func TestNomeDoIndicador(t *testing.T) {
	for in, want := range map[string]string{
		"pendenciaProcuracaoEcac":  "Procuração no e-CAC",
		"pendenciaNovaCoisaGrande": "Nova coisa grande",
	} {
		if got := nomeDoIndicador(in); got != want {
			t.Errorf("nomeDoIndicador(%q) = %q, quero %q", in, got, want)
		}
	}
}
