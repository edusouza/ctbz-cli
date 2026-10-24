package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	pathConsultaBB       = "/api/plataforma/novo-emissor/clientes/consulta/00000000000191"
	pathTomadores        = "/api/plataforma/novo-emissor/tomadores/init"
	pathCEPSe            = "/api/plataforma/novo-emissor/cep/logradouro?cep=01001000"
	pathPaises           = "/api/plataforma/notafiscal/emitir/buscarPaisesParaEmissao/"
	pathCadastroClientes = "/api/plataforma/novo-emissor/cadastro-clientes/"
	pathCadastroBB       = "/api/plataforma/novo-emissor/cadastro-clientes/00000000000191"
	pathCadastroExt      = "/api/plataforma/novo-emissor/cadastro-clientes/77"
)

// tomadoresFake responde como o escritaFake, mas o cadastro de um tomador que não está nos
// GETs responde 204, como o emissor faz com um documento que não está cadastrado.
func tomadoresFake(t *testing.T) *escritaFake {
	t.Helper()
	f := &escritaFake{gets: tomadoresGets(t)}
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Method == http.MethodGet {
			body, ok := f.gets[r.URL.RequestURI()]
			switch {
			case ok:
				io.WriteString(w, body)
			case strings.HasPrefix(r.URL.Path, pathCadastroClientes):
				w.WriteHeader(http.StatusNoContent)
			default:
				http.NotFound(w, r)
			}
			return
		}
		body, _ := io.ReadAll(r.Body)
		f.writes = append(f.writes, r.Method+" "+r.URL.RequestURI()+" "+string(body))
		if f.onWrite != nil {
			f.onWrite(f)
		}
	})
	return f
}

func tomadoresGets(t *testing.T) map[string]string {
	t.Helper()
	return map[string]string{
		pathConsultaBB: `{"cnpj":"00000000000191","razaoSocial":"BANCO EXEMPLO SA","nomeFantasia":"","dataAbertura":"01/01/1900","atividadePrincipal":null,"naturezaJuridica":null,
"logradouro":"Rua Exemplo","numero":"1","complemento":"","bairro":"Centro","cep":"01001000","municipio":"São Paulo","uf":"SP","email":"","telefone":"","situacaoCadastral":"ATIVA","optanteSimples":"NÃO"}`,
		pathCEPSe:       fixture(t, "cep"),
		pathPaises:      fixture(t, "paises_emissao"),
		pathCadastroBB:  fixture(t, "cadastro_cliente"),
		pathCadastroExt: `{"tipoCliente":"EXTERIOR","clienteExteriorDTO":{"razaoSocialOuNome":"EXAMPLE INC","email":"","endereco":{"complemento":"","logradouro":"Main St","numero":"100","cidade":"Springfield","pais":249,"descricaoPais":"Estados Unidos","simboloPais":"US"},"id":77}}`,
		pathTomadores:   `{"tomadores":[{"id":5,"nome":"BANCO EXEMPLO SA","cpfCnpj":"00000000000191"},{"id":77,"nome":"EXAMPLE INC","estrangeiro":true}],"emissaoSemTomador":false,"permiteEmissaoExterior":true}`,
	}
}

func TestTomadoresAdicionarCNPJ(t *testing.T) {
	f := tomadoresFake(t)
	delete(f.gets, pathCadastroBB) // documento novo
	out, stderr, code := execCLI(t, "", "notas", "tomadores", "adicionar", "--documento", "00.000.000/0001-91", "--email", "fin@exemplo.com", "--im", "123", "--yes", "-o", "json")
	want := `POST /api/plataforma/novo-emissor/clientes/salvar-cliente-nacional {"cpfCnpj":"00000000000191","razaoSocialOuNome":"BANCO EXEMPLO SA","telefone":"","email":"fin@exemplo.com","inscricaoMunicipal":"123",` +
		`"endereco":{"bairro":"Centro","cep":"01001000","codIbge":"3550308","complemento":"","logradouro":"Rua Exemplo","estado":"SP","numero":"1","cepInvalido":false}}`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, `"situacao": "cadastrado"`) || !strings.Contains(out, `"id": "5"`) {
		t.Errorf("saída:\n%s", out)
	}
}

