package cli

import (
	"strings"
	"testing"
)

const pathDadosAcesso = "/api/plataforma/empresa/dadosacesso/init"

func credenciaisFake(t *testing.T, dados string) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{pathDadosAcesso: dados})
	f.status = 204
	f.onWrite = func(f *escritaFake) {
		f.gets[pathDadosAcesso] = strings.Replace(f.gets[pathDadosAcesso], `"alertaPendenciaCredencial": true`, `"alertaPendenciaCredencial": false`, 1)
	}
	return f
}

func TestCredenciaisMascaradas(t *testing.T) {
	credenciaisFake(t, fixture(t, "dados_acesso"))
	out, stderr, code := execCLI(t, "", "empresa", "credenciais", "-o", "csv")
	want := "codigo_simples,usuario_prefeitura,senha_prefeitura,usuario_dataprev,senha_dataprev,redefinir_senha_prefeitura\npreenchido,usuario-exemplo,preenchido,,,true\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s", code, stderr, out)
	}
	if _, stderr, code := execCLI(t, "", "empresa", "credenciais", "--mostrar"); code != ExitUsage || !strings.Contains(stderr, "só funciona num terminal") {
		t.Errorf("--mostrar sem terminal: código %d, %s", code, stderr)
	}
}

func TestCredenciaisAtualizar(t *testing.T) {
	f := credenciaisFake(t, fixture(t, "dados_acesso"))
	// Uma linha por pergunta: código, usuário da prefeitura (Enter mantém) e senha.
	out, stderr, code := execCLI(t, "123456789012\n\nnova senha\n", "empresa", "credenciais", "atualizar", "--codigo-simples", "--prefeitura", "--yes", "-o", "csv")
	want := `POST /api/plataforma/empresa/dadosacesso/atualizarDadosAcesso {"chaveAcessoSimples":"123456789012","dataValidadeSenhaPrefeituraCuritiba":null,"id":1000000000000001,` +
		`"senhaDataprev":null,"senhaPrefeitura":"nova senha","usuarioDataprev":null,"usuarioPrefeitura":"usuario-exemplo"}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if strings.Contains(stderr, "nova senha") || strings.Contains(stderr, "123456789012") || strings.Contains(out, "nova senha") {
		t.Errorf("segredo na saída:\n%s\n%s", stderr, out)
	}
	if !strings.Contains(stderr, "Atualizar os dados de acesso: código do Simples, usuário e senha da prefeitura (risco alto)") ||
		!strings.Contains(stderr, "Usuário da prefeitura [Enter mantém usuario-exemplo]: ") || !strings.Contains(stderr, "Verificação de acesso iniciada") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if out != "acao,situacao,id,campos\natualizar credenciais,enviado,,\"código do Simples, usuário e senha da prefeitura\"\n" {
		t.Errorf("saída:\n%s", out)
	}
}

func TestCredenciaisAtualizarDryRunMascara(t *testing.T) {
	f := credenciaisFake(t, fixture(t, "dados_acesso"))
	out, _, code := execCLI(t, "usuario-dataprev\nsegredo-dataprev\n", "empresa", "credenciais", "atualizar", "--dataprev", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || strings.Contains(out, "segredo-dataprev") || strings.Contains(out, `"000000000000"`) ||
		!strings.Contains(out, `"senhaDataprev": "***"`) || !strings.Contains(out, `"chaveAcessoSimples": "***"`) || !strings.Contains(out, `"usuarioDataprev": "usuario-dataprev"`) {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), out)
	}
}

func TestCredenciaisAtualizarValidacoes(t *testing.T) {
	for _, tc := range []struct {
		stdin string
		args  []string
		want  string
	}{
		{"", nil, "escolha o que atualizar"},
		{"12345\n", []string{"--codigo-simples"}, "12 caracteres apenas numéricos"},
		{"\nab\n", []string{"--dataprev"}, "usuário e senha do Dataprev precisam de pelo menos 3 caracteres"},
	} {
		f := credenciaisFake(t, fixture(t, "dados_acesso"))
		_, stderr, code := execCLI(t, tc.stdin, append([]string{"empresa", "credenciais", "atualizar", "--yes"}, tc.args...)...)
		if code != ExitUsage || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestCredenciaisConfirmarPrefeitura(t *testing.T) {
	f := credenciaisFake(t, fixture(t, "dados_acesso"))
	out, stderr, code := execCLI(t, "", "empresa", "credenciais", "confirmar-prefeitura", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "PUT /api/plataforma/empresa/dadosacesso/confirmar-credencial-prefeitura-valida " ||
		!strings.Contains(stderr, "(risco médio)") || out != "acao,situacao,id\nconfirmar-prefeitura,confirmada,\n" {
		t.Fatalf("código %d, %q\n%s\n%s", code, f.writes, stderr, out)
	}
	_, stderr, code = execCLI(t, "", "empresa", "credenciais", "confirmar-prefeitura", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(stderr, "não pede para redefinir") {
		t.Errorf("sem alerta: código %d, %d escritas, %s", code, len(f.writes), stderr)
	}
}
