package cli

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// lerSegredo pede um segredo sem eco no terminal. Fora de um terminal (scripts), lê uma
// linha da entrada padrão. O valor nunca vai para argumentos, logs nem acoes.jsonl.
func lerSegredo(s streams, rotulo string) (string, error) {
	fmt.Fprint(s.err, rotulo)
	if f, ok := s.in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(s.err)
		return string(b), err
	}
	return lerLinha(s), nil
}

// lerTexto pede um valor visível; vazio devolve atual.
func lerTexto(s streams, rotulo, atual string) string {
	fmt.Fprint(s.err, rotulo)
	if v := strings.TrimSpace(lerLinha(s)); v != "" {
		return v
	}
	return atual
}
