package ctbz

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTracerRecebeATroca(t *testing.T) {
	var recebido string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		recebido = string(b)
		http.SetCookie(w, &http.Cookie{Name: "novo", Value: "valor"})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Cookies["sessao"] = &http.Cookie{Name: "sessao", Value: "abc"}
	var trocas []Exchange
	ctx := WithTracer(context.Background(), func(e Exchange) { trocas = append(trocas, e) })
	if _, err := c.Send(ctx, "POST", "x/y", strings.NewReader(`{"a":1}`), "application/json"); err != nil {
		t.Fatal(err)
	}
	if recebido != `{"a":1}` {
		t.Errorf("o servidor recebeu %q: o tracer não pode consumir o corpo", recebido)
	}
	if len(trocas) != 1 {
		t.Fatalf("%d trocas, quero 1", len(trocas))
	}
	e := trocas[0]
	if e.Method != "POST" || e.URL != srv.URL+"/api/plataforma/x/y" || e.Status != 201 || e.Err != nil {
		t.Errorf("troca %+v", e)
	}
	if e.Header.Get("Cookie") != "sessao=abc" || e.Header.Get("Content-Type") != "application/json" {
		t.Errorf("cabeçalhos enviados %v", e.Header)
	}
	if string(e.Body) != `{"a":1}` || string(e.RespBody) != `{"ok":true}` {
		t.Errorf("corpos %q / %q", e.Body, e.RespBody)
	}
	if !strings.Contains(e.RespHeader.Get("Set-Cookie"), "novo=valor") {
		t.Errorf("cabeçalhos recebidos %v", e.RespHeader)
	}
	if e.Elapsed <= 0 {
		t.Errorf("duração %v", e.Elapsed)
	}
}

func TestTracerRecebeOErroDeConexao(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close()
	var trocas []Exchange
	ctx := WithTracer(context.Background(), func(e Exchange) { trocas = append(trocas, e) })
	if _, err := NewClient(srv.URL).Do(ctx, "GET", "/x", nil, nil); err == nil {
		t.Fatal("quero erro de conexão")
	}
	if len(trocas) != 1 || trocas[0].Err == nil || trocas[0].Status != 0 {
		t.Errorf("trocas %+v", trocas)
	}
}

func TestSemTracerNadaMuda(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Write(b)
	}))
	defer srv.Close()
	resp, err := NewClient(srv.URL).Do(context.Background(), "POST", "/eco", strings.NewReader("corpo"), nil)
	if err != nil || string(resp.Body) != "corpo" {
		t.Fatalf("%v %q", err, resp.Body)
	}
}
