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

// pendenciaReclass monta uma resposta do init com uma pendência; campos sobrescreve os
// campos da pendência da fixture (o valor "-" remove o campo).
func pendenciaReclass(t *testing.T, campos map[string]string) string {
	t.Helper()
	p := map[string]string{
		"idPendencia":        `"1000000000000501"`,
		"idLancamento":       `1000000000000102`,
		"classificacaoAtual": `"Empréstimos e Financiamentos"`,
		"valor":              `50000`,
		"descricao":          `"TED RECEBIDA BANCO"`,
		"classificacoes": `[{"idClassificacao": 1000000000000041, "nome": "Empréstimos e Financiamentos", "socios": []},
{"idClassificacao": 1000000000000042, "nome": "Aporte de Capital", "socios": [{"idSocio": 1000000000000031, "nome": "FULANO DE TAL"}]}]`,
	}
	for k, v := range campos {
		if v == "-" {
			delete(p, k)
		} else {
			p[k] = v
		}
	}
	var partes []string
	for k, v := range p {
		partes = append(partes, `"`+k+`": `+v)
	}
	return "[{" + strings.Join(partes, ", ") + "}]"
}

// #428: um id ausente ou nulo no init vira erro de leitura antes do envio, nunca um POST com
// idLancamento 0 ou idPendencia null (ADR-0021).
func TestRotinasReclassificarPendenciaIncompleta(t *testing.T) {
	for _, tc := range []struct {
		name   string
		campos map[string]string
		args   []string
		want   string
	}{
		{"sem idPendencia", map[string]string{"idPendencia": "-"}, nil, "sem idPendencia"},
		{"idPendencia nulo", map[string]string{"idPendencia": "null"}, nil, "sem idPendencia"},
		{"idPendencia vazio", map[string]string{"idPendencia": `""`}, nil, "sem idPendencia"},
		{"sem idLancamento", map[string]string{"idLancamento": "-"}, nil, "sem idLancamento"},
		{"idLancamento nulo", map[string]string{"idLancamento": "null"}, nil, "sem idLancamento"},
		{"idLancamento zero", map[string]string{"idLancamento": "0"}, nil, "sem idLancamento"},
		{"sem opções", map[string]string{"classificacoes": "[]"}, nil, "sem opções de classificação"},
		{"opções nulas", map[string]string{"classificacoes": "null"}, nil, "sem opções de classificação"},
		{"opção sem id", map[string]string{"classificacoes": `[{"idClassificacao": null, "nome": "Aporte de Capital"}]`}, nil, "sem idClassificacao"},
		{"sócio sem id", map[string]string{"classificacoes": `[{"idClassificacao": 1000000000000042, "nome": "Aporte de Capital", "socios": [{"idSocio": null, "nome": "FULANO DE TAL"}]}]`}, nil, "sem idSocio"},
	} {
		for _, modo := range []string{"--yes", "--dry-run"} {
			t.Run(tc.name+" "+modo, func(t *testing.T) {
				f := newEscritaFake(t, map[string]string{pathReclassInit: pendenciaReclass(t, tc.campos)})
				args := []string{"rotinas", "reclassificar", "1000000000000501", "--classificacao", "Aporte de Capital", modo}
				out, stderr, code := execCLI(t, "", append(args, tc.args...)...)
				if code != ExitError || len(f.writes) != 0 || !strings.Contains(stderr, tc.want) || !strings.Contains(stderr, "nada foi enviado") {
					t.Errorf("código %d, escritas %v, %q (quero %q)", code, f.writes, stderr, tc.want)
				}
				if strings.Contains(out, "POST") {
					t.Errorf("--dry-run imprimiu a requisição:\n%s", out)
				}
			})
		}
	}
}

// #428: sem valor, o resumo diz "valor não informado" em vez de R$ 0,00; o valor não vai no
// corpo, então não impede a escrita.
func TestRotinasReclassificarValorNulo(t *testing.T) {
	for _, campos := range []map[string]string{{"valor": "null"}, {"valor": "-"}} {
		f := newEscritaFake(t, map[string]string{pathReclassInit: pendenciaReclass(t, campos)})
		_, stderr, code := execCLI(t, "", "rotinas", "reclassificar", "1000000000000501", "--classificacao", "1000000000000041", "--dry-run")
		if code != ExitOK || len(f.writes) != 0 || !strings.Contains(stderr, `"TED RECEBIDA BANCO" valor não informado → Empréstimos e Financiamentos`) || strings.Contains(stderr, "R$ 0,00") {
			t.Errorf("%v: código %d, %s", campos, code, stderr)
		}
	}
}

// #435: um nome que casa com mais de uma opção é erro de uso; o id ou o nome escrito
// exatamente como na lista escolhe sem ambiguidade.
func TestRotinasReclassificarNomeAmbiguo(t *testing.T) {
	opcoes := map[string]string{"classificacoes": `[{"idClassificacao": 1000000000000041, "nome": "Aporte de Capital"},
{"idClassificacao": 1000000000000042, "nome": "APORTE DE CAPITAL"},
{"idClassificacao": 1000000000000043, "nome": "Mútuo"},
{"idClassificacao": 1000000000000044, "nome": "Mútuo"},
{"idClassificacao": 1000000000000045, "nome": "1000000000000041"}]`}
	for _, tc := range []struct {
		classificacao string
		corpo         string // vazio: recusa por ambiguidade
		candidatos    []string
	}{
		{"aporte de capital", "", []string{`1000000000000041 "Aporte de Capital"`, `1000000000000042 "APORTE DE CAPITAL"`}},
		{"Mútuo", "", []string{`1000000000000043 "Mútuo"`, `1000000000000044 "Mútuo"`}},
		{"mútuo", "", []string{`1000000000000043 "Mútuo"`, `1000000000000044 "Mútuo"`}},
		{"Aporte de Capital", `"idClassificacao":1000000000000041`, nil},
		{"APORTE DE CAPITAL", `"idClassificacao":1000000000000042`, nil},
		{"1000000000000044", `"idClassificacao":1000000000000044`, nil},
		// O id vence uma opção cujo nome é igual a ele.
		{"1000000000000041", `"idClassificacao":1000000000000041`, nil},
	} {
		t.Run(tc.classificacao, func(t *testing.T) {
			f := newEscritaFake(t, map[string]string{pathReclassInit: pendenciaReclass(t, opcoes),
				"/api/plataforma/dashboard/v2/central-rotinas": `{"pendencias":{},"rotinas":[],"rotinasContabilizei":[]}`})
			_, stderr, code := execCLI(t, "", "rotinas", "reclassificar", "1000000000000501", "--classificacao", tc.classificacao, "--yes")
			if tc.corpo != "" {
				if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], tc.corpo) {
					t.Errorf("código %d, escritas %v\n%s", code, f.writes, stderr)
				}
				return
			}
			if code != ExitUsage || len(f.writes) != 0 || !strings.Contains(stderr, "ambígua") || !strings.Contains(stderr, "use o id") {
				t.Fatalf("código %d, escritas %v, %s", code, f.writes, stderr)
			}
			for _, c := range tc.candidatos {
				if !strings.Contains(stderr, c) {
					t.Errorf("falta o candidato %s em %s", c, stderr)
				}
			}
		})
	}
}
