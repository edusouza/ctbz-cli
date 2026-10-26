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

// senderProibido falha o teste se algo for enviado.
type senderProibido struct{ t *testing.T }

func (s senderProibido) Send(_ context.Context, method, path string, _ any, _ any) error {
	s.t.Errorf("enviou %s %s", method, path)
	return nil
}

func (s senderProibido) SendMultipart(_ context.Context, method, path string, _ []Campo, _ any) error {
	s.t.Errorf("enviou %s %s", method, path)
	return nil
}

func TestContratarParcelamentoPGFNSemQuantidade(t *testing.T) {
	for _, n := range []int{0, -1} {
		if _, err := ContratarParcelamento(context.Background(), senderProibido{t}, "pgfn-previdenciario", n, ""); err == nil || !strings.Contains(err.Error(), "quantidade de parcelas") {
			t.Errorf("%d parcelas: %v", n, err)
		}
	}
}
