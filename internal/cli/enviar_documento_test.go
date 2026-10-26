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

func TestDocumentosSemArquivo(t *testing.T) {
	f := documentoFake(t)
	f.gets["/api/plataforma/documentos/envio-documento/init?id=1000000000000603"] = fixture(t, "envio_documento_init")
	f.onWrite = func(f *escritaFake) {
		f.gets["/api/plataforma/documentos/envio-documento/init?id=1000000000000603"] = `{"documentos":[]}`
	}
	out, stderr, code := execCLI(t, "", "documentos", "sem-arquivo", "--pendencia", "1000000000000603", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `POST /api/plataforma/documentos/envio-documento/enviar/sem-arquivo [{"idPendencia":"1000000000000603","tipo":"ESTOQUE"}]` {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, `"situacao": "declarado"`) || !strings.Contains(stderr, "inconsistências na declaração anual") || !strings.Contains(stderr, "(risco alto)") {
		t.Errorf("saída:\n%s\n%s", out, stderr)
	}
	if _, stderr, code := execCLI(t, "", "documentos", "sem-arquivo", "--pendencia", "1000000000000601", "--yes"); code != ExitUsage || !strings.Contains(stderr, "não aceitam a declaração sem arquivo") {
		t.Errorf("tipo errado: código %d, %s", code, stderr)
	}
}

func TestDocumentosSemAplicacao(t *testing.T) {
	f := documentoFake(t)
	out, stderr, code := execCLI(t, "", "documentos", "sem-aplicacao", "--conta-bancaria", "7", "--pendencia", "1000000000000601", "--pendencia", "1000000000000602", "--yes", "-o", "json")
	want := `POST /api/plataforma/upload-documentos/extrato-aplicacao-financeira/enviar/sem-aplicacao-financeira {"tipoDocumento":"EXTRATO_APLICACAO_FINANCEIRA","competenciasPendentes":[{"id":"1000000000000601","mes":8,"ano":2026},{"id":"1000000000000602","mes":9,"ano":2026}],"idContaBancaria":7,"tipoPendencia":"EXTRATO_APLICACAO_FINANCEIRA"}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want || !strings.Contains(out, `"situacao": "declarado"`) {
		t.Fatalf("código %d, %q\n%s\n%s", code, f.writes, out, stderr)
	}
	// Sucesso parcial (resposta de erro com competenciasComErro): lista e termina com 1.
	f.status = 400
	f.resposta = `{"competenciasComErro":[{"mes":9,"ano":2026}]}`
	out, stderr, code = execCLI(t, "", "documentos", "sem-aplicacao", "--conta-bancaria", "7", "--pendencia", "1000000000000601", "--pendencia", "1000000000000602", "--yes", "-o", "json")
	if code != ExitError || !strings.Contains(out, `"situacao": "parcial"`) || !strings.Contains(stderr, "não aceitou 1 competência(s): 09/2026") {
		t.Errorf("parcial: código %d\n%s\n%s", code, out, stderr)
	}
	if _, stderr, code := execCLI(t, "", "documentos", "sem-aplicacao", "--conta-bancaria", "8", "--pendencia", "1000000000000601", "--yes"); code != ExitUsage || !strings.Contains(stderr, "é de outra conta") {
		t.Errorf("outra conta: código %d, %s", code, stderr)
	}
}

// initCompetencia é o init das pendências 601 e 602 de extrato de aplicação, com a
// competência da 602 trocada por comp602 (JSON do campo, ou "" para omiti-lo).
func initCompetencia(comp602 string) string {
	campo := ""
	if comp602 != "" {
		campo = `,"competencia":` + comp602
	}
	return `{"documentos":[
{"tipo":"EXTRATO_APLICACAO_FINANCEIRA","propriedades":{"idPendencia":"1000000000000601","idContaBancaria":7,"competencia":{"mes":8,"ano":2026}}},
{"tipo":"EXTRATO_APLICACAO_FINANCEIRA","propriedades":{"idPendencia":"1000000000000602","idContaBancaria":7` + campo + `}}]}`
}

// Pendência sem competência (ou com mês/ano inválidos) é erro de leitura antes de escrever,
// com --yes e com --dry-run: nunca sai {"mes":0,"ano":0} nem "00/0" (#628, #652).
func TestDocumentosCompetenciaInvalida(t *testing.T) {
	arq := arquivoPDF(t)
	const pathInit602 = "/api/plataforma/documentos/envio-documento/init?id=1000000000000602"
	comandos := map[string][]string{
		"enviar":             {"documentos", "enviar", arq, "--pendencia", "1000000000000602"},
		"enviar consolidado": {"documentos", "enviar", arq, "--pendencia", "1000000000000601", "--pendencia", "1000000000000602"},
		"sem-aplicacao":      {"documentos", "sem-aplicacao", "--conta-bancaria", "7", "--pendencia", "1000000000000602"},
		"sem-aplicacao duas": {"documentos", "sem-aplicacao", "--conta-bancaria", "7", "--pendencia", "1000000000000601,1000000000000602"},
	}
	competencias := map[string]string{
		"ausente":  "",
		"nula":     "null",
		"zerada":   `{"mes":0,"ano":0}`,
		"sem ano":  `{"mes":8}`,
		"mês 13":   `{"mes":13,"ano":2026}`,
		"mês zero": `{"mes":0,"ano":2026}`,
	}
	for nomeComp, comp := range competencias {
		for nomeCmd, args := range comandos {
			for _, modo := range []string{"--yes", "--dry-run"} {
				f := documentoFake(t)
				f.gets[pathInit602] = initCompetencia(comp)
				f.gets[pathEnvioInit2] = initCompetencia(comp)
				out, stderr, code := execCLI(t, "", append(append([]string{}, args...), modo)...)
				if code != ExitError || len(f.writes) != 0 || out != "" || !strings.Contains(stderr, "pendência 1000000000000602") ||
					!strings.Contains(stderr, "competência") || !strings.Contains(stderr, "painel") || strings.Contains(stderr, "00/0") {
					t.Errorf("%s, competência %s, %s: código %d, escritas %q\n%s\n%s", nomeCmd, nomeComp, modo, code, f.writes, out, stderr)
				}
			}
		}
	}
}

// Ids de --pendencia são aparados e deduplicados na ordem; id vazio é erro de uso (#619).
func TestDocumentosPendenciaRepetida(t *testing.T) {
	arq := arquivoPDF(t)
	const id = "1000000000000601"
	for _, ids := range [][]string{{"--pendencia", id, "--pendencia", id}, {"--pendencia", id + "," + id}, {"--pendencia", id + ", " + id + " "}} {
		f := documentoFake(t)
		_, stderr, code := execCLI(t, "", append(append([]string{"documentos", "enviar", arq}, ids...), "--yes")...)
		if code != ExitOK || len(f.writes) != 1 || !strings.HasPrefix(f.writes[0], "POST /api/plataforma/documentos/envio-documento/enviar ") ||
			!strings.Contains(stderr, "para 1 pendência(s) (08/2026)") {
			t.Errorf("enviar %q: código %d, %q\n%s", ids, code, f.writes, stderr)
		}
	}

	// Repetida entre outras: consolidado com cada pendência uma vez, na ordem.
	f := documentoFake(t)
	_, stderr, code := execCLI(t, "", "documentos", "enviar", arq, "--pendencia", id+",1000000000000602,"+id, "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.HasPrefix(f.writes[0], "POST /api/plataforma/documentos/envio-documento/enviar/consolidado ") ||
		strings.Count(f.writes[0], `"idPendencia":"`+id+`"`) != 1 || !strings.Contains(stderr, "para 2 pendência(s) (08/2026, 09/2026)") {
		t.Errorf("consolidado: código %d, %q\n%s", code, f.writes, stderr)
	}

	f = documentoFake(t)
	f.gets["/api/plataforma/documentos/envio-documento/init?id=1000000000000603"] = fixture(t, "envio_documento_init")
	_, stderr, code = execCLI(t, "", "documentos", "sem-arquivo", "--pendencia", "1000000000000603, 1000000000000603", "--yes")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `POST /api/plataforma/documentos/envio-documento/enviar/sem-arquivo [{"idPendencia":"1000000000000603","tipo":"ESTOQUE"}]` ||
		!strings.Contains(stderr, "em 1 pendência(s)") {
		t.Errorf("sem-arquivo: código %d, %q\n%s", code, f.writes, stderr)
	}

	f = documentoFake(t)
	_, stderr, code = execCLI(t, "", "documentos", "sem-aplicacao", "--conta-bancaria", "7", "--pendencia", id, "--pendencia", " "+id, "--yes")
	want := `POST /api/plataforma/upload-documentos/extrato-aplicacao-financeira/enviar/sem-aplicacao-financeira {"tipoDocumento":"EXTRATO_APLICACAO_FINANCEIRA","competenciasPendentes":[{"id":"1000000000000601","mes":8,"ano":2026}],"idContaBancaria":7,"tipoPendencia":"EXTRATO_APLICACAO_FINANCEIRA"}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Errorf("sem-aplicacao: código %d, %q\n%s", code, f.writes, stderr)
	}

	// Id vazio (vírgula sobrando ou só espaços): erro de uso, sem escrever.
	vazios := map[string][]string{
		"enviar":        {"documentos", "enviar", arq, "--pendencia", id + ","},
		"sem-arquivo":   {"documentos", "sem-arquivo", "--pendencia", " "},
		"sem-aplicacao": {"documentos", "sem-aplicacao", "--conta-bancaria", "7", "--pendencia", id + ",,1000000000000602"},
	}
	for nome, args := range vazios {
		f := documentoFake(t)
		_, stderr, code := execCLI(t, "", append(args, "--yes")...)
		if code != ExitUsage || len(f.writes) != 0 || !strings.Contains(stderr, "id vazio") {
			t.Errorf("%s com id vazio: código %d, %q\n%s", nome, code, f.writes, stderr)
		}
	}
}
