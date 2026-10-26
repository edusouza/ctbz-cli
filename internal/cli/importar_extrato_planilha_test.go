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

// O painel importa OFX pela tela "Importar extrato" mesmo quando o banco não o lista (o Inter
// lista PDF, XLS e CSV): a CLI envia com aviso e deixa a Contabilizei validar o arquivo (#1011).
func TestExtratosImportarOFXForaDaLista(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.ofx")
	f, nome := importarFake(t, mt, "ofx")
	bancoAceita(f, `"PDF","XLS","CSV"`)
	out, stderr, code := execCLI(t, "", append(baseImportar, arq, "--saldo-final", "1.234,56", "--yes", "-o", "json")...)
	if code != ExitOK || len(f.writes) != 2 {
		t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	if !strings.HasPrefix(f.writes[0], "POST /api/plataforma/upload-documentos/extrato/enviar/bucket") ||
		!strings.HasPrefix(f.writes[1], "POST /api/plataforma/movimentacao-financeira/eventoUploadExtrato") {
		t.Errorf("OFX fora da lista deve ir pela tela \"Importar extrato\":\n%q", f.writes)
	}
	want := "aviso: o banco desta conta não lista OFX entre os formatos aceitos (PDF, XLS ou CSV), mas o painel importa OFX " +
		"pela tela \"Importar extrato\"; se a Contabilizei recusar o arquivo, a mensagem diz o motivo"
	if !strings.Contains(stderr, want) {
		t.Errorf("stderr sem o aviso:\n%s", stderr)
	}
	for _, w := range []string{`"situacao": "importado"`, `"importacao": "automatica"`, `"arquivo": "` + nome + `"`, `"saldo_ultimo_dia": 1234.56`} {
		if !strings.Contains(out, w) {
			t.Errorf("saída sem %s:\n%s", w, out)
		}
	}
}

// Sem a lista de formatos (#596), o aviso não mostra parênteses vazios.
func TestExtratosImportarOFXSemListaDeFormatos(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.ofx")
	f, _ := importarFake(t, mt, "ofx")
	bancoAceita(f, ``)
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--saldo-final", "1.234,56", "--yes")...)
	want := "aviso: o banco desta conta não lista OFX entre os formatos aceitos, mas o painel importa OFX"
	if code != ExitOK || len(f.writes) != 2 || !strings.Contains(stderr, want) {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), stderr)
	}
}

// O aviso só aparece quando o OFX está fora da lista, e PDF fora da lista continua recusado.
func TestExtratosImportarOFXNaListaSemAviso(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.ofx")
	f, _ := importarFake(t, mt, "ofx")
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--saldo-final", "1.234,56", "--yes")...)
	if code != ExitOK || strings.Contains(stderr, "não lista OFX") {
		t.Errorf("código %d, stderr:\n%s", code, stderr)
	}
	pdf, _ := arquivoExtrato(t, "extrato.pdf")
	f, _ = importarFake(t, mt, "pdf")
	bancoAceita(f, `"XLS","CSV"`)
	_, stderr, code = execCLI(t, "", append(baseImportar, pdf, "--yes")...)
	if code != ExitUsage || len(f.writes) != 0 || !strings.Contains(stderr, "o banco desta conta aceita extratos em XLS ou CSV; este arquivo é .pdf") {
		t.Errorf("pdf fora da lista: código %d, %d escritas\n%s", code, len(f.writes), stderr)
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
		"confira na Central de Documentos do painel antes de enviar de novo (`ctbz extratos` continua PENDENTE mesmo depois de um envio manual)"
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
	// Recusado antes de qualquer escrita, com o extrato vazio ou com lançamentos (#920).
	for _, lancamentos := range []string{"0", "2"} {
		f, _ := substituirFake(t, lancamentos, true)
		arq, _ := arquivoExtrato(t, "extrato.csv")
		bancoAceita(f, `"PDF","CSV"`)
		_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...)
		if code != ExitUsage || len(f.writes) != 0 || !strings.Contains(stderr, "--substituir só vale para OFX e PDF") ||
			!strings.Contains(stderr, "em até 45 dias") || !strings.Contains(stderr, "ctbz extratos excluir") {
			t.Errorf("%s lançamentos: código %d, escritas %q\n%s", lancamentos, code, f.writes, stderr)
		}
	}
}

func TestExtratosImportarCSVSobreExtratoImportado(t *testing.T) {
	for _, situacao := range []string{"IMPORTADO", "PROCESSANDO"} {
		f, _ := substituirFake(t, "2", true)
		f.gets[pathExtratos] = extratosNaSituacao(situacao)
		arq, _ := arquivoExtrato(t, "extrato.xls")
		bancoAceita(f, `"PDF","XLS"`)
		_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
		// Recusa justificada (#921): não sugere --substituir, que o envio manual não aceita.
		if code != ExitError || len(f.writes) != 0 || strings.Contains(stderr, "--substituir") {
			t.Errorf("%s: código %d, escritas %q\n%s", situacao, code, f.writes, stderr)
		}
	}
	f, _ := substituirFake(t, "2", true)
	arq, _ := arquivoExtrato(t, "extrato.xls")
	bancoAceita(f, `"PDF","XLS"`)
	_, stderr, _ := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	if !strings.Contains(stderr, "poderia duplicar os lançamentos") || !strings.Contains(stderr, "Central de Documentos do painel") {
		t.Errorf("recusa sem a justificativa:\n%s", stderr)
	}
}

func TestExtratosImportarCSVReenvioAvisa(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.csv")
	f, _ := importarFake(t, mt, "csv")
	bancoAceita(f, `"CSV"`)
	f.gets[pathExtratos] = extratosNaSituacao("PENDENTE")
	aviso := "a competência continua PENDENTE em `ctbz extratos`. Se você já enviou este extrato, não envie de novo: " +
		"confira na Central de Documentos do painel"
	for envio := 1; envio <= 2; envio++ {
		_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
		if code != ExitOK || len(f.writes) != envio || !strings.Contains(stderr, aviso) ||
			!strings.Contains(stderr, "o que é o esperado: não envie o arquivo de novo") || strings.Contains(stderr, "Acompanhe com `ctbz extratos`") {
			t.Errorf("envio %d: código %d, escritas %d\n%s", envio, code, len(f.writes), stderr)
		}
	}
	// No --dry-run o aviso também aparece, antes de qualquer envio.
	if _, stderr, _ := execCLI(t, "", append(baseImportar, arq, "--dry-run")...); !strings.Contains(stderr, aviso) {
		t.Errorf("dry-run sem o aviso:\n%s", stderr)
	}
}
