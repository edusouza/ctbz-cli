package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// caixaFake simula o caixa de 09/2026: lista os lançamentos (mais os criados por escrita),
// as classificações e grava as escritas recebidas.
type caixaFake struct {
	mu       sync.Mutex
	lancs    []map[string]any
	writes   []string // "MÉTODO caminho corpo"
	onDelete func()
}

func newCaixaFake(t *testing.T) *caixaFake {
	t.Helper()
	old := now
	now = func() time.Time { return time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = old })
	f := &caixaFake{lancs: []map[string]any{
		{"id": 1, "data": 1789732800000, "descricao": "Mensalidade", "valor": -141.83, "situacao": "CONFIRMADO", "idContaUsuario": 5981343255101440, "confirmadoViaSistema": true},
		{"id": 2, "data": 1789700400000, "descricao": "Venda à vista", "valor": 1000, "situacao": "PENDENTE", "idContaUsuario": 1000000000000011, "confirmadoViaSistema": false},
	}}
	var cats struct {
		SerializedList json.RawMessage `json:"serializedList"`
	}
	data, err := os.ReadFile(filepath.Join("..", "api", "testdata", "caixa_categorias.json"))
	if err != nil || json.Unmarshal(data, &cats) != nil {
		t.Fatal(err)
	}
	comp := fixture(t, "contas_usuario_competencia")
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/plataforma/caixa/listpaginada/2026/9/1000/null":
			json.NewEncoder(w).Encode(map[string]any{"list": f.lancs, "total": len(f.lancs), "serializedList": cats.SerializedList})
		case r.Method == "GET" && r.URL.Path == "/api/plataforma/movimentacao-financeira/contas-usuario/2026/9":
			io.WriteString(w, comp)
		case r.Method == "GET":
			http.NotFound(w, r)
		default:
			body, _ := io.ReadAll(r.Body)
			f.writes = append(f.writes, r.Method+" "+r.URL.Path+" "+string(body))
			if r.Method == "DELETE" && f.onDelete != nil {
				f.onDelete()
			}
			if r.URL.Path == "/api/plataforma/caixa/lancamentousuario/novo/" {
				var req struct {
					LancamentoUsuario struct {
						ID        *int64  `json:"id"`
						Data      string  `json:"data"`
						Descricao string  `json:"descricao"`
						Valor     float64 `json:"valor"`
					} `json:"lancamentoUsuario"`
				}
				json.Unmarshal(body, &req)
				d, _ := time.Parse(time.RFC3339, req.LancamentoUsuario.Data)
				l := map[string]any{"id": 99, "data": d.UnixMilli(), "descricao": req.LancamentoUsuario.Descricao, "valor": req.LancamentoUsuario.Valor, "situacao": "PENDENTE"}
				if id := req.LancamentoUsuario.ID; id != nil {
					l["id"] = *id
					for i, x := range f.lancs {
						if x["id"] == int(*id) {
							f.lancs = append(f.lancs[:i], f.lancs[i+1:]...)
							break
						}
					}
				}
				f.lancs = append(f.lancs, l)
			}
		}
	})
	return f
}

