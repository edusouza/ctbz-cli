package cli

import (
	"strings"
	"testing"
)

const pathProlaboreParametros = "/api/plataforma/prolabore/init"

func gestaoInteligenteFake(t *testing.T, central, parametros string, depois [2]string) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{pathProlaboreCentral: central, pathProlaboreParametros: parametros})
	f.onWrite = func(f *escritaFake) {
		f.gets[pathProlaboreCentral], f.gets[pathProlaboreParametros] = depois[0], depois[1]
	}
	return f
}

func TestGestaoInteligenteAtivar(t *testing.T) {
	f := gestaoInteligenteFake(t, `{"tipoGerenciamento":"PERSONALIZADO","elegivelNoMotor":true}`, `{"isMotorFatorR":false}`,
		[2]string{`{"tipoGerenciamento":"INTELIGENTE","elegivelNoMotor":true}`, `{"isMotorFatorR":true}`})
	out, stderr, code := execCLI(t, "", "prolabore", "gestao-inteligente", "ativar", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "PUT /api/plataforma/prolabore/central/empresa/gestao-inteligente " {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "(risco alto)") || !strings.Contains(stderr, "definir o pró-labore de todos os sócios") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if out != "acao,situacao,id,gestao_inteligente\ngestao-inteligente ativar,ativada,,true\n" {
		t.Errorf("saída:\n%s", out)
	}
}

func TestGestaoInteligenteSair(t *testing.T) {
	f := gestaoInteligenteFake(t, `{"tipoGerenciamento":"PERSONALIZADO"}`, `{"isMotorFatorR":true}`,
		[2]string{`{"tipoGerenciamento":"PERSONALIZADO"}`, `{"isMotorFatorR":true}`})
	out, stderr, code := execCLI(t, "", "prolabore", "gestao-inteligente", "sair", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "PUT /api/plataforma/prolabore/excluir-empresa-motor-fator-r " {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "até o penúltimo dia do mês") {
		t.Errorf("stderr:\n%s", stderr)
	}
	// A releitura ainda mostra o motor: a CLI não afirma a saída.
	if !strings.Contains(out, "gestao-inteligente sair,enviado,,true") {
		t.Errorf("saída:\n%s", out)
	}
}

func TestGestaoInteligenteRecusas(t *testing.T) {
	f := gestaoInteligenteFake(t, `{"tipoGerenciamento":"PERSONALIZADO","elegivelNoMotor":false}`, `{"isMotorFatorR":false}`, [2]string{})
	_, stderr, code := execCLI(t, "", "prolabore", "gestao-inteligente", "ativar", "--yes")
	if code != ExitError || !strings.Contains(stderr, "não é elegível") || len(f.writes) != 0 {
		t.Errorf("não elegível: código %d, %s", code, stderr)
	}
	out, stderr, code := execCLI(t, "", "prolabore", "gestao-inteligente", "sair", "--yes", "-o", "csv")
	if code != ExitOK || !strings.Contains(stderr, "já está fora") || !strings.Contains(out, "sem mudança") || len(f.writes) != 0 {
		t.Errorf("já fora: código %d, %s\n%s", code, stderr, out)
	}
}
