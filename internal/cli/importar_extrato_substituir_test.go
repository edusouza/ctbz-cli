package cli

import (
	"fmt"
	"strings"
	"testing"
)

const pathExcluirExtrato9 = "/api/plataforma/movimentacao-financeira/extrato?idContaBancaria=7&ano=2026&mes=9"

// extratosNaSituacao é a lista de extratos do fake com a competência 09/2026 na situação dada.
func extratosNaSituacao(situacao string) string {
	return fmt.Sprintf(`[{"ano":2026,"mes":9,"idContaBancaria":7,"banco":"Banco","agencia":"1","numeroConta":"123456","situacao":%q,"statusIntegracao":null}]`, situacao)
}

// substituirFake prepara um extrato já importado com a quantidade de lançamentos dada. Ao
// excluir, a competência volta a PENDENTE; ao concluir a importação, fica IMPORTADO com 3.
func substituirFake(t *testing.T, lancamentos string, permiteExcluir bool) (*escritaFake, string) {
	t.Helper()
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	f.gets[pathExtratos] = extratosNaSituacao("IMPORTADO")
	f.gets[pathInfoExtrato9] = fmt.Sprintf(`{"permiteEditarLancamento":true,"permiteExclusaoExtrato":%t,"quantidadeLancamentos":%s}`, permiteExcluir, lancamentos)
	f.onWrite = func(f *escritaFake) {
		switch last := f.writes[len(f.writes)-1]; {
		case strings.HasPrefix(last, "DELETE "):
			f.gets[pathExtratos] = extratosNaSituacao("PENDENTE")
		case strings.Contains(last, "eventoUploadExtrato"):
			f.gets[pathExtratos] = extratosNaSituacao("IMPORTADO")
			f.gets[pathInfoExtrato9] = `{"permiteEditarLancamento":true,"permiteExclusaoExtrato":true,"quantidadeLancamentos":3}`
		}
	}
	return f, arq
}

func TestExtratosImportarSubstituirVazio(t *testing.T) {
	f, arq := substituirFake(t, "0", true)
	out, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes", "-o", "json")...)
	if code != ExitOK || len(f.writes) != 3 {
		t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
	}
	if f.writes[0] != "DELETE "+pathExcluirExtrato9+" " || !strings.Contains(f.writes[1], "enviar/bucket") || !strings.Contains(f.writes[2], "eventoUploadExtrato") {
		t.Errorf("ordem das escritas: excluir, enviar, concluir:\n%q", f.writes)
	}
	if !strings.Contains(stderr, "Substituir o extrato de 09/2026 da conta 7 (atual: 0 lançamentos) pelo arquivo extrato.pdf (PDF, 13 bytes) (risco médio)") {
		t.Errorf("resumo:\n%s", stderr)
	}
	if !strings.Contains(out, `"substituiu_extrato": true`) || !strings.Contains(out, `"quantidade_lancamentos": 3`) || !strings.Contains(out, `"situacao_extrato": "IMPORTADO"`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestExtratosImportarSubstituirComLancamentos(t *testing.T) {
	f, arq := substituirFake(t, "2", true)
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...)
	if code != ExitOK || len(f.writes) != 3 || !strings.Contains(stderr, "(atual: 2 lançamentos)") || !strings.Contains(stderr, "(risco alto)") ||
		!strings.Contains(stderr, "as classificações e informações adicionais das movimentações se perdem") {
		t.Errorf("código %d, escritas %d\n%s", code, len(f.writes), stderr)
	}
	// Sem terminal e sem --yes, nada é enviado.
	f, arq = substituirFake(t, "2", true)
	if _, _, code := execCLI(t, "", append(baseImportar, arq, "--substituir")...); code != ExitUsage || len(f.writes) != 0 {
		t.Errorf("sem --yes: código %d, escritas %d", code, len(f.writes))
	}
}

func TestExtratosImportarSubstituirDryRun(t *testing.T) {
	f, arq := substituirFake(t, "0", true)
	out, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--dry-run")...)
	if code != ExitOK || len(f.writes) != 0 {
		t.Fatalf("código %d, escritas %d\n%s", code, len(f.writes), stderr)
	}
	excluir, enviar, concluir := strings.Index(out, "movimentacao-financeira/extrato?idContaBancaria=7"), strings.Index(out, "enviar/bucket"), strings.Index(out, "eventoUploadExtrato")
	if excluir < 0 || enviar < excluir || concluir < enviar || !strings.Contains(out, "DELETE") {
		t.Errorf("a simulação deve mostrar excluir, enviar e concluir, nessa ordem:\n%s", out)
	}
}

func TestExtratosImportarSubstituirBloqueado(t *testing.T) {
	f, arq := substituirFake(t, "2", false)
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...)
	if code != ExitError || len(f.writes) != 0 || !strings.Contains(stderr, "não pode ser substituído: os contadores já concluíram a classificação") {
		t.Errorf("código %d, escritas %d\n%s", code, len(f.writes), stderr)
	}
	// Em processamento não há o que substituir ainda.
	f, arq = substituirFake(t, "0", true)
	f.gets[pathExtratos] = extratosNaSituacao("PROCESSANDO")
	if _, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...); code != ExitError || len(f.writes) != 0 || !strings.Contains(stderr, "em processamento") {
		t.Errorf("processando: código %d, escritas %d\n%s", code, len(f.writes), stderr)
	}
}

