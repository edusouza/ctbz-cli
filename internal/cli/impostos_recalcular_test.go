package cli

import (
	"io"
	"strings"
	"testing"
)

const pathRecalcInit = "/api/plataforma/impostos/v2/impostos-a-pagar/recalculo/init?idGuia=1000000000000001&origem=GUIAS&tipo=GUIA"

func recalculoFake(t *testing.T) *escritaFake {
	t.Helper()
	fixNow(t, "2026-10-04")
	f := impostosFake(t, "v5")
	f.gets[pathGuia1] = strings.Replace(fixture(t, "guia_detalhe"), `"PAGAR"`, `"RECALCULAR"`, 1)
	f.gets[pathRecalcInit] = fixture(t, "recalculo_init")
	return f
}

func TestImpostosRecalcularSoMostra(t *testing.T) {
	f := recalculoFake(t)
	out, stderr, code := execCLI(t, "", "impostos", "recalcular", "1000000000000001", "-o", "json")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, `"data_recomendada": "2026-10-09"`) ||
		!strings.Contains(out, `"valor_recalculo": 1234.56`) || !strings.Contains(stderr, "--vencimento 2026-10-09") {
		t.Errorf("código %d\n%s\n%s", code, out, stderr)
	}
}

func TestImpostosRecalcular(t *testing.T) {
	f := recalculoFake(t)
	f.resposta = `{"modalSucessoRecalculo":true,"guia":{"dataVencimento":"20/10/2026"}}`
	// Sem terminal e sem --yes: nada é enviado.
	if _, _, code := execCLI(t, "", "impostos", "recalcular", "1000000000000001", "--vencimento", "2026-10-20"); code != ExitUsage || len(f.writes) != 0 {
		t.Fatalf("sem --yes: código %d, %d escritas", code, len(f.writes))
	}
	out, stderr, code := execCLI(t, "", "impostos", "recalcular", "1000000000000001", "--vencimento", "2026-10-20", "--yes", "-o", "json")
	want := `PUT /api/plataforma/impostos/v2/impostos-a-pagar/guia/1000000000000001/v2/recalcular {"tipo":"GUIA","origem":"GUIAS","dataVencimento":"20/10/2026"}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
	}
	for _, w := range []string{"(risco alto)", "cobrado na sua próxima mensalidade", "Não dá para cancelar", "até 3 dias úteis"} {
		if !strings.Contains(stderr, w) {
			t.Errorf("stderr sem %q: %s", w, stderr)
		}
	}
	if !strings.Contains(out, `"situacao": "solicitado"`) || !strings.Contains(out, `"vencimento": "2026-10-20"`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestImpostosRecalcularConfirmoNoTerminal(t *testing.T) {
	f := recalculoFake(t)
	old := isTerminal
	t.Cleanup(func() { isTerminal = old })
	isTerminal = func(io.Reader) bool { return true }
	_, stderr, code := execCLI(t, "s\n", "impostos", "recalcular", "1000000000000001", "--vencimento", "2026-10-20")
	if code != ExitError || len(f.writes) != 0 || !strings.Contains(stderr, `Digite "confirmo"`) {
		t.Errorf("'s' não basta no risco alto: código %d, %s", code, stderr)
	}
}

// Sem cobrarRecalculo na resposta o formato real é desconhecido: a CLI não pode afirmar que é
// gratuito (#463).
func TestImpostosRecalcularCobrancaDesconhecida(t *testing.T) {
	for _, corpo := range []string{`"cobrarRecalculo": null,`, ``} {
		f := recalculoFake(t)
		f.resposta = `{"modalSucessoRecalculo":true,"guia":{"dataVencimento":"20/10/2026"}}`
		f.gets[pathRecalcInit] = strings.Replace(fixture(t, "recalculo_init"), `"cobrarRecalculo": true,`, corpo, 1)
		out, _, _ := execCLI(t, "", "impostos", "recalcular", "1000000000000001", "-o", "json")
		if !strings.Contains(out, `"cobrar_recalculo": null`) {
			t.Errorf("leitura (%q): cobrar_recalculo deveria ser null:\n%s", corpo, out)
		}
		out, stderr, code := execCLI(t, "", "impostos", "recalcular", "1000000000000001", "--vencimento", "2026-10-20", "--yes", "-o", "json")
		if code != ExitOK || !strings.Contains(stderr, "não informou se será cobrado") || strings.Contains(stderr, "Sem cobrança") {
			t.Errorf("consequência (%q): código %d\n%s", corpo, code, stderr)
		}
		if !strings.Contains(out, `"cobrado": null`) {
			t.Errorf("resultado (%q): cobrado deveria ser null:\n%s", corpo, out)
		}
	}
}

// Valor recalculado nulo não pode aparecer como "R$ 0,00" na confirmação (#462).
func TestImpostosRecalcularValorDesconhecido(t *testing.T) {
	f := recalculoFake(t)
	f.resposta = `{"modalSucessoRecalculo":true,"guia":{"dataVencimento":"20/10/2026"}}`
	f.gets[pathRecalcInit] = strings.Replace(fixture(t, "recalculo_init"), `"valorRecalculo": 1234.56,`, `"valorRecalculo": null,`, 1)
	_, stderr, code := execCLI(t, "", "impostos", "recalcular", "1000000000000001", "--vencimento", "2026-10-20", "--yes")
	if code != ExitOK || !strings.Contains(stderr, "valor estimado indisponível") || strings.Contains(stderr, "R$ 0,00") {
		t.Errorf("código %d\n%s", code, stderr)
	}
}

// Sem confirmação do novo vencimento na resposta, a CLI não pode apresentar a data pedida como
// se fosse a do painel: relê a guia e separa o que foi pedido do que foi confirmado (#464).
func TestImpostosRecalcularSemConfirmacaoDoVencimento(t *testing.T) {
	f := recalculoFake(t)
	f.resposta = `{"modalSucessoRecalculo":true}`
	out, stderr, code := execCLI(t, "", "impostos", "recalcular", "1000000000000001", "--vencimento", "2026-10-20", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 {
		t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
	}
	for _, w := range []string{`"vencimento": null`, `"vencimento_pedido": "2026-10-20"`, `"status": [`, `"RECALCULADA"`} {
		if !strings.Contains(out, w) {
			t.Errorf("saída sem %s:\n%s", w, out)
		}
	}
	if !strings.Contains(stderr, "não confirmou o novo vencimento") {
		t.Errorf("stderr sem aviso:\n%s", stderr)
	}
}

func TestImpostosRecalcularValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
		prep func(*escritaFake)
	}{
		{[]string{"--vencimento", "20/10/2026"}, ExitUsage, "AAAA-MM-DD", nil},
		{[]string{"--vencimento", "2026-10-01"}, ExitUsage, "já passou", nil},
		{[]string{"--vencimento", "2026-10-11"}, ExitUsage, "indisponível", nil},
		{[]string{"--vencimento", "2026-10-20"}, ExitError, "marcada como não paga", func(f *escritaFake) { f.gets[pathGuia1] = fixture(t, "guia_detalhe") }},
		{[]string{"--vencimento", "2026-10-20"}, ExitError, "pendências críticas", func(f *escritaFake) {
			f.gets[pathCentralRt] = strings.Replace(centralLimpa, "false", "true", 1)
		}},
	} {
		f := recalculoFake(t)
		if tc.prep != nil {
			tc.prep(f)
		}
		_, stderr, code := execCLI(t, "", append([]string{"impostos", "recalcular", "1000000000000001", "--yes"}, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, %q (quero %d, %q)", tc.args, code, len(f.writes), stderr, tc.code, tc.want)
		}
	}
}
