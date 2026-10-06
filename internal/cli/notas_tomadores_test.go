package cli

import (
	"strings"
	"testing"
)

func TestNotasTomadores(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/novo-emissor/tomadores/init": `{"emissaoSemTomador":true,"permiteEmissaoExterior":true,"tomadores":[
{"id":1,"nome":"ACME LTDA","cpfCnpj":"11222333000181","email":"a@acme.com","endereco":{"municipio":{"nome":"Curitiba"},"uf":{"id":"PR"}}},
{"id":2,"nome":"John Doe","cpfCnpj":"","estrangeiro":true,"endereco":{"municipio":"Lisboa"}}]}`,
		"/api/plataforma/novo-emissor/clientes/consulta/00000000000191": fixture(t, "consulta_cnpj"),
	}))

	out, stderr, code := execCLI(t, "", "notas", "tomadores", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	want := "nome,documento,email,telefone,inscricao_municipal,municipio,uf,exterior,id\n" +
		"ACME LTDA,11222333000181,a@acme.com,,,Curitiba,PR,false,1\n" +
		"John Doe,,,,,Lisboa,,true,2\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}

	out, stderr, code = execCLI(t, "", "notas", "tomadores", "consulta", "00.000.000/0001-91", "-o", "json")
	if code != ExitOK {
		t.Fatalf("consulta: código %d: %s", code, stderr)
	}
	for _, want := range []string{`"atividade_principal": "6422-1/00 Bancos múltiplos, com carteira comercial"`,
		`"abertura": "1966-08-01"`, `"situacao_cadastral": "ATIVA"`, `"uf": "DF"`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON sem %s:\n%s", want, out)
		}
	}
	if _, _, code := execCLI(t, "", "notas", "tomadores", "consulta", "123"); code != ExitUsage {
		t.Errorf("CNPJ inválido: código %d, quero %d", code, ExitUsage)
	}
}
