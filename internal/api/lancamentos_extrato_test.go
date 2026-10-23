package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestBuscarLancamentosExtrato(t *testing.T) {
	ls, err := BuscarLancamentosExtrato(context.Background(), fixtureGetter{PathLancamentosExtrato(7, 2026, 9, 1): "lancamentos_extrato"}, 7, 2026, 9)
	if err != nil {
		t.Fatal(err)
	}
	if len(ls) != 3 || ls[0].ID != 1000000000000101 || ls[1].IDContaUsuario != nil || *ls[2].IDLancamentoPai != 1000000000000100 {
		t.Fatalf("lançamentos = %+v", ls)
	}
	if got := DiaEmBrasilia(ls[0].Data.Time).Format("2006-01-02"); got != "2026-09-18" {
		t.Errorf("data epoch = %s", got)
	}
	if got := DiaEmBrasilia(ls[2].Data.Time).Format("2006-01-02"); got != "2026-09-20" {
		t.Errorf("data texto = %s", got)
	}
	if s := ls[1].MelhorSugestao(); s == nil || s.Descricao != "Pagamento de Fornecedores" || s.Probabilidade != 0.64 {
		t.Errorf("melhor sugestão = %+v", s)
	}
	if s := ls[2].MelhorSugestao(); s != nil {
		t.Errorf("sem sugestões: %+v", s)
	}
}

// paginasGetter devolve páginas (arrays) cheias até a última. Com repete, ignora o número da
// página e devolve sempre a primeira; com infinito, nunca termina.
type paginasGetter struct {
	total            int
	repete, infinito bool
	paginasLidas     *int
}

func (g paginasGetter) GetJSON(_ context.Context, path string, v any) error {
	var pagina int
	fmt.Sscanf(path[strings.LastIndex(path, "=")+1:], "%d", &pagina)
	if g.paginasLidas != nil {
		*g.paginasLidas = pagina
	}
	if g.repete {
		pagina = 1
	}
	list := []LancamentoExtrato{}
	for i := (pagina - 1) * RegistrosPorPaginaExtrato; (g.infinito || i < g.total) && i < pagina*RegistrosPorPaginaExtrato; i++ {
		list = append(list, LancamentoExtrato{ID: int64(i + 1)})
	}
	b, _ := json.Marshal(list)
	return json.Unmarshal(b, v)
}

func TestBuscarLancamentosExtratoPaginas(t *testing.T) {
	for _, total := range []int{0, 99, 100, 250} {
		ls, err := BuscarLancamentosExtrato(context.Background(), paginasGetter{total: total}, 1, 2026, 9)
		if err != nil || len(ls) != total || ls == nil {
			t.Errorf("total %d: %d lançamentos, %v", total, len(ls), err)
		}
	}
	// O servidor ignora a página e repete a primeira: erro na repetição, em vez de lista truncada.
	var lidas int
	ls, err := BuscarLancamentosExtrato(context.Background(), paginasGetter{total: 300, repete: true, paginasLidas: &lidas}, 1, 2026, 9)
	if err == nil || !strings.Contains(err.Error(), "página 2 de lançamentos do extrato de 09/2026 repete a anterior") || ls != nil || lidas != 2 {
		t.Errorf("página repetida: %d lançamentos em %d páginas, %v", len(ls), lidas, err)
	}
	// Ignorar a página num mês com menos de uma página cheia não é problema: para na primeira.
	ls, err = BuscarLancamentosExtrato(context.Background(), paginasGetter{total: 40, repete: true, paginasLidas: &lidas}, 1, 2026, 9)
	if err != nil || len(ls) != 40 || lidas != 1 {
		t.Errorf("página repetida incompleta: %d lançamentos em %d páginas, %v", len(ls), lidas, err)
	}
	// Páginas cheias sem fim: erro no limite, em vez de laço infinito ou lista truncada.
	_, err = BuscarLancamentosExtrato(context.Background(), paginasGetter{infinito: true, paginasLidas: &lidas}, 1, 2026, 9)
	if err == nil || !strings.Contains(err.Error(), "mais de 100 páginas") || lidas != maxPaginasExtrato {
		t.Errorf("sem fim: %d páginas, %v", lidas, err)
	}
}

func TestBuscarLancamentosExtratoObjeto(t *testing.T) {
	g := fixtureGetter{PathLancamentosExtrato(7, 2026, 9, 1): "extrato_info"} // um objeto, não um array
	if _, err := BuscarLancamentosExtrato(context.Background(), g, 7, 2026, 9); err == nil {
		t.Error("um objeto no lugar do array deveria falhar")
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
