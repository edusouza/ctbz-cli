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
		{"suporte@mail.contabilizei.com.br", `{"ativo":true,"limite":2}`, ExitUsage, "e-mails da Contabilizei"},
		{"Suporte@CONTABILIZEI.com", `{"ativo":true,"limite":2}`, ExitUsage, "e-mails da Contabilizei"},
		{`"a b"@example.com`, `{"ativo":true,"limite":2}`, ExitUsage, "e-mail inválido"},
		{"a@[10.0.0.1]", `{"ativo":true,"limite":2}`, ExitUsage, "e-mail inválido"},
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

func TestUsuariosSemControle(t *testing.T) {
	// Textos da API sem caracteres de controle no resumo, nos avisos e na tabela (#816).
	f := usuariosFake(t, fixture(t, "status_multiusuario"))
	f.gets[pathUsuarios] = `[{"id":102,"nome":"\u001b[2K\rFORJADO","email":"b@example.com\u001b[1A","tipo":"USUARIO_SECUNDARIO","status":"INATIVO\u0007"}]`
	f.onWrite = nil
	out, stderr, code := execCLI(t, "", "usuarios", "-o", "csv")
	if code != ExitOK || strings.ContainsAny(out, "\x1b\r\a") || !strings.Contains(out, "102,[2KFORJADO,b@example.com[1A,USUARIO_SECUNDARIO,INATIVO\n") {
		t.Errorf("lista: código %d, %q\n%s", code, out, stderr)
	}
	_, stderr, code = execCLI(t, "", "usuarios", "ativar", "102", "--yes")
	if code != ExitOK || len(f.writes) != 1 || strings.ContainsAny(stderr, "\x1b\r\a") || !strings.Contains(stderr, "Ativar o acesso de [2KFORJADO (b@example.com[1A, INATIVO)") {
		t.Errorf("ativar: código %d, %q", code, stderr)
	}
	_, stderr, code = execCLI(t, "", "usuarios", "desativar", "102", "--yes")
	if code != ExitOK || len(f.writes) != 1 || strings.ContainsAny(stderr, "\x1b\r\a") || !strings.Contains(stderr, "já está INATIVO") {
		t.Errorf("desativar: código %d, %q", code, stderr)
	}
}

func TestUsuariosCamposAusentes(t *testing.T) {
	// Usuário sem id, tipo ou status é erro de leitura, nunca uma escrita (#818, ADR-0021).
	for _, tc := range []struct {
		nome, lista string
		args        []string
		want        string
	}{
		{"sem id", `[{"nome":"X","email":"x@example.com","tipo":"USUARIO_SECUNDARIO","status":"INATIVO"}]`, []string{"ativar", ""}, "sem id"},
		{"id nulo", `[{"id":null,"email":"x@example.com","tipo":"USUARIO_SECUNDARIO","status":"ATIVO"}]`, []string{"desativar", ""}, "sem id"},
		{"sem tipo", `[{"id":101,"nome":"X","email":"x@example.com","status":"ATIVO"}]`, []string{"desativar", "101"}, "sem tipo"},
		{"sem tipo ativar", `[{"id":101,"email":"x@example.com","status":"INATIVO"}]`, []string{"ativar", "101"}, "sem tipo"},
		{"sem status", `[{"id":102,"email":"x@example.com","tipo":"USUARIO_SECUNDARIO"}]`, []string{"ativar", "102"}, "sem status"},
		{"sem status desativar", `[{"id":102,"email":"x@example.com","tipo":"USUARIO_SECUNDARIO","status":""}]`, []string{"desativar", "102"}, "sem status"},
		{"tipo desconhecido", `[{"id":105,"email":"x@example.com","tipo":"USUARIO_GESTOR","status":"ATIVO"}]`, []string{"desativar", "105"}, "tipo USUARIO_GESTOR"},
		{"lista vazia", `[]`, []string{"desativar", "102"}, "usuário não encontrado"},
	} {
		for _, flag := range []string{"--yes", "--dry-run"} {
			f := usuariosFake(t, fixture(t, "status_multiusuario"))
			f.gets[pathUsuarios], f.onWrite = tc.lista, nil
			out, stderr, code := execCLI(t, "", append([]string{"usuarios"}, append(tc.args, flag)...)...)
			if code != ExitError || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 || strings.Contains(out+stderr, "idUsuarioEmpresa") {
				t.Errorf("%s %s: código %d, %d escritas, %q\n%s (quero %q)", tc.nome, flag, code, len(f.writes), stderr, out, tc.want)
			}
		}
	}
}