func TestExtratosImportarSubstituirFalhaDepoisDeExcluir(t *testing.T) {
	f, arq := substituirFake(t, "0", true)
	f.cair = "enviar/bucket"
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...)
	f.mu.Lock()
	escritas := len(f.writes)
	f.mu.Unlock()
	if code != ExitError || escritas != 2 {
		t.Fatalf("código %d, escritas %d\n%s", code, escritas, stderr)
	}
	for _, want := range []string{
		"o extrato antigo de 09/2026 foi excluído e a competência está PENDENTE, mas o novo arquivo não foi importado",
		"o extrato de 09/2026 continua PENDENTE: nada foi importado e é seguro rodar o mesmo comando de novo",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr sem %q:\n%s", want, stderr)
		}
	}
}

func TestExtratosImportarSubstituirSemExtrato(t *testing.T) {
	arq, mt := arquivoExtrato(t, "extrato.pdf")
	f, _ := importarFake(t, mt, "pdf")
	out, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes", "-o", "json")...)
	if code != ExitOK || len(f.writes) != 2 || strings.HasPrefix(f.writes[0], "DELETE") ||
		!strings.Contains(stderr, "Nada a substituir: o extrato de 09/2026 não foi importado") || !strings.Contains(out, `"substituiu_extrato": false`) {
		t.Errorf("código %d, escritas %q\nstderr %s\nsaída %s", code, f.writes, stderr, out)
	}
}

func TestExtratosImportarSubstituirConclusaoSemResposta(t *testing.T) {
	f, arq := substituirFake(t, "0", true)
	f.cair = "eventoUploadExtrato"
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...)
	f.mu.Lock()
	escritas := append([]string(nil), f.writes...)
	f.mu.Unlock()
	if code != ExitError || len(escritas) != 3 || !strings.HasPrefix(escritas[0], "DELETE ") ||
		!strings.Contains(escritas[1], "enviar/bucket") || !strings.Contains(escritas[2], "eventoUploadExtrato") {
		t.Fatalf("código %d, escritas %q\n%s", code, escritas, stderr)
	}
	for _, want := range []string{
		"o extrato antigo de 09/2026 foi excluído e não se sabe se o novo arquivo foi importado (não use --substituir antes de conferir",
		"confira com `ctbz extratos lancamentos --conta-bancaria 7 --competencia 2026-09` antes de repetir",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr sem %q:\n%s", want, stderr)
		}
	}
	for _, nao := range []string{"voltou a ficar pendente", "de novo", "seguro"} {
		if strings.Contains(stderr, nao) {
			t.Errorf("stderr com %q:\n%s", nao, stderr)
		}
	}
}

