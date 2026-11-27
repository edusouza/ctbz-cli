package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

const idFaturaTeste = "0f8fad5b-d9cb-469f-a165-70867728950e"

// segundaViaFake serve o card da fatura (com a ação dada), o GET de segunda via (contado) e as
// faturas públicas do checkout. publica muda a resposta da v2.
type segundaViaFake struct {
	acao    string // botaoAdicionalAcao (JSON: "null" ou "\"request:…\"")
	botao   string // botaoAcao (JSON); padrão: um redirect
	publica string
	status  int // status das faturas públicas
	// resumo e statusResumo mudam a resposta da v1 (itens e situação) quando preenchidos.
	resumo       string
	statusResumo int
	geradas      atomic.Int32
}

func newSegundaViaFake(t *testing.T) *segundaViaFake {
	t.Helper()
	f := &segundaViaFake{acao: `"request:dashboard/nova-fatura/202610"`, botao: `"redirect:/painel-de-controle/#/pagar-com-cartao"`, publica: fixture(t, "fatura_publica"), status: http.StatusOK,
		resumo: fixture(t, "fatura_publica_resumo")}
	// A fixture vence em 25/10/2026: fixa "hoje" antes do vencimento.
	antes := now
	now = func() time.Time { return time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = antes })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/plataforma/" + api.PathAppBar:
			io.WriteString(w, "{}")
		case "/api/plataforma/dashboard/fatura":
			io.WriteString(w, `{"total":1234.56,"competencia":"Outubro, 2026","status":"Vencida","vencimento":null,"mensagemAdicional":"",`+
				`"possuiCartaoPrincipal":false,"botaoTitulo":"Pagar com cartão","botaoAcao":`+f.botao+`,`+
				`"botaoAdicionalTitulo":"Gerar segunda via s/ juros","botaoAdicionalAcao":`+f.acao+`}`)
		case "/api/plataforma/dashboard/v1/mensalidade":
			io.WriteString(w, fixture(t, "mensalidade"))
		case "/api/plataforma/dashboard/nova-fatura/202610":
			f.geradas.Add(1)
			io.WriteString(w, `{"url":"`+srvURL(r)+`/checkout/#/fatura-contabilizei?idFatura=`+idFaturaTeste+`"}`)
		case api.PathFaturaPublica(idFaturaTeste):
			w.WriteHeader(f.status)
			io.WriteString(w, f.publica)
		case api.PathFaturaPublicaResumo(idFaturaTeste):
			if f.statusResumo != 0 {
				w.WriteHeader(f.statusResumo)
			} else {
				w.WriteHeader(f.status)
			}
			io.WriteString(w, f.resumo)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	withSession(t, srv)
	return f
}

func srvURL(r *http.Request) string { return "http://" + r.Host }

func TestSegundaVia(t *testing.T) {
	f := newSegundaViaFake(t)
	out, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", "--yes", "-o", "json")
	if code != ExitOK || f.geradas.Load() != 1 {
		t.Fatalf("código %d, %d segunda(s) via(s)\n%s", code, f.geradas.Load(), stderr)
	}
	for _, want := range []string{
		`"acao": "segunda-via"`, `"situacao": "gerada"`, `"id": "202610"`, `"competencia": "10/2026"`,
		`"vencimento": "2026-10-25"`, `"total": 1234.56`, `"status": "Pendente"`,
		`"descricao": "Multa e juros da fatura anterior"`, `"valor": 34.56`,
		`"pix_copia_e_cola": "00020101021226`, `"linha_digitavel": "00190000090000000000000000000000000000000000000"`,
		`"codigo_barras": "00190000000000000000000000000000000000000000"`,
		`"link": "http://127.0.0.1`, `/checkout/#/fatura-contabilizei?idFatura=` + idFaturaTeste,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída sem %s:\n%s", want, out)
		}
	}
	if strings.Contains(out, "token") || strings.Contains(out, "data:image") {
		t.Errorf("a saída não pode ter o token nem o QR:\n%s", out)
	}
	if !strings.Contains(stderr, "Gerar a segunda via da mensalidade de 10/2026") || !strings.Contains(stderr, "ctbz mensalidade situacao") {
		t.Errorf("stderr:\n%s", stderr)
	}
	store, _ := ctbz.DefaultStore()
	raw, _ := os.ReadFile(filepath.Join(store.Dir, "acoes.jsonl"))
	if !strings.Contains(string(raw), `"caminho":"/api/plataforma/dashboard/nova-fatura/202610"`) || strings.Contains(string(raw), idFaturaTeste) {
		t.Errorf("registro de ações:\n%s", raw)
	}
}

