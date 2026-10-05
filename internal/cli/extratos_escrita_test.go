package cli

import (
	"net/http"
	"strings"
	"testing"
)

const (
	pathLancsExtrato = "/api/plataforma/movimentacao-financeira/lancamento-usuario/paginado?idContaBancaria=7&ano=2026&mes=9&registroPorPagina=100&pagina=1"
	pathInfoExtrato  = "/api/plataforma/movimentacao-financeira/extrato/info?idContaBancaria=7&ano=2026&mes=9"
)

func extratoFake(t *testing.T) *escritaFake {
	t.Helper()
	return newEscritaFake(t, map[string]string{
		pathLancsExtrato: fixture(t, "lancamentos_extrato"),
		pathInfoExtrato:  fixture(t, "extrato_info"),
		"/api/plataforma/movimentacao-financeira/contas-usuario/2026/9?origem=EXTRATO": fixture(t, "contas_usuario_competencia"),
		"/api/plataforma/movimentacao-financeira/contasUsuario": `[{"id":1000000000000011,"descricao":"Receita de Serviços"},
{"id":1000000000000012,"descricao":"Pagamento de Fornecedores"},{"id":6199733752168448,"descricao":"Distribuição de lucros"}]`,
	})
}

func TestExtratosLancamentos(t *testing.T) {
	extratoFake(t)
	out, stderr, code := execCLI(t, "", "extratos", "lancamentos", "--conta-bancaria", "7", "--competencia", "2026-09", "-o", "csv")
	want := "id,data,descricao,valor,classificacao,id_classificacao,id_lancamento_pai\n" +
		"1000000000000101,2026-09-18,PIX RECEBIDO FULANO,1500.00,Receita de Serviços,1000000000000011,\n" +
		"1000000000000102,2026-09-19,PAGAMENTO BOLETO,-1234.56,,,\n" +
		"1000000000000103,2026-09-20,Aluguel,-1200.00,Pagamento de Fornecedores,1000000000000012,1000000000000100\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s\nquero:\n%s", code, stderr, out, want)
	}
	if _, _, code := execCLI(t, "", "extratos", "lancamentos", "--competencia", "2026-09"); code != ExitUsage {
		t.Errorf("sem --conta-bancaria: código %d", code)
	}
}

func TestExtratosClassificar(t *testing.T) {
	f := extratoFake(t)
	f.onWrite = func(f *escritaFake) {
		f.gets[pathLancsExtrato] = strings.Replace(f.gets[pathLancsExtrato], `"idContaUsuario": null`, `"idContaUsuario": 1000000000000012`, 1)
	}
	args := []string{"extratos", "classificar", "1000000000000102", "--conta-bancaria", "7", "--competencia", "2026-09"}
	out, stderr, code := execCLI(t, "", append(args, "--conta", "pagamento de fornecedores", "--yes", "-o", "json")...)
	if code != ExitOK || len(f.writes) != 1 {
		t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
	}
	want := `PUT /api/plataforma/movimentacao-financeira/classificar "{\"idLancamentoUsuario\":1000000000000102,\"idContaUsuario\":1000000000000012,\"idSocio\":null}"`
	if f.writes[0] != want {
		t.Errorf("escrita:\n%s\nquero:\n%s", f.writes[0], want)
	}
	if !strings.Contains(stderr, `Classificar o lançamento 1000000000000102 ("PAGAMENTO BOLETO", -R$ 1.234,56): sem classificação → Pagamento de Fornecedores`) ||
		!strings.Contains(out, `"situacao": "classificado"`) {
		t.Errorf("stderr %s\nsaída %s", stderr, out)
	}
}

func TestExtratosClassificarSocio(t *testing.T) {
	f := extratoFake(t)
	args := []string{"extratos", "classificar", "1000000000000102", "--conta-bancaria", "7", "--competencia", "2026-09", "--yes", "--conta", "6199733752168448"}
	if _, stderr, code := execCLI(t, "", args...); code != ExitUsage || !strings.Contains(stderr, "exige --socio") {
		t.Errorf("sem sócio: código %d, %s", code, stderr)
	}
	_, stderr, code := execCLI(t, "", append(args, "--socio", "1000000000000031")...)
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `\"idSocio\":1000000000000031`) {
		t.Fatalf("código %d, %v, %s", code, f.writes, stderr)
	}
	// O fake não muda a classificação: a CLI avisa.
	if !strings.Contains(stderr, "ainda mostra a classificação antiga") {
		t.Errorf("stderr: %s", stderr)
	}
}

func TestExtratosClassificarValidacoes(t *testing.T) {
	base := []string{"extratos", "classificar", "--conta-bancaria", "7", "--competencia", "2026-09", "--yes"}
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"1000000000000102"}, ExitUsage, "informe --conta"},
		{[]string{"1000000000000102", "--conta", "1000000000000011"}, ExitUsage, "não aceita para despesa"},
		{[]string{"1000000000000101", "--conta", "1000000000000011"}, ExitUsage, "já está classificado"},
		{[]string{"1000000000000102", "--conta", "1000000000000012", "--socio", "1"}, ExitUsage, "não usa sócio"},
		{[]string{"1000000000000102", "--conta", "6199733752168448", "--socio", "1"}, ExitUsage, "sócio \"1\" não encontrado"},
		{[]string{"999", "--conta", "1000000000000012"}, ExitError, "não encontrado"},
	} {
		f := extratoFake(t)
		_, stderr, code := execCLI(t, "", append(base, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, stderr %q (quero %d, %q)", tc.args, code, len(f.writes), stderr, tc.code, tc.want)
		}
	}
	f := extratoFake(t)
	f.gets[pathInfoExtrato] = `{"permiteEditarLancamento":false}`
	if _, stderr, code := execCLI(t, "", append(base, "1000000000000102", "--conta", "1000000000000012")...); code != ExitError || !strings.Contains(stderr, "não podem mais ser alterados") {
		t.Errorf("período fechado: código %d, %s", code, stderr)
	}
}

func TestExtratosClassificar406(t *testing.T) {
	f := extratoFake(t)
	f.status = http.StatusNotAcceptable
	f.resposta = `{"detalhes":[{"identificador":"exception/erro-negocial-406","detalhe":"x"}]}`
	_, stderr, code := execCLI(t, "", "extratos", "classificar", "1000000000000102", "--conta-bancaria", "7", "--competencia", "2026-09", "--conta", "1000000000000012", "--yes")
	if code != ExitError || !strings.Contains(stderr, "os contadores já classificaram as movimentações deste período") {
		t.Errorf("código %d, %s", code, stderr)
	}
}
