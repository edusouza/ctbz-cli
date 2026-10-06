package api

import (
	"encoding/json"
	"testing"
)

func TestComQuantidades(t *testing.T) {
	p := ProdutoBruto{Bruto: json.RawMessage(`{"id":11,"quantidadeTotal":5,"quantidadeConsumo":5,"extra":{"a":[1]}}`)}
	got, err := p.ComQuantidades(Quantidades{Estoque: 2, Consumo: 2.5, Prestacao: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"extra":{"a":[1]},"id":11,"quantidadeAtivo":0,"quantidadeConsumo":2.5,"quantidadeEstoque":2,"quantidadeInsumo":0,"quantidadePrestacao":0.5,"quantidadeTotal":5}`
	if string(got) != want {
		t.Errorf("got %s\nwant %s", got, want)
	}
	q := Quantidades{Estoque: 2, Consumo: 2.5, Prestacao: 0.5}
	if q.Soma() != 5 || q.Negativa() || !(Quantidades{Insumo: -1}).Negativa() {
		t.Errorf("Soma/Negativa: %v %v", q.Soma(), q.Negativa())
	}
}
