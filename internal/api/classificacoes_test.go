package api

import (
	"context"
	"testing"
)

func TestVinculoExigido(t *testing.T) {
	for desc, want := range map[string]string{
		"Impostos - Simples Nacional":                 VinculoGuiaImposto,
		"Impostos - IRRF":                             VinculoGuiaImposto,
		"Impostos - Taxas Municipais":                 "",
		"Sócios - Distribuição de Lucros Antecipados": VinculoSocio,
		"Pagamento de Fornecedores":                   "",
	} {
		if got := VinculoExigido(desc); got != want {
			t.Errorf("VinculoExigido(%q) = %q, quero %q", desc, got, want)
		}
	}
}

func TestBuscarCaixaCategorias(t *testing.T) {
	cs, err := BuscarCaixaCategorias(context.Background(), fixtureGetter{PathCaixa(2026, 9): "caixa_categorias"}, 2026, 9)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 8 || cs[6].IDTexto() != "1000000000000021" || cs[7].IDTexto() != "1000000000000031" || !cs[0].Entrada {
		t.Errorf("categorias = %+v", cs)
	}
}
