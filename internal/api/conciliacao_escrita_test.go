package api

import (
	"sort"
	"testing"
)

func TestMotivosConciliacao(t *testing.T) {
	if len(MotivosConciliacao) != 28 {
		t.Errorf("%d motivos, a documentação lista 28", len(MotivosConciliacao))
	}
	if !sort.SliceIsSorted(MotivosConciliacao, func(i, j int) bool { return MotivosConciliacao[i].Codigo < MotivosConciliacao[j].Codigo }) {
		t.Error("motivos fora de ordem")
	}
	if m, ok := MotivoPorCodigo("AFAC"); !ok || !m.ExigeSocio {
		t.Errorf("AFAC = %+v, %v", m, ok)
	}
	if _, ok := MotivoPorCodigo("OUTRO"); ok {
		t.Error("OUTRO não chama a API")
	}
}
