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

func TestVersaoConfirmacao(t *testing.T) {
	for rollout, want := range map[string]string{"v5": "v5", "v4": "v3", "v3": "v3", "": "v3"} {
		if got := VersaoConfirmacao(rollout); got != want {
			t.Errorf("VersaoConfirmacao(%q) = %s, quero %s", rollout, got, want)
		}
	}
}

func TestRecalculo(t *testing.T) {
	path := PathRecalculoInit(1000000000000001, "GUIAS", "GUIA")
	if path != "impostos/v2/impostos-a-pagar/recalculo/init?idGuia=1000000000000001&origem=GUIAS&tipo=GUIA" {
		t.Errorf("caminho = %s", path)
	}
	r, err := BuscarRecalculo(context.Background(), fixtureGetter{path: "recalculo_init"}, 1000000000000001, "GUIAS", "GUIA")
	if err != nil || r.DataRecomendada != "2026-10-09" || !r.CobrarRecalculo || len(r.DatasIndisponiveis) != 3 || *r.ValorRecalculo != 1234.56 {
		t.Errorf("recálculo = %+v, %v", r, err)
	}
}
