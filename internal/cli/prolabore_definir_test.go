package cli

import (
	"strings"
	"testing"
)

const (
	pathProlaboreCentral = "/api/plataforma/prolabore/central/init"
	pathGestaoSocio      = "/api/plataforma/prolabore/central/gestao/1000000000000001"
)

func prolaboreFake(t *testing.T, gestao, dashboard string) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{
		pathProlaboreCentral:                  fixture(t, "prolabore_central"),
		pathGestaoSocio:                       gestao,
		"/api/plataforma/dashboard/prolabore": dashboard,
	})
	f.onWrite = func(f *escritaFake) {
		f.gets[pathProlaboreCentral] = strings.Replace(f.gets[pathProlaboreCentral], `"valorProlabore": 1000`, `"valorProlabore": 3000`, 1)
		f.gets[pathGestaoSocio] = strings.Replace(f.gets[pathGestaoSocio], `"INTELIGENTE"`, `"PERSONALIZADO"`, 1)
	}
	return f
}

func TestProlaboreDefinirPersonalizado(t *testing.T) {
	f := prolaboreFake(t, fixture(t, "gestao_socio"), `{"podeAlterar":false}`)
	out, stderr, code := execCLI(t, "", "prolabore", "definir", "--socio", "1000000000000001", "--tipo", "personalizado", "--valor", "3.000,00", "--yes", "-o", "json")
	want := `PUT /api/plataforma/prolabore/central/gestao/1000000000000001 {"tipoGerenciamento":"PERSONALIZADO","valorProlabore":3000,"valorProlaboreMinimo":null,"modoEdicaoProlaboreMin":false,"sairGestaoInteligente":true}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	for _, s := range []string{"Mudar o pró-labore de FULANO DE TAL de INTELIGENTE (R$ 1.000,00) para PERSONALIZADO de R$ 3.000,00 (risco alto)",
		"A empresa sai da gestão inteligente", "vale a partir do próximo mês"} {
		if !strings.Contains(stderr, s) {
			t.Errorf("stderr sem %q:\n%s", s, stderr)
		}
	}
	for _, s := range []string{`"situacao": "alterado"`, `"gestao_anterior": "INTELIGENTE"`, `"valor_anterior": 1000.00`, `"gestao": "PERSONALIZADO"`, `"valor": 3000.00`} {
		if !strings.Contains(out, s) {
			t.Errorf("saída sem %s:\n%s", s, out)
		}
	}
}

func TestProlaboreDefinirTetoComOutrosSocios(t *testing.T) {
	gestao := strings.Replace(fixture(t, "gestao_socio"), `"qtdSocioGestaoInteligente": 1`, `"qtdSocioGestaoInteligente": 2`, 1)
	f := prolaboreFake(t, gestao, fixture(t, "prolabore_dashboard"))
	out, stderr, code := execCLI(t, "", "prolabore", "definir", "--socio", "1000000000000001", "--tipo", "teto-inss", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `{"tipoGerenciamento":"TETO_INSS","valorProlabore":8157.41,`) ||
		!strings.HasSuffix(f.writes[0], `"sairGestaoInteligente":false}`) {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "permanece na gestão inteligente") || strings.Contains(stderr, "próximo mês") {
		t.Errorf("stderr:\n%s", stderr)
	}
	// A gestão relida não é TETO_INSS (o fake muda para PERSONALIZADO): a CLI não afirma a mudança.
	if !strings.Contains(out, `"situacao": "enviado"`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestProlaboreDefinirValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"--tipo", "teto-inss"}, ExitUsage, "informe --socio"},
		{[]string{"--socio", "1", "--tipo", "maximo"}, ExitUsage, "--tipo deve ser salario-minimo"},
		{[]string{"--socio", "1", "--tipo", "personalizado"}, ExitUsage, "exige --valor"},
		{[]string{"--socio", "1", "--tipo", "teto-inss", "--valor", "10"}, ExitUsage, "--valor só vale com --tipo personalizado"},
		{[]string{"--socio", "1", "--tipo", "teto-inss", "--minimo", "10"}, ExitUsage, "--minimo só vale com --tipo inteligente"},
		{[]string{"--socio", "1", "--tipo", "personalizado", "--valor", "abc"}, ExitUsage, "--valor: valor inválido"},
		{[]string{"--socio", "99", "--tipo", "teto-inss"}, ExitError, "sócio 99 não encontrado"},
		{[]string{"--socio", "1000000000000001", "--tipo", "personalizado", "--valor", "1000"}, ExitError, "maior ou igual ao salário mínimo vigente: 1518.00"},
	} {
		f := prolaboreFake(t, fixture(t, "gestao_socio"), "{}")
		_, stderr, code := execCLI(t, "", append([]string{"prolabore", "definir", "--yes"}, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestProlaboreDefinirDryRun(t *testing.T) {
	f := prolaboreFake(t, fixture(t, "gestao_socio"), "{}")
	out, _, code := execCLI(t, "", "prolabore", "definir", "--socio", "1000000000000001", "--tipo", "inteligente", "--minimo", "2000", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, `"valorProlaboreMinimo": 2000`) {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), out)
	}
}
