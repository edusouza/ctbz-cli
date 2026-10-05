package cli

import (
	"fmt"
	"strconv"
	"strings"
)

// parseValor lê um valor em reais digitado pelo usuário e devolve centavos. Aceita
// "150,25", "1.234,56", "150.25" e "150"; com vírgula, o ponto separa milhares.
// Escritas somam valores em centavos (inteiros), nunca em float64.
func parseValor(s string) (int64, error) {
	t := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "R$"))
	neg := strings.HasPrefix(t, "-")
	t = strings.TrimSpace(strings.TrimPrefix(t, "-"))
	if strings.Contains(t, ",") {
		t = strings.ReplaceAll(t, ".", "")
		t = strings.Replace(t, ",", ".", 1)
	}
	inteiro, frac, _ := strings.Cut(t, ".")
	if inteiro == "" || len(frac) > 2 || strings.ContainsAny(inteiro+frac, "+-.,") {
		return 0, fmt.Errorf("valor inválido %q: use 150,25 ou 150.25", s)
	}
	frac += strings.Repeat("0", 2-len(frac))
	c, err := strconv.ParseInt(inteiro+frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("valor inválido %q: use 150,25 ou 150.25", s)
	}
	if neg {
		c = -c
	}
	return c, nil
}

// centavos converte um valor da API (float64 em reais) para centavos, arredondando.
func centavos(v float64) int64 {
	if v < 0 {
		return -int64(-v*100 + 0.5)
	}
	return int64(v*100 + 0.5)
}

// reais converte centavos para o número enviado à API.
func reais(c int64) float64 { return float64(c) / 100 }

// formatarCentavos mostra centavos como "R$ 1.234,56".
func formatarCentavos(c int64) string {
	sinal := ""
	if c < 0 {
		sinal, c = "-", -c
	}
	inteiro := strconv.FormatInt(c/100, 10)
	var b strings.Builder
	for i, r := range inteiro {
		if i > 0 && (len(inteiro)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return fmt.Sprintf("%sR$ %s,%02d", sinal, b.String(), c%100)
}
