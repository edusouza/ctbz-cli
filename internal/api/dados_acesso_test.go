package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNovoFormularioDadosAcesso(t *testing.T) {
	bruto := json.RawMessage(`{"chaveAcessoSimples":"000000000000","usuarioPrefeitura":"u","senhaPrefeitura":"a","extra":1,"senhaDataprev":null}`)
	got, err := NovoFormularioDadosAcesso(bruto, json.RawMessage("7"), map[string]string{CampoSenhaPrefeitura: "b"})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"chaveAcessoSimples":"000000000000","extra":1,"id":7,"senhaDataprev":null,"senhaPrefeitura":"b","usuarioPrefeitura":"u"}`
	if string(got) != want {
		t.Errorf("got %s\nwant %s", got, want)
	}
	if got, _ := NovoFormularioDadosAcesso(json.RawMessage(`{"id":3}`), json.RawMessage("7"), nil); string(got) != `{"id":3}` {
		t.Errorf("id do formulário: %s", got)
	}
	for _, id := range []json.RawMessage{nil, json.RawMessage(""), json.RawMessage("null")} {
		if _, err := NovoFormularioDadosAcesso(json.RawMessage(`{}`), id, nil); err == nil || !strings.Contains(err.Error(), "id da empresa") {
			t.Errorf("sem id (%q): %v", id, err)
		}
	}
	// O id da sessão vai byte a byte, sem passar por float64 (#827).
	got, err = NovoFormularioDadosAcesso(json.RawMessage(`{"id":null}`), json.RawMessage("9007199254740993"), nil)
	if err != nil || string(got) != `{"id":9007199254740993}` {
		t.Errorf("id grande: %s, %v", got, err)
	}
}
