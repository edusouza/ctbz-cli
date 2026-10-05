package cli

import (
	"strings"
	"testing"
)

func TestParseParte(t *testing.T) {
	p, err := parseParte(" Aluguel : 1.000,00 : Pagamento de Fornecedores : 31 ")
	if err != nil || p != (parteFlag{descricao: "Aluguel", valor: 100000, conta: "Pagamento de Fornecedores", socio: "31"}) {
		t.Errorf("parseParte = %+v, %v", p, err)
	}
	for _, in := range []string{"Aluguel:10", "a:b:c:d:e", ":10:1", "x:10:", "x:0:1", "x:-5:1", "x:abc:1"} {
		if _, err := parseParte(in); err == nil {
			t.Errorf("parseParte(%q) deveria falhar", in)
		}
	}
}

func TestExtratosDesmembrar(t *testing.T) {
	f := extratoFake(t)
	f.onWrite = func(f *escritaFake) {
		f.gets[pathLancsExtrato] = `{"total":2,"list":[
{"id":201,"data":1789786800000,"descricao":"Fornecedor","valor":-1000,"idContaUsuario":1000000000000012,"idLancamentoPai":1000000000000102},
{"id":202,"data":1789786800000,"descricao":"Lucros","valor":-234.56,"idContaUsuario":6199733752168448,"idLancamentoPai":1000000000000102}]}`
	}
	out, stderr, code := execCLI(t, "", "extratos", "desmembrar", "1000000000000102", "--conta-bancaria", "7", "--competencia", "2026-09",
		"--parte", "Fornecedor:1.000,00:Pagamento de Fornecedores", "--parte", "Lucros:234,56:6199733752168448:1000000000000031", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 {
		t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
	}
	want := `PUT /api/plataforma/movimentacao-financeira/desmembrar "{\"idLancamentoPai\":1000000000000102,\"lancamentosFilho\":[` +
		`{\"descricao\":\"Fornecedor\",\"valor\":-1000,\"idContaUsuario\":1000000000000012,\"idVinculo\":null},` +
		`{\"descricao\":\"Lucros\",\"valor\":-234.56,\"idContaUsuario\":6199733752168448,\"idVinculo\":1000000000000031}]}"`
	if f.writes[0] != want {
		t.Errorf("escrita:\n%s\nquero:\n%s", f.writes[0], want)
	}
	if !strings.Contains(stderr, `em 2 partes: "Fornecedor" -R$ 1.000,00 (Pagamento de Fornecedores); "Lucros" -R$ 234,56 (Distribuição de lucros / FULANO DE TAL)`) {
		t.Errorf("resumo: %s", stderr)
	}
	if !strings.Contains(out, `"situacao": "desmembrado"`) || !strings.Contains(out, `"id": 202`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestExtratosDesmembrarValidacoes(t *testing.T) {
	base := []string{"extratos", "desmembrar", "--conta-bancaria", "7", "--competencia", "2026-09", "--yes"}
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"1000000000000102", "--parte", "a:1234,56:1000000000000012"}, ExitUsage, "pelo menos 2"},
		{[]string{"1000000000000102", "--parte", "a:1000:1000000000000012", "--parte", "b:200:1000000000000012"}, ExitUsage, "soma das partes (R$ 1.200,00) é diferente"},
		{[]string{"1000000000000102", "--parte", "a:0:1000000000000012", "--parte", "b:1234,56:1000000000000012"}, ExitUsage, "maior que zero"},
		{[]string{"1000000000000102", "--parte", "a:1000:1000000000000011", "--parte", "b:234,56:1000000000000012"}, ExitUsage, "não aceita para despesa"},
		{[]string{"1000000000000102", "--parte", "a:1000:6199733752168448", "--parte", "b:234,56:1000000000000012"}, ExitUsage, "exige --socio"},
		{[]string{"1000000000000103", "--parte", "a:1000:1000000000000012", "--parte", "b:200:1000000000000012"}, ExitError, "já é parte do desmembramento"},
	} {
		f := extratoFake(t)
		_, stderr, code := execCLI(t, "", append(base, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, stderr %q (quero %d, %q)", tc.args, code, len(f.writes), stderr, tc.code, tc.want)
		}
	}
}

func TestExtratosDesfazerDesmembramento(t *testing.T) {
	f := extratoFake(t)
	f.onWrite = func(f *escritaFake) {
		f.gets[pathLancsExtrato] = `{"total":1,"list":[{"id":1000000000000100,"data":1789786800000,"descricao":"Original","valor":-1200,"idContaUsuario":null}]}`
	}
	// Pelo id de uma parte: o pedido vai para o original.
	out, stderr, code := execCLI(t, "", "extratos", "desfazer-desmembramento", "1000000000000103", "--conta-bancaria", "7", "--competencia", "2026-09", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 ||
		f.writes[0] != "DELETE /api/plataforma/movimentacao-financeira/desmembrar/desfazer/1000000000000100?idLancamentoPai=1000000000000100 " {
		t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, `"situacao": "desfeito"`) || !strings.Contains(out, `"id": 1000000000000100`) || !strings.Contains(stderr, "(1 parte(s) voltam a ser um lançamento só)") {
		t.Errorf("saída:\n%s\n%s", out, stderr)
	}
	f = extratoFake(t)
	if _, stderr, code := execCLI(t, "", "extratos", "desfazer-desmembramento", "1000000000000101", "--conta-bancaria", "7", "--competencia", "2026-09", "--yes"); code != ExitError || !strings.Contains(stderr, "não é um lançamento desmembrado") || len(f.writes) != 0 {
		t.Errorf("não desmembrado: código %d, %s", code, stderr)
	}
}
