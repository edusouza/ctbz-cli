package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/edusouza/ctbz-cli/internal/api"
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

// bancosAmbiguos tem os casos da lista real: par nacional/internacional (077 e 077-1), o mesmo
// código em dois bancos (341) e um código que só existe com sufixo (104-0).
var bancosAmbiguos = []api.Banco{
	{ID: 11, Codigo: "077-1", Nome: "Banco Exemplo - Internacional"},
	{ID: 12, Codigo: "077", Nome: "Banco Exemplo S.A."},
	{ID: 21, Codigo: "341", Nome: "Banco Repetido"},
	{ID: 22, Codigo: "341", Nome: "Banco Repetido S.A."},
	{ID: 31, Codigo: "104-0", Nome: "Banco Com Dígito"},
	{ID: 41, Codigo: "218-1", Nome: "Outro - Internacional"},
	{ID: 42, Codigo: "218-2", Nome: "Outro - Digital"},
}

func TestAcharBanco(t *testing.T) {
	for _, tc := range []struct {
		v    string
		id   int64
		erro string
	}{
		{"077", 12, ""},   // código exato, não o internacional
		{"077-1", 11, ""}, // o internacional só pelo código completo
		{"11", 11, ""},    // id
		{" 104 ", 31, ""}, // sem o sufixo, quando só há um
		{"104-0", 31, ""},
		{"341", 0, "corresponde a 2 bancos"},
		{"218", 0, "corresponde a 2 bancos"},
		{"999", 0, "não encontrado"},
	} {
		b, err := acharBanco(bancosAmbiguos, tc.v)
		switch {
		case tc.erro == "" && (err != nil || b.ID != tc.id):
			t.Errorf("%q: %+v, %v (quero id %d)", tc.v, b, err, tc.id)
		case tc.erro != "" && (err == nil || !strings.Contains(err.Error(), tc.erro)):
			t.Errorf("%q: %+v, %v (quero erro %q)", tc.v, b, err, tc.erro)
		}
	}
	_, err := acharBanco(bancosAmbiguos, "341")
	if !strings.Contains(err.Error(), "21  341  Banco Repetido\n") || !strings.Contains(err.Error(), "22  341  Banco Repetido S.A.") {
		t.Errorf("a lista de candidatos deve ter id, código e nome: %v", err)
	}
	if b, ok := bancoAtual(bancosAmbiguos, "341", "banco repetido s.a."); !ok || b.ID != 22 {
		t.Errorf("banco atual pelo código e nome: %+v %v", b, ok)
	}
}

// bancosColisao tem um valor que é id de um banco e código exato de outro (#870).
var bancosColisao = []api.Banco{
	{ID: 237, Codigo: "999", Nome: "Banco A"},
	{ID: 5, Codigo: "237", Nome: "Banco B"},
	{ID: 104, Codigo: "104", Nome: "Banco C"}, // id e código do mesmo banco não é colisão
}

func TestAcharBancoColisaoIDCodigo(t *testing.T) {
	_, err := acharBanco(bancosColisao, "237")
	if err == nil || !strings.Contains(err.Error(), "corresponde a 2 bancos") ||
		!strings.Contains(err.Error(), "237  999  Banco A") || !strings.Contains(err.Error(), "5  237  Banco B") {
		t.Errorf("colisão id/código deve listar os dois bancos: %v", err)
	}
	var u usageError
	if !errors.As(err, &u) {
		t.Errorf("a colisão deve ser erro de uso: %T", err)
	}
	if b, err := acharBanco(bancosColisao, "5"); err != nil || b.ID != 5 {
		t.Errorf("id inequívoco: %+v, %v", b, err)
	}
	if b, err := acharBanco(bancosColisao, "104"); err != nil || b.ID != 104 {
		t.Errorf("id e código do mesmo banco: %+v, %v", b, err)
	}
}

func TestContasBancariasAdicionarColisaoIDCodigo(t *testing.T) {
	f := contasFake(t)
	f.gets[pathContasBancarias] = `{"bancos":[{"id":237,"codigo":"999","nome":"Banco A"},{"id":5,"codigo":"237","nome":"Banco B"}],"contasBancarias":[]}`
	_, stderr, code := execCLI(t, "", "contas-bancarias", "adicionar", "--banco", "237", "--agencia", "1234", "--conta", "12345-6",
		"--abertura", "2024-01-01", "--declaro-conta-pj", "--yes")
	if code != ExitUsage || len(f.writes) != 0 || !strings.Contains(stderr, "237  999  Banco A") || !strings.Contains(stderr, "5  237  Banco B") {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), stderr)
	}
}