func TestUsuariosIDBruto(t *testing.T) {
	// O id volta exatamente como listado: número acima de 2^53 e texto (#818).
	lista := `[{"id":9007199254740993,"email":"a@example.com","tipo":"USUARIO_SECUNDARIO","status":"INATIVO"},` +
		`{"id":"abc-7","email":"b@example.com","tipo":"USUARIO_SECUNDARIO","status":"ATIVO"}]`
	f := usuariosFake(t, fixture(t, "status_multiusuario"))
	f.gets[pathUsuarios], f.onWrite = lista, nil
	out, stderr, code := execCLI(t, "", "usuarios", "-o", "csv")
	if code != ExitOK || out != "id,nome,email,tipo,status\n9007199254740993,,a@example.com,USUARIO_SECUNDARIO,INATIVO\nabc-7,,b@example.com,USUARIO_SECUNDARIO,ATIVO\n" {
		t.Errorf("lista: código %d, %s\n%s", code, stderr, out)
	}
	for _, tc := range []struct{ args, want string }{
		{"ativar 9007199254740993", `PUT /api/multiusuario/usuarios/ativar {"idUsuarioEmpresa":9007199254740993,"ativo":true}`},
		{"desativar abc-7", `PUT /api/multiusuario/usuarios/ativar {"idUsuarioEmpresa":"abc-7","ativo":false}`},
	} {
		f.writes = nil
		_, stderr, code := execCLI(t, "", append([]string{"usuarios"}, append(strings.Fields(tc.args), "--yes")...)...)
		if code != ExitOK || len(f.writes) != 1 || f.writes[0] != tc.want {
			t.Errorf("%s: código %d, %q\n%s", tc.args, code, f.writes, stderr)
		}
	}
	// O texto com aspas não é o id listado.
	f.writes = nil
	if _, _, code := execCLI(t, "", "usuarios", "desativar", `"abc-7"`, "--yes"); code != ExitError || len(f.writes) != 0 {
		t.Errorf("id com aspas: código %d, %q", code, f.writes)
	}
}

func TestUsuariosJSON(t *testing.T) {
	// O id sai como hoje: texto, inclusive o numérico (ADR-0017).
	usuariosFake(t, fixture(t, "status_multiusuario"))
	out, stderr, code := execCLI(t, "", "usuarios", "-o", "json")
	if code != ExitOK || !strings.Contains(out, `"id": "101"`) || !strings.Contains(out, `"nome": null`) {
		t.Errorf("código %d, %s\n%s", code, stderr, out)
	}
}

func TestUsuariosConvidarLimite(t *testing.T) {
	const principal = `{"id":101,"email":"dono@example.com","tipo":"USUARIO_PRINCIPAL","status":"ATIVO"}`
	for _, tc := range []struct {
		nome, outro    string
		code, escritas int
		want           string
	}{
		// O limite conta os secundários ativos, como o painel (#972).
		{"secundário ativo", `{"id":105,"email":"x@example.com","tipo":"USUARIO_SECUNDARIO","status":"ATIVO"}`, ExitError, 0, "número máximo de 1 usuários"},
		{"secundário inativo", `{"id":105,"email":"x@example.com","tipo":"USUARIO_SECUNDARIO","status":"INATIVO"}`, ExitOK, 1, ""},
		{"convite pendente", `{"id":105,"email":"x@example.com","tipo":"USUARIO_SECUNDARIO","status":"CONVITE_ENVIADO"}`, ExitOK, 1, ""},
		// Tipo ou situação vazia ou desconhecida: não dá para conferir o limite (#818, #972).
		{"tipo desconhecido", `{"id":105,"email":"x@example.com","tipo":"USUARIO_GESTOR","status":"ATIVO"}`, ExitError, 0, `tipo "USUARIO_GESTOR"`},
		{"sem situação", `{"id":106,"email":"x@example.com","tipo":"USUARIO_SECUNDARIO"}`, ExitError, 0, `situação ""`},
		{"sem tipo", `{"id":107,"email":"x@example.com","status":"ATIVO"}`, ExitError, 0, `tipo ""`},
	} {
		f := usuariosFake(t, `{"ativo":true,"limite":1}`)
		f.gets[pathUsuarios] = "[" + principal + "," + tc.outro + "]"
		f.onWrite = nil
		_, stderr, code := execCLI(t, "", "usuarios", "convidar", "nova@example.com", "--yes")
		if code != tc.code || len(f.writes) != tc.escritas || !strings.Contains(stderr, tc.want) {
			t.Errorf("%s: código %d, %d escritas\n%s", tc.nome, code, len(f.writes), stderr)
		}
		if desconhecido := strings.Contains(tc.want, `"`); desconhecido != strings.Contains(stderr, "convide pelo painel") {
			t.Errorf("%s: só a recusa por dado desconhecido manda usar o painel:\n%s", tc.nome, stderr)
		}
	}
}
