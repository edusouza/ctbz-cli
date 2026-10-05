package api

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// entradaCatalogo é uma linha de docs/api/catalogo.md.
type entradaCatalogo struct {
	base, metodo string
	molde        *regexp.Regexp
}

var (
	reSecaoBase  = regexp.MustCompile("^## `(/api/[^`]+/)`")
	reLinhaTabel = regexp.MustCompile("^\\| (GET|POST|PUT|PATCH|DELETE) \\| `([^`]+)` \\|$")
)

// lerCatalogo interpreta o catálogo gerado: "{}" casa um segmento, e caminhos terminados em
// "/" ou "=" (ou com query) casam por prefixo, porque o front concatena o resto.
func lerCatalogo(texto string) []entradaCatalogo {
	var out []entradaCatalogo
	base := ""
	sc := bufio.NewScanner(strings.NewReader(texto))
	for sc.Scan() {
		linha := sc.Text()
		if m := reSecaoBase.FindStringSubmatch(linha); m != nil {
			base = m[1]
			continue
		}
		m := reLinhaTabel.FindStringSubmatch(linha)
		if m == nil || base == "" {
			continue
		}
		caminho := strings.TrimPrefix(m[2], "/")
		prefixo := strings.HasSuffix(caminho, "/") || strings.HasSuffix(caminho, "=") || strings.Contains(caminho, "?")
		caminho, _, _ = strings.Cut(caminho, "?")
		expr := "^" + strings.ReplaceAll(regexp.QuoteMeta(caminho), regexp.QuoteMeta("{}"), "[^/]+")
		if !prefixo {
			expr += "/?$"
		}
		out = append(out, entradaCatalogo{base: base, metodo: m[1], molde: regexp.MustCompile(expr)})
	}
	return out
}

// naoCatalogadas devolve as requisições ("MÉTODO caminho") que nenhuma linha do catálogo cobre.
func naoCatalogadas(catalogo []entradaCatalogo, reqs []requisicao) []string {
	var faltam []string
	for _, r := range reqs {
		achou := false
		for _, e := range catalogo {
			rel, ok := strings.CutPrefix(r.Caminho, e.base)
			if ok && e.metodo == r.Metodo && e.molde.MatchString(rel) {
				achou = true
				break
			}
		}
		if !achou {
			faltam = append(faltam, r.Metodo+" "+r.Caminho)
		}
	}
	return faltam
}

func TestNaoCatalogadas(t *testing.T) {
	cat := lerCatalogo("## `/api/plataforma/` (3)\n\n### caixa\n\n| Método | Caminho |\n|---|---|\n" +
		"| POST | `caixa/lancamentousuario/novo/` |\n" +
		"| DELETE | `/caixa/lancamentousuario/remover/` |\n" +
		"| PUT | `impostos/v5/impostos-a-pagar/guia/{}/confirmar-pagamento` |\n" +
		"| GET | `contabancaria/listar` |\n" +
		"## `/api/emissor/` (1)\n\n| POST | `notasentrada/manifestar/` |\n")
	reqs := []requisicao{
		{Metodo: "POST", Caminho: "/api/plataforma/caixa/lancamentousuario/novo/"},
		{Metodo: "DELETE", Caminho: "/api/plataforma/caixa/lancamentousuario/remover/2026/9/123"},
		{Metodo: "PUT", Caminho: "/api/plataforma/impostos/v5/impostos-a-pagar/guia/77/confirmar-pagamento"},
		{Metodo: "POST", Caminho: "/api/emissor/notasentrada/manifestar/"},
		{Metodo: "PUT", Caminho: "/api/plataforma/impostos/v5/impostos-a-pagar/guia/77/outra-coisa"},
		{Metodo: "POST", Caminho: "/api/plataforma/contabancaria/listar"},
		{Metodo: "POST", Caminho: "/api/legado/caixa/lancamentousuario/novo/"},
	}
	got := naoCatalogadas(cat, reqs)
	want := []string{
		"PUT /api/plataforma/impostos/v5/impostos-a-pagar/guia/77/outra-coisa",
		"POST /api/plataforma/contabancaria/listar",
		"POST /api/legado/caixa/lancamentousuario/novo/",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("não catalogadas:\n%s\nquero:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// TestEscritasNoCatalogo confere se cada escrita da CLI (pelos goldens) ainda aparece no
// catálogo. Roda no monitoramento, depois de regenerar o catálogo a partir do front:
//
//	CTBZ_CATALOGO_ESCRITAS=1 go test ./internal/api -run EscritasNoCatalogo
func TestEscritasNoCatalogo(t *testing.T) {
	if os.Getenv("CTBZ_CATALOGO_ESCRITAS") == "" {
		t.Skip("defina CTBZ_CATALOGO_ESCRITAS=1 (usado pelo monitoramento com o catálogo regenerado)")
	}
	texto, err := os.ReadFile(filepath.Join("..", "..", "docs", "api", "catalogo.md"))
	if err != nil {
		t.Fatal(err)
	}
	cat := lerCatalogo(string(texto))
	files, _ := filepath.Glob(goldenRequisicoes("*"))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var reqs []requisicao
		if err := json.Unmarshal(data, &reqs); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, falta := range naoCatalogadas(cat, reqs) {
			t.Errorf("%s: escrita ausente do catálogo: %s", filepath.Base(f), falta)
		}
	}
}