func TestSegundaViaSoOCodigo(t *testing.T) {
	newSegundaViaFake(t)
	out, _, code := execCLI(t, "", "mensalidade", "segunda-via", "--pix", "--yes")
	pix := "00020101021226000000000000000000000000000000BR.GOV.BCB.PIX5204000053039865802BR5913EMPRESA EXEMPLO6009SAO PAULO62070503***6304ABCD\n"
	if code != ExitOK || out != pix {
		t.Errorf("--pix: código %d, saída %q", code, out)
	}
	out, _, code = execCLI(t, "", "mensalidade", "segunda-via", "--linha", "--yes")
	if code != ExitOK || out != "00190000090000000000000000000000000000000000000\n" {
		t.Errorf("--linha: código %d, saída %q", code, out)
	}
	_, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", "--pix", "--linha", "--yes")
	if code != ExitUsage || !strings.Contains(stderr, "--pix ou --linha") {
		t.Errorf("--pix com --linha: código %d\n%s", code, stderr)
	}
}

func TestSegundaViaSemOMeioPedido(t *testing.T) {
	f := newSegundaViaFake(t)
	f.publica = `{"dataVencimento":"2026-10-25","total":10,"meioDePagamento":{"pix":{"textoPagamento":"","exibir":false},"boleto":null}}`
	for _, c := range []struct{ flag, want string }{
		{"--pix", "a segunda via foi gerada, mas a fatura não oferece Pix"},
		{"--linha", "a segunda via foi gerada, mas a fatura não oferece boleto"},
	} {
		out, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", c.flag, "--yes")
		if code != ExitError || out != "" || !strings.Contains(stderr, c.want) || !strings.Contains(stderr, "/checkout/#/fatura-contabilizei") {
			t.Errorf("%s: código %d, saída %q\n%s", c.flag, code, out, stderr)
		}
	}
}

func TestSegundaViaDryRunEConfirmacao(t *testing.T) {
	f := newSegundaViaFake(t)
	out, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", "--dry-run")
	if code != ExitOK || f.geradas.Load() != 0 || !strings.Contains(out, "GET /api/plataforma/dashboard/nova-fatura/202610") {
		t.Errorf("dry-run: código %d, %d geradas\n%s\n%s", code, f.geradas.Load(), out, stderr)
	}
	_, stderr, code = execCLI(t, "", "mensalidade", "segunda-via")
	if code != ExitUsage || f.geradas.Load() != 0 || !strings.Contains(stderr, "use --yes") {
		t.Errorf("sem terminal: código %d, %d geradas\n%s", code, f.geradas.Load(), stderr)
	}
}

func TestSegundaViaSemAcaoNoCard(t *testing.T) {
	const naoOferece = "o card da mensalidade não oferece segunda via agora"
	const inesperado = "o card da mensalidade pediu um caminho inesperado"
	for _, c := range []struct{ acao, botao, want string }{
		{"null", "", naoOferece + " (a fatura pode estar em dia; veja ctbz mensalidade)"},
		{`"redirect:/painel-de-controle/#/x"`, "", naoOferece},
		{`"request:https://example.com/x"`, "", inesperado},
		{`"request:dashboard/../conta"`, "", inesperado},
		// Outra operação do BFF não pode ser enviada sob o resumo de segunda via (#929).
		{`"request:dashboard/conta"`, "", inesperado},
		{`"request:dashboard/pagar-com-cartao"`, "", inesperado},
		{"null", `"request:dashboard/pagar-com-cartao"`, inesperado},
		{"null", `"request:dashboard/conta"`, inesperado},
		// Sem o mês no caminho o ID e o rótulo do resumo ficariam desconhecidos.
		{`"request:dashboard/nova-fatura"`, "", inesperado},
		{`"request:dashboard/nova-fatura/202613"`, "", inesperado},
	} {
		f := newSegundaViaFake(t)
		f.acao = c.acao
		if c.botao != "" {
			f.botao = c.botao
		}
		_, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", "--yes")
		if code != ExitError || f.geradas.Load() != 0 || !strings.Contains(stderr, c.want) {
			t.Errorf("%s / %s: código %d, %d geradas\n%s", c.acao, c.botao, code, f.geradas.Load(), stderr)
		}
	}
}

// Com a fatura em dia só o botaoAcao traz uma ação; se for a segunda via legítima, vale.
func TestSegundaViaAceitaABotaoAcaoComOMes(t *testing.T) {
	f := newSegundaViaFake(t)
	f.acao, f.botao = "null", `"request:dashboard/nova-fatura/202610"`
	_, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", "--yes")
	if code != ExitOK || f.geradas.Load() != 1 {
		t.Errorf("código %d, %d geradas\n%s", code, f.geradas.Load(), stderr)
	}
}

func TestSegundaViaCompetencia(t *testing.T) {
	f := newSegundaViaFake(t)
	_, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", "--competencia", "2026-09", "--yes")
	if code != ExitError || f.geradas.Load() != 0 || !strings.Contains(stderr, "o card oferece a segunda via de 10/2026, não de 09/2026") {
		t.Errorf("competência diferente: código %d, %d geradas\n%s", code, f.geradas.Load(), stderr)
	}
	if _, stderr, code = execCLI(t, "", "mensalidade", "segunda-via", "--competencia", "2026-10", "--yes"); code != ExitOK || f.geradas.Load() != 1 {
		t.Errorf("competência igual: código %d\n%s", code, stderr)
	}
	if _, stderr, code = execCLI(t, "", "mensalidade", "segunda-via", "--competencia", "10/2026", "--yes"); code != ExitUsage {
		t.Errorf("competência inválida: código %d\n%s", code, stderr)
	}
}

