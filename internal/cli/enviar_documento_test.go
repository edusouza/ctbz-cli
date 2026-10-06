package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	pathEnvioInit1 = "/api/plataforma/documentos/envio-documento/init?id=1000000000000601"
	pathEnvioInit2 = "/api/plataforma/documentos/envio-documento/init?id=1000000000000601&id=1000000000000602"
	pathEnvioInit3 = "/api/plataforma/documentos/envio-documento/init?id=1000000000000601&id=1000000000000603"
)

func documentoFake(t *testing.T) *escritaFake {
	t.Helper()
	ini := fixture(t, "envio_documento_init")
	return newEscritaFake(t, map[string]string{
		pathEnvioInit1: ini, pathEnvioInit2: ini, pathEnvioInit3: ini,
		"/api/plataforma/dashboard/v2/central-rotinas": `{"pendencias":{},"rotinasContabilizei":[],"rotinas":[{"tipo":"DOC","prazo":"2026-10-10","status":"EM_ABERTO","automatica":false,
"propriedades":{"documentosPendentes":["1000000000000601"]}}]}`,
	})
}

func arquivoPDF(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "aplicacao.pdf")
	os.WriteFile(p, []byte("%PDF-1.4"), 0o600)
	return p
}

func TestDocumentosPendentes(t *testing.T) {
	documentoFake(t)
	out, stderr, code := execCLI(t, "", "documentos", "pendentes", "-o", "csv")
	want := "id_pendencia,tipo,competencia,conta,tipo_investimento\n" +
		"1000000000000601,EXTRATO_APLICACAO_FINANCEIRA,08/2026,Banco Exemplo ag. 1234 conta 123456,RENDA_FIXA\n" +
		"1000000000000602,EXTRATO_APLICACAO_FINANCEIRA,09/2026,Banco Exemplo ag. 1234 conta 123456,RENDA_FIXA\n" +
		"1000000000000603,ESTOQUE,12/2025,,\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s\nquero:\n%s", code, stderr, out, want)
	}
}

func TestDocumentosEnviar(t *testing.T) {
	arq := arquivoPDF(t)
	f := documentoFake(t)
	f.onWrite = func(f *escritaFake) { f.gets[pathEnvioInit1] = `{"documentos":[]}` }
	out, stderr, code := execCLI(t, "", "documentos", "enviar", arq, "--pendencia", "1000000000000601", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || !strings.HasPrefix(f.writes[0], "POST /api/plataforma/documentos/envio-documento/enviar ") {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	for _, w := range []string{`name="tipoDocumento"` + "\r\n\r\nEXTRATO_APLICACAO_FINANCEIRA", `{"mes":8,"ano":2026}`,
		`{"valor":null,"idPendencia":"1000000000000601","periodo":null,"idContaBancaria":7,"tipoInvestimento":"RENDA_FIXA"}`, "%PDF-1.4"} {
		if !strings.Contains(f.writes[0], w) {
			t.Errorf("envio sem %q", w)
		}
	}
	if !strings.Contains(out, `"situacao": "resolvido"`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestDocumentosEnviarConsolidado(t *testing.T) {
	arq := arquivoPDF(t)
	f := documentoFake(t)
	_, stderr, code := execCLI(t, "", "documentos", "enviar", arq, "--pendencia", "1000000000000601", "--pendencia", "1000000000000602", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.HasPrefix(f.writes[0], "POST /api/plataforma/documentos/envio-documento/enviar/consolidado ") ||
		!strings.Contains(f.writes[0], `"idPendencia":"1000000000000602"`) || !strings.Contains(stderr, "para 2 pendência(s) (08/2026, 09/2026)") {
		t.Errorf("código %d, %q\n%s", code, f.writes, stderr)
	}
}

func TestDocumentosEnviarErros(t *testing.T) {
	arq := arquivoPDF(t)
	f := documentoFake(t)
	if _, stderr, code := execCLI(t, "", "documentos", "enviar", arq, "--pendencia", "1000000000000601", "--pendencia", "1000000000000603", "--yes"); code != ExitUsage || !strings.Contains(stderr, "tipos diferentes") {
		t.Errorf("tipos diferentes: código %d, %s", code, stderr)
	}
	if _, _, code := execCLI(t, "", "documentos", "enviar", arq, "--yes"); code != ExitUsage {
		t.Errorf("sem pendência: código %d", code)
	}
	f.status = 400
	f.resposta = "(ERRO-NA-VALIDACAO-DE-DOCUMENTO-API-CONTABIL) O arquivo não é um extrato de aplicação"
	_, stderr, code := execCLI(t, "", "documentos", "enviar", arq, "--pendencia", "1000000000000601", "--yes")
	if code != ExitError || !strings.Contains(stderr, "recusou o documento: requisição recusada: O arquivo não é um extrato de aplicação") || strings.Contains(stderr, "ERRO-NA-VALIDACAO") {
		t.Errorf("erro com prefixo: código %d, %s", code, stderr)
	}
}
