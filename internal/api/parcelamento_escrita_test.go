package api

import (
	"context"
	"strings"
	"testing"
)

func TestValidarParcelamento(t *testing.T) {
	for _, tc := range []struct{ tipo, neg, erro string }{
		{"simples-nacional", "", ""},
		{"pgfn-previdenciario", "", ""},
		{"especializado", "DIVIDA_ATIVA", ""},
		{"especializado", "", "exige --negociacao"},
		{"especializado", "OUTRA", "exige --negociacao"},
		{"simples-nacional", "VENCIDOS", "só vale"},
		{"mei", "", "desconhecido"},
	} {
		err := ValidarParcelamento(tc.tipo, tc.neg)
		if (tc.erro == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), tc.erro)) {
			t.Errorf("%s/%s: %v, quero %q", tc.tipo, tc.neg, err, tc.erro)
		}
	}
	if got := strings.Join(TiposParcelamento(), ","); got != "pgfn-nao-previdenciario,pgfn-previdenciario,pgfn-simples-nacional,simples-nacional,especializado" {
		t.Errorf("tipos = %s", got)
	}
}

func TestSimularParcelamento(t *testing.T) {
	if p := PathSimulacaoParcelamento("especializado", "DIVIDA_ATIVA"); p != "impostos/parcelamento/negociacao-especializada/init?tipoNegociacao=DIVIDA_ATIVA" {
		t.Errorf("caminho = %s", p)
	}
	path := PathSimulacaoParcelamento("pgfn-previdenciario", "")
	s, err := SimularParcelamento(context.Background(), fixtureGetter{path: "simulacao_parcelamento"}, "pgfn-previdenciario", "")
	if err != nil || len(s.Parcelas) != 2 || s.Parcelas[1].Quantidade != 24 || !s.CobrarServicoAdicional || len(s.Negociacao.Impostos) != 1 {
		t.Errorf("simulação = %+v, %v", s, err)
	}
}
