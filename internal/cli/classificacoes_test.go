package cli

import (
	"strings"
	"testing"
)

func classificacoesAPI(t *testing.T) {
	t.Helper()
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/caixa/listpaginada/2026/9/1000/null":                          fixture(t, "caixa_categorias"),
		"/api/plataforma/movimentacao-financeira/contas-usuario/2026/9?origem=CAIXA":   fixture(t, "contas_usuario_competencia"),
		"/api/plataforma/movimentacao-financeira/contas-usuario/2026/9?origem=EXTRATO": fixture(t, "contas_usuario_competencia"),
		"/api/plataforma/movimentacao-financeira/contasUsuario": `[{"id":6199733752168448,"descricao":"Distribuição de lucros","classificacao":"DESPESA","situacao":"ATIVO"},
{"id":1000000000000011,"descricao":"Receita de Serviços","classificacao":"RECEITA","situacao":"ATIVO"}]`,
	}))
}

func TestCaixaContas(t *testing.T) {
	classificacoesAPI(t)
	out, stderr, code := execCLI(t, "", "caixa", "contas", "--competencia", "2026-09", "-o", "csv")
	want := "id,descricao,tipo,vinculo\n" +
		"1000000000000011,Receita de Serviços,recebimento,\n" +
		"1000000000000012,Pagamento de Fornecedores,pagamento,\n" +
		"1000000000000013,Impostos - Simples Nacional,pagamento,guia\n" +
		"1000000000000014,Sócios - Distribuição de Lucros Antecipados,pagamento,sócio\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, stderr %q\n%s\nquero:\n%s", code, stderr, out, want)
	}
	out, _, _ = execCLI(t, "", "caixa", "contas", "--competencia", "2026-09", "--recebimento", "-o", "csv")
	if strings.Count(out, "\n") != 2 || !strings.Contains(out, "Receita de Serviços") {
		t.Errorf("--recebimento:\n%s", out)
	}
	out, _, _ = execCLI(t, "", "caixa", "contas", "--competencia", "2026-09", "--vinculos", "-o", "csv")
	want = "id,descricao,tipo\n0,SEM GUIA,guia\n1000000000000021,Simples Nacional 08/2026,guia\n1000000000000031,FULANO DE TAL,sócio\n"
	if out != want {
		t.Errorf("--vinculos:\n%s\nquero:\n%s", out, want)
	}
	for _, args := range [][]string{
		{"caixa", "contas"},
		{"caixa", "contas", "--competencia", "09/2026"},
		{"caixa", "contas", "--competencia", "2026-09", "--recebimento", "--pagamento"},
	} {
		if _, _, code := execCLI(t, "", args...); code != ExitUsage {
			t.Errorf("%v: código %d, quero %d", args, code, ExitUsage)
		}
	}
}

func TestExtratosContas(t *testing.T) {
	classificacoesAPI(t)
	out, stderr, code := execCLI(t, "", "extratos", "contas", "--competencia", "2026-09", "-o", "csv")
	want := "id,descricao,classificacao,exige_socio\n" +
		"1000000000000011,Receita de Serviços,RECEITA,false\n" +
		"1000000000000012,,DESPESA,false\n" +
		"5981343255101440,,DESPESA,false\n" +
		"1000000000000013,,DESPESA,false\n" +
		"1000000000000014,,DESPESA,false\n" +
		"6199733752168448,Distribuição de lucros,DESPESA,true\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, stderr %q\n%s\nquero:\n%s", code, stderr, out, want)
	}
	out, _, _ = execCLI(t, "", "extratos", "contas", "--competencia", "2026-09", "--receita", "-o", "csv")
	if strings.Count(out, "\n") != 2 {
		t.Errorf("--receita:\n%s", out)
	}
}
