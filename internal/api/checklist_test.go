package api

import (
	"context"
	"testing"
)

func TestBuscarChecklist(t *testing.T) {
	c, err := BuscarChecklist(context.Background(), fixtureGetter{PathChecklistInit: "checklist_onboarding"})
	if err != nil {
		t.Fatal(err)
	}
	e, ok := c.Etapa(EtapaGerarProcuracao)
	if !ok || e.Concluida() || len(c.Etapas) != 5 {
		t.Errorf("etapa = %+v, %v", e, ok)
	}
	if e, _ := c.Etapa("CADASTRO_CONTA_PJ"); !e.Concluida() {
		t.Errorf("CADASTRO_CONTA_PJ deveria estar concluída")
	}
	if _, ok := c.Etapa("NAO_EXISTE"); ok {
		t.Error("etapa inexistente")
	}
}
