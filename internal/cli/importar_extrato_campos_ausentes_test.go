package cli

import (
	"strings"
	"testing"
)

// Campo ausente na resposta de validação nunca afrouxa a checagem em silêncio (ADR-0021, #596):
// a importação para antes de qualquer escrita. Null em statusIntegracao é uma resposta válida.
func TestExtratosImportarRespostaDeValidacaoIncompleta(t *testing.T) {
	ofx, mt := arquivoExtrato(t, "extrato.ofx")
	const permite = "/api/plataforma/movimentacao-financeira/permiteImportacao/7/2026/9"
	for _, tc := range []struct {
		nome, chave, resp, querer string
	}{
		{"temLancamentoPendente ausente", permite,
			`{"periodosDeImportacao":[{"ano":2026,"mes":8,"mensagem":"x"}]}`, "não informou se há importação pendente em 08/2026"},
		{"formatos ausentes", pathUploadInit,
			`{"bancos":[{"id":33,"codigo":"341","contasBancarias":[{"id":7,"statusIntegracao":"NAO_INTEGRADA"}]}]}`, "não informou os formatos"},
		{"statusIntegracao ausente", pathUploadInit,
			`{"bancos":[{"id":33,"codigo":"341","formatosDisponiveisUpload":["OFX"],"contasBancarias":[{"id":7}]}]}`, "não informou se a conta é integrada"},
	} {
		f, _ := importarFake(t, mt, "ofx")
		f.gets[tc.chave] = tc.resp
		_, stderr, code := execCLI(t, "", append(append(baseImportar, ofx), "--yes")...)
		if code != ExitError || !strings.Contains(stderr, tc.querer) || len(f.writes) != 0 {
			t.Errorf("%s: código %d, %d escritas, %q (quero %q)", tc.nome, code, len(f.writes), stderr, tc.querer)
		}
	}
	f, _ := importarFake(t, mt, "ofx")
	f.gets[pathUploadInit] = `{"bancos":[{"id":33,"codigo":"341","formatosDisponiveisUpload":["OFX"],"contasBancarias":[{"id":7,"statusIntegracao":null}]}]}`
	if _, stderr, code := execCLI(t, "", append(append(baseImportar, ofx), "--saldo-final", "1.234,56", "--yes")...); code != ExitOK {
		t.Errorf("statusIntegracao null deve seguir: código %d\n%s", code, stderr)
	}
}

// Lista de formatos vazia (presente) é um banco sem formatos: a mensagem não sai com a lista em branco.
func TestExtratosImportarBancoSemFormatos(t *testing.T) {
	pdf, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	bancoAceita(f, ``)
	_, stderr, code := execCLI(t, "", append(append(baseImportar, pdf), "--yes")...)
	if code != ExitUsage || !strings.Contains(stderr, "não lista nenhum formato de extrato") || strings.Contains(stderr, "em , ") || len(f.writes) != 0 {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), stderr)
	}
}
