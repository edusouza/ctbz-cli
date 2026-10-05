package cli

import (
	"strings"
	"testing"
)

const pathConcInit = "/api/plataforma/conciliacao-fiscal/v2/init"

func conciliacaoFake(t *testing.T) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{
		pathConcInit: `{"qtdNotasFiscaisPendentes":2,"qtdRecebimentosPendentes":3,"qtdConciliacoesAutomaticasMesAnterior":0,"competenciaMesAnterior":"2026-09"}`,
		"/api/plataforma/conciliacao-fiscal/conciliar/receitas": `[{"id":987,"numero":"12","valor":100}]`,
	})
	f.onWrite = func(f *escritaFake) {
		if strings.Contains(f.writes[len(f.writes)-1], "resolver-pendencia") {
			f.gets[pathConcInit] = `{"qtdNotasFiscaisPendentes":2,"qtdRecebimentosPendentes":2,"qtdConciliacoesAutomaticasMesAnterior":0,"competenciaMesAnterior":"2026-09"}`
		}
	}
	return f
}

func TestConciliacaoMotivos(t *testing.T) {
	t.Setenv("CTBZ_HOME", t.TempDir())
	out, _, code := execCLI(t, "", "pendencias", "conciliacao", "motivos", "-o", "csv")
	if code != ExitOK || strings.Count(out, "\n") != 29 || !strings.Contains(out, "AFAC,Valor adiantado para virar capital social (AFAC),true") {
		t.Errorf("código %d\n%s", code, out)
	}
}

func TestConciliacaoCandidatosEDetalhes(t *testing.T) {
	f := conciliacaoFake(t)
	out, _, code := execCLI(t, "", "pendencias", "conciliacao", "candidatos", "--notas", "-o", "csv")
	if code != ExitOK || !strings.Contains(out, "987") {
		t.Errorf("candidatos: código %d\n%s", code, out)
	}
	if _, _, code := execCLI(t, "", "pendencias", "conciliacao", "candidatos"); code != ExitUsage {
		t.Errorf("candidatos sem tipo: código %d", code)
	}
	f.resposta = `{"justificativa":"REEMBOLSO"}`
	// detalhes é um POST de leitura: não passa pela camada de escrita (nem confirmação, nem registro).
	_, stderr, code := execCLI(t, "", "pendencias", "conciliacao", "detalhes", "--recebimento", "1,2", "--nota", "3")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `POST /api/plataforma/conciliacao-fiscal/pendencia/detalhes {"idsRecebimento":[1,2],"idsReceita":[3]}` {
		t.Errorf("detalhes: código %d, escritas %q, %s", code, f.writes, stderr)
	}
}

func TestConciliacaoResolver(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		corpo string
	}{
		{"vincular", []string{"123", "124", "--vincular", "nota:987"},
			`{"idPendencia":[123,124],"contraparte":[{"origem":"NOTAFISCAL","id":987}],"tipoResolucaoPendencia":null}`},
		{"motivo com sócio", []string{"123", "--motivo", "emprestimo_do_socio_a_empresa", "--socio", "55"},
			`{"idPendencia":[123],"contraparte":[],"tipoResolucaoPendencia":"EMPRESTIMO_DO_SOCIO_A_EMPRESA","idVinculo":55}`},
		{"vincular com motivo", []string{"123", "--vincular", "MOVIMENTACAO:456", "--motivo", "DESCONTO_CONCEDIDO", "--nota-ativacao", "1234"},
			`{"idPendencia":[123],"contraparte":[{"origem":"MOVIMENTACAO","id":456}],"tipoResolucaoPendencia":"DESCONTO_CONCEDIDO","numeroNotaAtivacaoContabil":"1234"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := conciliacaoFake(t)
			out, stderr, code := execCLI(t, "", append([]string{"pendencias", "conciliacao", "resolver", "--yes", "-o", "json"}, tc.args...)...)
			if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "POST /api/plataforma/conciliacao-fiscal/conciliar/resolver-pendencia "+tc.corpo {
				t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
			}
			if !strings.Contains(out, `"situacao": "resolvido"`) || !strings.Contains(out, `"pendencias_restantes": 4`) || !strings.Contains(stderr, "(risco alto)") {
				t.Errorf("saída:\n%s\n%s", out, stderr)
			}
		})
	}
}

func TestConciliacaoResolverValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"123"}, "--vincular ORIGEM:ID ou --motivo"},
		{[]string{"abc", "--motivo", "REEMBOLSO"}, "inválido"},
		{[]string{"123", "--vincular", "OUTRA:1"}, "NOTAFISCAL ou MOVIMENTACAO"},
		{[]string{"123", "--vincular", "NOTAFISCAL"}, "use NOTAFISCAL:ID"},
		{[]string{"123", "--motivo", "OUTRO"}, "motivo desconhecido"},
		{[]string{"123", "--motivo", "AFAC"}, "exige --socio"},
		{[]string{"123", "--motivo", "REEMBOLSO", "--socio", "1"}, "não usa sócio"},
		{[]string{"123", "--vincular", "NOTAFISCAL:1", "--socio", "1"}, "só valem com --motivo"},
	} {
		f := conciliacaoFake(t)
		_, stderr, code := execCLI(t, "", append([]string{"pendencias", "conciliacao", "resolver", "--yes"}, tc.args...)...)
		if code != ExitUsage || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}
