package cli

import (
	"strings"
	"testing"
)

func TestCaixa(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/caixa/listpaginada/2026/9/1000/null": `{"total":3,"cursor":"1","list":[
{"id":1,"data":1789732800000,"descricao":"Pagamento Contabilizei","valor":-141.83,"situacao":"CONFIRMADO","idContaUsuario":10,"confirmadoViaSistema":true},
{"id":2,"data":1789819200000,"descricao":"Recebimento cliente","valor":1000,"situacao":"PENDENTE","idContaUsuario":null}]}`,
		"/api/plataforma/movimentacao-financeira/contasUsuario": `[{"id":10,"descricao":"Mensalidade de contabilidade",
"descricaoContaContabil":"Serviços contábeis","classificacao":"DESPESA","situacao":"ATIVO"}]`,
	}))
	out, stderr, code := execCLI(t, "", "caixa", "2026-09", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	want := "data,descricao,conta,classificacao,valor,situacao,automatico,id\n" +
		"2026-09-18,Pagamento Contabilizei,Mensalidade de contabilidade,DESPESA,-141.83,CONFIRMADO,true,1\n" +
		"2026-09-19,Recebimento cliente,,,1000.00,PENDENTE,false,2\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}
	if !strings.Contains(stderr, "o mês tem 3 lançamentos; a Contabilizei devolveu 2") {
		t.Errorf("aviso de lançamentos faltando: %q", stderr)
	}
	_, stderr, _ = execCLI(t, "", "caixa", "2026-09")
	if !strings.Contains(stderr, "Total do mês: R$ 858,17 em 2 lançamento(s)") {
		t.Errorf("total no stderr: %q", stderr)
	}
}

func TestCaixaFixture(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/caixa/listpaginada/2026/9/1000/null":   fixture(t, "caixa"),
		"/api/plataforma/movimentacao-financeira/contasUsuario": fixture(t, "contas_usuario"),
	}))
	out, stderr, code := execCLI(t, "", "caixa", "2026-09", "-o", "json")
	if code != ExitOK || stderr != "" || !strings.Contains(out, `"conta": "Assistência Odontológica"`) {
		t.Errorf("código %d, stderr %q:\n%s", code, stderr, out)
	}
}
