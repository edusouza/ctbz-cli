package cli

import (
	"strings"
	"testing"
)

const pathConcInit = "/api/plataforma/conciliacao-fiscal/v2/init"

// pathConcPendencias é a primeira página das pendências abertas de um tipo, na janela de
// 12 meses até 2026-10-03 (o dia fixado por conciliacaoFake).
func pathConcPendencias(tipo string) string {
	return "/api/plataforma/conciliacao-fiscal/v2/pendencias?pagina=1&periodoFinal=2026-10-03&periodoInicial=2025-10-03" +
		"&status=PENDENTE&tipoPendencia=" + tipo + "&totalPagina=10"
}

var (
	pathConcRecebimentos = pathConcPendencias("RECEBIMENTO_SEM_RECEITA")
	pathConcNotas        = pathConcPendencias("RECEITA_SEM_RECEBIMENTO")
)

// conciliacaoFake tem os recebimentos sem nota 123 e 124 e a nota sem recebimento 321.
func conciliacaoFake(t *testing.T) *escritaFake {
	t.Helper()
	fixNow(t, "2026-10-03")
	f := newEscritaFake(t, map[string]string{
		pathConcInit: `{"qtdNotasFiscaisPendentes":2,"qtdRecebimentosPendentes":3,"qtdConciliacoesAutomaticasMesAnterior":0,"competenciaMesAnterior":"2026-09"}`,
		"/api/plataforma/conciliacao-fiscal/conciliar/receitas": `[{"id":987,"numero":"12","valor":100}]`,
		pathConcRecebimentos: `{"pagina":[{"id":123,"dataRecebimento":"2026-09-10","valor":100.5,"descricao":"PIX CLIENTE FICTICIO\u001b[2J"},` +
			`{"id":124,"dataRecebimento":"2026-09-12","valor":50,"observacao":null,"contas":[1]}],"totalRegistros":2,"totalPaginas":1}`,
		pathConcNotas: `{"pagina":[{"id":321,"numeroNota":"10","dataEmissao":"2026-08-01","valor":200}],"totalRegistros":1,"totalPaginas":1}`,
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
		name   string
		args   []string
		corpo  string
		resumo string
	}{
		{"vincular", []string{"123", "124", "--vincular", "nota:987"},
			`{"idPendencia":[123,124],"contraparte":[{"origem":"NOTAFISCAL","id":987}],"tipoResolucaoPendencia":null}`,
			"(recebimentos sem nota fiscal) — 123 [dataRecebimento 2026-09-10, valor 100.5"},
		{"motivo com sócio", []string{"123", "--motivo", "emprestimo_do_socio_a_empresa", "--socio", "55"},
			`{"idPendencia":[123],"contraparte":[],"tipoResolucaoPendencia":"EMPRESTIMO_DO_SOCIO_A_EMPRESA","idVinculo":55}`,
			"(recebimentos sem nota fiscal) — 123 [dataRecebimento 2026-09-10"},
		{"vincular com motivo", []string{"321", "--vincular", "MOVIMENTACAO:456", "--motivo", "DESCONTO_CONCEDIDO", "--nota-ativacao", "1234"},
			`{"idPendencia":[321],"contraparte":[{"origem":"MOVIMENTACAO","id":456}],"tipoResolucaoPendencia":"DESCONTO_CONCEDIDO","numeroNotaAtivacaoContabil":"1234"}`,
			"(notas sem recebimento) — 321 [numeroNota 10, dataEmissao 2026-08-01, valor 200]"},
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
			if !strings.Contains(stderr, tc.resumo) {
				t.Errorf("o resumo deveria mostrar %q:\n%s", tc.resumo, stderr)
			}
		})
	}
}

// O resumo da confirmação mostra o tipo e os campos de cada pendência lida da API (item, data,
// valor), sem caracteres de controle vindos da resposta, e não só os ids digitados.
func TestConciliacaoResolverResumo(t *testing.T) {
	f := conciliacaoFake(t)
	out, stderr, code := execCLI(t, "", "pendencias", "conciliacao", "resolver", "--dry-run", "-o", "json", "123", "124", "--vincular", "NOTAFISCAL:987")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, "resolver-pendencia") {
		t.Fatalf("código %d, escritas %q\n%s\n%s", code, f.writes, out, stderr)
	}
	want := "Resolver 2 pendência(s) de conciliação fiscal (recebimentos sem nota fiscal) — " +
		"123 [dataRecebimento 2026-09-10, valor 100.5, descricao PIX CLIENTE FICTICIO[2J]; " +
		"124 [dataRecebimento 2026-09-12, valor 50]: vincular NOTAFISCAL 987 (risco alto)"
	if !strings.Contains(stderr, want) {
		t.Errorf("resumo:\n%s\nquero:\n%s", stderr, want)
	}
}