func TestTomadoresAdicionarCPFEExterior(t *testing.T) {
	f := tomadoresFake(t)
	_, stderr, code := execCLI(t, "", "notas", "tomadores", "adicionar", "--documento", "529.982.247-25", "--nome", "Fulano de Tal", "--cep", "01001-000", "--numero", "10", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `"cpfCnpj":"52998224725","razaoSocialOuNome":"Fulano de Tal","telefone":"","email":"","inscricaoMunicipal":null`) ||
		!strings.Contains(f.writes[0], `"logradouro":"Rua Exemplo","estado":"SP","numero":"10"`) {
		t.Fatalf("CPF: código %d, %q\n%s", code, f.writes, stderr)
	}
	_, stderr, code = execCLI(t, "", "notas", "tomadores", "adicionar", "--exterior", "--nome", "Example Inc", "--pais", "us", "--cidade", "Springfield", "--logradouro", "Main St", "--numero", "100", "--yes")
	want := `POST /api/plataforma/novo-emissor/clientes/salvar-cliente-exterior {"razaoSocialOuNome":"Example Inc","email":"","endereco":{"complemento":"","logradouro":"Main St","numero":"100","cidade":"Springfield","pais":249,"descricaoPais":"Estados Unidos","simboloPais":"US"}}`
	if code != ExitOK || len(f.writes) != 2 || f.writes[1] != want {
		t.Errorf("exterior: código %d, %q\n%s", code, f.writes, stderr)
	}
}

func TestTomadoresAdicionarValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, "informe --documento (nacional) ou --exterior"},
		{[]string{"--documento", "00000000000192"}, "documento inválido"},
		{[]string{"--documento", "52998224725"}, "faltam dados do tomador: informe --bairro, --cep, --logradouro, --nome, --numero"},
		{[]string{"--documento", "52998224725", "--im", "1"}, "pessoa física não tem inscrição municipal"},
		{[]string{"--exterior", "--nome", "X", "--pais", "ZZ"}, "país \"ZZ\" não encontrado"},
		{[]string{"--exterior", "--nome", "X", "--telefone", "1"}, "não tem inscrição municipal, telefone"},
	} {
		f := tomadoresFake(t)
		_, stderr, code := execCLI(t, "", append([]string{"notas", "tomadores", "adicionar", "--yes"}, tc.args...)...)
		if code != ExitUsage || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestTomadoresEditar(t *testing.T) {
	f := tomadoresFake(t)
	_, stderr, code := execCLI(t, "", "notas", "tomadores", "editar", "00.000.000/0001-91", "--email", "novo@exemplo.com", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(f.writes[0], `"razaoSocialOuNome":"FULANO DE TAL","telefone":"","email":"novo@exemplo.com"`) {
		t.Fatalf("nacional: código %d, %q\n%s", code, f.writes, stderr)
	}
	_, stderr, code = execCLI(t, "", "notas", "tomadores", "editar", "77", "--cidade", "Boston", "--yes")
	if code != ExitOK || len(f.writes) != 2 || !strings.Contains(f.writes[1], `"cidade":"Boston"`) || !strings.HasSuffix(f.writes[1], `"id":77}`) {
		t.Fatalf("exterior: código %d, %q\n%s", code, f.writes, stderr)
	}
	if _, stderr, code := execCLI(t, "", "notas", "tomadores", "editar", "99", "--email", "a@b.c", "--yes"); code != ExitError || len(f.writes) != 2 {
		t.Errorf("inexistente: código %d, %s", code, stderr)
	}
}

func TestTomadoresRemover(t *testing.T) {
	f := tomadoresFake(t)
	f.onWrite = func(f *escritaFake) {
		f.gets[pathTomadores] = `{"tomadores":[{"id":77,"nome":"EXAMPLE INC","estrangeiro":true}],"emissaoSemTomador":false,"permiteEmissaoExterior":true}`
	}
	out, stderr, code := execCLI(t, "", "notas", "tomadores", "remover", "00.000.000/0001-91", "--yes", "-o", "json")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "DELETE /api/plataforma/autopilot/clientes/00000000000191 " {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(out, `"situacao": "removido"`) || !strings.Contains(stderr, "Excluir o tomador BANCO EXEMPLO SA (00000000000191). Esta ação não pode ser desfeita") {
		t.Errorf("saída:\n%s\n%s", out, stderr)
	}
	// Exterior pelo id; nacional pelo id usa o documento.
	execCLI(t, "", "notas", "tomadores", "remover", "77", "--yes")
	if len(f.writes) != 2 || f.writes[1] != "DELETE /api/plataforma/autopilot/clientes/77 " {
		t.Errorf("exterior: %q", f.writes)
	}
	f = tomadoresFake(t)
	execCLI(t, "", "notas", "tomadores", "remover", "5", "--yes")
	if len(f.writes) != 1 || f.writes[0] != "DELETE /api/plataforma/autopilot/clientes/00000000000191 " {
		t.Errorf("nacional pelo id: %q", f.writes)
	}
	if _, _, code := execCLI(t, "", "notas", "tomadores", "remover", "123", "--yes"); code != ExitError {
		t.Errorf("inexistente: código %d", code)
	}
}

// O cadastro do tomador responde 204 quando ele não existe: a leitura não pode tratar o 204
// como sucesso.
func TestTomadoresEditarNaoEncontrado(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathCadastroBB {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	withSession(t, srv)

	_, stderr, code := execCLI(t, "", "notas", "tomadores", "editar", "00000000000191", "--email", "a@exemplo.com", "--yes")
	if code != ExitError || !strings.Contains(stderr, "tomador 00000000000191 não encontrado") {
		t.Errorf("tomador inexistente: código %d, %s", code, stderr)
	}
}

// Cadastrar um documento que já existe sobrescreveria o cadastro com a Receita e as flags,
// apagando a IM e o que foi corrigido à mão (o emissor salva pelo documento): recusa e
// orienta o editar, sem escrita, inclusive com --yes e --dry-run (#660).
func TestTomadoresAdicionarDocumentoJaCadastrado(t *testing.T) {
	for _, tc := range []struct{ documento, normalizado, flag string }{
		{"00.000.000/0001-91", "00000000000191", "--yes"},
		{"00.000.000/0001-91", "00000000000191", "--dry-run"},
		{"12.abc.345/01de-35", "12ABC34501DE35", "--yes"}, // CNPJ alfanumérico
	} {
		f := tomadoresFake(t)
		f.gets[pathCadastroClientes+"12ABC34501DE35"] = strings.ReplaceAll(f.gets[pathCadastroBB], "00000000000191", "12ABC34501DE35")
		out, stderr, code := execCLI(t, "", "notas", "tomadores", "adicionar", "--documento", tc.documento, "--email", "x@exemplo.com", tc.flag)
		if code != ExitUsage || len(f.writes) != 0 || out != "" ||
			!strings.Contains(stderr, "tomador "+tc.normalizado+" já está cadastrado") ||
			!strings.Contains(stderr, "ctbz notas tomadores editar "+tc.normalizado) {
			t.Errorf("%v: código %d, %d escritas, saída %q\n%s", tc, code, len(f.writes), out, stderr)
		}
	}
}

// Documento novo (o cadastro responde 204) segue para a escrita; erro na leitura do cadastro
// para antes de escrever.
func TestTomadoresAdicionarConfereCadastro(t *testing.T) {
	f := tomadoresFake(t)
	delete(f.gets, pathCadastroBB)
	out, stderr, code := execCLI(t, "", "notas", "tomadores", "adicionar", "--documento", "00000000000191", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, "salvar-cliente-nacional") || !strings.Contains(stderr, "Cadastrar o tomador BANCO EXEMPLO SA") {
		t.Errorf("documento novo, dry-run: código %d, %d escritas\n%s\n%s", code, len(f.writes), out, stderr)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("escrita inesperada: %s %s", r.Method, r.URL)
		}
		http.Error(w, "erro", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	withSession(t, srv)
	if _, stderr, code := execCLI(t, "", "notas", "tomadores", "adicionar", "--documento", "00000000000191", "--yes"); code != ExitError || !strings.Contains(stderr, "conferindo se o tomador 00000000000191 já está cadastrado") {
		t.Errorf("cadastro ilegível: código %d\n%s", code, stderr)
	}
}
