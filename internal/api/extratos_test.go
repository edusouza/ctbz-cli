package api

import (
	"context"
	"testing"
)

func TestBuscarExtratosExibirBotaoImportar(t *testing.T) {
	es, err := BuscarExtratos(context.Background(), fixtureGetter{PathExtratos: "extratos"})
	if err != nil {
		t.Fatal(err)
	}
	botao := map[string]*bool{}
	for i := range es {
		botao[es[i].Situacao] = es[i].ExibirBotaoImportar
	}
	if botao["ABERTO"] != nil || botao["IMPORTADO"] == nil || *botao["IMPORTADO"] || botao["PROCESSANDO"] == nil || !*botao["PROCESSANDO"] {
		t.Errorf("exibirBotaoImportar por situação = %v", botao)
	}
}

func TestExtratoAguardaArquivo(t *testing.T) {
	sim, nao := true, false
	casos := []struct {
		nome string
		e    *Extrato
		want bool
	}{
		{"nil", nil, false},
		{"PROCESSANDO sem o campo", &Extrato{Situacao: "PROCESSANDO"}, false},
		{"PROCESSANDO com false", &Extrato{Situacao: "PROCESSANDO", ExibirBotaoImportar: &nao}, false},
		{"PROCESSANDO com true", &Extrato{Situacao: "PROCESSANDO", ExibirBotaoImportar: &sim}, true},
		{"PENDENTE com true", &Extrato{Situacao: "PENDENTE", ExibirBotaoImportar: &sim}, false},
		{"IMPORTADO com true", &Extrato{Situacao: "IMPORTADO", ExibirBotaoImportar: &sim}, false},
	}
	for _, c := range casos {
		if got := c.e.AguardaArquivo(); got != c.want {
			t.Errorf("%s: AguardaArquivo() = %v, quero %v", c.nome, got, c.want)
		}
	}
}
