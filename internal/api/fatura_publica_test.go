package api

import "testing"

// Só o caminho da segunda via (dashboard/nova-fatura/AAAAMM) é aceito: qualquer outra ação
// "request:" do card seria enviada sob o resumo "gerar a segunda via" (#929).
func TestCaminhoSegundaVia(t *testing.T) {
	for _, c := range []struct {
		valor string
		ok    bool
	}{
		{"dashboard/nova-fatura/202610", true},
		{"dashboard/nova-fatura/202601", true},
		{"dashboard/nova-fatura/202612", true},
		{"dashboard/conta", false},
		{"dashboard/pagar-com-cartao", false},
		{"dashboard/nova-fatura", false},
		{"dashboard/nova-fatura/", false},
		{"dashboard/nova-fatura/20261", false},
		{"dashboard/nova-fatura/2026100", false},
		{"dashboard/nova-fatura/202613", false},
		{"dashboard/nova-fatura/202600", false},
		{"dashboard/nova-fatura/202610/extra", false},
		{"/dashboard/nova-fatura/202610", false},
		{"dashboard/nova-fatura/202610?x=1", false},
		{"dashboard/../conta", false},
		{"https://example.com/x", false},
		{"", false},
	} {
		got, err := CaminhoSegundaVia(c.valor)
		if (err == nil) != c.ok {
			t.Errorf("CaminhoSegundaVia(%q) = %q, %v; ok esperado %v", c.valor, got, err, c.ok)
		}
		if c.ok && got != c.valor {
			t.Errorf("CaminhoSegundaVia(%q) = %q", c.valor, got)
		}
	}
}