func TestContasBancariasEditarColisaoIDCodigo(t *testing.T) {
	f := contasFake(t)
	f.gets[pathContasBancarias] = `{"bancos":[{"id":237,"codigo":"999","nome":"Banco A"},{"id":5,"codigo":"237","nome":"Banco B"}],
"contasBancarias":[{"id":1000000000000001,"nomeBanco":"Banco B","codigoBanco":"237","agencia":"1234","contaCorrente":"12345-6",
"vlrSaldoInicial":100,"dataSaldoInicial":1704855600000}]}`
	// --banco 237 é ambíguo.
	_, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--banco", "237", "--yes")
	if code != ExitUsage || len(f.writes) != 0 || !strings.Contains(stderr, "corresponde a 2 bancos") {
		t.Errorf("--banco ambíguo: código %d, %d escritas\n%s", code, len(f.writes), stderr)
	}
	// Sem --banco, o código da conta (237) nunca é lido como id do Banco A: infere o Banco B (id 5).
	_, stderr, code = execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--abertura", "2024-02-01", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `"bancoId":5,`) {
		t.Errorf("banco atual: código %d, %q\n%s", code, f.writes, stderr)
	}
}

func TestBancoAtualExigeCodigoENomeUnicos(t *testing.T) {
	bancos := []api.Banco{
		{ID: 11, Codigo: "077-1", Nome: "Banco Exemplo - Internacional"},
		{ID: 21, Codigo: "341", Nome: "Banco X"},
		{ID: 22, Codigo: "341", Nome: "Banco X"},
		{ID: 31, Codigo: "104", Nome: "Banco Y"},
	}
	for _, tc := range []struct {
		nome, codigo, banco string
		id                  int64
		ok                  bool
	}{
		{"só código sem sufixo na lista", "077", "Banco Exemplo S.A.", 0, false},
		{"mesmo código e nome em dois bancos", "341", "Banco X", 0, false},
		{"código único com nome diferente", "104", "Outro Nome", 0, false},
		{"código vazio", "", "Banco Y", 0, false},
		{"código e nome conferem com um só banco", "104", "banco y", 31, true},
		{"código completo com sufixo", "077-1", "Banco Exemplo - Internacional", 11, true},
	} {
		t.Run(tc.nome, func(t *testing.T) {
			if b, ok := bancoAtual(bancos, tc.codigo, tc.banco); ok != tc.ok || b.ID != tc.id {
				t.Errorf("bancoAtual(%q, %q) = %+v, %v; quero id %d, %v", tc.codigo, tc.banco, b, ok, tc.id, tc.ok)
			}
		})
	}
}

// Conta sem bancoId e sem um único banco da lista que confira: editar sem --banco recusa em vez
// de reenviar a conta em outro banco, em --yes e em --dry-run (#871).
func TestContasBancariasEditarSemBancoIDNaoInferido(t *testing.T) {
	const conta = `"contasBancarias":[{"id":1000000000000001,"nomeBanco":"Banco Exemplo S.A.","codigoBanco":"077","agencia":"1234","contaCorrente":"12345-6",
"vlrSaldoInicial":100,"dataSaldoInicial":1704855600000}]`
	for _, tc := range []struct{ nome, bancos, flag string }{
		{"só o internacional", `{"id":11,"codigo":"077-1","nome":"Banco Exemplo - Internacional"}`, "11"},
		{"dois iguais", `{"id":21,"codigo":"077","nome":"Banco Exemplo S.A."},{"id":22,"codigo":"077","nome":"Banco Exemplo S.A."}`, "22"},
	} {
		t.Run(tc.nome, func(t *testing.T) {
			f := contasFake(t)
			f.gets[pathContasBancarias] = `{"bancos":[` + tc.bancos + `],` + conta + `}`
			for _, modo := range []string{"--yes", "--dry-run"} {
				out, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--abertura", "2024-02-01", modo)
				if code != ExitUsage || len(f.writes) != 0 || out != "" || !strings.Contains(stderr, "informe --banco") {
					t.Errorf("%s: código %d, escritas %q, saída %q\n%s", modo, code, f.writes, out, stderr)
				}
			}
			// Com --banco explícito, a edição segue.
			_, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--banco", tc.flag, "--yes")
			if code != ExitOK || len(f.writes) != 1 {
				t.Errorf("com --banco: código %d, escritas %q\n%s", code, f.writes, stderr)
			}
		})
	}
}

func TestContasBancariasEditarSemBancoIDInferidoMostraBancoNoResumo(t *testing.T) {
	f := contasFake(t)
	f.gets[pathContasBancarias] = strings.Replace(contasBancariasJSON, `"nomeBanco":"Banco Exemplo","codigoBanco":"341-7"`, `"nomeBanco":"banco exemplo","codigoBanco":"341-7"`, 1)
	_, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--abertura", "2024-02-01", "--dry-run")
	// Banco inferido pela lista (código e nome) aparece no resumo com o nome e o código da lista.
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(stderr, "→ Banco Exemplo (341-7) agência 1234") {
		t.Errorf("código %d\n%s", code, stderr)
	}
}

