package cli

import (
	"strings"
	"testing"
)

const (
	pathUsuarios           = "/api/multiusuario/usuarios-e-invites/consultar"
	pathStatusMultiusuario = "/api/multiusuario/status/servico"
)

func usuariosFake(t *testing.T, status string) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{pathUsuarios: fixture(t, "usuarios_empresa"), pathStatusMultiusuario: status})
	f.onWrite = func(f *escritaFake) {
		f.gets[pathUsuarios] = strings.Replace(f.gets[pathUsuarios], `"status": "INATIVO"`, `"status": "ATIVO"`, 1)
		f.gets[pathUsuarios] = strings.Replace(f.gets[pathUsuarios], `]`, `,{"id":104,"email":"nova@example.com","tipo":"USUARIO_SECUNDARIO","status":"CONVITE_ENVIADO"}]`, 1)
	}
	return f
}

func TestUsuarios(t *testing.T) {
	usuariosFake(t, fixture(t, "status_multiusuario"))
	out, stderr, code := execCLI(t, "", "usuarios", "-o", "csv")
	want := "id,nome,email,tipo,status\n101,FULANO DE TAL,fulano@example.com,USUARIO_PRINCIPAL,ATIVO\n" +
		"102,BELTRANO DE TAL,beltrano@example.com,USUARIO_SECUNDARIO,INATIVO\n103,,convidado@example.com,USUARIO_SECUNDARIO,CONVITE_ENVIADO\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s", code, stderr, out)
	}
}

func TestUsuariosConvidar(t *testing.T) {
	f := usuariosFake(t, fixture(t, "status_multiusuario"))
	out, stderr, code := execCLI(t, "", "usuarios", "convidar", "Nova@Example.com", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `POST /api/multiusuario/invites/enviar {"email":"nova@example.com"}` {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "(risco alto)") || !strings.Contains(stderr, "inclusive financeiras") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if out != "acao,situacao,id,email\nconvidar,CONVITE_ENVIADO,104,nova@example.com\n" {
		t.Errorf("saída:\n%s", out)
	}
}

func TestUsuariosConvidarRecusas(t *testing.T) {
	for _, tc := range []struct {
		email, status string
		code          int
		want          string
	}{
		{"nao-e-email", `{"ativo":true,"limite":2}`, ExitUsage, "e-mail inválido"},
		{"Fulano <f@example.com>", `{"ativo":true,"limite":2}`, ExitUsage, "e-mail inválido"},
		{"a@localhost", `{"ativo":true,"limite":2}`, ExitUsage, "e-mail inválido"},
		{"suporte@contabilizei.com.br", `{"ativo":true,"limite":2}`, ExitUsage, "e-mails da Contabilizei"},
		{"nova@example.com", `{"ativo":false,"limite":2}`, ExitError, "indisponível"},
		{"beltrano@example.com", `{"ativo":true,"limite":2}`, ExitError, "já está na lista de usuários (INATIVO)"},
		{"nova@example.com", `{"ativo":true,"limite":0}`, ExitError, "número máximo de 0 usuários"},
	} {
		f := usuariosFake(t, tc.status)
		_, stderr, code := execCLI(t, "", "usuarios", "convidar", tc.email, "--yes")
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%s: código %d, %d escritas, %q (quero %q)", tc.email, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestUsuariosAtivarDesativar(t *testing.T) {
	f := usuariosFake(t, fixture(t, "status_multiusuario"))
	out, stderr, code := execCLI(t, "", "usuarios", "ativar", "102", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `PUT /api/multiusuario/usuarios/ativar {"idUsuarioEmpresa":102,"ativo":true}` {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "Ativar o acesso de BELTRANO DE TAL (beltrano@example.com, INATIVO) (risco médio)") || out != "acao,situacao,id\nativar,ATIVO,102\n" {
		t.Errorf("stderr:\n%s\nsaída:\n%s", stderr, out)
	}

	f = usuariosFake(t, fixture(t, "status_multiusuario"))
	_, stderr, code = execCLI(t, "", "usuarios", "desativar", "103", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `"ativo":false`) || !strings.Contains(stderr, "não se sabe se desativar cancela") {
		t.Errorf("desativar convite: código %d, %q\n%s", code, f.writes, stderr)
	}

	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"desativar", "101"}, ExitError, "não é possível desativar o administrador"},
		{[]string{"ativar", "103"}, ExitError, "o convite ainda não foi aceito"},
		{[]string{"ativar", "999"}, ExitError, "usuário não encontrado"},
		{[]string{"desativar", "102"}, ExitOK, "já está INATIVO"},
	} {
		f := usuariosFake(t, fixture(t, "status_multiusuario"))
		_, stderr, code := execCLI(t, "", append([]string{"usuarios"}, append(tc.args, "--yes")...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}
