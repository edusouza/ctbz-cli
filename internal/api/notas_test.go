package api

import (
	"encoding/json"
	"testing"
)

func TestStatusNotaFormatos(t *testing.T) {
	for _, tc := range []struct {
		name, in string
		want     StatusNota
		rotulo   string
	}{
		{"objeto", `{"descricao":"NFS-e processada com sucesso","text":"Emitida","badgeType":"success"}`,
			StatusNota{Descricao: "NFS-e processada com sucesso", Text: "Emitida", BadgeType: "success"}, "Emitida"},
		{"objeto sem text", `{"descricao":"NFS-e cancelada"}`, StatusNota{Descricao: "NFS-e cancelada"}, "NFS-e cancelada"},
		{"texto", `"AUTORIZADA"`, StatusNota{Text: "AUTORIZADA"}, "AUTORIZADA"},
		{"null", `null`, StatusNota{}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var s StatusNota
			if err := json.Unmarshal([]byte(tc.in), &s); err != nil {
				t.Fatal(err)
			}
			if s != tc.want || s.Rotulo() != tc.rotulo {
				t.Errorf("got %+v (rótulo %q), quero %+v (rótulo %q)", s, s.Rotulo(), tc.want, tc.rotulo)
			}
		})
	}
	var s StatusNota
	if err := json.Unmarshal([]byte(`7`), &s); err == nil {
		t.Error("status numérico deveria falhar")
	}
}

// TestNotaEmitidaStatusObjeto usa o formato de 02/11/2026, em que statusNotaFiscal passou a
// vir como objeto e dataEmissao como array (#1052).
func TestNotaEmitidaStatusObjeto(t *testing.T) {
	data := `{"list":[{"id":1000000000000201,"numero":"5","cpfCnpjTomador":"","nomeRazaoTomador":"TOMADOR NÃO INFORMADO",
"dataEmissao":[2026,1,30,19,1,46],"valorServico":250,"situacaoNota":"PROCESSADO_SUCESSO",
"statusNotaFiscal":{"descricao":"NFS-e processada com sucesso","text":"Emitida","badgeType":"success"},
"errosNotaFiscal":null,"dataEmissaoFormatada":"30/01/2026","permiteReplicacao":false,"tomadorExterior":false}],"total":1}`
	var l ListaNotas
	if err := json.Unmarshal([]byte(data), &l); err != nil {
		t.Fatal(err)
	}
	if len(l.List) != 1 || l.List[0].StatusNotaFiscal.Rotulo() != "Emitida" || l.List[0].SituacaoNota != "PROCESSADO_SUCESSO" {
		t.Errorf("lista: %+v", l)
	}
}
