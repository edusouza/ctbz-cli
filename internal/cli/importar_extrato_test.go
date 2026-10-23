package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// arquivoExtrato cria um arquivo com data de modificação fixa (entra no nome enviado).
func arquivoExtrato(t *testing.T, nome string) (string, int64) {
	t.Helper()
	p := filepath.Join(t.TempDir(), nome)
	if err := os.WriteFile(p, []byte("OFXHEADER:100"), 0o600); err != nil {
		t.Fatal(err)
	}
	mt := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	os.Chtimes(p, mt, mt)
	return p, mt.UnixMilli()
}

func importarFake(t *testing.T, mtime int64, ext string) (*escritaFake, string) {
	t.Helper()
	old := now
	now = func() time.Time { return time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = old })
	nome := fmt.Sprintf("00000000000000_2026_9_123456_%d.%s", mtime, ext)
	f := newEscritaFake(t, map[string]string{
		"/api/plataforma/upload-documentos/extrato/v2/init":                     fixture(t, "upload_extrato_init"),
		"/api/plataforma/movimentacao-financeira/permiteImportacao/7/2026/9":    fixture(t, "permite_importacao"),
		"/api/plataforma/contabancaria/list":                                    `{"contasBancarias":[{"id":7,"nomeBanco":"Banco","agencia":"1","contaCorrente":"12345-6","dataSaldoInicial":0}]}`,
		"/api/plataforma/movimentacao-financeira/info-extrato/7/2026/9/" + nome: fixture(t, "info_extrato"),
		"/api/plataforma/movimentacao-financeira/v2/extratos":                   `[{"ano":2026,"mes":9,"idContaBancaria":7,"banco":"Banco","agencia":"1","numeroConta":"123456","situacao":"ABERTO","statusIntegracao":null}]`,
	})
	return f, nome
}

var baseImportar = []string{"extratos", "importar", "--conta-bancaria", "7", "--competencia", "2026-09"}

