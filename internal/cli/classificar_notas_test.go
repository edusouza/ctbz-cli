package cli

import (
	"strings"
	"testing"
)

const (
	pathAClassificar  = "/api/emissor/classificacaonotas/listar/?mes=9&ano=2026&empresa=&tipo=0&limite=10&cursor=&offset=0"
	pathClassificadas = "/api/emissor/classificacaonotas/listar/?mes=9&ano=2026&empresa=&tipo=1&limite=10&cursor=&offset=0"
)

func classificarFake(t *testing.T) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{
		pathAClassificar: `{"total":2,"cursor":null,"list":[
{"id":701,"razaoSocial":"Fornecedor Fictício","valor":150.25,"campoDesconhecido":{"x":1}},
{"id":702,"razaoSocial":"Outro Fornecedor","valor":10}]}`,
		pathClassificadas: `{"total":0,"cursor":null,"list":[]}`,
		"/api/emissor/classificacaonotas/listarprodutos/701": fixture(t, "produtos_nota"),
	})
	// Depois do envio, a nota 701 passa às classificadas.
	f.onWrite = func(f *escritaFake) {
		f.gets[pathAClassificar] = `{"total":1,"cursor":null,"list":[{"id":702,"razaoSocial":"Outro Fornecedor","valor":10}]}`
		f.gets[pathClassificadas] = `{"total":1,"cursor":null,"list":[{"id":701,"razaoSocial":"Fornecedor Fictício","valor":150.25,"tipoClassificacao":"USO_CONSUMO"}]}`
	}
	return f
}

func TestClassificarLote(t *testing.T) {
	f := classificarFake(t)
	out, stderr, code := execCLI(t, "", "notas", "entrada", "classificar", "701", "702", "--como", "uso-consumo", "--mes", "2026-09", "--yes", "-o", "csv")
	want := `POST /api/emissor/classificacaonotas/salvarloteclassificacao/USO_CONSUMO [{"id":701,"razaoSocial":"Fornecedor Fictício","valor":150.25,"campoDesconhecido":{"x":1}},{"id":702,"razaoSocial":"Outro Fornecedor","valor":10}]`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "Classificar 2 nota(s) como uso-consumo: 701 de Fornecedor Fictício (R$ 150,25)") || !strings.Contains(stderr, "(risco médio)") {
		t.Errorf("stderr:\n%s", stderr)
	}
	wantOut := "acao,situacao,id,classificacao\nclassificar,USO_CONSUMO,701,uso-consumo\nclassificar,enviado,702,uso-consumo\n"
	if out != wantOut {
		t.Errorf("saída:\n%s\nesperado:\n%s", out, wantOut)
	}
}

func TestClassificarValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"701"}, ExitUsage, "informe --como: estoque, insumo, uso-consumo, ativo-imobilizado, prestacao-servico"},
		{[]string{"701", "--como", "revenda"}, ExitUsage, `--como deve ser estoque, insumo`},
		{[]string{"701", "--como", "estoque", "--mes", "2026/09"}, ExitUsage, "--mes deve ser AAAA-MM"},
		{[]string{"999", "--como", "estoque", "--mes", "2026-09"}, ExitError, "nota 999 não está a classificar em 09/2026"},
	} {
		f := classificarFake(t)
		_, stderr, code := execCLI(t, "", append([]string{"notas", "entrada", "classificar", "--yes"}, tc.args...)...)
		if code != tc.code || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestProdutosNota(t *testing.T) {
	classificarFake(t)
	out, stderr, code := execCLI(t, "", "notas", "entrada", "produtos", "701", "-o", "csv")
	want := "id,descricao,ncm,valor,quantidade,estoque,insumo,uso_consumo,ativo_imobilizado,prestacao_servico\n" +
		"11,Papel A4 500 folhas,48025610,120.50,5,0,0,5,0,0\n" +
		"12,Toner para impressora,84439933,29.75,1,1,0,0,0,0\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s\nesperado:\n%s", code, stderr, out, want)
	}
}