func TestExtratosImportarSubstituirUploadRecusado(t *testing.T) {
	f, arq := substituirFake(t, "0", true)
	excluir := f.onWrite
	f.onWrite = func(f *escritaFake) {
		excluir(f)
		f.status = 200
		if strings.Contains(f.writes[len(f.writes)-1], "enviar/bucket") {
			f.status = 400
		}
	}
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...)
	want := "o extrato antigo de 09/2026 foi excluído e a competência está PENDENTE, mas o novo arquivo não foi importado " +
		"(rode ctbz extratos importar de novo, sem precisar de --substituir)"
	if code != ExitError || len(f.writes) != 2 || !strings.Contains(stderr, want) {
		t.Errorf("código %d, escritas %d\n%s", code, len(f.writes), stderr)
	}
}

// uploadRecusadoDepoisDeExcluir: DELETE 200, upload 400 e, depois do upload, a competência
// relida na situação dada ("" = não aparece; "erro" = a releitura falha) (#901).
func uploadRecusadoDepoisDeExcluir(t *testing.T, situacao string) (string, int, int) {
	t.Helper()
	f, arq := substituirFake(t, "0", true)
	excluir := f.onWrite
	f.onWrite = func(f *escritaFake) {
		excluir(f)
		f.status = 200
		if strings.Contains(f.writes[len(f.writes)-1], "enviar/bucket") {
			f.status = 400
			switch situacao {
			case "erro":
				delete(f.gets, pathExtratos)
			case "":
				f.gets[pathExtratos] = "[]"
			default:
				f.gets[pathExtratos] = extratosNaSituacao(situacao)
			}
		}
	}
	_, stderr, code := execCLI(t, "", append(baseImportar, arq, "--substituir", "--yes")...)
	return stderr, code, len(f.writes)
}

func TestExtratosImportarSubstituirRecusadoEProcessando(t *testing.T) {
	stderr, code, escritas := uploadRecusadoDepoisDeExcluir(t, "PROCESSANDO")
	want := "o extrato antigo de 09/2026 foi excluído e a competência está PROCESSANDO; espere ela sair desse estado e confira com " +
		"`ctbz extratos` antes de importar de novo (sem --substituir)"
	if code != ExitError || escritas != 2 || !strings.Contains(stderr, want) {
		t.Errorf("código %d, escritas %d\n%s", code, escritas, stderr)
	}
	for _, nao := range []string{"pendente", "PENDENTE", "rode ctbz extratos importar de novo"} {
		if strings.Contains(stderr, nao) {
			t.Errorf("com PROCESSANDO, stderr não pode ter %q:\n%s", nao, stderr)
		}
	}
	if !strings.Contains(stderr, "não reconheceu o arquivo") && !strings.Contains(stderr, "HTTP 400") {
		t.Errorf("stderr sem a causa da recusa:\n%s", stderr)
	}
}

func TestExtratosImportarSubstituirRecusadoOutrasSituacoes(t *testing.T) {
	for _, tc := range []struct{ situacao, want string }{
		{"IMPORTADO", "o extrato antigo de 09/2026 foi excluído, mas a competência aparece como IMPORTADO; confira com " +
			"`ctbz extratos lancamentos --conta-bancaria 7 --competencia 2026-09` antes de importar de novo"},
		{"", "o extrato antigo de 09/2026 foi excluído e a competência não aparece entre os extratos, mas o novo arquivo não foi importado " +
			"(rode ctbz extratos importar de novo, sem precisar de --substituir)"},
		{"erro", "o extrato antigo de 09/2026 foi excluído e o novo arquivo não foi importado; não foi possível reler a competência ("},
	} {
		stderr, code, escritas := uploadRecusadoDepoisDeExcluir(t, tc.situacao)
		if code != ExitError || escritas != 2 || !strings.Contains(stderr, tc.want) {
			t.Errorf("%q: código %d, escritas %d\n%s", tc.situacao, code, escritas, stderr)
		}
	}
}
