package cli

import (
	"strings"
	"testing"
)

const pathReclassInit = "/api/plataforma/documentos/reclassificar/init?id=1000000000000501"

func reclassificarFake(t *testing.T) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{
		pathReclassInit: fixture(t, "reclassificar_init"),
		"/api/plataforma/dashboard/v2/central-rotinas": `{"pendencias":{},"rotinasContabilizei":[],"rotinas":[{"tipo":"RECLASSIFICACAO","prazo":"2026-10-10",
"status":"EM_ABERTO","automatica":false,"propriedades":{"documentosPendentes":["1000000000000501"]}}]}`,
	})
	return f
}

func TestRotinasReclassificarListaOpcoes(t *testing.T) {
	f := reclassificarFake(t)
	out, stderr, code := execCLI(t, "", "rotinas", "reclassificar", "1000000000000501", "-o", "csv")
	want := "id_pendencia,descricao,valor,classificacao_atual,id_classificacao,classificacao,socios\n" +
		"1000000000000501,TED RECEBIDA BANCO,50000.00,Empréstimos e Financiamentos,1000000000000041,Empréstimos e Financiamentos,\n" +
		"1000000000000501,TED RECEBIDA BANCO,50000.00,Empréstimos e Financiamentos,1000000000000042,Aporte de Capital,1000000000000031 FULANO DE TAL\n" +
		"1000000000000501,TED RECEBIDA BANCO,50000.00,Empréstimos e Financiamentos,1000000000000043,Mútuo de Sócio,1000000000000031 FULANO DE TAL; 1000000000000032 BELTRANO DE TAL\n"
	if code != ExitOK || out != want || len(f.writes) != 0 {
		t.Errorf("código %d, %s\n%s\nquero:\n%s", code, stderr, out, want)
	}
}

func TestRotinasReclassificar(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		corpo string
	}{
		{"sem sócio", []string{"--classificacao", "Empréstimos e Financiamentos"},
			`[{"idPendencia":"1000000000000501","idClassificacao":1000000000000041,"idSocio":null,"idLancamento":1000000000000102}]`},
		{"sócio único escolhido sozinho", []string{"--classificacao", "1000000000000042"},
			`[{"idPendencia":"1000000000000501","idClassificacao":1000000000000042,"idSocio":1000000000000031,"idLancamento":1000000000000102}]`},
		{"sócio informado", []string{"--classificacao", "mútuo de sócio", "--socio", "1000000000000032"},
			`"idSocio":1000000000000032`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := reclassificarFake(t)
			f.onWrite = func(f *escritaFake) {
				f.gets["/api/plataforma/dashboard/v2/central-rotinas"] = `{"pendencias":{},"rotinas":[],"rotinasContabilizei":[]}`
			}
			out, stderr, code := execCLI(t, "", append([]string{"rotinas", "reclassificar", "1000000000000501", "--yes", "-o", "json"}, tc.args...)...)
			if code != ExitOK || len(f.writes) != 1 || !strings.HasPrefix(f.writes[0], "POST /api/plataforma/documentos/reclassificar ") || !strings.Contains(f.writes[0], tc.corpo) {
				t.Fatalf("código %d, escritas %v\n%s", code, f.writes, stderr)
			}
			if !strings.Contains(out, `"situacao": "concluido"`) || !strings.Contains(out, `"id": "1000000000000501"`) {
				t.Errorf("saída:\n%s", out)
			}
			if !strings.Contains(stderr, `Reclassificar e concluir 1 pendência(s) da Central de Rotinas: "TED RECEBIDA BANCO" R$ 50.000,00 →`) {
				t.Errorf("resumo: %s", stderr)
			}
		})
	}
}

func TestRotinasReclassificarValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--classificacao", "Inexistente"}, "não é opção"},
		{[]string{"--classificacao", "1000000000000043"}, "exige --socio"},
		{[]string{"--classificacao", "1000000000000041", "--socio", "1"}, "não usa sócio"},
		{[]string{"--classificacao", "1000000000000042", "--socio", "9"}, "não é opção"},
		{[]string{"--socio", "1"}, "só vale com --classificacao"},
	} {
		f := reclassificarFake(t)
		_, stderr, code := execCLI(t, "", append([]string{"rotinas", "reclassificar", "1000000000000501", "--yes"}, tc.args...)...)
		if code != ExitUsage || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
	if _, _, code := execCLI(t, "", "rotinas", "reclassificar"); code != ExitUsage {
		t.Errorf("sem ids: código %d", code)
	}
	// A pendência continua na Central de Rotinas: situação "enviado".
	f := reclassificarFake(t)
	out, _, _ := execCLI(t, "", "rotinas", "reclassificar", "1000000000000501", "--classificacao", "1000000000000041", "--yes", "-o", "json")
	if len(f.writes) != 1 || !strings.Contains(out, `"situacao": "enviado"`) {
		t.Errorf("pendência ainda aberta:\n%s", out)
	}
}