func TestSegundaViaLeituraFalhaDepoisDeGerar(t *testing.T) {
	f := newSegundaViaFake(t)
	f.status = http.StatusInternalServerError
	_, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", "--yes")
	if code != ExitError || f.geradas.Load() != 1 {
		t.Fatalf("código %d, %d geradas\n%s", code, f.geradas.Load(), stderr)
	}
	if !strings.Contains(stderr, "a segunda via foi gerada, mas não foi possível ler a fatura") ||
		!strings.Contains(stderr, "não gere outra; abra o link para pagar: http://127.0.0.1") {
		t.Errorf("stderr:\n%s", stderr)
	}
	// O UUID só aparece no link, não no erro da leitura.
	if strings.Count(stderr, idFaturaTeste) != 1 {
		t.Errorf("o UUID aparece fora do link:\n%s", stderr)
	}
}

func TestMensalidadeMostraAsAcoesDoCard(t *testing.T) {
	newSegundaViaFake(t)
	out, _, code := execCLI(t, "", "mensalidade", "-o", "json")
	for _, want := range []string{`"botao_titulo": "Pagar com cartão"`, `"botao_acao": "redirect:/painel-de-controle/#/pagar-com-cartao"`,
		`"botao_adicional_titulo": "Gerar segunda via s/ juros"`, `"botao_adicional_acao": "request:dashboard/nova-fatura/202610"`} {
		if code != ExitOK || !strings.Contains(out, want) {
			t.Errorf("código %d, saída sem %s:\n%s", code, want, out)
		}
	}
}

// --pix e --linha existem para ir direto a um pagamento: o código só sai se a fatura ainda
// puder ser paga (#931).
func TestSegundaViaCodigoSoDeFaturaPagavel(t *testing.T) {
	semTotal := `{"dataVencimento":"2026-10-25","total":null,"meioDePagamento":{"pix":{"textoPagamento":"PIXCOPIA","exibir":true},"boleto":{"textoPagamento":"LINHA","exibir":true}}}`
	totalZero := strings.Replace(semTotal, `"total":null`, `"total":0`, 1)
	vencida := strings.Replace(semTotal, `"total":null`, `"total":10`, 1)
	vencida = strings.Replace(vencida, "2026-10-25", "2026-10-19", 1)
	for _, c := range []struct{ nome, publica, resumo, want string }{
		{"paga", "", `{"resumo":{"itens":[]},"status":{"id":"PAGA","descricao":"Paga"}}`, "já foi paga"},
		{"cancelada", "", `{"resumo":{"itens":[]},"status":{"id":"CANCELADA"}}`, "foi cancelada"},
		{"total nulo", semTotal, "", "sem valor a pagar"},
		{"total zero", totalZero, "", "sem valor a pagar"},
		{"vencimento passado", vencida, "", "venceu em 19/10/2026"},
	} {
		for _, flag := range []string{"--pix", "--linha"} {
			f := newSegundaViaFake(t)
			if c.publica != "" {
				f.publica = c.publica
			}
			if c.resumo != "" {
				f.resumo = c.resumo
			}
			out, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", flag, "--yes")
			if code != ExitError || out != "" || !strings.Contains(stderr, c.want) || !strings.Contains(stderr, "/checkout/#/fatura-contabilizei") {
				t.Errorf("%s %s: código %d, saída %q\n%s", c.nome, flag, code, out, stderr)
			}
		}
	}
}

// Sem o resumo v1 (que os modos de pipe não precisam), o código que a v2 trouxe ainda sai, com
// aviso: falhar aqui faria o usuário repetir a escrita e gerar outra segunda via.
func TestSegundaViaResumoComErroNaoImpedeOCodigo(t *testing.T) {
	f := newSegundaViaFake(t)
	f.statusResumo = http.StatusInternalServerError
	out, stderr, code := execCLI(t, "", "mensalidade", "segunda-via", "--pix", "--yes")
	if code != ExitOK || !strings.HasPrefix(out, "00020101021226") || strings.Count(out, "\n") != 1 {
		t.Errorf("--pix: código %d, saída %q\n%s", code, out, stderr)
	}
	if !strings.Contains(stderr, "aviso: não foi possível ler a situação da fatura") || strings.Count(stderr, idFaturaTeste) != 1 {
		t.Errorf("stderr:\n%s", stderr)
	}
	out, stderr, code = execCLI(t, "", "mensalidade", "segunda-via", "--yes", "-o", "json")
	if code != ExitOK || !strings.Contains(out, `"pix_copia_e_cola": "00020101021226`) || !strings.Contains(stderr, "aviso: não foi possível ler a situação da fatura") {
		t.Errorf("modo completo: código %d\n%s\n%s", code, out, stderr)
	}
}
