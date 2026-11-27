package cli

import (
	"strings"
	"testing"
)

// --substituir com OFX fora da lista do banco exclui o extrato antes de um envio que a própria
// CLI avisa poder ser recusado: é recusado com código 2, sem nenhuma escrita, também no
// --dry-run (#1016). Sem --substituir, o OFX segue com o aviso.
func TestExtratosImportarSubstituirOFXForaDaLista(t *testing.T) {
	for _, flags := range [][]string{{"--yes"}, {"--dry-run"}, {"--yes", "--saldo-final", "1.234,56"}} {
		f, _ := substituirFake(t, "2", true)
		ofx, _ := arquivoExtrato(t, "extrato.ofx")
		bancoAceita(f, `"PDF","XLS","CSV"`)
		_, stderr, code := execCLI(t, "", append(append(baseImportar, ofx, "--substituir"), flags...)...)
		if code != ExitUsage || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %q\n%s", flags, code, f.writes, stderr)
		}
		for _, want := range []string{"--substituir", "não lista OFX", "ctbz extratos excluir", "sem --substituir"} {
			if !strings.Contains(stderr, want) {
				t.Errorf("%v: erro sem %q:\n%s", flags, want, stderr)
			}
		}
	}
	// Banco que lista OFX: --substituir segue normalmente.
	f, _ := substituirFake(t, "2", true)
	ofx, _ := arquivoExtrato(t, "extrato.ofx")
	bancoAceita(f, `"OFX","PDF"`)
	if _, stderr, code := execCLI(t, "", append(baseImportar, ofx, "--substituir", "--dry-run")...); code != ExitOK || strings.Contains(stderr, "não lista OFX") {
		t.Errorf("OFX listado: código %d\n%s", code, stderr)
	}
}
