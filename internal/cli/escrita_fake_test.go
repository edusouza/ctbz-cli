package cli

import (
	"io"
	"net/http"
	"sync"
	"testing"
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
		w.WriteHeader(f.status)
		io.WriteString(w, f.resposta)
	})
	return f
}
