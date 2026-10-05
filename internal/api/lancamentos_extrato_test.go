package api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestBuscarLancamentosExtrato(t *testing.T) {
	ls, err := BuscarLancamentosExtrato(context.Background(), fixtureGetter{PathLancamentosExtrato(7, 2026, 9, 1): "lancamentos_extrato"}, 7, 2026, 9)
	if err != nil {
		t.Fatal(err)
	}
	if len(ls) != 3 || ls[1].IDContaUsuario != nil || *ls[2].IDLancamentoPai != 1000000000000100 {
		t.Fatalf("lançamentos = %+v", ls)
	}
	if got := DiaEmBrasilia(ls[0].Data.Time).Format("2006-01-02"); got != "2026-09-18" {
		t.Errorf("data epoch = %s", got)
	}
	if got := DiaEmBrasilia(ls[2].Data.Time).Format("2006-01-02"); got != "2026-09-20" {
		t.Errorf("data texto = %s", got)
	}
}

// paginasGetter devolve páginas cheias até a última.
type paginasGetter struct{ total int }

func (g paginasGetter) GetJSON(_ context.Context, path string, v any) error {
	var pagina int
	fmt.Sscanf(path[len(path)-2:], "=%d", &pagina)
	var list []LancamentoExtrato
	for i := (pagina - 1) * RegistrosPorPaginaExtrato; i < g.total && i < pagina*RegistrosPorPaginaExtrato; i++ {
		list = append(list, LancamentoExtrato{ID: int64(i)})
	}
	b, _ := json.Marshal(map[string]any{"list": list, "total": g.total})
	return json.Unmarshal(b, v)
}

func TestBuscarLancamentosExtratoPaginas(t *testing.T) {
	for _, total := range []int{0, 99, 100, 250} {
		ls, err := BuscarLancamentosExtrato(context.Background(), paginasGetter{total}, 1, 2026, 9)
		if err != nil || len(ls) != total {
			t.Errorf("total %d: %d lançamentos, %v", total, len(ls), err)
		}
	}
}

func TestDataInvalida(t *testing.T) {
	var d Data
	for _, in := range []string{`"ontem"`, `true`} {
		if err := json.Unmarshal([]byte(in), &d); err == nil {
			t.Errorf("%s deveria falhar", in)
		}
	}
	if err := json.Unmarshal([]byte(`null`), &d); err != nil || !d.IsZero() {
		t.Errorf("null: %v %v", d, err)
	}
}
