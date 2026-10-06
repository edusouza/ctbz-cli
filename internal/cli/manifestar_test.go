package cli

import (
	"io"
	"strings"
	"testing"
)

func manifestarFake(t *testing.T) *escritaFake {
	t.Helper()
	q := "?mes=9&ano=2026&empresa=&qtdPagina=10&cursor="
	f := newEscritaFake(t, map[string]string{
		"/api/emissor/notasentrada/listar/0" + q: `{"total":2,"cursor":null,"list":[
{"id":701,"razaoSocial":"Fornecedor Fictício","valor":150.25,"situacao":{"id":"PENDENTE"}},
{"id":702,"razaoSocial":"Outro Fornecedor","valor":10,"situacao":{"id":"CIENCIA"}}]}`,
		"/api/emissor/notasentrada/listar/1" + q: `{"total":1,"cursor":null,"list":[{"id":703,"razaoSocial":"Já Manifestada","valor":1,"situacao":{"id":"CONFIRMADA"}}]}`,
	})
	f.resposta = `[{"id":701,"situacao":{"id":"CONFIRMADA"}}]`
	return f
}

func TestManifestarConfirmar(t *testing.T) {
	f := manifestarFake(t)
	out, stderr, code := execCLI(t, "", "notas", "entrada", "manifestar", "701", "703", "--confirmar", "--mes", "2026-09", "--yes", "-o", "json")
	want := `POST /api/emissor/notasentrada/manifestar/ {"tipoManifestacao":"CONFIRMACAO","origem":"PLATAFORMA","idNotas":[701]}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	for _, s := range []string{"Ignorada: nota 703 está CONFIRMADA", "(risco alto)", "não dá para desfazer", "Fornecedor Fictício (R$ 150,25)", avisoManifestacao} {
		if !strings.Contains(stderr, s) {
			t.Errorf("stderr sem %q:\n%s", s, stderr)
		}
	}
	if !strings.Contains(out, `"situacao": "CONFIRMADA"`) || !strings.Contains(out, `"id": "701"`) || !strings.Contains(out, `"manifestacao": "CONFIRMACAO"`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestManifestarNaoRealizadaECiencia(t *testing.T) {
	f := manifestarFake(t)
	_, stderr, code := execCLI(t, "", "notas", "entrada", "manifestar", "702", "--nao-realizada", "--justificativa", "Mercadoria extraviada", "--mes", "2026-09", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.HasSuffix(f.writes[0], `"idNotas":[702],"justificativa":"Mercadoria extraviada"}`) {
		t.Fatalf("não realizada: código %d, %q\n%s", code, f.writes, stderr)
	}

	old := isTerminal
	isTerminal = func(io.Reader) bool { return true }
	t.Cleanup(func() { isTerminal = old })
	_, stderr, code = execCLI(t, "s\n", "notas", "entrada", "manifestar", "701", "--ciencia", "--mes", "2026-09")
	if code != ExitOK || len(f.writes) != 2 || !strings.Contains(f.writes[1], `"tipoManifestacao":"CIENCIA"`) || !strings.Contains(stderr, "(risco médio)") {
		t.Errorf("ciência: código %d, %q\n%s", code, f.writes, stderr)
	}
}

func TestManifestarValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"701"}, ExitUsage, "escolha uma manifestação"},
		{[]string{"701", "--ciencia", "--confirmar"}, ExitUsage, "escolha uma manifestação"},
		{[]string{"701", "--nao-realizada", "--justificativa", "curta"}, ExitUsage, "de 15 a 255 caracteres (tem 5)"},
		{[]string{"701", "--confirmar", "--justificativa", "Mercadoria extraviada"}, ExitUsage, "só vale com --nao-realizada"},
		{[]string{"701", "--confirmar", "--mes", "09/2026"}, ExitUsage, "--mes deve ser AAAA-MM"},
		{[]string{"999", "--confirmar"}, ExitError, "nota 999 não encontrada em 09/2026"},
		{[]string{"703", "--confirmar"}, ExitError, "nenhuma das notas pode ser manifestada"},
	} {
		f := manifestarFake(t)
		args := append([]string{"notas", "entrada", "manifestar", "--yes"}, tc.args...)
		if !strings.Contains(strings.Join(tc.args, " "), "--mes") {
			args = append(args, "--mes", "2026-09")
		}
		_, stderr, code := execCLI(t, "", args...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestManifestarDryRun(t *testing.T) {
	f := manifestarFake(t)
	out, stderr, code := execCLI(t, "", "notas", "entrada", "manifestar", "701", "--desconhecer", "--mes", "2026-09", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, `"tipoManifestacao": "DESCONHECIMENTO"`) || !strings.Contains(stderr, "nada foi enviado") {
		t.Errorf("código %d, %d escritas\n%s\n%s", code, len(f.writes), out, stderr)
	}
}