// resolver confere os ids contra as pendências abertas antes de qualquer escrita: com --yes
// nada é enviado e com --dry-run nenhuma requisição é impressa.
func TestConciliacaoResolverConfereAsPendencias(t *testing.T) {
	semPendencias := `{"pagina":[],"totalRegistros":0,"totalPaginas":0}`
	for _, tc := range []struct {
		name string
		args []string
		gets map[string]string // substitui respostas de conciliacaoFake
		code int
		want string
	}{
		{"id inexistente", []string{"999", "--motivo", "REEMBOLSO"}, nil, ExitUsage,
			"999 não está entre as pendências abertas (notas sem recebimento ou recebimentos sem nota fiscal) dos últimos 12 meses"},
		{"só os ids inexistentes são listados", []string{"123", "998", "999", "--vincular", "NOTAFISCAL:987"}, nil, ExitUsage,
			"998, 999 não estão entre as pendências abertas (recebimentos sem nota fiscal) dos últimos 12 meses"},
		{"nota vinculada a nota", []string{"321", "--vincular", "NOTAFISCAL:987"}, nil, ExitUsage,
			"321 não está entre as pendências abertas (recebimentos sem nota fiscal)"},
		{"recebimento vinculado a movimentação", []string{"123", "--vincular", "MOVIMENTACAO:456"}, nil, ExitUsage,
			"123 não está entre as pendências abertas (notas sem recebimento)"},
		{"sem pendências abertas", []string{"123", "--motivo", "REEMBOLSO"},
			map[string]string{pathConcRecebimentos: semPendencias, pathConcNotas: semPendencias}, ExitUsage,
			"123 não está entre as pendências abertas"},
		{"id repetido", []string{"123", " 123", "--motivo", "REEMBOLSO"}, nil, ExitUsage, "id de pendência repetido: 123"},
		{"contraparte repetida", []string{"123", "--vincular", "NOTAFISCAL:987", "--vincular", "nota:987"}, nil, ExitUsage,
			"contraparte repetida: NOTAFISCAL:987"},
		{"origens misturadas", []string{"123", "--vincular", "NOTAFISCAL:987", "--vincular", "MOVIMENTACAO:456"}, nil, ExitUsage,
			"não misture NOTAFISCAL e MOVIMENTACAO"},
		{"tipos misturados", []string{"123", "321", "--motivo", "REEMBOLSO"}, nil, ExitUsage,
			"as pendências misturam recebimentos sem nota fiscal (123) e notas sem recebimento (321)"},
		{"id nos dois tipos", []string{"123", "--motivo", "REEMBOLSO"},
			map[string]string{pathConcNotas: `{"pagina":[{"id":123,"valor":1}],"totalRegistros":1,"totalPaginas":1}`}, ExitUsage,
			"123 aparece entre as notas sem recebimento e entre os recebimentos sem nota fiscal"},
		{"item sem id", []string{"123", "--vincular", "NOTAFISCAL:987"},
			map[string]string{pathConcRecebimentos: `{"pagina":[{"idPendencia":123,"valor":1}],"totalRegistros":1,"totalPaginas":1}`}, ExitError,
			"pendência de conciliação sem id numérico na resposta"},
		{"leitura das pendências falha", []string{"123", "--vincular", "NOTAFISCAL:987"},
			map[string]string{pathConcRecebimentos: `{"pagina":`}, ExitError, ""},
	} {
		for _, modo := range []string{"--yes", "--dry-run"} {
			t.Run(tc.name+" "+modo, func(t *testing.T) {
				f := conciliacaoFake(t)
				for k, v := range tc.gets {
					f.gets[k] = v
				}
				out, stderr, code := execCLI(t, "", append([]string{"pendencias", "conciliacao", "resolver", modo}, tc.args...)...)
				if code != tc.code || !strings.Contains(stderr, tc.want) {
					t.Errorf("código %d (quero %d), stderr %q (quero %q)", code, tc.code, stderr, tc.want)
				}
				if len(f.writes) != 0 || out != "" {
					t.Errorf("nada deveria ser enviado nem impresso: escritas %q, saída %q", f.writes, out)
				}
			})
		}
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
