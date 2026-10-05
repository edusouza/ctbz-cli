package api

import (
	"context"
	"testing"
	"time"
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

func TestDatasBrasilia(t *testing.T) {
	if got := DataISOBrasilia(time.Date(2026, 9, 15, 22, 0, 0, 0, time.UTC)); got != "2026-09-15T03:00:00.000Z" {
		t.Errorf("DataISOBrasilia = %s", got)
	}
	// 1789732800000 = 2026-09-18T12:00:00Z; 1789700400000 = 2026-09-18T03:00:00Z (meia-noite em Brasília).
	for ms, want := range map[int64]string{1789732800000: "2026-09-18", 1789700400000: "2026-09-18", 1789700399000: "2026-09-17"} {
		if got := DiaEmBrasilia(time.UnixMilli(ms)).Format("2006-01-02"); got != want {
			t.Errorf("DiaEmBrasilia(%d) = %s, quero %s", ms, got, want)
		}
	}
}
