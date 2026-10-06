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
		{[]string{"701"}, ExitUsage, "informe --como (estoque, insumo, uso-consumo, ativo-imobilizado, prestacao-servico) ou --produto"},
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

func TestClassificarPorProduto(t *testing.T) {
	f := classificarFake(t)
	out, stderr, code := execCLI(t, "", "notas", "entrada", "classificar", "701", "--produto", "11:estoque=2, uso-consumo=2.5,insumo=0.5", "--mes", "2026-09", "--yes", "-o", "csv")
	want := `POST /api/emissor/classificacaonotas/salvarclassificacao/ [` +
		`{"descricao":"Papel A4 500 folhas","id":11,"ncm":"48025610","quantidadeAtivo":0,"quantidadeConsumo":2.5,"quantidadeEstoque":2,"quantidadeInsumo":0.5,"quantidadePrestacao":0,"quantidadeTotal":5,"valor":120.5},` +
		`{"descricao":"Toner para impressora","id":12,"ncm":"84439933","quantidadeAtivo":0,"quantidadeConsumo":0,"quantidadeEstoque":1,"quantidadeInsumo":0,"quantidadePrestacao":0,"quantidadeTotal":1,"valor":29.75}]`
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != want {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "item 11 Papel A4 500 folhas: estoque 2, insumo 0.5, uso-consumo 2.5; item 12 Toner para impressora: estoque 1") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if out != "acao,situacao,id,classificacao\nclassificar,USO_CONSUMO,701,por produto\n" {
		t.Errorf("saída:\n%s", out)
	}
}

func TestClassificarPorProdutoValidacoes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"701", "--como", "estoque", "--produto", "11:estoque=5"}, "use --como ou --produto, não os dois"},
		{[]string{"701", "702", "--produto", "11:estoque=5"}, "--produto classifica uma nota por vez"},
		{[]string{"701", "--produto", "11"}, "--produto deve ser ITEM:opcao=qtd"},
		{[]string{"701", "--produto", "11:revenda=5"}, "--produto 11: opção deve ser estoque"},
		{[]string{"701", "--produto", "11:estoque=2,5"}, "quantidades usam ponto decimal"},
		{[]string{"701", "--produto", "11:estoque=x"}, "quantidade inválida"},
		{[]string{"701", "--produto", "11:estoque=NaN"}, "quantidade inválida"},
		{[]string{"701", "--produto", "11:estoque=4"}, "item 11 (Papel A4 500 folhas): distribua toda a quantidade (4 de 5)"},
		{[]string{"701", "--produto", "11:estoque=6,insumo=-1"}, "a classificação não pode ser negativa"},
		{[]string{"701", "--produto", "11:estoque=5", "--produto", "11:insumo=5"}, "item 11 repetido"},
		{[]string{"701", "--produto", "11:estoque=5", "--produto", "99:insumo=1"}, "a nota 701 não tem o item 99"},
	} {
		f := classificarFake(t)
		_, stderr, code := execCLI(t, "", append([]string{"notas", "entrada", "classificar", "--yes", "--mes", "2026-09"}, tc.args...)...)
		if code != ExitUsage || !strings.Contains(stderr, tc.want) || len(f.writes) != 0 {
			t.Errorf("%v: código %d, %d escritas, %q (quero %q)", tc.args, code, len(f.writes), stderr, tc.want)
		}
	}
}

func TestReclassificar(t *testing.T) {
	f := classificarFake(t)
	f.gets[pathClassificadas] = `{"total":1,"cursor":null,"list":[{"id":701,"razaoSocial":"Fornecedor Fictício","valor":150.25,"tipoClassificacao":"USO_CONSUMO"}]}`
	f.onWrite = func(f *escritaFake) {
		f.gets[pathClassificadas] = `{"total":1,"cursor":null,"list":[{"id":701,"razaoSocial":"Fornecedor Fictício","valor":150.25,"tipoClassificacao":"PROCESSANDO"}]}`
	}
	out, stderr, code := execCLI(t, "", "notas", "entrada", "reclassificar", "701", "--como", "ativo-imobilizado", "--mes", "2026-09", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || !strings.HasPrefix(f.writes[0], "POST /api/emissor/classificacaonotas/reclassificar?idNfe=701 [") ||
		strings.Count(f.writes[0], `"quantidadeAtivo":`) != 2 || !strings.Contains(f.writes[0], `"quantidadeAtivo":5,"quantidadeConsumo":0`) ||
		!strings.Contains(f.writes[0], `"quantidadeAtivo":1,"quantidadeConsumo":0,"quantidadeEstoque":0`) {
		t.Fatalf("código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "(hoje USO_CONSUMO)") || out != "acao,situacao,id,classificacao\nreclassificar,PROCESSANDO,701,ativo-imobilizado\n" {
		t.Errorf("stderr:\n%s\nsaída:\n%s", stderr, out)
	}

	_, stderr, code = execCLI(t, "", "notas", "entrada", "reclassificar", "702", "--como", "estoque", "--mes", "2026-09", "--yes")
	if code != ExitError || !strings.Contains(stderr, "nota 702 não está classificada em 09/2026") {
		t.Errorf("não classificada: código %d, %s", code, stderr)
	}
}

func TestReclassificarDryRun(t *testing.T) {
	f := classificarFake(t)
	f.gets[pathClassificadas] = `{"total":1,"cursor":null,"list":[{"id":701,"razaoSocial":"Fornecedor Fictício","valor":150.25}]}`
	out, _, code := execCLI(t, "", "notas", "entrada", "reclassificar", "701", "--produto", "11:estoque=5", "--mes", "2026-09", "--dry-run")
	if code != ExitOK || len(f.writes) != 0 || !strings.Contains(out, "reclassificar?idNfe=701") {
		t.Errorf("código %d, %d escritas\n%s", code, len(f.writes), out)
	}
}
