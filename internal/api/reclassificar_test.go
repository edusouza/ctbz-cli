package api

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestPathReclassificarInit(t *testing.T) {
	if got := PathReclassificarInit([]string{"1", "a b"}); got != "documentos/reclassificar/init?id=1&id=a+b" {
		t.Errorf("caminho = %s", got)
	}
}

func TestBuscarReclassificacoes(t *testing.T) {
	ps, err := BuscarReclassificacoes(context.Background(), fixtureGetter{PathReclassificarInit([]string{"1000000000000501"}): "reclassificar_init"}, []string{"1000000000000501"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 || ps[0].IDTexto() != "1000000000000501" || len(ps[0].Classificacoes) != 3 || len(ps[0].Classificacoes[2].Socios) != 2 {
		t.Errorf("pendências = %+v", ps)
	}
}

func TestIDsPendentes(t *testing.T) {
	var r RotinaEmpresa
	json.Unmarshal([]byte(`{"propriedades":{"documentosPendentes":["a1",22,{"idPendencia":"p3"},{"id":44},{"outro":1},null]}}`), &r)
	if got := strings.Join(r.IDsPendentes(), ","); got != "a1,22,p3,44" {
		t.Errorf("IDsPendentes = %s", got)
	}
	if (RotinaEmpresa{}).IDsPendentes() != nil {
		t.Error("sem propriedades deveria ser nil")
	}
}
