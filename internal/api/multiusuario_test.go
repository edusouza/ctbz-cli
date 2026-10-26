package api

import (
	"encoding/json"
	"testing"
)

func TestUsuarioEmpresaIDTexto(t *testing.T) {
	// O id bruto vira o texto da lista; o que não é número nem texto fica vazio (#818).
	for bruto, want := range map[string]string{
		``: "", `null`: "", `true`: "", `{}`: "", `[1]`: "", `""`: "",
		`102`: "102", `9007199254740993`: "9007199254740993", `-1`: "-1",
		`"abc-7"`: "abc-7", `"a\"b"`: `a"b`, ` 7 `: "7",
	} {
		if got := (UsuarioEmpresa{ID: json.RawMessage(bruto)}).IDTexto(); got != want {
			t.Errorf("%q: %q, quero %q", bruto, got, want)
		}
	}
}
