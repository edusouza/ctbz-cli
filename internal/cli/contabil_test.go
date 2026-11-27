package cli

import (
	"strings"
	"testing"
	"time"
)

func TestRazaoDataEmMeiaNoiteUTC(t *testing.T) {
	// #1048: a API manda a data como epoch à meia-noite UTC; num fuso a oeste, o dia local seria o anterior.
	antes := time.Local
	time.Local = time.FixedZone("BRT", -3*3600)
	t.Cleanup(func() { time.Local = antes })
	fixNow(t, "2026-10-03")
	julho := `[{"id":"1.01.01.01.00","descricao":"Caixa Geral","nivel":5,"listaLancamento":[
{"id":1,"data":1784592000000,"descricao":"Capital social","debito":1000,"credito":null,"saldoExercicio":1000,"contaContrapartida":null}]}]`
	withSession(t, fakeAPI(t, map[string]string{"/api/plataforma/relatorios-ms/gerarrazaotipoa/2026/7": julho}))
	out, stderr, code := execCLI(t, "", "razao", "--de", "2026-07", "-o", "csv")
	if code != ExitOK || !strings.Contains(out, "2026-07-21,") {
		t.Errorf("código %d: %s\n%s", code, stderr, out)
	}
}

func TestBalancete(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/relatorios-ms/gerar-balancete/2026/8": `[
{"id":"1","idContaPai":null,"descricao":"ATIVO","nivel":1,"tipo":"S","saldoAnterior":0,"totalDebito":1815,"totalCredito":1419.83,"saldoExercicio":395.17},
{"id":"1.01.01.01.00","idContaPai":"1.01.01.01","descricao":"Caixa Geral","nivel":5,"tipo":"A","saldoAnterior":0,"totalDebito":1139,"totalCredito":1280.83,"saldoExercicio":-141.83}]`,
		"/api/plataforma/relatorios-ms/gerar-balancete/2026/9": fixture(t, "balancete"),
	}))
	out, stderr, code := execCLI(t, "", "balancete", "2026-08", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	want := "conta,descricao,nivel,saldo_anterior,debitos,creditos,saldo\n" +
		"1,ATIVO,1,0.00,1815.00,1419.83,395.17\n" +
		"1.01.01.01.00,Caixa Geral,5,0.00,1139.00,1280.83,-141.83\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}
	out, _, _ = execCLI(t, "", "balancete", "2026-08")
	if !strings.Contains(out, "        Caixa Geral") { // nível 5: quatro recuos de dois espaços
		t.Errorf("tabela deveria recuar a conta de nível 5:\n%s", out)
	}
	if out, _, code := execCLI(t, "", "balancete", "2026-09", "-o", "json"); code != ExitOK || !strings.Contains(out, `"descricao": "CIRCULANTE"`) {
		t.Errorf("fixture (código %d):\n%s", code, out)
	}
	if _, _, code := execCLI(t, "", "balancete", "08/2026"); code != ExitUsage {
		t.Errorf("mês inválido: código %d, quero %d", code, ExitUsage)
	}
}

func TestBalanco(t *testing.T) {
	balanco := `[
{"id":"1","descricao":"ATIVO","nivel":1,"classificacaoConta":"ATIVO","saldoExercicio":395.17,"saldoExercicioAnterior":0},
{"id":"2","descricao":"PASSIVO","nivel":1,"classificacaoConta":"PASSIVO","saldoExercicio":-395.17,"saldoExercicioAnterior":0},
{"id":"3","descricao":"RECEITAS","nivel":1,"classificacaoConta":"RESULTADO","saldoExercicio":100,"saldoExercicioAnterior":0}]`
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/relatorios-ms/gerarbalanco/2025/12": balanco,
		"/api/plataforma/relatorios-ms/gerarbalanco/2026/9":  fixture(t, "balanco"),
	}))
	out, stderr, code := execCLI(t, "", "balanco", "2025", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	want := "conta,descricao,nivel,grupo,saldo,saldo_exercicio_anterior\n" +
		"1,ATIVO,1,ATIVO,395.17,0.00\n" +
		"2,PASSIVO,1,PASSIVO,-395.17,0.00\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}
	if out, _, code := execCLI(t, "", "balanco", "2026-09", "-o", "json"); code != ExitOK || !strings.Contains(out, `"grupo": "ATIVO"`) {
		t.Errorf("fixture (código %d):\n%s", code, out)
	}
	if _, _, code := execCLI(t, "", "balanco", "set/2025"); code != ExitUsage {
		t.Errorf("período inválido: código %d, quero %d", code, ExitUsage)
	}
}

func TestRazao(t *testing.T) {
	fixNow(t, "2026-10-03")
	agosto := `[{"id":"1.01.01.01.00","descricao":"Caixa Geral","nivel":5,"listaLancamento":[
{"id":1,"data":1784635200000,"descricao":"Capital social","debito":1000,"credito":null,"saldoExercicio":1000,
 "contaContrapartida":{"id":"2.07.01.01.00","descricao":"Capital Social Realizado no País"}}]},
{"id":"2.01.01","descricao":"Fornecedores","nivel":3,"listaLancamento":[
{"id":2,"data":1786536000000,"descricao":"Pagamento","debito":null,"credito":50.5,"saldoExercicio":-50.5,"contaContrapartida":null}]}]`
	withSession(t, fakeAPI(t, map[string]string{
		"/api/plataforma/relatorios-ms/gerarrazaotipoa/2026/8": agosto,
		"/api/plataforma/relatorios-ms/gerarrazaotipoa/2026/9": fixture(t, "razao"),
	}))
	out, stderr, code := execCLI(t, "", "razao", "--de", "2026-08", "--ate", "2026-08", "--conta", "1.01", "-o", "csv")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	want := "data,conta,conta_descricao,historico,contrapartida,debito,credito,saldo\n" +
		"2026-07-21,1.01.01.01.00,Caixa Geral,Capital social,2.07.01.01.00 Capital Social Realizado no País,1000.00,,1000.00\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}
	out, _, code = execCLI(t, "", "razao", "--de", "2026-08", "--ate", "2026-09", "-o", "json")
	if code != ExitOK || !strings.Contains(out, `"conta": "2.01.01"`) || !strings.Contains(out, "FULANO DE TAL") {
		t.Errorf("dois meses (código %d):\n%s", code, out)
	}
}
