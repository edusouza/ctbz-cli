package api

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"testing"
)

func TestEncodeBody(t *testing.T) {
	for _, tc := range []struct {
		name string
		body any
		want string
	}{
		{"nil", nil, ""},
		{"bytes", []byte(`{"a":1}`), `{"a":1}`},
		{"raw", json.RawMessage(`[1,2]`), `[1,2]`},
		{"struct", struct {
			ID    int64   `json:"id"`
			Valor float64 `json:"valor"`
		}{7, -150.25}, `{"id":7,"valor":-150.25}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EncodeBody(tc.body)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Errorf("EncodeBody = %s, quero %s", got, tc.want)
			}
		})
	}
	if _, err := EncodeBody(make(chan int)); err == nil {
		t.Error("corpo não serializável deveria falhar")
	}
}

func TestEncodeMultipart(t *testing.T) {
	body, ct, err := EncodeMultipart([]Campo{
		{Nome: "competencia", Valor: "2026-09"},
		{Nome: "arquivo", Arquivo: "extrato.semtipo", Conteudo: []byte("OFXHEADER")},
		{Nome: "pdf", Arquivo: "extrato.PDF", Conteudo: []byte("%PDF")},
		{Nome: "bin", Arquivo: "dados", Conteudo: []byte{0}, Tipo: "text/csv"},
	})
	if err != nil {
		t.Fatal(err)
	}
	mt, params, err := mime.ParseMediaType(ct)
	if err != nil || mt != "multipart/form-data" {
		t.Fatalf("Content-Type %q: %v", ct, err)
	}
	r := multipart.NewReader(strings.NewReader(string(body)), params["boundary"])
	type parte struct{ nome, arquivo, tipo, conteudo string }
	var got []parte
	for {
		p, err := r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(p)
		got = append(got, parte{p.FormName(), p.FileName(), p.Header.Get("Content-Type"), string(data)})
	}
	want := []parte{
		{"competencia", "", "", "2026-09"},
		{"arquivo", "extrato.semtipo", "application/octet-stream", "OFXHEADER"},
		{"pdf", "extrato.PDF", "application/pdf", "%PDF"},
		{"bin", "dados", "text/csv", "\x00"},
	}
	if len(got) != len(want) {
		t.Fatalf("partes = %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parte %d = %+v, quero %+v", i, got[i], want[i])
		}
	}
}

func TestDecodeResponse(t *testing.T) {
	if err := DecodeResponse([]byte("qualquer coisa"), nil); err != nil {
		t.Errorf("v nil: %v", err)
	}
	var s string
	if err := DecodeResponse([]byte("OK"), &s); err != nil || s != "OK" {
		t.Errorf("texto: %q, %v", s, err)
	}
	v := struct {
		ID int64 `json:"id"`
	}{ID: 1}
	if err := DecodeResponse([]byte("  \n"), &v); err != nil || v.ID != 1 {
		t.Errorf("corpo vazio: %+v, %v", v, err)
	}
	if err := DecodeResponse([]byte(`{"id":42}`), &v); err != nil || v.ID != 42 {
		t.Errorf("JSON: %+v, %v", v, err)
	}
	if err := DecodeResponse([]byte(`<html>`), &v); err == nil {
		t.Error("HTML deveria falhar ao decodificar em struct")
	}
}
