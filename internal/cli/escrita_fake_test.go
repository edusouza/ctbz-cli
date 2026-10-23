package cli

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// escritaFake responde GETs fixos (por caminho com query) e grava as escritas recebidas,
// respondendo com status e corpo configuráveis. Serve para os testes de comandos de escrita.
type escritaFake struct {
	mu       sync.Mutex
	gets     map[string]string
	writes   []string // "MÉTODO caminho?query corpo"
	status   int
	resposta string
	onWrite  func(*escritaFake)
	// atraso segura a resposta das escritas cujo caminho contém atrasar (simula timeout).
	atrasar string
	atraso  time.Duration
	// cair derruba a conexão, sem resposta, nas escritas cujo caminho contém o trecho.
	cair string
}

func newEscritaFake(t *testing.T, gets map[string]string) *escritaFake {
	t.Helper()
	f := &escritaFake{gets: gets, status: http.StatusOK}
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Method == http.MethodGet {
			body, ok := f.gets[r.URL.RequestURI()]
			if !ok {
				http.NotFound(w, r)
				return
			}
			io.WriteString(w, body)
			return
		}
		body, _ := io.ReadAll(r.Body)
		f.writes = append(f.writes, r.Method+" "+r.URL.RequestURI()+" "+string(body))
		if f.onWrite != nil {
			f.onWrite(f)
		}
		if f.cair != "" && strings.Contains(r.URL.Path, f.cair) {
			if conn, _, err := w.(http.Hijacker).Hijack(); err == nil {
				conn.Close()
			}
			return
		}
		if f.atraso > 0 && strings.Contains(r.URL.Path, f.atrasar) {
			atraso := f.atraso
			f.mu.Unlock()
			time.Sleep(atraso)
			f.mu.Lock()
		}
		w.WriteHeader(f.status)
		io.WriteString(w, f.resposta)
	})
	return f
}
