package cli

import "testing"

func TestValidarDocumento(t *testing.T) {
	for in, want := range map[string]string{
		"00.000.000/0001-91": "00000000000191", // Banco do Brasil
		"12ABC34501DE35":     "12ABC34501DE35", // exemplo da Receita (CNPJ alfanumérico)
		"12abc34501de35":     "12ABC34501DE35",
		"529.982.247-25":     "52998224725",
	} {
		got, err := validarDocumento(in)
		if err != nil || got != want {
			t.Errorf("validarDocumento(%q) = %q, %v; quero %q", in, got, err, want)
		}
	}
	for _, in := range []string{"00000000000192", "11111111111", "11111111111111", "52998224724", "123", "12ABC34501DE3A", ""} {
		if got, err := validarDocumento(in); err == nil {
			t.Errorf("validarDocumento(%q) = %q, quero erro", in, got)
		}
	}
}
