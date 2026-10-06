package cli

import (
	"io"
	"strings"
	"testing"
)

const pathDadosLogin = "/api/plataforma/conta-usuario/init"

func contaFake(t *testing.T) *escritaFake {
	t.Helper()
	t.Setenv("CTBZ_OTP_CMD", "")
	f := newEscritaFake(t, map[string]string{pathDadosLogin: fixture(t, "dados_login")})
	f.resposta = `{"tempo":60}`
	return f
}

func TestConta(t *testing.T) {
	contaFake(t)
	out, stderr, code := execCLI(t, "", "conta", "-o", "csv")
	if code != ExitOK || out != "email,telefone,metodo_2fa\nfu****@example.com,(**)*****-1234,EMAIL\n" {
		t.Errorf("código %d, %s\n%s", code, stderr, out)
	}
}

func TestContaAlterarComOTPCmd(t *testing.T) {
	f := contaFake(t)
	out, stderr, code := execCLI(t, "", "conta", "alterar", "--email", "Novo@Example.com", "--telefone", "(11)98765-4321", "--otp-cmd", "echo 123456", "--yes", "-o", "csv")
	want := []string{
		`POST /api/plataforma/conta-usuario/enviar-token-otp {"metodoEnvio":"EMAIL"}`,
		`POST /api/plataforma/conta-usuario/alterar-dados {"email":"novo@example.com","telefone":"11987654321","codigo":"123456","metodo":"EMAIL"}`,
	}
	if code != ExitOK || strings.Join(f.writes, "\n") != strings.Join(want, "\n") {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "Trocar o login de fu****@example.com / (**)*****-1234 para novo@example.com / (11)98765-4321 (risco alto)") || !strings.Contains(stderr, "CTBZ_USER") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if !strings.HasPrefix(out, "acao,situacao,id,email,telefone,metodo_2fa\nalterar,alterado,,") {
		t.Errorf("saída:\n%s", out)
	}
}

func TestContaSenhaNoTerminal(t *testing.T) {
	f := contaFake(t)
	old := isTerminal
	isTerminal = func(io.Reader) bool { return true }
	t.Cleanup(func() { isTerminal = old })
	// Senha, confirmação, "confirmo" e o código, uma linha cada.
	out, stderr, code := execCLI(t, "Nova#Senha1\nNova#Senha1\nconfirmo\n12345\n654321\n", "conta", "senha", "--via", "sms", "-o", "csv")
	want := []string{
		`POST /api/plataforma/conta-usuario/enviar-token-otp {"metodoEnvio":"SMS"}`,
		`POST /api/plataforma/conta-usuario/alterar-senha {"senha":"Nova#Senha1","codigo":"654321","metodo":"SMS"}`,
	}
	if code != ExitOK || strings.Join(f.writes, "\n") != strings.Join(want, "\n") {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if strings.Contains(stderr, "Nova#Senha1") || strings.Contains(out, "Nova#Senha1") || !strings.Contains(stderr, "código inválido: digite os 6 dígitos") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if out != "acao,situacao,id\nsenha,alterada,\n" {
		t.Errorf("saída:\n%s", out)
	}
}

func TestContaCodigoIncorretoEEspera(t *testing.T) {
	f := contaFake(t)
	f.onWrite = func(f *escritaFake) {
		if len(f.writes) == 2 {
			f.status = 404
		}
	}
	_, stderr, code := execCLI(t, "", "conta", "alterar", "--email", "a@example.com", "--telefone", "11987654321", "--otp-cmd", "echo 111111", "--yes")
	if code != ExitError || !strings.Contains(stderr, "o código está incorreto ou inválido; nada foi alterado") {
		t.Errorf("404: código %d, %s", code, stderr)
	}
	f = contaFake(t)
	f.status, f.resposta = 422, `{"error":42}`
	_, stderr, code = execCLI(t, "", "conta", "alterar", "--email", "a@example.com", "--telefone", "11987654321", "--otp-cmd", "echo 111111", "--yes")
	if code != ExitError || len(f.writes) != 1 || !strings.Contains(stderr, "aguarde 42 segundos") {
		t.Errorf("422: código %d, %q, %s", code, f.writes, stderr)
	}
}

func TestContaDryRunEApp(t *testing.T) {
	f := newEscritaFake(t, map[string]string{pathDadosLogin: `{"email":"x","telefone":"y","metodo":"APP"}`})
	out, _, code := execCLI(t, "Nova#Senha1\nNova#Senha1\n", "conta", "senha", "--dry-run")
	// No app autenticador não há envio de código; a senha e o código aparecem como ***.
	if code != ExitOK || len(f.writes) != 0 || strings.Contains(out, "enviar-token-otp") || strings.Contains(out, "Nova#Senha1") ||
		!strings.Contains(out, `"senha": "***"`) || !strings.Contains(out, `"codigo": "***"`) || !strings.Contains(out, `"metodo": "APP"`) {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), out)
	}
}

func TestContaValidacoes(t *testing.T) {
	for _, tc := range []struct {
		stdin string
		args  []string
		want  string
	}{
		{"", []string{"alterar", "--email", "a@example.com"}, "informe --email e --telefone"},
		{"", []string{"alterar", "--email", "a@", "--telefone", "11987654321"}, "e-mail inválido"},
		{"", []string{"alterar", "--email", "a@example.com", "--telefone", "1198765432"}, "telefone inválido"},
		{"", []string{"alterar", "--email", "a@example.com", "--telefone", "11987654321", "--via", "pombo"}, "--via deve ser email ou sms"},
		{"", []string{"alterar", "--email", "a@example.com", "--telefone", "11987654321"}, "sem terminal, informe --otp-cmd"},
		{"curta\n", []string{"senha"}, "pelo menos 8 caracteres"},
		{"semespecial1A\n", []string{"senha"}, "caractere especial"},
		{"Nova#Senha1\nOutra#Senha1\n", []string{"senha"}, "as senhas não conferem"},
	} {
		f := contaFake(t)
		_, stderr, code := execCLI(t, tc.stdin, append([]string{"conta"}, append(tc.args, "--yes")...)...)
		if code != ExitUsage || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}
