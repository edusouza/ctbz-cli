package cli

import (
	"fmt"
	"strings"
)

// normalizarDocumento tira máscara (pontos, barra, hífen, espaços) e põe letras em maiúsculas
// (CNPJ alfanumérico).
func normalizarDocumento(s string) string {
	return strings.ToUpper(strings.NewReplacer(".", "", "/", "", "-", "", " ", "").Replace(strings.TrimSpace(s)))
}

// validarDocumento confere CPF (11 dígitos) ou CNPJ (14 posições, inclusive o alfanumérico)
// pelos dígitos verificadores e devolve o documento normalizado.
func validarDocumento(s string) (string, error) {
	d := normalizarDocumento(s)
	switch {
	case len(d) == 11 && soDigitos(d) && dvCPF(d):
		return d, nil
	case len(d) == 14 && dvCNPJ(d):
		return d, nil
	}
	return "", fmt.Errorf("documento inválido %q: informe um CPF ou CNPJ válido", s)
}

func soDigitos(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func repetido(s string) bool { return strings.Count(s, s[:1]) == len(s) }

func dvCPF(d string) bool {
	if repetido(d) {
		return false
	}
	for _, n := range []int{9, 10} {
		soma := 0
		for i := 0; i < n; i++ {
			soma += int(d[i]-'0') * (n + 1 - i)
		}
		dv := (soma * 10) % 11 % 10
		if dv != int(d[n]-'0') {
			return false
		}
	}
	return true
}

// dvCNPJ valida CNPJ numérico e alfanumérico: cada posição vale o código ASCII menos 48, e os
// dois últimos são dígitos verificadores numéricos.
func dvCNPJ(d string) bool {
	if !soDigitos(d[12:]) || repetido(d) {
		return false
	}
	for _, c := range d[:12] {
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'Z') {
			return false
		}
	}
	pesos := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	for _, n := range []int{12, 13} {
		soma := 0
		for i := 0; i < n; i++ {
			soma += int(d[i]-'0') * pesos[len(pesos)-n+i]
		}
		dv := 11 - soma%11
		if dv >= 10 {
			dv = 0
		}
		if dv != int(d[n]-'0') {
			return false
		}
	}
	return true
}
