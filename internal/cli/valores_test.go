package cli

import "testing"

func TestParseValor(t *testing.T) {
	for in, want := range map[string]int64{
		"150,25": 15025, "1.234,56": 123456, "150.25": 15025, "150": 15000, "0,5": 50, "R$ 10,00": 1000,
		"-3,10": -310, " 7.5 ": 750, "1.234.567,89": 123456789,
	} {
		got, err := parseValor(in)
		if err != nil || got != want {
			t.Errorf("parseValor(%q) = %d, %v; quero %d", in, got, err, want)
		}
	}
	for _, in := range []string{"", "abc", "1,234,5", "1.2.3", "10,123", "--1", "1e3", ",5"} {
		if v, err := parseValor(in); err == nil {
			t.Errorf("parseValor(%q) = %d, quero erro", in, v)
		}
	}
}

func TestCentavos(t *testing.T) {
	for v, want := range map[float64]int64{150.25: 15025, -141.83: -14183, 0.1 + 0.2: 30, 1000: 100000} {
		if got := centavos(v); got != want {
			t.Errorf("centavos(%v) = %d, quero %d", v, got, want)
		}
	}
	if reais(-15025) != -150.25 {
		t.Error("reais")
	}
	for c, want := range map[int64]string{15025: "R$ 150,25", -123456789: "-R$ 1.234.567,89", 5: "R$ 0,05"} {
		if got := formatarCentavos(c); got != want {
			t.Errorf("formatarCentavos(%d) = %s, quero %s", c, got, want)
		}
	}
}
