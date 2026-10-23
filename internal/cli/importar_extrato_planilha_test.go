package cli

import (
	"strings"
	"testing"
)

const (
	pathUploadInit       = "/api/plataforma/upload-documentos/extrato/v2/init"
	envioCentralDocsPOST = "POST /api/plataforma/upload-documentos/extrato/enviar "
)

// bancoAceita faz o banco da conta 7 declarar os formatos dados.
func bancoAceita(f *escritaFake, formatos string) {
	f.gets[pathUploadInit] = `{"bancos":[{"id":55,"codigo":"077","formatosDisponiveisUpload":[` + formatos +
		`],"contasBancarias":[{"id":7,"statusIntegracao":"NAO_INTEGRADA"}]}]}`
}

func TestExtratosImportarCSVPelaCentralDeDocumentos(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.csv")
	f, _ := importarFake(t, mt, "csv")
	bancoAceita(f, `"PDF","XLS","CSV"`)
	f.onWrite = func(f *escritaFake) { f.gets[pathExtratos] = extratosNaSituacao("PENDENTE") }
	out, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes", "-o", "json")...)
	if code != ExitOK || len(f.writes) != 1 {
		t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	for _, want := range []string{
		`name="documento"; filename="extrato.csv"` + "\r\nContent-Type: text/csv\r\n\r\nOFXHEADER:100",
		"name=\"competencia\"\r\n\r\n09/2026",
		"name=\"idContaBancaria\"\r\n\r\n7",
	} {
		if !strings.HasPrefix(f.writes[0], envioCentralDocsPOST) || !strings.Contains(f.writes[0], want) {
			t.Errorf("envio sem %q:\n%s", want, f.writes[0])
		}
	}
	if !strings.Contains(stderr, "Importar o extrato extrato.csv (CSV, 13 bytes) da conta 7 em 09/2026 pela Central de Documentos (risco médio)") {
		t.Errorf("resumo:\n%s", stderr)
	}
	if !strings.Contains(stderr, "a Contabilizei importa este extrato manualmente, em até 45 dias") {
		t.Errorf("stderr sem o aviso da importação manual:\n%s", stderr)
	}
	for _, want := range []string{`"situacao": "enviado"`, `"importacao": "manual"`, `"arquivo": "extrato.csv"`, `"situacao_extrato": "PENDENTE"`} {
		if !strings.Contains(out, want) {
			t.Errorf("saída sem %s:\n%s", want, out)
		}
	}
}

func TestExtratosImportarXLSDryRun(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.XLS")
	f, _ := importarFake(t, mt, "xls")
	bancoAceita(f, `"PDF","XLS","CSV"`)
	out, _, code := execCLI(t, "", append(baseImportar, arq, "--dry-run", "-o", "json")...)
	if code != ExitOK || len(f.writes) != 0 {
		t.Fatalf("código %d, escritas %q", code, f.writes)
	}
	for _, want := range []string{`"caminho": "/api/plataforma/upload-documentos/extrato/enviar"`, `"valor": "09/2026"`,
		`"valor": "arquivo extrato.XLS (13 bytes)"`} {
		if !strings.Contains(out, want) {
			t.Errorf("simulação sem %s:\n%s", want, out)
		}
	}
	if strings.Contains(out, "eventoUploadExtrato") {
		t.Errorf("CSV/XLS não passam pela conclusão da importação:\n%s", out)
	}
}

func TestExtratosImportarFormatoQueOBancoNaoAceita(t *testing.T) {
	csv, mt := arquivoExtrato(t, "extrato.csv")
	semExtensao, _ := arquivoExtrato(t, "extrato")
	for _, tc := range []struct {
		arq, formatos, want string
		code                int
	}{
		{csv, `"OFX","PDF"`, "o banco desta conta aceita extratos em OFX ou PDF; este arquivo é .csv", ExitUsage},
		{csv, `"PDF","XLS"`, "o banco desta conta aceita extratos em PDF ou XLS; este arquivo é .csv", ExitUsage},
		{semExtensao, `"OFX","PDF"`, "o arquivo do extrato não tem extensão", ExitUsage},
	} {
		f, _ := importarFake(t, mt, "csv")
		bancoAceita(f, tc.formatos)
		_, stderr, code := execCLI(t, "", append(baseImportar, tc.arq, "--yes")...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%s com %s: código %d, %d escritas\n%s(quero %q)", tc.arq, tc.formatos, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestExtratosImportarCSVRecusado(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.csv")
	f, _ := importarFake(t, mt, "csv")
	bancoAceita(f, `"CSV"`)
	f.status = 400
	f.resposta = `"Arquivo inválido"`
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	if code != ExitError || len(f.writes) != 1 || !strings.Contains(stderr, "a Contabilizei recusou o arquivo: ") ||
		!strings.Contains(stderr, "Arquivo inválido") {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), stderr)
	}
}

func TestExtratosImportarCSVSemResposta(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.csv")
	f, _ := importarFake(t, mt, "csv")
	bancoAceita(f, `"CSV"`)
	statusNaEscrita(f, "extrato/enviar", 502)
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	want := "o servidor respondeu com erro (HTTP 502) no envio do arquivo, que pode ou não ter chegado à Contabilizei; " +
		"confira com `ctbz extratos` e na Central de Documentos do painel antes de enviar de novo"
	if code != ExitError || len(f.writes) != 1 || !strings.Contains(stderr, want) {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), stderr)
	}
}

func TestExtratosImportarCSVSaldoFinal(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.csv")
	f, _ := importarFake(t, mt, "csv")
	bancoAceita(f, `"CSV"`)
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--saldo-final", "1", "--yes")...)
	if code != ExitUsage || !strings.Contains(stderr, "--saldo-final só vale para extratos OFX") || len(f.writes) != 0 {
		t.Errorf("código %d\n%s", code, stderr)
	}
}

func TestExtratosImportarSubstituirComCSV(t *testing.T) {
	f, _ := substituirFake(t, "0", true)
	arq, _ := arquivoExtrato(t, "extrato.csv")
	bancoAceita(f, `"PDF","CSV"`)
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...)
	if code != ExitOK || len(f.writes) != 2 || f.writes[0] != "DELETE "+pathExcluirExtrato9+" " || !strings.HasPrefix(f.writes[1], envioCentralDocsPOST) {
		t.Errorf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "pelo arquivo extrato.csv (CSV, 13 bytes), pela Central de Documentos") {
		t.Errorf("resumo:\n%s", stderr)
	}
}
