package cli

import (
	"net/http"
	"strings"
	"testing"
)

const (
	pathSimPGFN = "/api/plataforma/impostos/parcelamento/negociacao-automatica/pgfn-previdenciario/init"
	pathSimSN   = "/api/plataforma/impostos/parcelamento/negociacao-automatica/simples-nacional/init"
	pathSimEsp  = "/api/plataforma/impostos/parcelamento/negociacao-especializada/init?tipoNegociacao=VENCIDOS"
)

func parcelamentoFake(t *testing.T) *escritaFake {
	t.Helper()
	sim := fixture(t, "simulacao_parcelamento")
	return newEscritaFake(t, map[string]string{
		pathSimPGFN: sim, pathSimSN: sim, pathSimEsp: sim,
		"/api/plataforma/impostos/parcelamento/negociacao-especializada/detalhes/init/T-77": `{"status":"EM_ANALISE"}`,
	})
}

func TestParcelamentoContratar(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		escrita  string
		resposta string
		saida    string
	}{
		{"pgfn com parcelas", []string{"pgfn-previdenciario", "--parcelas", "24"},
			`POST /api/plataforma/impostos/parcelamento/negociacao-automatica/pgfn-previdenciario/contratar {"quantidadeParcelas":24}`, "", `"parcelas": 24`},
		{"pgfn padrão = primeira opção", []string{"pgfn-previdenciario"},
			`{"quantidadeParcelas":12}`, "", `"entrada": 500.00`},
		{"simples sem corpo", []string{"simples-nacional"},
			`POST /api/plataforma/impostos/parcelamento/negociacao-automatica/simples-nacional/contratar `, "", `"tipo": "simples-nacional"`},
		{"especializado", []string{"especializado", "--negociacao", "VENCIDOS"},
			`POST /api/plataforma/impostos/parcelamento/negociacao-especializada/contratar {"tipoNegociacao":"VENCIDOS"}`, `{"idTicket":"T-77"}`, `"status": "EM_ANALISE"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := parcelamentoFake(t)
			f.resposta = tc.resposta
			out, stderr, code := execCLI(t, "", append(append([]string{"impostos", "parcelamento", "contratar"}, tc.args...), "--yes", "-o", "json")...)
			if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], tc.escrita) {
				t.Fatalf("código %d, escritas %q\n%s", code, f.writes, stderr)
			}
			if !strings.Contains(out, `"situacao": "solicitado"`) || !strings.Contains(out, tc.saida) {
				t.Errorf("saída:\n%s", out)
			}
			for _, w := range []string{"(risco alto)", "confissão de dívida", "cancelado automaticamente", "Selic", "serviço adicional R$ 99,90; emissão de guia R$ 15,00"} {
				if !strings.Contains(stderr, w) {
					t.Errorf("stderr sem %q: %s", w, stderr)
				}
			}
		})
	}
}

func TestParcelamentoContratarValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"pgfn-previdenciario", "--parcelas", "36"}, ExitUsage, "36 parcelas não é uma opção; use 12, 24"},
		{[]string{"simples-nacional", "--parcelas", "12"}, ExitUsage, "só vale para os tipos pgfn"},
		{[]string{"especializado"}, ExitUsage, "exige --negociacao"},
	} {
		f := parcelamentoFake(t)
		_, stderr, code := execCLI(t, "", append(append([]string{"impostos", "parcelamento", "contratar"}, tc.args...), "--yes")...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, escritas %d, %q", tc.args, code, len(f.writes), stderr)
		}
	}
	// Sem débitos: nada é contratado.
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("escrita sem débitos: %s %s", r.Method, r.URL)
		}
		w.WriteHeader(560)
	})
	if _, stderr, code := execCLI(t, "", "impostos", "parcelamento", "contratar", "simples-nacional", "--yes"); code != ExitError || !strings.Contains(stderr, "nenhum débito") {
		t.Errorf("sem débitos: código %d, %s", code, stderr)
	}
}

// Uma opção sem quantidade (ou com 0) ou sem os valores do resumo é erro de leitura
// (ADR-0021): nada é contratado, nem com --yes, e o --dry-run não mostra a requisição.
func TestParcelamentoContratarOpcaoIncompleta(t *testing.T) {
	const (
		semQtd     = `{"parcelas":[{"valorEntrada":500,"valorDemaisParcelas":250.5}]}`
		qtdZero    = `{"parcelas":[{"quantidade":0,"valorEntrada":500,"valorDemaisParcelas":250.5}]}`
		umaSemQtd  = `{"parcelas":[{"quantidade":12,"valorEntrada":500,"valorDemaisParcelas":250.5},{"valorEntrada":300,"valorDemaisParcelas":140.25}]}`
		semEntrada = `{"parcelas":[{"quantidade":12,"valorEntrada":null,"valorDemaisParcelas":250.5}]}`
		semDemais  = `{"parcelas":[{"quantidade":12,"valorEntrada":500},{"quantidade":24,"valorEntrada":300}]}`
	)
	const msgQtd, msgValor = "sem a quantidade de parcelas; contrate pelo painel", "não trouxe o valor das parcelas da opção de 12 parcelas; contrate pelo painel"
	for _, tc := range []struct {
		name, sim string
		args      []string
		want      string
	}{
		{"sem quantidade", semQtd, []string{"pgfn-previdenciario", "--yes"}, msgQtd},
		{"sem quantidade com --parcelas", semQtd, []string{"pgfn-previdenciario", "--parcelas", "12", "--yes"}, msgQtd},
		{"sem quantidade em --dry-run", semQtd, []string{"pgfn-previdenciario", "--dry-run"}, msgQtd},
		{"quantidade 0", qtdZero, []string{"pgfn-previdenciario", "--yes"}, msgQtd},
		{"outra opção sem quantidade", umaSemQtd, []string{"pgfn-previdenciario", "--parcelas", "12", "--yes"}, msgQtd},
		{"simples sem quantidade no resumo", semQtd, []string{"simples-nacional", "--yes"}, msgQtd},
		{"entrada nula", semEntrada, []string{"pgfn-previdenciario", "--yes"}, msgValor},
		{"entrada nula em --dry-run", semEntrada, []string{"pgfn-previdenciario", "--dry-run"}, msgValor},
		{"demais ausente", semDemais, []string{"pgfn-previdenciario", "--parcelas", "12", "--yes"}, msgValor},
		{"simples com entrada nula", semEntrada, []string{"simples-nacional", "--yes"}, msgValor},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newEscritaFake(t, map[string]string{pathSimPGFN: tc.sim, pathSimSN: tc.sim})
			out, stderr, code := execCLI(t, "", append([]string{"impostos", "parcelamento", "contratar"}, tc.args...)...)
			if code != ExitError || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
				t.Fatalf("código %d, escritas %q, stderr %q", code, f.writes, stderr)
			}
			if strings.Contains(out+stderr, "POST") || strings.Contains(stderr, "confissão de dívida") {
				t.Errorf("mostrou a requisição ou o resumo:\nstdout %s\nstderr %s", out, stderr)
			}
		})
	}
	// Só a opção escolhida precisa dos valores: a de 24 está completa.
	f := newEscritaFake(t, map[string]string{pathSimPGFN: `{"parcelas":[{"quantidade":12,"valorEntrada":null},{"quantidade":24,"valorEntrada":300,"valorDemaisParcelas":140.25}]}`})
	if _, stderr, code := execCLI(t, "", "impostos", "parcelamento", "contratar", "pgfn-previdenciario", "--parcelas", "24", "--yes"); code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `{"quantidadeParcelas":24}`) {
		t.Errorf("opção 24 completa: código %d, escritas %q, %s", code, f.writes, stderr)
	}
}

// Custo adicional sem valor aparece como "não informado" na confirmação, não como R$ 0,00.
func TestParcelamentoContratarCustoNaoInformado(t *testing.T) {
	sim := `{"parcelas":[{"quantidade":12,"valorEntrada":500,"valorDemaisParcelas":250.5}],"cobrarServicoAdicional":true,"valorServicoAdicional":null}`
	f := newEscritaFake(t, map[string]string{pathSimPGFN: sim})
	_, stderr, code := execCLI(t, "", "impostos", "parcelamento", "contratar", "pgfn-previdenciario", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(stderr, "serviço adicional não informado; emissão de guia não informado") || strings.Contains(stderr, "R$ 0,00") {
		t.Errorf("código %d, stderr %s", code, stderr)
	}
}
