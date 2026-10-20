package api

import "testing"

func TestNovaDistribuicaoSocio(t *testing.T) {
	d := NovaDistribuicaoSocio(7, 3000005, 6000)
	if d.Valor != "30000.05" || d.Percentual != "60.00" || d.IDSocio != 7 {
		t.Errorf("%+v", d)
	}
	if got := duasCasas(-5); got != "-0.05" {
		t.Errorf("negativo: %s", got)
	}
}
