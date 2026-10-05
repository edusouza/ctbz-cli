package cli

import (
	"strings"
	"testing"
)

const (
	pathContasBancarias = "/api/plataforma/contabancaria/list"
	pathDetalheConta    = "/api/plataforma/contabancaria/detalhes-da-conta/init/1000000000000001"
	contasBancariasJSON = `{"bancos":[{"id":33,"codigo":"341-7","nome":"Banco Exemplo"},{"id":44,"codigo":"237","nome":"Outro Banco"}],
"contasBancarias":[{"id":1000000000000001,"nomeBanco":"Banco Exemplo","codigoBanco":"341-7","agencia":"1234","contaCorrente":"12345-6",
"vlrSaldoInicial":100,"dataSaldoInicial":1704855600000,"statusIntegracao":"NAO_INTEGRADA","fluxoIntegracao":null}]}`
)

func contasFake(t *testing.T) *escritaFake {
	t.Helper()
	fixNow(t, "2026-10-04")
	return newEscritaFake(t, map[string]string{
		pathContasBancarias: contasBancariasJSON,
		pathDetalheConta:    `{"permiteEditar":true,"permiteExcluir":true}`,
	})
}

func TestContasBancariasBancos(t *testing.T) {
	contasFake(t)
	out, _, code := execCLI(t, "", "contas-bancarias", "bancos", "-o", "csv")
	if code != ExitOK || out != "id,codigo,nome\n33,341-7,Banco Exemplo\n44,237,Outro Banco\n" {
		t.Errorf("código %d\n%s", code, out)
	}
}

func TestContasBancariasAdicionar(t *testing.T) {
	f := contasFake(t)
	f.onWrite = func(f *escritaFake) {
		f.gets[pathContasBancarias] = strings.Replace(contasBancariasJSON, `"contasBancarias":[`,
			`"contasBancarias":[{"id":77,"nomeBanco":"Outro Banco","agencia":"0001","contaCorrente":"99887","dataSaldoInicial":0},`, 1)
	}
	out, stderr, code := execCLI(t, "", "contas-bancarias", "adicionar", "--banco", "237", "--agencia", "0001", "--conta", "9988-7",
		"--abertura", "2024-01-10", "--saldo-inicial", "1.000,50", "--declaro-conta-pj", "--yes", "-o", "json")
	want := `POST /api/plataforma/contabancaria/salvar {"bancoId":44,"agencia":"0001","contaCorrente":"99887","dataSaldoInicial":"2024-01-10","vlrSaldoInicial":1000.5}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, `"situacao": "cadastrado"`) || !strings.Contains(out, `"id": 77`) || !strings.Contains(stderr, "desde quando a Contabilizei vai pedir os extratos") {
		t.Errorf("saída:\n%s\n%s", out, stderr)
	}
}

func TestContasBancariasAdicionarValidacoes(t *testing.T) {
	base := []string{"contas-bancarias", "adicionar", "--yes"}
	ok := []string{"--banco", "341", "--agencia", "1", "--conta", "1-1", "--abertura", "2024-01-10"}
	for _, tc := range []struct {
		args []string
		want string
	}{
		{ok, "--declaro-conta-pj"},
		{[]string{"--declaro-conta-pj", "--banco", "341"}, "informe --banco, --agencia"},
		{append([]string{"--declaro-conta-pj", "--banco", "999"}, ok[2:]...), "banco \"999\" não encontrado"},
		{[]string{"--declaro-conta-pj", "--banco", "341", "--agencia", "1", "--conta", "1", "--abertura", "2027-01-01"}, "futura"},
	} {
		f := contasFake(t)
		_, stderr, code := execCLI(t, "", append(base, tc.args...)...)
		if code != ExitUsage || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestContasBancariasEditar(t *testing.T) {
	f := contasFake(t)
	out, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--abertura", "2024-02-01", "--yes", "-o", "json")
	want := `POST /api/plataforma/contabancaria/salvar {"id":1000000000000001,"bancoId":33,"agencia":"1234","contaCorrente":"123456","dataSaldoInicial":"2024-02-01","vlrSaldoInicial":100}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, `"situacao": "editado"`) {
		t.Errorf("saída:\n%s", out)
	}
	f = contasFake(t)
	f.gets[pathDetalheConta] = `{"permiteEditar":false,"permiteExcluir":false,"motivoPermissoesExclusaoEEdicao":"Extratos já classificados."}`
	if _, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--agencia", "9", "--yes"); code != ExitError || !strings.Contains(stderr, "não permite editar esta conta: Extratos já classificados.") {
		t.Errorf("bloqueada: código %d, %s", code, stderr)
	}
	if _, _, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--yes"); code != ExitUsage {
		t.Errorf("sem alteração: código %d", code)
	}
}

func TestContasBancariasRemover(t *testing.T) {
	f := contasFake(t)
	f.onWrite = func(f *escritaFake) { f.gets[pathContasBancarias] = `{"bancos":[],"contasBancarias":[]}` }
	out, stderr, code := execCLI(t, "", "contas-bancarias", "remover", "1000000000000001", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "DELETE /api/plataforma/contabancaria/excluir/1000000000000001 " {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, `"situacao": "removido"`) || !strings.Contains(stderr, "Banco Exemplo agência 1234 conta 12345-6 (risco alto)") || !strings.Contains(stderr, "vínculos") {
		t.Errorf("saída:\n%s\n%s", out, stderr)
	}
	f = contasFake(t)
	f.gets[pathDetalheConta] = `{"permiteEditar":true,"permiteExcluir":false,"motivoPermissoesExclusaoEEdicao":"Conta integrada."}`
	if _, stderr, code := execCLI(t, "", "contas-bancarias", "remover", "1000000000000001", "--yes"); code != ExitError || len(f.writes) != 0 || !strings.Contains(stderr, "não permite excluir esta conta: Conta integrada.") {
		t.Errorf("bloqueada: código %d, %s", code, stderr)
	}
}
