package cli

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

// writeServer responde ao GET de conferência da sessão com appbarStatus e entrega as
// demais requisições a handler, contando quantas chegaram.
func writeServer(t *testing.T, appbarStatus int, handler http.HandlerFunc) *atomic.Int32 {
	t.Helper()
	var writes atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/plataforma/"+api.PathAppBar {
			w.WriteHeader(appbarStatus)
			io.WriteString(w, "{}")
			return
		}
		writes.Add(1)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	withSession(t, srv)
	return &writes
}

func testSender() sessionSender {
	return sessionSender{s: streams{in: strings.NewReader(""), out: io.Discard, err: io.Discard}, origem: origemEscrita{comando: "teste"}}
}

func TestSenderJSON(t *testing.T) {
	writes := writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != "PUT" || r.URL.Path != "/api/plataforma/movimentacao-financeira/classificar" ||
			r.Header.Get("Content-Type") != "application/json" || string(body) != `{"id":7}` {
			t.Errorf("requisição inesperada: %s %s %q %s", r.Method, r.URL.Path, r.Header.Get("Content-Type"), body)
		}
		io.WriteString(w, `{"id":7,"situacao":"CLASSIFICADO"}`)
	})
	var resp struct {
		Situacao string `json:"situacao"`
	}
	err := testSender().Send(context.Background(), "PUT", "movimentacao-financeira/classificar",
		map[string]int{"id": 7}, &resp)
	if err != nil || resp.Situacao != "CLASSIFICADO" || writes.Load() != 1 {
		t.Fatalf("err %v, resp %+v, %d escrita(s)", err, resp, writes.Load())
	}
}

func TestSenderTextAndEmptyBody(t *testing.T) {
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "" || r.ContentLength > 0 {
			t.Errorf("DELETE sem corpo enviou Content-Type %q, %d bytes", r.Header.Get("Content-Type"), r.ContentLength)
		}
		io.WriteString(w, "OK")
	})
	var txt string
	if err := testSender().Send(context.Background(), "DELETE", "caixa/lancamentousuario/remover/2026/9/1", nil, &txt); err != nil || txt != "OK" {
		t.Fatalf("err %v, texto %q", err, txt)
	}
}

func TestSenderMultipart(t *testing.T) {
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		f, h, err := r.FormFile("arquivo")
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(f)
		if r.FormValue("competencia") != "2026-09" || h.Filename != "extrato.ofx" || string(data) != "OFX" {
			t.Errorf("formulário inesperado: %v %q %q", r.MultipartForm.Value, h.Filename, data)
		}
	})
	err := testSender().SendMultipart(context.Background(), "POST", "upload-documentos/extrato/enviar", []api.Campo{
		{Nome: "competencia", Valor: "2026-09"},
		{Nome: "arquivo", Arquivo: "extrato.ofx", Conteudo: []byte("OFX")},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

// Uma escrita nunca é repetida: nem com 5xx, nem com 401, nem com erro de rede.
func TestSenderSingleAttempt(t *testing.T) {
	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{"5xx", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}, "erro no servidor da Contabilizei (HTTP 502)"},
		{"401", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}, "a escrita não foi aplicada"},
		{"rede", func(w http.ResponseWriter, _ *http.Request) {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				conn.Close()
			}
		}, "pode ou não ter sido aplicada"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := writeServer(t, http.StatusOK, tc.handler)
			t.Setenv("CTBZ_OTP_CMD", "echo 123456") // re-login disponível, mas não pode ser usado depois do envio
			err := testSender().Send(context.Background(), "POST", "caixa/lancamentousuario/novo/", map[string]int{"id": 1}, nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("erro = %v, quero %q", err, tc.want)
			}
			if n := writes.Load(); n != 1 {
				t.Errorf("%d requisições de escrita, quero 1", n)
			}
		})
	}
}

func TestSenderExpiredSessionBeforeSend(t *testing.T) {
	writes := writeServer(t, http.StatusUnauthorized, func(http.ResponseWriter, *http.Request) {})
	err := testSender().Send(context.Background(), "POST", "caixa/lancamentousuario/novo/", map[string]int{}, nil)
	if !errors.Is(err, ctbz.ErrUnauthorized) {
		t.Errorf("erro = %v, quero sessão expirada", err)
	}
	if writes.Load() != 0 {
		t.Errorf("escrita enviada com a sessão expirada")
	}
}

func TestAPICommandWriteUsesSender(t *testing.T) {
	writes := writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != "POST" || string(body) != `{"a":1}` || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("requisição inesperada: %s %q %s", r.Method, r.Header.Get("Content-Type"), body)
		}
		w.WriteHeader(http.StatusUnauthorized)
	})
	t.Setenv("CTBZ_OTP_CMD", "echo 123456")
	_, stderr, code := execCLI(t, "", "api", "-X", "post", "-d", `{"a":1}`, "caminho/qualquer")
	if code != ExitError || !strings.Contains(stderr, "a escrita não foi aplicada") {
		t.Errorf("código %d, stderr %q", code, stderr)
	}
	if writes.Load() != 1 {
		t.Errorf("%d requisições de escrita, quero 1", writes.Load())
	}
}
