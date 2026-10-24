package cli

import (
	"fmt"
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

func TestLucrosDistribuirPercentuaisSomam100(t *testing.T) {
	tres := `{"ano":2025,"saldo":3000,"totalDistribuido":0,"podeAlterar":true,"dataLimite":"2026-12-31",
"lucrosSocios":[{"id":1,"valor":0},{"id":2,"valor":0},{"id":3,"valor":0}]}`
	f := distribuirFake(t, tres, fixture(t, "restricoes_informe"))
	_, stderr, code := execCLI(t, "", "lucros", "distribuir", "--socio", "1=1000,00", "--socio", "2=1000,00", "--socio", "3=1000,00", "--yes")
	want := `[{"idSocio":1,"percentual":"33.34","valor":"1000.00"},{"idSocio":2,"percentual":"33.33","valor":"1000.00"},{"idSocio":3,"percentual":"33.33","valor":"1000.00"}]`
	if code != ExitOK || len(f.writes) != 1 || !strings.HasSuffix(f.writes[0], want) {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "FULANO 33,34% (R$ 1.000,00)") {
		t.Errorf("o resumo deve mostrar o percentual enviado:\n%s", stderr)
	}
}

func TestPercentuaisDe(t *testing.T) {
	for _, tc := range []struct {
		cents []int64
		total int64
		want  []int64
	}{
		{[]int64{100000, 100000, 100000}, 300000, []int64{3334, 3333, 3333}},
		{[]int64{33330, 66671}, 100001, []int64{3333, 6667}},
		{[]int64{50001, 50000}, 100001, []int64{5000, 5000}},
		{[]int64{0, 100001}, 100001, []int64{0, 10000}},
		{[]int64{1, 1, 1}, 3, []int64{3334, 3333, 3333}},
	} {
		got := percentuaisDe(tc.cents, tc.total)
		var soma int64
		for _, p := range got {
			soma += p
		}
		if fmt.Sprint(got) != fmt.Sprint(tc.want) || soma != 10000 {
			t.Errorf("percentuaisDe(%v, %d) = %v (soma %d), quero %v", tc.cents, tc.total, got, soma, tc.want)
		}
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
		// #770: um campo ausente do lucro total é erro de leitura, não zero.
		{"saldo nulo", strings.Replace(distribuicaoJSON, `"saldo":400.01`, `"saldo":null`, 1), restricoesOK, []string{"--socio", "1=100%"}, ExitError, "a Contabilizei não informou o saldo de lucros do exercício 2025; distribua pelo painel"},
		{"saldo ausente", strings.Replace(distribuicaoJSON, `"saldo":400.01,`, ``, 1), restricoesOK, []string{"--socio", "1=100%"}, ExitError, "não informou o saldo de lucros"},
		{"total distribuído nulo", strings.Replace(distribuicaoJSON, `"totalDistribuido":600`, `"totalDistribuido":null`, 1), restricoesOK, []string{"--socio", "1=100%"}, ExitError, "a Contabilizei não informou o total já distribuído do exercício 2025; distribua pelo painel"},
		{"total distribuído ausente", strings.Replace(distribuicaoJSON, `"totalDistribuido":600,`, ``, 1), restricoesOK, []string{"--socio", "1=100%"}, ExitError, "não informou o total já distribuído"},
		// #777: o id de cada item é o idSocio da escrita; ausente ou repetido, recusa.
		{"sócio sem id", strings.Replace(distribuicaoJSON, `{"id":1,"valor":600}`, `{"valor":600}`, 1), restricoesOK, []string{"--socio", "2=100%"}, ExitError, "a Contabilizei listou um sócio da distribuição sem id; distribua pelo painel"},
		{"sócio com id nulo", strings.Replace(distribuicaoJSON, `{"id":1,`, `{"id":null,`, 1), restricoesOK, []string{"--socio", "2=100%"}, ExitError, "sócio da distribuição sem id"},
		{"sócio com id vazio", strings.Replace(distribuicaoJSON, `{"id":1,`, `{"id":"",`, 1), restricoesOK, []string{"--socio", "2=100%"}, ExitError, "sócio da distribuição sem id"},
		{"sócio com id objeto", strings.Replace(distribuicaoJSON, `{"id":1,`, `{"id":{"id":1},`, 1), restricoesOK, []string{"--socio", "2=100%"}, ExitError, "sócio da distribuição sem id"},
		{"id repetido", strings.Replace(distribuicaoJSON, `{"id":2,`, `{"id":1,`, 1), restricoesOK, []string{"--socio", "1=100%"}, ExitError, "a Contabilizei listou o sócio 1 mais de uma vez na distribuição; distribua pelo painel"},
		{"id repetido como texto", strings.Replace(distribuicaoJSON, `{"id":2,`, `{"id":"1",`, 1), restricoesOK, []string{"--socio", "1=100%"}, ExitError, "listou o sócio 1 mais de uma vez"},
	} {
		f := distribuirFake(t, tc.distribuicao, tc.restricoes)
		_, stderr, code := execCLI(t, "", append([]string{"lucros", "distribuir", "--yes"}, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%s: código %d, %d escritas, %q (quero %q)", tc.nome, code, len(f.writes), stderr, tc.want)
		}
	}
}

// tresSociosJSON é um exercício com R$ 1.000,01 de lucro e três sócios.
const tresSociosJSON = `{"ano":2025,"saldo":1000.01,"totalDistribuido":0,"podeAlterar":true,"dataLimite":"2026-12-31",
"lucrosSocios":[{"id":1,"valor":0},{"id":2,"valor":0},{"id":3,"valor":0}]}`

// O ajuste de centavo só cobre o arredondamento dos percentuais: até meio centavo por sócio
// com percentual não nulo, nunca abaixo de zero nem para quem pediu 0% (#779).
func TestLucrosDistribuirAjusteDeCentavo(t *testing.T) {
	for _, tc := range []struct {
		nome, distribuicao string
		socios             []string
		want               string // corpo enviado, ou o erro de soma
	}{
		// Cenários da issue, com R$ 1.000,01: antes enviavam -0,01 e 0,01 a quem pediu 0%.
		{"valor negativo", distribuicaoJSON, []string{"1=0%", "2=1000,02"}, "a soma das partes (R$ 1.000,02) deve ser igual ao lucro total do exercício (R$ 1.000,01)"},
		{"centavo para 0%", tresSociosJSON, []string{"1=500,00", "2=500,00", "3=0%"}, "a soma das partes (R$ 1.000,00) deve ser igual ao lucro total do exercício (R$ 1.000,01)"},
		// Dois centavos a mais com um só sócio em percentual não vêm de arredondamento.
		{"diferença maior que o arredondamento", distribuicaoJSON, []string{"1=500,02", "2=50%"}, "a soma das partes (R$ 1.000,03) deve ser igual ao lucro total do exercício (R$ 1.000,01)"},
		{"centavo em reais", distribuicaoJSON, []string{"1=500,00", "2=500,00"}, "a soma das partes (R$ 1.000,00) deve ser igual ao lucro total do exercício (R$ 1.000,01)"},
		// Arredondamento de verdade: 50% de 1.000,01 = 500,005 → 500,01 duas vezes.
		{"dois percentuais", distribuicaoJSON, []string{"1=50%", "2=50%"}, `[{"idSocio":1,"percentual":"50.00","valor":"500.01"},{"idSocio":2,"percentual":"50.00","valor":"500.00"}]`},
		{"valor e percentual", distribuicaoJSON, []string{"1=500,00", "2=50%"}, `[{"idSocio":1,"percentual":"50.00","valor":"500.00"},{"idSocio":2,"percentual":"50.00","valor":"500.01"}]`},
		// O 0% fica com zero; o centavo vai para quem pediu percentual não nulo.
		{"com 0%", tresSociosJSON, []string{"1=50%", "2=50%", "3=0%"}, `[{"idSocio":1,"percentual":"50.00","valor":"500.01"},{"idSocio":2,"percentual":"50.00","valor":"500.00"},{"idSocio":3,"percentual":"0.00","valor":"0.00"}]`},
		// R$ 0,05 em 30/50/12/8%: 2 + 3 + 1 + 0 = 6 centavos. O último (8%) está em zero e não
		// pode perder o centavo: sai de quem arredondou para cima (12%).
		{"sem valor negativo", `{"ano":2025,"saldo":0.05,"totalDistribuido":0,"podeAlterar":true,"dataLimite":"2026-12-31",
"lucrosSocios":[{"id":1},{"id":2},{"id":3},{"id":4}]}`, []string{"1=30%", "2=50%", "3=12%", "4=8%"}, `{"idSocio":1,"percentual":"40.00","valor":"0.02"},{"idSocio":2,"percentual":"60.00","valor":"0.03"},{"idSocio":3,"percentual":"0.00","valor":"0.00"},{"idSocio":4,"percentual":"0.00","valor":"0.00"}]`},
	} {
		f := distribuirFake(t, tc.distribuicao, fixture(t, "restricoes_informe"))
		args := []string{"lucros", "distribuir", "--yes"}
		for _, so := range tc.socios {
			args = append(args, "--socio", so)
		}
		_, stderr, code := execCLI(t, "", args...)
		if strings.HasPrefix(tc.want, "a soma") {
			if code != ExitUsage || len(f.writes) != 0 || !strings.Contains(stderr, tc.want) {
				t.Errorf("%s: código %d, escritas %q\n%s", tc.nome, code, f.writes, stderr)
			}
			continue
		}
		if code != ExitOK || len(f.writes) != 1 || !strings.HasSuffix(f.writes[0], tc.want) {
			t.Errorf("%s: código %d, escritas %q (quero %s)\n%s", tc.nome, code, f.writes, tc.want, stderr)
		}
	}
}

// Com dado ausente da API, o --dry-run também não mostra requisição (#770, #777).
func TestLucrosDistribuirDryRunRecusaDadoAusente(t *testing.T) {
	for _, d := range []string{
		strings.Replace(distribuicaoJSON, `"saldo":400.01`, `"saldo":null`, 1),
		strings.Replace(distribuicaoJSON, `{"id":1,`, `{`, 1),
	} {
		f := distribuirFake(t, d, fixture(t, "restricoes_informe"))
		out, stderr, code := execCLI(t, "", "lucros", "distribuir", "--socio", "2=100%", "--dry-run")
		if code != ExitError || out != "" || len(f.writes) != 0 || strings.Contains(stderr, "salvarconfiguracaocliente") {
			t.Errorf("código %d, escritas %q\nstdout:\n%s\nstderr:\n%s", code, f.writes, out, stderr)
		}
	}
}
