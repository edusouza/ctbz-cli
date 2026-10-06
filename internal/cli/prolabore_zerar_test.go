package cli

import (
	"strings"
	"testing"
)

func zerarFake(t *testing.T, central string) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{pathProlaboreCentral: central})
	f.onWrite = func(f *escritaFake) {
		c := f.gets[pathProlaboreCentral]
		if strings.Contains(c, `"zerarProlabore":false`) {
			f.gets[pathProlaboreCentral] = strings.Replace(c, `"zerarProlabore":false`, `"zerarProlabore":true`, 1)
		} else {
			f.gets[pathProlaboreCentral] = strings.Replace(c, `"zerarProlabore":true`, `"zerarProlabore":false`, 1)
		}
	}
	return f
}

func TestProlaboreZerarOn(t *testing.T) {
	f := zerarFake(t, `{"tipoGerenciamento":"PERSONALIZADO","zerarProlabore":false,"socios":[{"id":1},{"id":2}]}`)
	out, stderr, code := execCLI(t, "", "prolabore", "zerar-sem-faturamento", "on", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `PATCH /api/plataforma/prolabore/central/empresa/zerar-prolabore {"zerarProlabore":true}` {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "(vale para os 2 sócios) (risco alto)") || !strings.Contains(stderr, "não contribuem para o INSS") || strings.Contains(stderr, "gestão inteligente") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if out != "acao,situacao,id,zerar\nzerar-sem-faturamento,ligada,,true\n" {
		t.Errorf("saída:\n%s", out)
	}
}

func TestProlaboreZerarGestaoInteligenteEOff(t *testing.T) {
	f := zerarFake(t, `{"tipoGerenciamento":"INTELIGENTE","zerarProlabore":false,"socios":[{"id":1}]}`)
	_, stderr, code := execCLI(t, "", "prolabore", "zerar-sem-faturamento", "on", "--yes")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `PATCH /api/plataforma/prolabore/central/empresa/zerar-prolabore-gt ` ||
		!strings.Contains(stderr, "vale a partir deste mês") {
		t.Fatalf("GT: código %d, %q\n%s", code, f.writes, stderr)
	}
	_, stderr, code = execCLI(t, "", "prolabore", "zerar-sem-faturamento", "off", "--yes")
	if code != ExitOK || len(f.writes) != 2 || f.writes[1] != `PATCH /api/plataforma/prolabore/central/empresa/zerar-prolabore {"zerarProlabore":false}` ||
		!strings.Contains(stderr, "(risco médio)") {
		t.Errorf("off: código %d, %q\n%s", code, f.writes, stderr)
	}
}

func TestProlaboreZerarSemMudancaEUso(t *testing.T) {
	f := zerarFake(t, `{"zerarProlabore":true,"socios":[]}`)
	out, stderr, code := execCLI(t, "", "prolabore", "zerar-sem-faturamento", "on", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(stderr, "já está ligada") || !strings.Contains(out, "sem mudança") {
		t.Errorf("sem mudança: código %d, %d escritas, %s\n%s", code, len(f.writes), stderr, out)
	}
	if _, _, code := execCLI(t, "", "prolabore", "zerar-sem-faturamento", "sim"); code != ExitUsage {
		t.Errorf("argumento inválido: código %d", code)
	}
}
