package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

const historicoInitJSON = `{"status":"ATIVO","competencia":"10/2026","dataProximoPagamento":"2026-11-10",
	"habilitado":true,"ativado":true,"adyenClientKey":"SEGREDO-CHAVE","cartao":{"final":"4242"}}`

const pathFaturasPagas = "/api/pagamentos/faturas?cnpj=00000000000000&confirmadas=true"

func withHistorico(t *testing.T, faturas string) {
	t.Helper()
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/payments/recorrencia/init": historicoInitJSON,
		pathFaturasPagas: faturas,
	}))
}

// faturasComSegredos acrescenta à fixture campos que a resposta real traz e que nunca podem
// sair na saída: códigos de pagamento e dados da empresa.
func faturasComSegredos(t *testing.T) string {
	return strings.Replace(fixture(t, "faturas_pagas"), `"tipo": "COMPETENCIA",`,
		`"tipo": "COMPETENCIA", "linhaDigitavel": "SEGREDO-LINHA", "qrCodeTextPix": "SEGREDO-PIX", "cliente": {"razaoSocial": "SEGREDO-EMPRESA"},`, -1)
}

func TestMensalidadeHistoricoFormats(t *testing.T) {
	withHistorico(t, faturasComSegredos(t))

	out, stderr, code := execCLI(t, "", "mensalidade", "historico", "-o", "json")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	var v struct {
		Ativo      bool   `json:"debito_automatico_ativo"`
		Proximo    string `json:"proximo_pagamento"`
		Pagamentos []struct {
			Data        *string `json:"data"`
			Competencia string  `json:"competencia"`
			Vencimento  string  `json:"vencimento"`
			Valor       float64 `json:"valor"`
			MultaJuros  float64 `json:"multa_juros"`
			Forma       *string `json:"forma_pagamento"`
			Status      string  `json:"status"`
		} `json:"pagamentos"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("stdout não é JSON: %v\n%s", err, out)
	}
	if !v.Ativo || v.Proximo != "2026-11-10" || len(v.Pagamentos) != 3 {
		t.Fatalf("JSON inesperado: %s", out)
	}
	set, ago, jul := v.Pagamentos[0], v.Pagamentos[1], v.Pagamentos[2]
	if set.Competencia != "09/2026" || ago.Competencia != "08/2026" || jul.Competencia != "07/2026" {
		t.Errorf("a mais recente vem primeiro: %s", out)
	}
	if set.Valor != 141.83 || set.MultaJuros != 2.83 || set.Forma == nil || *set.Forma != "PIX" || set.Status != "Paga" || set.Data == nil || *set.Data != "2026-09-18" {
		t.Errorf("setembro: %+v", set)
	}
	if ago.MultaJuros != 0 || *ago.Forma != "BOLETO" || ago.Vencimento != "2026-08-14" {
		t.Errorf("agosto: %+v", ago)
	}
	if jul.Valor != 0 || jul.Data != nil || jul.Forma != nil {
		t.Errorf("fatura zerada: %+v", jul)
	}
	if stderr != "" {
		t.Errorf("stderr deveria estar vazio, veio %q", stderr)
	}
	for _, segredo := range []string{"SEGREDO", "adyen", "4242", "cartao"} {
		if strings.Contains(out, segredo) {
			t.Errorf("saída JSON vazou %q:\n%s", segredo, out)
		}
	}

	out, stderr, code = execCLI(t, "", "mensalidade", "historico")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	for _, want := range []string{"Débito automático de impostos ativo:", "10/11/2026", "Faturas pagas:", "R$ 141,83", "R$ 2,83", "PIX", "Paga", "18/09/2026"} {
		if !strings.Contains(out, want) {
			t.Errorf("tabela sem %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "SEGREDO") {
		t.Errorf("tabela vazou segredo:\n%s", out)
	}

	out, stderr, code = execCLI(t, "", "mensalidade", "historico", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	if lines := strings.Split(strings.TrimSpace(out), "\n"); len(lines) != 2 || !strings.HasPrefix(lines[0], "debito_automatico_ativo,") {
		t.Errorf("CSV inesperado:\n%s", out)
	}
}

func TestMensalidadeHistoricoSemPagamentos(t *testing.T) {
	withHistorico(t, `[]`)
	out, stderr, code := execCLI(t, "", "mensalidade", "historico", "-o", "json")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	if !strings.Contains(out, `"pagamentos": []`) {
		t.Errorf("esperava lista vazia:\n%s", out)
	}
}

func TestMensalidadeHistoricoErroNasFaturas(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{"/api/plataforma/payments/recorrencia/init": historicoInitJSON}))
	_, stderr, code := execCLI(t, "", "mensalidade", "historico", "-o", "json")
	if code != ExitError || !strings.Contains(stderr, "/api/pagamentos/faturas") {
		t.Errorf("código %d: %s", code, stderr)
	}
}

func TestMensalidadeHistoricoRejeitaArgumento(t *testing.T) {
	t.Setenv("CTBZ_HOME", t.TempDir())
	_, _, code := execCLI(t, "", "mensalidade", "historico", "extra")
	if code != ExitUsage {
		t.Errorf("código %d, quero %d", code, ExitUsage)
	}
}