func TestContasBancariasAdicionarCodigoExato(t *testing.T) {
	f := contasFake(t)
	f.gets[pathContasBancarias] = `{"bancos":[{"id":11,"codigo":"077-1","nome":"Banco Exemplo - Internacional"},{"id":12,"codigo":"077","nome":"Banco Exemplo S.A."},
{"id":21,"codigo":"341","nome":"Banco Repetido"},{"id":22,"codigo":"341","nome":"Banco Repetido S.A."}],"contasBancarias":[]}`
	args := []string{"contas-bancarias", "adicionar", "--agencia", "0001", "--conta", "12345-6", "--abertura", "2026-08-01", "--declaro-conta-pj", "--dry-run"}
	out, stderr, code := execCLI(t, "", append(args, "--banco", "077")...)
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(stderr, "Cadastrar a conta Banco Exemplo S.A. (077) agência 0001") || !strings.Contains(out, `"bancoId": 12`) {
		t.Errorf("código %d, escritas %v\nstderr %s\nsaída %s", code, f.writes, stderr, out)
	}
	_, stderr, code = execCLI(t, "", append(args, "--banco", "341")...)
	if code != ExitUsage || !strings.Contains(stderr, "corresponde a 2 bancos; informe o id em --banco") || !strings.Contains(stderr, "22  341  Banco Repetido S.A.") {
		t.Errorf("ambíguo: código %d, %s", code, stderr)
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

// A edição reenvia a abertura e o saldo da conta: sem esses dados na lista (zero, null, ausente
// ou uma data implausível como 1969-12-31), recusa em vez de enviar uma data ou um saldo errados (#573).
func TestContasBancariasEditarSemAberturaOuSaldo(t *testing.T) {
	const data, saldo = `"dataSaldoInicial":1704855600000`, `"vlrSaldoInicial":100,`
	for _, tc := range []struct {
		nome, de, para, quer string
	}{
		{"data zero", data, `"dataSaldoInicial":0`, "informe --abertura AAAA-MM-DD"},
		{"data nula", data, `"dataSaldoInicial":null`, "informe --abertura AAAA-MM-DD"},
		{"data ausente", `,` + data, ``, "informe --abertura AAAA-MM-DD"},
		{"data implausível", data, `"dataSaldoInicial":1000`, "informe --abertura AAAA-MM-DD"},
		{"data antes de 1990", data, `"dataSaldoInicial":315543600000`, "informe --abertura AAAA-MM-DD"},
		{"saldo nulo", saldo, `"vlrSaldoInicial":null,`, "informe --saldo-inicial"},
		{"saldo ausente", saldo, ``, "informe --saldo-inicial"},
	} {
		t.Run(tc.nome, func(t *testing.T) {
			f := contasFake(t)
			f.gets[pathContasBancarias] = strings.Replace(contasBancariasJSON, tc.de, tc.para, 1)
			for _, modo := range []string{"--yes", "--dry-run"} {
				out, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--agencia", "9", modo)
				if code != ExitUsage || len(f.writes) != 0 || out != "" || !strings.Contains(stderr, tc.quer) {
					t.Errorf("%s: código %d, escritas %q, saída %q\n%s", modo, code, f.writes, out, stderr)
				}
			}
		})
	}

	// Com a flag que falta, envia o valor informado e mantém o resto.
	f := contasFake(t)
	f.gets[pathContasBancarias] = strings.Replace(strings.Replace(contasBancariasJSON, data, `"dataSaldoInicial":0`, 1), saldo, `"vlrSaldoInicial":null,`, 1)
	if _, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--agencia", "9", "--yes"); code != ExitUsage ||
		!strings.Contains(stderr, "informe --abertura AAAA-MM-DD e --saldo-inicial") {
		t.Errorf("sem abertura nem saldo: código %d, %s", code, stderr)
	}
	out, stderr, code := execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--agencia", "9", "--abertura", "2024-02-01", "--saldo-inicial", "0", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, `"dataSaldoInicial": "2024-02-01"`) || !strings.Contains(out, `"vlrSaldoInicial": 0`) {
		t.Errorf("dry-run com as flags: código %d, escritas %q\n%s\n%s", code, f.writes, out, stderr)
	}
	_, stderr, code = execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--agencia", "9", "--abertura", "2024-02-01", "--saldo-inicial", "0", "--yes")
	want := `POST /api/plataforma/contabancaria/salvar {"id":1000000000000001,"bancoId":33,"agencia":"9","contaCorrente":"123456","dataSaldoInicial":"2024-02-01","vlrSaldoInicial":0}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Errorf("com as flags: código %d, %q\n%s", code, f.writes, stderr)
	}

	// Com abertura e saldo na lista, só --agencia muda e o resto é reenviado como veio.
	f = contasFake(t)
	_, stderr, code = execCLI(t, "", "contas-bancarias", "editar", "1000000000000001", "--agencia", "9", "--yes")
	want = `POST /api/plataforma/contabancaria/salvar {"id":1000000000000001,"bancoId":33,"agencia":"9","contaCorrente":"123456","dataSaldoInicial":"2024-01-10","vlrSaldoInicial":100}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Errorf("conta completa: código %d, %q\n%s", code, f.writes, stderr)
	}
}
