package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// A página 2 usa uma fixture que é um array vazio: a leitura só termina numa página vazia.
func TestBuscarLancamentosExtrato(t *testing.T) {
	ls, err := BuscarLancamentosExtrato(context.Background(), fixtureGetter{PathLancamentosExtrato(7, 2026, 9, 1): "lancamentos_extrato", PathLancamentosExtrato(7, 2026, 9, 2): "chamados_em_andamento"}, 7, 2026, 9)
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
	tamanho          int // tamanho real da página do servidor; 0 = RegistrosPorPaginaExtrato
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
	tam := g.tamanho
	if tam == 0 {
		tam = RegistrosPorPaginaExtrato
	}
	list := []LancamentoExtrato{}
	for i := (pagina - 1) * tam; (g.infinito || i < g.total) && i < pagina*tam; i++ {
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
	// Ignorar a página também é erro num mês pequeno: só a página vazia encerra a leitura.
	ls, err = BuscarLancamentosExtrato(context.Background(), paginasGetter{total: 40, repete: true, paginasLidas: &lidas}, 1, 2026, 9)
	if err == nil || !strings.Contains(err.Error(), "repete a anterior") || ls != nil {
		t.Errorf("página repetida incompleta: %d lançamentos em %d páginas, %v", len(ls), lidas, err)
	}
	// O servidor limita a página abaixo de 100: a leitura segue até a página vazia (#867).
	ls, err = BuscarLancamentosExtrato(context.Background(), paginasGetter{total: 120, tamanho: 50, paginasLidas: &lidas}, 1, 2026, 9)
	if err != nil || len(ls) != 120 || lidas != 4 {
		t.Errorf("página limitada a 50: %d lançamentos em %d páginas, %v", len(ls), lidas, err)
	}
	// Páginas cheias sem fim: erro no limite, em vez de laço infinito ou lista truncada.
	_, err = BuscarLancamentosExtrato(context.Background(), paginasGetter{infinito: true, paginasLidas: &lidas}, 1, 2026, 9)
	if err == nil || !strings.Contains(err.Error(), "mais de 100 páginas") || lidas != maxPaginasExtrato {
		t.Errorf("sem fim: %d páginas, %v", lidas, err)
	}
}

// semIDGetter devolve n lançamentos sem a chave idLancamentoUsuario.
type semIDGetter struct{ n int }

func (g semIDGetter) GetJSON(_ context.Context, path string, v any) error {
	if !strings.HasSuffix(path, "&pagina=1") {
		return json.Unmarshal([]byte(`[]`), v)
	}
	itens := make([]string, g.n)
	for i := range itens {
		itens[i] = fmt.Sprintf(`{"data":1789700400000,"descricao":"L%d","valor":10}`, i)
	}
	return json.Unmarshal([]byte("["+strings.Join(itens, ",")+"]"), v)
}

func TestBuscarLancamentosExtratoSemID(t *testing.T) {
	for _, n := range []int{3, RegistrosPorPaginaExtrato} {
		ls, err := BuscarLancamentosExtrato(context.Background(), semIDGetter{n: n}, 7, 2026, 9)
		if err == nil || !strings.Contains(err.Error(), "sem idLancamentoUsuario") || ls != nil {
			t.Errorf("%d itens sem id: %d lançamentos, %v", n, len(ls), err)
		}
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
