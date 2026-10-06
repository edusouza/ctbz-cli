package cli

import (
	"strings"
	"testing"
)

const (
	pathDistribuicao = "/api/plataforma/informerendimento/recuperardadosdistribuicaocliente"
	pathRestricoes   = "/api/plataforma/informerendimento/v2/2025/restricoes"
)

// distribuicaoJSON é um exercício de 2025 com R$ 1.000,01 de lucro e dois sócios.
const distribuicaoJSON = `{"ano":2025,"saldo":400.01,"totalDistribuido":600,"podeAlterar":true,"exercicioFechado":true,"dataLimite":"2026-12-31",
"lucrosSocios":[{"id":1,"valor":600},{"id":2,"socio":"BELTRANO","valor":0}]}`

func distribuirFake(t *testing.T, distribuicao, restricoes string) *escritaFake {
	t.Helper()
	fixNow(t, "2026-10-04")
	f := newEscritaFake(t, map[string]string{
		pathDistribuicao:         distribuicao,
		pathRestricoes:           restricoes,
		"/api/legado/socio/list": `[{"id":1,"nome":"FULANO"},{"id":2,"nome":"OUTRO NOME"}]`,
	})
	f.onWrite = func(f *escritaFake) {
		f.gets[pathDistribuicao] = `{"ano":2025,"saldo":0,"totalDistribuido":1000.01,"podeAlterar":true,"lucrosSocios":[{"id":1,"valor":333.34},{"id":2,"valor":666.67}]}`
	}
	return f
}

func TestLucrosDistribuirPercentual(t *testing.T) {
	f := distribuirFake(t, distribuicaoJSON, fixture(t, "restricoes_informe"))
	out, stderr, code := execCLI(t, "", "lucros", "distribuir", "--socio", "1=33,33%", "--socio", "2=66,67%", "--yes", "-o", "csv")
	// 33,33% de R$ 1.000,01 = 333,30 e 66,67% = 666,71 → soma 1.000,01 sem ajuste.
	want := `POST /api/plataforma/informerendimento/salvarconfiguracaocliente [{"idSocio":1,"percentual":"33.33","valor":"333.30"},{"idSocio":2,"percentual":"66.67","valor":"666.71"}]`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "Distribuir R$ 1.000,01 de lucro do exercício 2025: FULANO 33,33% (R$ 333,30); BELTRANO 66,67% (R$ 666,71) (risco alto)") {
		t.Errorf("stderr:\n%s", stderr)
	}
	// A releitura tem outros valores: a CLI não afirma o registro.
	if out != "acao,situacao,id,socio,percentual,valor\ndistribuir,enviado,1,FULANO,33.33,333.30\ndistribuir,enviado,2,BELTRANO,66.67,666.71\n" {
		t.Errorf("saída:\n%s", out)
	}
}

func TestLucrosDistribuirValorEArredondamento(t *testing.T) {
	f := distribuirFake(t, distribuicaoJSON, fixture(t, "restricoes_informe"))
	out, stderr, code := execCLI(t, "", "lucros", "distribuir", "--socio", "1=333,34", "--socio", "2=666,67", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `{"idSocio":1,"percentual":"33.33","valor":"333.34"},{"idSocio":2,"percentual":"66.67","valor":"666.67"}`) {
		t.Fatalf("valor: código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, "distribuir,registrado,1,FULANO,33.33,333.34") {
		t.Errorf("saída:\n%s", out)
	}
	// 50% + 50% de 1.000,01 = 500,01 (arredondado) + 500,01 = 1.000,02: o último percentual absorve o centavo.
	_, stderr, code = execCLI(t, "", "lucros", "distribuir", "--socio", "1=50%", "--socio", "2=50%", "--yes")
	if code != ExitOK || len(f.writes) != 2 || !strings.Contains(f.writes[1], `"valor":"500.01"},{"idSocio":2,"percentual":"50.00","valor":"500.00"}`) {
		t.Errorf("arredondamento: código %d, %q\n%s", code, f.writes, stderr)
	}
	// Um sócio só, com tudo: o outro vai com zero.
	_, _, code = execCLI(t, "", "lucros", "distribuir", "--socio", "2=100%", "--yes")
	if code != ExitOK || len(f.writes) != 3 || !strings.Contains(f.writes[2], `{"idSocio":1,"percentual":"0.00","valor":"0.00"},{"idSocio":2,"percentual":"100.00","valor":"1000.01"}`) {
		t.Errorf("um sócio: código %d, %q", code, f.writes)
	}
}

func TestLucrosDistribuirRecusas(t *testing.T) {
	restricoesOK := `{"restricoes":{"pendenciaDocumental":{"possuiPendencia":false},"debitosFederais":{"possuiPendencia":false}}}`
	for _, tc := range []struct {
		nome, distribuicao, restricoes string
		args                           []string
		code                           int
		want                           string
	}{
		{"sem --socio", distribuicaoJSON, restricoesOK, nil, ExitUsage, "informe a parte de cada sócio"},
		{"formato", distribuicaoJSON, restricoesOK, []string{"--socio", "1"}, ExitUsage, "--socio deve ser ID=60%"},
		{"percentual", distribuicaoJSON, restricoesOK, []string{"--socio", "1=120%"}, ExitUsage, "parte inválida"},
		{"repetido", distribuicaoJSON, restricoesOK, []string{"--socio", "1=10%", "--socio", "1=90%"}, ExitUsage, "sócio 1 repetido"},
		{"soma", distribuicaoJSON, restricoesOK, []string{"--socio", "1=60%", "--socio", "2=30%"}, ExitUsage, "a soma das partes (R$ 900,01) deve ser igual ao lucro total do exercício (R$ 1.000,01)"},
		{"sócio de fora", distribuicaoJSON, restricoesOK, []string{"--socio", "9=100%"}, ExitUsage, "sócio 9 não está na distribuição; os sócios dela são 1 (FULANO), 2 (BELTRANO) (veja ctbz lucros)"},
		{"sem exercício", `{"ano":0,"podeAlterar":true}`, restricoesOK, []string{"--socio", "1=100%"}, ExitError, "não há exercício aberto"},
		{"não pode alterar", `{"ano":2025,"podeAlterar":false,"motivoNaoPodeAlterar":"Exercício encerrado"}`, restricoesOK, []string{"--socio", "1=100%"}, ExitError, "não pode mais ser alterada: Exercício encerrado"},
		{"prazo", strings.Replace(distribuicaoJSON, "2026-12-31", "2026-04-30", 1), restricoesOK, []string{"--socio", "1=100%"}, ExitError, "o prazo para distribuir os lucros de 2025 acabou em 30/04/2026"},
		{"restrições", distribuicaoJSON, `{"restricoes":{"pendenciaDocumental":{"possuiPendencia":true},"debitosFederais":{"possuiPendencia":true}}}`, []string{"--socio", "1=100%"}, ExitError, "tem restrições (pendência documental e débitos federais)"},
	} {
		f := distribuirFake(t, tc.distribuicao, tc.restricoes)
		_, stderr, code := execCLI(t, "", append([]string{"lucros", "distribuir", "--yes"}, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%s: código %d, %d escritas, %q (quero %q)", tc.nome, code, len(f.writes), stderr, tc.want)
		}
	}
}
