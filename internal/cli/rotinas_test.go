package cli

import (
	"strings"
	"testing"
)

const centralRotinasJSON = `{"pendencias":{"pendenciasCriticas":{},"outrasPendencias":{}},
"rotinas":[
 {"tipo":"IMPORTACAO_EXTRATO","prazo":"2026-09-05","status":"REALIZADA","automatica":false,"propriedades":{"mesReferencia":"agosto"}},
 {"tipo":"IMPORTACAO_EXTRATO","prazo":"2026-10-05","status":"EM_ABERTO","automatica":false,"propriedades":{"mesReferencia":"setembro"}},
 {"tipo":"IMPOSTO","prazo":"2026-10-02","status":"EM_ABERTO","automatica":false,"propriedades":{"tituloModal":"DARF Unificada"}},
 {"tipo":"VENCIMENTO_MENSALIDADE","prazo":"2026-10-15","status":"EM_ABERTO","automatica":false,"propriedades":{"valorPagamento":15.9}},
 {"tipo":"NOVA_ROTINA","prazo":"2026-10-20","status":"REALIZADA","automatica":true,"propriedades":null}],
"rotinasContabilizei":[
 {"titulo":"EFD-Reinf (Escrituração Fiscal Digital) - R2099 (Previdenciário)","prazo":"2026-10-15","status":"EM_ABERTO","conteudo":{"sigla":"EFD-REINF_R2099"}},
 {"titulo":"eSocial (Escrituração Digital)","prazo":"2026-11-15","status":"EM_ABERTO","conteudo":{"sigla":"ESOCIAL"}}]}`

func TestRotinas(t *testing.T) {
	fixNow(t, "2026-10-03")
	withSession(t, fakeAPI(t, map[string]string{"/api/plataforma/dashboard/v2/central-rotinas": centralRotinasJSON}))

	out, stderr, code := execCLI(t, "", "rotinas", "--fail-on-vencidas", "-o", "csv")
	if code != ExitAttention {
		t.Errorf("com rotina vencida, código %d, quero %d: %s", code, ExitAttention, stderr)
	}
	want := "responsavel,rotina,prazo,status,valor,alerta,pendencias\n" +
		"empresa,Importar extrato bancário de setembro,2026-10-05,EM_ABERTO,,próxima,\n" +
		"empresa,DARF Unificada,2026-10-02,EM_ABERTO,,vencida,\n" +
		"empresa,Mensalidade da Contabilizei,2026-10-15,EM_ABERTO,15.90,,\n" +
		"empresa,Nova rotina,2026-10-20,REALIZADA,,,\n" +
		"contabilizei,EFD-Reinf - R2099 (Previdenciário),2026-10-15,EM_ABERTO,,,\n"
	if out != want {
		t.Errorf("CSV:\n%s\nesperado:\n%s", out, want)
	}

	out, _, code = execCLI(t, "", "rotinas", "--mes", "2026-11", "--fail-on-vencidas", "-o", "csv")
	if code != ExitOK || !strings.Contains(out, "contabilizei,eSocial,2026-11-15") || strings.Count(out, "\n") != 2 {
		t.Errorf("--mes 2026-11 (código %d):\n%s", code, out)
	}

	if _, _, code := execCLI(t, "", "rotinas", "--mes", "10/2026"); code != ExitUsage {
		t.Errorf("--mes inválido: código %d, quero %d", code, ExitUsage)
	}
}

func TestRotinasFixture(t *testing.T) {
	fixNow(t, "2026-09-01")
	withSession(t, fakeAPI(t, map[string]string{"/api/plataforma/dashboard/v2/central-rotinas": fixture(t, "central_rotinas")}))
	out, stderr, code := execCLI(t, "", "rotinas", "-o", "json")
	if code != ExitOK {
		t.Fatalf("código %d: %s", code, stderr)
	}
	for _, want := range []string{`"rotina": "Mensalidade da Contabilizei"`, `"valor": 1234.56`, `"rotina": "eSocial"`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON sem %s:\n%s", want, out)
		}
	}
}

func TestRotinasPendencias(t *testing.T) {
	fixNow(t, "2026-10-03")
	withSession(t, fakeAPI(t, map[string]string{"/api/plataforma/dashboard/v2/central-rotinas": `{"pendencias":{},"rotinasContabilizei":[],
"rotinas":[{"tipo":"RECLASSIFICACAO","prazo":"2026-10-10","status":"EM_ABERTO","automatica":false,
"propriedades":{"tituloModal":"Alterar lançamento bancário","documentosPendentes":["p1","p2"]}}]}`}))
	out, stderr, code := execCLI(t, "", "rotinas", "-o", "csv")
	if code != ExitOK || !strings.HasSuffix(out, "empresa,Alterar lançamento bancário,2026-10-10,EM_ABERTO,,próxima,p1; p2\n") {
		t.Errorf("código %d, %s\n%s", code, stderr, out)
	}
}