func TestExtratosImportarOFX(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.ofx")
	f, nome := importarFake(t, mt, "ofx")
	out, stderr, code := execCLI(t, "", append(append(baseImportar, arq), "--saldo-final", "1.234,56", "--yes", "-o", "json")...)
	if code != ExitOK || len(f.writes) != 2 {
		t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	for _, w := range []string{"POST /api/plataforma/upload-documentos/extrato/enviar/bucket", "name=\"nomeArquivo\"\r\n\r\n" + nome, "name=\"idConta\"\r\n\r\n7", "OFXHEADER:100"} {
		if !strings.Contains(f.writes[0], w) {
			t.Errorf("upload sem %q:\n%s", w, f.writes[0])
		}
	}
	want := `POST /api/plataforma/movimentacao-financeira/eventoUploadExtrato {"ano":2026,"mes":9,"cnpj":"00000000000000","idContaBancaria":7,"nomeArquivoStorage":"` + nome +
		`","respostas":[{"tipo":"SALDO_ULTIMO_MES","data":"2026-10-04T10:00:00.000Z","valor":1234.56}]}`
	if f.writes[1] != want {
		t.Errorf("evento:\n%s\nquero:\n%s", f.writes[1], want)
	}
	if !strings.Contains(out, `"situacao": "importado"`) || !strings.Contains(out, `"situacao_extrato": "ABERTO"`) || !strings.Contains(out, `"saldo_ultimo_dia": 1234.56`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestExtratosImportarSaldoDiferente(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.ofx")
	f, _ := importarFake(t, mt, "ofx")
	_, stderr, code := execCLI(t, "", append(append(baseImportar, arq), "--saldo-final", "1000", "--yes")...)
	if code != ExitError || len(f.writes) != 1 || !strings.Contains(stderr, "diferente de --saldo-final R$ 1.000,00 (importação não concluída)") {
		t.Errorf("código %d, %d escritas, %s", code, len(f.writes), stderr)
	}
}

func TestExtratosImportarTerminal(t *testing.T) {
	old := isTerminal
	isTerminal = func(io.Reader) bool { return true }
	t.Cleanup(func() { isTerminal = old })
	arq, mt := arquivoExtrato(t, "extrato.ofx")
	f, _ := importarFake(t, mt, "ofx")
	_, stderr, code := execCLI(t, "s\ns\n", append(baseImportar, arq)...)
	if code != ExitOK || len(f.writes) != 2 || !strings.Contains(stderr, "Saldo do último dia lido pela Contabilizei: R$ 1.234,56. Confere com o extrato? [s/N]") {
		t.Errorf("aceito: código %d, %d escritas, %s", code, len(f.writes), stderr)
	}
	f, _ = importarFake(t, mt, "ofx")
	_, stderr, code = execCLI(t, "s\nn\n", append(baseImportar, arq)...)
	if code != ExitError || len(f.writes) != 1 || !strings.Contains(stderr, "saldo não confirmado") {
		t.Errorf("recusado: código %d, %d escritas, %s", code, len(f.writes), stderr)
	}
}

func TestExtratosImportarPDFEDryRun(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	out, _, code := execCLI(t, "", append(append(baseImportar, arq), "--dry-run")...)
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, "documento: arquivo extrato.pdf (13 bytes)") || !strings.Contains(out, `"respostas": []`) {
		t.Errorf("dry-run: código %d\n%s", code, out)
	}
	_, _, code = execCLI(t, "", append(append(baseImportar, arq), "--yes")...)
	if code != ExitOK || len(f.writes) != 2 || !strings.HasSuffix(f.writes[1], `"respostas":[]}`) {
		t.Errorf("pdf: código %d, %q", code, f.writes)
	}
}

func TestExtratosImportarValidacoes(t *testing.T) {
	ofx, mt := arquivoExtrato(t, "extrato.ofx")
	txt, _ := arquivoExtrato(t, "extrato.txt")
	pdf, _ := arquivoExtrato(t, "extrato.pdf")
	for _, tc := range []struct {
		args []string
		code int
		want string
		prep func(*escritaFake)
	}{
		{[]string{txt}, ExitUsage, ".ofx ou .pdf", nil},
		{[]string{pdf, "--saldo-final", "1"}, ExitUsage, "só vale para extratos OFX", nil},
		{[]string{ofx, "--conta-bancaria", "8"}, ExitError, "a conta é integrada", nil},
		{[]string{pdf, "--conta-bancaria", "9"}, ExitUsage, "aceita extratos em OFX, não PDF", nil},
		{[]string{ofx}, ExitError, "há importação pendente em 08/2026", func(f *escritaFake) {
			f.gets["/api/plataforma/movimentacao-financeira/permiteImportacao/7/2026/9"] = `{"periodosDeImportacao":[{"ano":2026,"mes":8,"mensagem":"Importe agosto","temLancamentoPendente":true}]}`
		}},
	} {
		f, _ := importarFake(t, mt, "ofx")
		if tc.prep != nil {
			tc.prep(f)
		}
		_, stderr, code := execCLI(t, "", append(append(baseImportar, tc.args...), "--yes")...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestExtratosImportarErroConhecido(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.ofx")
	f, _ := importarFake(t, mt, "ofx")
	f.status = 400
	f.resposta = `{"detalhes":[{"identificador":"exception/movimentacao-financeira-903","detalhe":"{\"conta\":\"1\"}"}]}`
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	if code != ExitError || len(f.writes) != 1 || !strings.Contains(stderr, "o extrato não corresponde à conta escolhida") {
		t.Errorf("código %d, %s", code, stderr)
	}
}

func TestExtratosExcluir(t *testing.T) {
	f := extratoFake(t)
	f.gets["/api/plataforma/movimentacao-financeira/v2/extratos"] = `[{"ano":2026,"mes":9,"idContaBancaria":7,"banco":"B","agencia":"1","numeroConta":"1","situacao":"PENDENTE","statusIntegracao":null}]`
	out, stderr, code := execCLI(t, "", "extratos", "excluir", "--conta-bancaria", "7", "--competencia", "2026-09", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "DELETE /api/plataforma/movimentacao-financeira/extrato?idContaBancaria=7&ano=2026&mes=9 " {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, `"situacao_extrato": "PENDENTE"`) || !strings.Contains(stderr, "(risco alto)") || !strings.Contains(stderr, "importação volta a ficar pendente") {
		t.Errorf("saída:\n%s\n%s", out, stderr)
	}
	f = extratoFake(t)
	f.gets[pathInfoExtrato] = `{"permiteEditarLancamento":true,"permiteExclusaoExtrato":false}`
	if _, stderr, code := execCLI(t, "", "extratos", "excluir", "--conta-bancaria", "7", "--competencia", "2026-09", "--yes"); code != ExitError || len(f.writes) != 0 || !strings.Contains(stderr, "já concluíram a classificação") {
		t.Errorf("bloqueado: código %d, %s", code, stderr)
	}
}

// semRespostaNoEnvio faz a escrita cujo caminho contém trecho estourar o prazo do cliente.
func semRespostaNoEnvio(t *testing.T, f *escritaFake, trecho string) {
	t.Helper()
	old := timeoutUpload
	timeoutUpload = 50 * time.Millisecond
	t.Cleanup(func() { timeoutUpload = old })
	f.atrasar, f.atraso = trecho, 300*time.Millisecond
}

func TestExtratosImportarTimeoutNoUploadPendente(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	f.gets["/api/plataforma/movimentacao-financeira/v2/extratos"] = strings.Replace(f.gets["/api/plataforma/movimentacao-financeira/v2/extratos"], "ABERTO", "PENDENTE", 1)
	semRespostaNoEnvio(t, f, "enviar/bucket")
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	if code != ExitError || len(f.writes) != 1 || !strings.Contains(f.writes[0], "enviar/bucket") {
		t.Fatalf("código %d, escritas %d\n%s", code, len(f.writes), stderr)
	}
	for _, w := range []string{"o extrato de 09/2026 continua PENDENTE: nada foi importado e é seguro rodar o mesmo comando de novo", "Client.Timeout"} {
		if !strings.Contains(stderr, w) {
			t.Errorf("stderr sem %q:\n%s", w, stderr)
		}
	}
}

func TestExtratosImportarTimeoutNoUploadImportado(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	// O servidor aplicou o envio, mas a resposta não chegou.
	f.onWrite = func(f *escritaFake) {
		f.gets[pathExtratos] = strings.Replace(f.gets[pathExtratos], "ABERTO", "IMPORTADO", 1)
	}
	semRespostaNoEnvio(t, f, "enviar/bucket")
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	want := "o extrato de 09/2026 aparece como IMPORTADO; confira com `ctbz extratos lancamentos --conta-bancaria 7 --competencia 2026-09` antes de repetir"
	if code != ExitError || len(f.writes) != 1 || !strings.Contains(stderr, want) || strings.Contains(stderr, "seguro") {
		t.Errorf("código %d, escritas %d\n%s", code, len(f.writes), stderr)
	}
}

func TestExtratosImportarTimeoutNoUploadSemReleitura(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	f.onWrite = func(f *escritaFake) { delete(f.gets, pathExtratos) }
	semRespostaNoEnvio(t, f, "enviar/bucket")
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	if code != ExitError || !strings.Contains(stderr, "pode ou não ter sido aplicada; confira antes de repetir") || !strings.Contains(stderr, "não foi possível reler a situação do extrato") {
		t.Errorf("código %d\n%s", code, stderr)
	}
}

func TestExtratosImportarConexaoCaiAoConcluir(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	f.cair = "eventoUploadExtrato"
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	want := "o arquivo foi enviado, mas a conexão falhou ao concluir a importação, que pode ou não ter sido aplicada; " +
		"confira com `ctbz extratos lancamentos --conta-bancaria 7 --competencia 2026-09` antes de repetir"
	// Sem resposta, nenhuma requisição posterior sincroniza com o servidor: lê sob a trava.
	f.mu.Lock()
	escritas := len(f.writes)
	f.mu.Unlock()
	if code != ExitError || escritas != 2 || !strings.Contains(stderr, want) || strings.Contains(stderr, "não foi concluída") {
		t.Errorf("código %d, escritas %d\n%s", code, escritas, stderr)
	}
}

const (
	pathExtratos     = "/api/plataforma/movimentacao-financeira/v2/extratos"
	pathInfoExtrato9 = "/api/plataforma/movimentacao-financeira/extrato/info?idContaBancaria=7&ano=2026&mes=9"
)

// comSituacao troca a situação da competência 09/2026 na lista de extratos do fake.
func comSituacao(f *escritaFake, situacao string) {
	f.gets[pathExtratos] = strings.Replace(f.gets[pathExtratos], `"situacao":"ABERTO"`, `"situacao":"`+situacao+`"`, 1)
}

func TestExtratosImportarCompetenciaJaImportada(t *testing.T) {
	for _, tc := range []struct {
		nome, situacao, info, want string
	}{
		{"vazia", "IMPORTADO", `{"permiteEditarLancamento":true,"quantidadeLancamentos":0}`,
			"o extrato de 09/2026 já aparece como IMPORTADO, mas sem nenhum lançamento (o processamento do extrato de outro mês pode tê-lo criado vazio); " +
				"para trocá-lo por este arquivo, rode o mesmo comando com --substituir"},
		{"com lançamentos", "IMPORTADO", `{"permiteEditarLancamento":true,"quantidadeLancamentos":2}`,
			"o extrato de 09/2026 já foi importado (veja `ctbz extratos lancamentos --conta-bancaria 7 --competencia 2026-09`); " +
				"para trocá-lo por este arquivo, rode o mesmo comando com --substituir (risco alto"},
		{"sem contagem", "IMPORTADO", `{"permiteEditarLancamento":true}`, "o extrato de 09/2026 já foi importado"},
		{"processando", "PROCESSANDO", "", "o extrato de 09/2026 ainda está em processamento"},
	} {
		t.Run(tc.nome, func(t *testing.T) {
			arq, mt := arquivoExtrato(t, "extrato.pdf")
			f, _ := importarFake(t, mt, "pdf")
			comSituacao(f, tc.situacao)
			if tc.info != "" {
				f.gets[pathInfoExtrato9] = tc.info
			}
			// Vale também no --dry-run: nada seria enviado.
			for _, extra := range []string{"--yes", "--dry-run"} {
				_, stderr, code := execCLI(t, "", append(baseImportar, arq, extra)...)
				if code != ExitError || len(f.writes) != 0 || !strings.Contains(stderr, tc.want) {
					t.Errorf("%s: código %d, escritas %d\n%s", extra, code, len(f.writes), stderr)
				}
			}
		})
	}
}

func TestExtratosImportarTerminaSemLancamentos(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	f.gets[pathInfoExtrato9] = `{"permiteEditarLancamento":true,"quantidadeLancamentos":0}`
	f.onWrite = func(f *escritaFake) { comSituacao(f, "IMPORTADO") }
	out, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes", "-o", "json")...)
	if code != ExitOK || len(f.writes) != 2 || !strings.Contains(stderr, "aviso: o extrato de 09/2026 ficou IMPORTADO sem nenhum lançamento") ||
		!strings.Contains(out, `"quantidade_lancamentos": 0`) || !strings.Contains(out, `"situacao_extrato": "IMPORTADO"`) {
		t.Errorf("código %d, escritas %d\nstderr %s\nsaída %s", code, len(f.writes), stderr, out)
	}
	f.gets[pathInfoExtrato9] = `{"permiteEditarLancamento":true,"quantidadeLancamentos":2}`
	// Volta ao estado anterior para importar de novo.
	f.gets[pathExtratos] = strings.Replace(f.gets[pathExtratos], `"situacao":"IMPORTADO"`, `"situacao":"ABERTO"`, 1)
	out, stderr, _ = execCLI(t, "", append(baseImportar, arq, "--yes", "-o", "json")...)
	if strings.Contains(stderr, "sem nenhum lançamento") || !strings.Contains(out, `"quantidade_lancamentos": 2`) {
		t.Errorf("com lançamentos:\nstderr %s\nsaída %s", stderr, out)
	}
}

func TestExtratosImportarTerminaProcessando(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	f.onWrite = func(f *escritaFake) { comSituacao(f, "PROCESSANDO") }
	out, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes", "-o", "json")...)
	if code != ExitOK || !strings.Contains(stderr, "A Contabilizei está processando o extrato; em alguns minutos confira") ||
		!strings.Contains(out, `"quantidade_lancamentos": null`) {
		t.Errorf("código %d\nstderr %s\nsaída %s", code, stderr, out)
	}
}

func TestExtratosImportarArquivoNaoReconhecido(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	f.status = 400
	f.resposta = `{"mensagem":"(ERRO-NA-VALIDACAO-DE-DOCUMENTO-API-CONTABIL) arquivo inválido"}`
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--yes")...)
	if code != ExitError || !strings.Contains(stderr, "não reconheceu o arquivo como extrato bancário; tente o OFX do banco ou o PDF baixado direto do internet banking") {
		t.Errorf("código %d\n%s", code, stderr)
	}
}
