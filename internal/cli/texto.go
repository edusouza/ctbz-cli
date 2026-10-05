package cli

import (
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/edusouza/ctbz-cli/internal/output"
)

var (
	reTag     = regexp.MustCompile(`<[^>]*>`)
	reEspacos = regexp.MustCompile(`\s+`)
)

// textoSimples tira as tags de um trecho de HTML vindo da API e junta os espaços.
func textoSimples(s string) output.Text {
	s = html.UnescapeString(reTag.ReplaceAllString(s, " "))
	return output.Text(strings.TrimSpace(reEspacos.ReplaceAllString(s, " ")))
}

var (
	reInvisivel    = regexp.MustCompile(`(?is)<(head|style|script)\b.*?</(head|style|script)\s*>`)
	reFimDeBloco   = regexp.MustCompile(`(?i)<br\s*/?>|</(p|div|li|tr|h[1-6]|table|ul|ol)\s*>`)
	reFimDeCelula  = regexp.MustCompile(`(?i)</t[dh]\s*>`)
	reEspacoLinha  = regexp.MustCompile(`[ \t\f\v\x{a0}]+`)
	reLinhasVazias = regexp.MustCompile(`\n{3,}`)
)

// htmlParaTexto converte um documento HTML (ex.: contrato) em texto legível: descarta
// cabeçalho, estilos e scripts, quebra linha no fim de parágrafos e blocos, separa células de tabela e junta espaços.
func htmlParaTexto(doc string) string {
	s := reInvisivel.ReplaceAllString(doc, "")
	s = reFimDeBloco.ReplaceAllString(s, "\n")
	s = reFimDeCelula.ReplaceAllString(s, " ")
	s = html.UnescapeString(reTag.ReplaceAllString(s, ""))
	linhas := strings.Split(strings.ReplaceAll(s, "\r", ""), "\n")
	for i, l := range linhas {
		linhas[i] = strings.TrimSpace(reEspacoLinha.ReplaceAllString(l, " "))
	}
	s = reLinhasVazias.ReplaceAllString(strings.Join(linhas, "\n"), "\n\n")
	return semControle(strings.TrimSpace(s)) + "\n"
}

// semControle tira caracteres de controle (exceto quebra de linha e tabulação), para que um
// texto vindo da API não mande sequências de escape ao terminal.
func semControle(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			return r
		}
		return -1
	}, s)
}
