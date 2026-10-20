package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNovoFormularioDadosAcesso(t *testing.T) {
	bruto := json.RawMessage(`{"chaveAcessoSimples":"000000000000","usuarioPrefeitura":"u","senhaPrefeitura":"a","extra":1,"senhaDataprev":null}`)
	got, err := NovoFormularioDadosAcesso(bruto, 7, map[string]string{CampoSenhaPrefeitura: "b"})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"chaveAcessoSimples":"000000000000","extra":1,"id":7,"senhaDataprev":null,"senhaPrefeitura":"b","usuarioPrefeitura":"u"}`
	if string(got) != want {
		t.Errorf("got %s\nwant %s", got, want)
	}
	if got, _ := NovoFormularioDadosAcesso(json.RawMessage(`{"id":3}`), 7, nil); string(got) != `{"id":3}` {
		t.Errorf("id do formulário: %s", got)
	}
	if _, err := NovoFormularioDadosAcesso(json.RawMessage(`{}`), nil, nil); err == nil || !strings.Contains(err.Error(), "id da empresa") {
		t.Errorf("sem id: %v", err)
	}
}
