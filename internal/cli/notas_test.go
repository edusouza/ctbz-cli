package cli

import (
	"strings"
	"testing"
)

const pathNotas = "/api/plataforma/novo-emissor/v2/listagem/notas/filtro?"

func TestNotas(t *testing.T) {
	fixNow(t, "2026-10-03")
	withSession(t, fakeAPI(t, map[string]string{
		pathNotas + "pagina=1&limite=10&ano=2026&mes=9": `{"total":11,"list":[
{"id":1,"numero":101,"cpfCnpjTomador":"11.222.333/0001-81","nomeRazaoTomador":"ACME LTDA","valorServico":1000,
 "dataEmissao":1788523200000,"situacaoNota":"EMITIDA",
 "statusNotaFiscal":{"descricao":"NFS-e processada com sucesso","text":"Emitida","badgeType":"success"}},
{"id":2,"numero":"102","cpfCnpjTomador":"123.456.789-01","nomeRazaoTomador":"FULANO","valorServico":500.5,
 "dataEmissaoFormatada":"15/09/2026","statusNotaFiscal":"CANCELADA"},
{"id":3},{"id":4},{"id":5},{"id":6},{"id":7},{"id":8},{"id":9},{"id":10}]}`,
		pathNotas + "pagina=2&limite=10&ano=2026&mes=9":  `{"total":11,"list":[{"id":11,"numero":111,"valorServico":0.5}]}`,
		pathNotas + "pagina=1&limite=10&ano=2026&mes=10": fixture(t, "notas_emitidas"),
	}))

	out, stderr, code := execCLI(t, "", "notas", "--de", "2026-09", "--ate", "2026-10", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// statusNotaFiscal vem como objeto (101 e a fixture) ou como texto (102); dataEmissao
	// como epoch, como array (a fixture) ou ausente (102, que usa dataEmissaoFormatada).
	if len(lines) != 13 || lines[0] != "numero,emissao,tomador,documento,valor,status,situacao" ||
		lines[1] != "101,2026-09-04,ACME LTDA,11222333000181,1000.00,Emitida,EMITIDA" ||
		lines[2] != "102,2026-09-15,FULANO,12345678901,500.50,CANCELADA," ||
		lines[11] != "111,,,,0.50,," ||
		lines[12] != "5,2026-10-30,TOMADOR NÃO INFORMADO,,250.00,Emitida,PROCESSADO_SUCESSO" {
		t.Errorf("CSV:\n%s", out)
	}

	_, stderr, _ = execCLI(t, "", "notas", "--de", "2026-09", "--ate", "2026-10")
	if !strings.Contains(stderr, "Total: R$ 1.751,00 em 12 nota(s)") {
		t.Errorf("total no stderr: %q", stderr)
	}
}

func TestNotasFiltros(t *testing.T) {
	fixNow(t, "2026-10-03")
	vazia := `{"list":[],"total":0}`
	withSession(t, fakeAPI(t, map[string]string{
		pathNotas + "pagina=1&limite=10&ano=2026&mes=10&documento=11222333000181": vazia,
		pathNotas + "pagina=1&limite=10&ano=2026&mes=10&nomeTomador=ACME+LTDA":    vazia,
		pathNotas + "pagina=1&limite=10&ano=2026&mes=10&numeroNota=101":           vazia,
	}))
	for _, args := range [][]string{
		{"--tomador", "11.222.333/0001-81"},
		{"--tomador", "ACME LTDA"},
		{"--numero", "101"},
	} {
		if _, stderr, code := execCLI(t, "", append([]string{"notas", "-o", "json"}, args...)...); code != ExitOK {
			t.Errorf("%v: código %d: %s", args, code, stderr)
		}
	}
}

func TestNotasPeriodoInvalido(t *testing.T) {
	fixNow(t, "2026-10-03")
	withSession(t, fakeAPI(t, nil))
	for _, args := range [][]string{
		{"--de", "2026/09"},
		{"--de", "2026-10", "--ate", "2026-09"},
		{"--de", "2024-01", "--ate", "2026-10"},
		{"--tomador", "x", "--numero", "1"},
	} {
		if _, _, code := execCLI(t, "", append([]string{"notas"}, args...)...); code != ExitUsage {
			t.Errorf("%v: código %d, quero %d", args, code, ExitUsage)
		}
	}
}