func TestCaixaAdicionar(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		corpo string
		saida string
	}{
		{"pagamento com guia", []string{"--pagamento", "--valor", "150,25", "--conta", "Impostos - Simples Nacional", "--guia", "1000000000000021", "--data", "2026-09-15"},
			`{"ano":"2026","mes":9,"lancamentoUsuario":{"data":"2026-09-15T03:00:00.000Z","descricao":"Teste","idContaUsuario":1000000000000013,"idVinculo":"1000000000000021","valor":-150.25}}`,
			`"valor": -150.25`},
		{"recebimento sem vínculo e dia padrão", []string{"--recebimento", "--valor", "1.000,00", "--conta", "1000000000000011"},
			`{"ano":"2026","mes":9,"lancamentoUsuario":{"data":"2026-09-01T03:00:00.000Z","descricao":"Teste","idContaUsuario":1000000000000011,"idVinculo":null,"valor":1000}}`,
			`"valor": 1000.00`},
		{"sem guia", []string{"--pagamento", "--valor", "80", "--conta", "1000000000000013", "--sem-guia", "--data", "2026-09-30"},
			`"idVinculo":"0","valor":-80`, `"acao": "adicionar"`},
		{"sócio", []string{"--pagamento", "--valor", "5000", "--conta", "1000000000000014", "--socio", "1000000000000031", "--data", "2026-09-30"},
			`"idVinculo":1000000000000031,"valor":-5000`, `"situacao": "adicionado"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCaixaFake(t)
			args := append([]string{"caixa", "adicionar", "--competencia", "2026-09", "--descricao", "Teste", "--yes", "-o", "json"}, tc.args...)
			out, stderr, code := execCLI(t, "", args...)
			if code != ExitOK || len(f.writes) != 1 {
				t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
			}
			if !strings.Contains(f.writes[0], tc.corpo) {
				t.Errorf("corpo:\n%s\nquero conter:\n%s", f.writes[0], tc.corpo)
			}
			if !strings.Contains(out, tc.saida) || !strings.Contains(out, `"id": 99`) {
				t.Errorf("saída:\n%s", out)
			}
		})
	}
}

func TestCaixaAdicionarValidacoes(t *testing.T) {
	base := []string{"caixa", "adicionar", "--competencia", "2026-09", "--descricao", "X", "--yes"}
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--valor", "10", "--conta", "1000000000000012"}, "--recebimento ou --pagamento"},
		{[]string{"--pagamento", "--recebimento", "--valor", "10", "--conta", "1000000000000012"}, "--recebimento ou --pagamento"},
		{[]string{"--pagamento", "--valor", "0", "--conta", "1000000000000012"}, "maior que zero"},
		{[]string{"--pagamento", "--valor", "-5", "--conta", "1000000000000012"}, "maior que zero"},
		{[]string{"--pagamento", "--valor", "10", "--conta", "1000000000000012", "--data", "2026-10-05"}, "futura"},
		{[]string{"--pagamento", "--valor", "10", "--conta", "1000000000000011"}, "não aceita para pagamento"},
		{[]string{"--pagamento", "--valor", "10", "--conta", "1000000000000015"}, "não aceita"},    // exibir=false
		{[]string{"--pagamento", "--valor", "10", "--conta", "5981343255101440"}, "não aceita"},    // vetada para pagamento
		{[]string{"--pagamento", "--valor", "10", "--conta", "1000000000000013"}, "exige --guia"},  // imposto
		{[]string{"--pagamento", "--valor", "10", "--conta", "1000000000000014"}, "exige --socio"}, // distribuição
		{[]string{"--pagamento", "--valor", "10", "--conta", "1000000000000012", "--sem-guia"}, "não aceita guia"},
		{[]string{"--pagamento", "--valor", "10", "--conta", "1000000000000013", "--guia", "123"}, "não encontrado"},
		{[]string{"--pagamento", "--valor", "10", "--conta", "1000000000000013", "--guia", "1000000000000021", "--sem-guia"}, "não os dois"},
		{[]string{"--pagamento", "--valor", "10"}, "informe --valor, --conta e --descricao"},
	} {
		f := newCaixaFake(t)
		_, stderr, code := execCLI(t, "", append(base, tc.args...)...)
		if code != ExitUsage || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, stderr %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestCaixaAdicionarDryRun(t *testing.T) {
	f := newCaixaFake(t)
	out, stderr, code := execCLI(t, "", "caixa", "adicionar", "--competencia", "2026-09", "--data", "2026-09-15", "--pagamento",
		"--valor", "150,25", "--conta", "1000000000000012", "--descricao", "Material", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, `"valor": -150.25`) ||
		!strings.Contains(stderr, `Adicionar pagamento de R$ 150,25 em 15/09/2026 no caixa de 09/2026: "Material" (Pagamento de Fornecedores)`) {
		t.Errorf("código %d, escritas %v\nstdout:\n%s\nstderr:\n%s", code, f.writes, out, stderr)
	}
}

func TestCaixaEditar(t *testing.T) {
	f := newCaixaFake(t)
	out, stderr, code := execCLI(t, "", "caixa", "editar", "2", "--competencia", "2026-09", "--valor", "900", "--descricao", "Venda balcão", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 {
		t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
	}
	want := `{"ano":"2026","mes":9,"lancamentoUsuario":{"data":"2026-09-18T03:00:00.000Z","descricao":"Venda balcão","id":2,"idContaUsuario":1000000000000011,"idVinculo":null,"valor":900}}`
	if !strings.HasSuffix(f.writes[0], want) {
		t.Errorf("corpo:\n%s\nquero:\n%s", f.writes[0], want)
	}
	if !strings.Contains(stderr, `Editar o lançamento 2 do caixa de 09/2026: valor R$ 1.000,00 → R$ 900,00; descrição "Venda à vista" → "Venda balcão"`) {
		t.Errorf("resumo: %s", stderr)
	}
	if !strings.Contains(out, `"situacao": "editado"`) || !strings.Contains(out, `"id": 2`) || !strings.Contains(out, `"valor": 900.00`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestCaixaEditarTrocaTipoEConta(t *testing.T) {
	f := newCaixaFake(t)
	_, stderr, code := execCLI(t, "", "caixa", "editar", "2", "--competencia", "2026-09", "--pagamento",
		"--conta", "Impostos - Simples Nacional", "--sem-guia", "--data", "2026-09-20", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `"data":"2026-09-20T03:00:00.000Z"`) ||
		!strings.Contains(f.writes[0], `"idContaUsuario":1000000000000013,"idVinculo":"0","valor":-1000`) {
		t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "tipo recebimento → pagamento; data 18/09/2026 → 20/09/2026; conta Receita de Serviços → Impostos - Simples Nacional; vínculo: SEM GUIA") {
		t.Errorf("resumo: %s", stderr)
	}
}

func TestCaixaEditarValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"2"}, ExitUsage, "informe o que alterar"},
		{[]string{"2", "--valor", "1000"}, ExitUsage, "nada muda"},
		{[]string{"1", "--valor", "10"}, ExitError, "feito pelo sistema"},
		{[]string{"77", "--valor", "10"}, ExitError, "não encontrado"},
		{[]string{"abc", "--valor", "10"}, ExitUsage, "inválido"},
		{[]string{"2", "--pagamento"}, ExitUsage, "não aceita para pagamento"}, // a conta atual é de recebimento
		{[]string{"2", "--recebimento", "--pagamento"}, ExitUsage, "não os dois"},
	} {
		f := newCaixaFake(t)
		_, stderr, code := execCLI(t, "", append([]string{"caixa", "editar", "--competencia", "2026-09", "--yes"}, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, stderr %q (quero %d, %q)", tc.args, code, len(f.writes), stderr, tc.code, tc.want)
		}
	}
}

func TestCaixaEditarDryRunMostraAntesEDepois(t *testing.T) {
	f := newCaixaFake(t)
	out, stderr, code := execCLI(t, "", "caixa", "editar", "2", "--competencia", "2026-09", "--valor", "1,50", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(stderr, "valor R$ 1.000,00 → R$ 1,50") || !strings.Contains(out, `"id": 2`) {
		t.Errorf("código %d\nstdout:\n%s\nstderr:\n%s", code, out, stderr)
	}
}

func TestCaixaRemover(t *testing.T) {
	f := newCaixaFake(t)
	out, stderr, code := execCLI(t, "", "caixa", "remover", "2", "--competencia", "2026-09", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, "DELETE /api/plataforma/caixa/lancamentousuario/remover/2026/9/2") ||
		!strings.Contains(stderr, `Excluir permanentemente o lançamento 2 do caixa de 09/2026: "Venda à vista" de 18/09/2026 no valor de R$ 1.000,00 (risco médio)`) {
		t.Fatalf("--dry-run: código %d\n%s\n%s", code, out, stderr)
	}
	out, stderr, code = execCLI(t, "", "caixa", "remover", "2", "--competencia", "2026-09", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "DELETE /api/plataforma/caixa/lancamentousuario/remover/2026/9/2 " {
		t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	// O fake mantém o lançamento: a CLI avisa que ele ainda aparece.
	if !strings.Contains(out, `"situacao": "enviado"`) || !strings.Contains(stderr, "ainda aparece") {
		t.Errorf("saída:\n%s\n%s", out, stderr)
	}
	if !strings.Contains(stderr, `Para recriar: ctbz caixa adicionar --competencia 2026-09 --data 2026-09-18 --recebimento --valor 1.000,00 --conta 1000000000000011 --descricao "Venda à vista"`) {
		t.Errorf("comando para recriar: %s", stderr)
	}
	for _, want := range []string{`"acao": "remover"`, `"id": 2`, `"id_conta": 1000000000000011`, `"valor": 1000.00`, `"data": "2026-09-18"`} {
		if !strings.Contains(out, want) {
			t.Errorf("saída sem %s:\n%s", want, out)
		}
	}
}

func TestCaixaRemoverRemovido(t *testing.T) {
	f := newCaixaFake(t)
	f.onDelete = func() { f.lancs = f.lancs[:1] }
	out, _, code := execCLI(t, "", "caixa", "remover", "2", "--competencia", "2026-09", "--yes", "-o", "json")
	if code != ExitOK || !strings.Contains(out, `"situacao": "removido"`) {
		t.Errorf("código %d:\n%s", code, out)
	}
	_, stderr, code := execCLI(t, "", "caixa", "remover", "1", "--competencia", "2026-09", "--yes")
	if code != ExitError || !strings.Contains(stderr, "feito pelo sistema") || len(f.writes) != 1 {
		t.Errorf("lançamento do sistema: código %d, %s", code, stderr)
	}
}
