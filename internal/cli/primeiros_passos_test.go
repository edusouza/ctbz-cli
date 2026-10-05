package cli

import (
	"net/http"
	"strings"
	"testing"
)

func TestPrimeirosPassos(t *testing.T) {
	withSession(t, fakeAPI(t, map[string]string{pathChecklist: fixture(t, "checklist_onboarding")}))
	out, stderr, code := execCLI(t, "", "primeiros-passos", "-o", "csv")
	want := "etapa,tarefa,situacao,concluida,conclusao_manual\n" +
		"CADASTRO_CONTA_PJ,Abrir a conta PJ,CONCLUIDA,true,false\n" +
		"CADASTRE_PRO_LABORE,Cadastrar o pró-labore,PENDENTE,false,false\n" +
		"GERAR_PROCURACAO_VIRTUAL,Gerar a procuração eletrônica (ctbz pendencias procuracao),PENDENTE,false,false\n" +
		"REUNIAO_BOAS_VINDAS,Reunião de boas-vindas,PENDENTE,false,true\n" +
		"LIVE_EMISSAO_NOTAS_FISCAIS,Live: emissão de notas fiscais,CONCLUIDA,true,false\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s\nquero:\n%s", code, stderr, out, want)
	}
}

func TestPrimeirosPassosConcluir(t *testing.T) {
	f := newEscritaFake(t, nil)
	out, stderr, code := execCLI(t, "", "primeiros-passos", "concluir", "reuniao_boas_vindas", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `POST /api/plataforma/checklist-onboarding/aside/concluir-etapa {"etapa":"REUNIAO_BOAS_VINDAS"}` ||
		!strings.Contains(out, `"situacao": "concluido"`) {
		t.Fatalf("código %d, %q\n%s\n%s", code, f.writes, out, stderr)
	}
	f.status = http.StatusNotModified
	out, stderr, code = execCLI(t, "", "primeiros-passos", "concluir", "LIVE_VIDA_COM_CNPJ", "--yes", "-o", "json")
	if code != ExitOK || !strings.Contains(out, `"situacao": "ja_concluido"`) || !strings.Contains(stderr, "já estava concluída") {
		t.Errorf("304: código %d\n%s\n%s", code, out, stderr)
	}
	for _, args := range [][]string{{"concluir", "CADASTRO_CONTA_PJ"}, {"concluir", "OUTRA"}} {
		if _, _, code := execCLI(t, "", append([]string{"primeiros-passos"}, append(args, "--yes")...)...); code != ExitUsage {
			t.Errorf("%v: código %d", args, code)
		}
	}
}

func TestPrimeirosPassosDispensarReativar(t *testing.T) {
	f := newEscritaFake(t, nil)
	execCLI(t, "", "primeiros-passos", "dispensar", "--yes")
	out, _, code := execCLI(t, "", "primeiros-passos", "reativar", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 2 || f.writes[0] != "PATCH /api/plataforma/checklist-onboarding/exibir-primeiros-passos/dispensar " ||
		f.writes[1] != "PATCH /api/plataforma/checklist-onboarding/reativar-tarefas " || !strings.Contains(out, `"situacao": "reativado"`) {
		t.Errorf("código %d, %q\n%s", code, f.writes, out)
	}
}
