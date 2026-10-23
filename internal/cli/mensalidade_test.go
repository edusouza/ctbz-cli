package cli

import (
	"strings"
	"testing"

	"github.com/edusouza/ctbz-cli/internal/output"
)

func TestMensalidade(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/dashboard/fatura":         fixture(t, "fatura"),
		"/api/plataforma/dashboard/v1/mensalidade": fixture(t, "mensalidade"),
	}))
	out, stderr, code := execCLI(t, "", "mensalidade", "--fail-on-atraso", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	want := "competencia,valor,vencimento,situacao,observacao,competencia_anterior_atrasada," +
		"botao_titulo,botao_acao,botao_adicional_titulo,botao_adicional_acao\n" +
		"10/2026,1234.56,,Em aberto,+ 1 Serviço adicional,false," +
		"Pagar com cartão,redirect:/painel-de-controle/#/pagar-com-cartao,Gerar segunda via s/ juros,request:dashboard/nova-fatura/202610\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}
}

func TestMensalidadeEmAtraso(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/dashboard/fatura": `{"total":null,"competencia":"Setembro, 2026","status":"Vencida","vencimento":"15/09/2026"}`,
		"/api/plataforma/dashboard/v1/mensalidade": `{"status":"ATRASADA","valor":139,"dataVencimento":1789700000000,` +
			`"competenciaAnteriorAtrasada":true}`,
	}))
	out, _, code := execCLI(t, "", "mensalidade", "--fail-on-atraso", "-o", "json")
	if code != ExitAttention {
		t.Errorf("código %d, quero %d", code, ExitAttention)
	}
	for _, want := range []string{`"valor": 139.00`, `"vencimento": "2026-09-15"`, `"competencia_anterior_atrasada": true`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON sem %s:\n%s", want, out)
		}
	}
}

func TestDataDeValor(t *testing.T) {
	for _, c := range []struct {
		in   any
		want any
	}{
		{nil, nil},
		{"20/10/2026", "2026-10-20"},
		{"amanhã", "amanhã"},
		{float64(1792497600000), "2026-10-20"}, // 20/10/2026 12:00 UTC
		{true, "true"},
	} {
		got := dataDeValor(c.in)
		if d, ok := got.(output.Date); ok {
			got = d.Format("2006-01-02")
		}
		if got != c.want {
			t.Errorf("dataDeValor(%v) = %v, quero %v", c.in, got, c.want)
		}
	}
}

func TestMensalidadeSituacao(t *testing.T) {
	path := "/api/plataforma/inadimplencia/consultasituacaomensalidadeempresa"
	withSession(t, fakeAPI(t, map[string]string{path: "OK"}))
	out, stderr, code := execCLI(t, "", "mensalidade", "situacao", "--fail-on-inadimplencia", "-o", "csv")
	if code != ExitOK || out != "em_dia,situacao\ntrue,OK\n" {
		t.Errorf("em dia (código %d): %q %s", code, out, stderr)
	}

	withSession(t, fakeAPI(t, map[string]string{path: "INADIMPLENTE"}))
	out, _, code = execCLI(t, "", "mensalidade", "situacao", "--fail-on-inadimplencia", "-o", "csv")
	if code != ExitAttention || out != "em_dia,situacao\nfalse,INADIMPLENTE\n" {
		t.Errorf("inadimplente (código %d): %q", code, out)
	}
}
