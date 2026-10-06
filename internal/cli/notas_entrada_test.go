package cli

import "testing"

func TestNotasEntrada(t *testing.T) {
	fixNow(t, "2026-10-03")
	base := "/api/emissor/notasentrada/listar/0?mes=10&ano=2026&empresa=&qtdPagina=10&cursor="
	withSession(t, fakeAPI(t, map[string]string{
		base: `{"total":2,"cursor":"c1","list":[{"chave":"41170911198802000174551200000006991580485674","cnpjEmitente":"11122233000188",
"dataEmissao":1788523200000,"id":5684961520648192,"razaoSocial":"Empresa Fictícia - ME","situacao":{"id":"CIENCIA","descricao":"Ciência"},"valor":4000}]}`,
		base + "c1": `{"total":2,"cursor":null,"list":[{"id":2,"razaoSocial":"Outra","valor":10.5}]}`,
		"/api/emissor/classificacaonotas/listar/?mes=9&ano=2026&empresa=ACME&tipo=1&limite=10&cursor=&offset=0": fixture(t, "notas_entrada_classificacao"),
	}))

	out, stderr, code := execCLI(t, "", "notas", "entrada", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	want := "emissao,emitente,cnpj_emitente,valor,situacao,chave,id,classificacao\n" +
		"2026-09-04,Empresa Fictícia - ME,11122233000188,4000.00,Ciência,41170911198802000174551200000006991580485674,5684961520648192,\n" +
		",Outra,,10.50,,,2,\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}

	out, _, code = execCLI(t, "", "notas", "entrada", "--lista", "classificadas", "--mes", "2026-09", "--emitente", "ACME", "-o", "json")
	if code != ExitOK || out != "[]\n" {
		t.Errorf("classificadas (código %d): %q", code, out)
	}
	for _, args := range [][]string{{"--lista", "todas"}, {"--mes", "09/2026"}} {
		if _, _, code := execCLI(t, "", append([]string{"notas", "entrada"}, args...)...); code != ExitUsage {
			t.Errorf("%v: código %d, quero %d", args, code, ExitUsage)
		}
	}
}
