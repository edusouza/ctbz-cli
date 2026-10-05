package ctbz

import (
	"strings"
	"testing"
)

func TestNewWriteError(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   string
	}{
		{400, "Competência fechada", "requisição recusada: Competência fechada"},
		{400, `{"detalhes":[{"identificador":"exception/erro-negocial-001","detalhe":"Já existe um lançamento"}]}`,
			"requisição recusada: Já existe um lançamento"},
		{403, `{"message":"Não é possível realizar a assinatura como admin!"}`,
			"sem permissão: Não é possível realizar a assinatura como admin!"},
		{404, "<!DOCTYPE html><html>…</html>", "não encontrado: o servidor respondeu uma página HTML (caminho incompleto ou inexistente)"},
		{406, `{"detalhes":[{"identificador":"exception/erro-negocial-406","detalhe":"Já classificado"}]}`,
			"dados recusados pelo servidor: Já classificado"},
		{422, `{"error":"inválido"}`, "dados recusados pelo servidor: inválido"},
		{422, `{"error":30}`, `dados recusados pelo servidor: {"error":30}`},
		{560, `{"detalhes":[{"detalhe":"Empresa não possui guias para simulação"}]}`,
			"recusado pela regra de negócio da Contabilizei: Empresa não possui guias para simulação"},
		{500, "", "erro no servidor da Contabilizei"},
		{502, "<html>bad gateway</html>", "erro no servidor da Contabilizei: o servidor respondeu uma página HTML"},
		{400, `"texto como string JSON"`, "requisição recusada: texto como string JSON"},
		{302, "", "redirecionado pelo servidor (sessão inválida?)"},
	} {
		err := NewWriteError("POST", "caixa/x", &Response{Status: tc.status, Body: []byte(tc.body)})
		if err.Message != tc.want {
			t.Errorf("%d %s:\n got %q\nwant %q", tc.status, tc.body, err.Message, tc.want)
		}
	}
}

func TestWriteErrorTruncates(t *testing.T) {
	err := NewWriteError("PUT", "a/b", &Response{Status: 400, Body: []byte(strings.Repeat("x ", 500))})
	if n := len([]rune(err.Message)); n > 330 {
		t.Errorf("mensagem com %d caracteres", n)
	}
	if got := err.Error(); !strings.HasPrefix(got, "PUT a/b: requisição recusada: x x") || !strings.HasSuffix(got, "… (HTTP 400)") {
		t.Errorf("Error() = %q", got)
	}
}

func TestWriteErrorIdentificador(t *testing.T) {
	err := NewWriteError("POST", "x", &Response{Status: 400, Body: []byte(`{"detalhes":[{"identificador":"exception/movimentacao-financeira-901","detalhe":"Só OFX"}]}`)})
	if err.Identificador != "exception/movimentacao-financeira-901" || err.Message != "requisição recusada: Só OFX" {
		t.Errorf("erro = %+v", err)
	}
	if NewWriteError("POST", "x", &Response{Status: 400, Body: []byte("texto")}).Identificador != "" {
		t.Error("texto puro não tem identificador")
	}
}
