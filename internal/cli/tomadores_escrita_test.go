package cli

import (
	"strings"
	"testing"
)

const (
	pathConsultaBB  = "/api/plataforma/novo-emissor/clientes/consulta/00000000000191"
	pathTomadores   = "/api/plataforma/novo-emissor/tomadores/init"
	pathCEPSe       = "/api/plataforma/novo-emissor/cep/logradouro?cep=01001000"
	pathPaises      = "/api/plataforma/notafiscal/emitir/buscarPaisesParaEmissao/"
	pathCadastroBB  = "/api/plataforma/novo-emissor/cadastro-clientes/00000000000191"
	pathCadastroExt = "/api/plataforma/novo-emissor/cadastro-clientes/77"
)

func tomadoresFake(t *testing.T) *escritaFake {
	t.Helper()
	return newEscritaFake(t, map[string]string{
		pathConsultaBB: `{"cnpj":"00000000000191","razaoSocial":"BANCO EXEMPLO SA","nomeFantasia":"","dataAbertura":"01/01/1900","atividadePrincipal":null,"naturezaJuridica":null,
"logradouro":"Rua Exemplo","numero":"1","complemento":"","bairro":"Centro","cep":"01001000","municipio":"São Paulo","uf":"SP","email":"","telefone":"","situacaoCadastral":"ATIVA","optanteSimples":"NÃO"}`,
		pathCEPSe:       fixture(t, "cep"),
		pathPaises:      fixture(t, "paises_emissao"),
		pathCadastroBB:  fixture(t, "cadastro_cliente"),
		pathCadastroExt: `{"tipoCliente":"EXTERIOR","clienteExteriorDTO":{"razaoSocialOuNome":"EXAMPLE INC","email":"","endereco":{"complemento":"","logradouro":"Main St","numero":"100","cidade":"Springfield","pais":249,"descricaoPais":"Estados Unidos","simboloPais":"US"},"id":77}}`,
		pathTomadores:   `{"tomadores":[{"id":5,"nome":"BANCO EXEMPLO SA","cpfCnpj":"00000000000191"},{"id":77,"nome":"EXAMPLE INC","estrangeiro":true}],"emissaoSemTomador":false,"permiteEmissaoExterior":true}`,
	})
}

func TestTomadoresAdicionarCNPJ(t *testing.T) {
	f := tomadoresFake(t)
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
