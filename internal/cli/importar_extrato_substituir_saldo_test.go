package cli

import (
	"os"
	"strings"
	"testing"
	"time"
)

const ofxComSaldo = `OFXHEADER:100
DATA:OFXSGML
<OFX><BANKMSGSRSV1><STMTTRNRS><STMTRS><BANKTRANLIST></BANKTRANLIST>
<LEDGERBAL><BALAMT>1234.56<DTASOF>20260930</LEDGERBAL>
</STMTRS></STMTTRNRS></BANKMSGSRSV1></OFX>
`

// ofxSubstituir cria um OFX com o conteúdo dado, para um extrato já importado com lançamentos.
func ofxSubstituir(t *testing.T, conteudo string) (*escritaFake, string) {
	t.Helper()
	arq, mt := arquivoExtrato(t, "extrato.ofx")
	if err := os.WriteFile(arq, []byte(conteudo), 0o600); err != nil {
		t.Fatal(err)
	}
	d := time.UnixMilli(mt)
	os.Chtimes(arq, d, d)
	f, _ := importarFake(t, mt, "ofx")
	f.gets[pathExtratos] = extratosNaSituacao("IMPORTADO")
	f.gets[pathInfoExtrato9] = `{"permiteEditarLancamento":true,"permiteExclusaoExtrato":true,"quantidadeLancamentos":3}`
	return f, arq
}

// Com --substituir, o saldo do próprio OFX é conferido com --saldo-final antes do DELETE: se
// diferem, nada é excluído nem enviado (#880).
func TestExtratosImportarSubstituirSaldoDoOFXAntesDeExcluir(t *testing.T) {
	f, arq := ofxSubstituir(t, ofxComSaldo)
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--saldo-final", "100,00", "--yes")...)
	if code != ExitUsage || len(f.writes) != 0 {
		t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	for _, want := range []string{"R$ 1.234,56", "R$ 100,00", "nada foi excluído", "ctbz extratos excluir"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr sem %q:\n%s", want, stderr)
		}
	}
	// Igual: segue (dry-run só mostra as requisições). Sem saldo legível no arquivo, também segue.
	for _, c := range []struct{ conteudo, saldo string }{{ofxComSaldo, "1.234,56"}, {"OFXHEADER:100", "100,00"}} {
		f, arq := ofxSubstituir(t, c.conteudo)
		if _, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--saldo-final", c.saldo, "--dry-run")...); code != ExitOK || len(f.writes) != 0 {
			t.Errorf("%q: código %d\n%s", c.saldo, code, stderr)
		}
	}
}

// A confirmação diz que a exclusão vem antes do envio e que, se o arquivo ou o saldo forem
// recusados, o mês fica sem extrato (#880), com o extrato atual vazio ou com lançamentos.
func TestExtratosImportarSubstituirAvisaQueRecusaDeixaMesVazio(t *testing.T) {
	for _, lanc := range []string{"0", "2"} {
		f, arq := substituirFake(t, lanc, true)
		_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--dry-run")...)
		if code != ExitOK || len(f.writes) != 0 {
			t.Fatalf("%s lançamentos: código %d\n%s", lanc, code, stderr)
		}
		want := "a exclusão vem antes do envio: se a Contabilizei recusar o arquivo ou o saldo não for confirmado, o mês fica sem extrato"
		if !strings.Contains(stderr, want) {
			t.Errorf("%s lançamentos: stderr sem %q:\n%s", lanc, want, stderr)
		}
	}
}
