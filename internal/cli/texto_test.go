package cli

import "testing"

func TestHTMLParaTexto(t *testing.T) {
	doc := `<html><head><title>PROPOSTA</title><style>p.MsoNormal {margin:0}</style></head>
<body><h1>CONTRATO</h1><p class=MsoNormal>Cláusula&nbsp;1ª &ndash; do   objeto</p>
<p>Linha<br>quebrada</p><script>alert(1)</script>


<ul><li>item &amp; outro</li></ul><table><tr><th>Faixa</th><td>Valor</td></tr></table></body></html>`
	want := "CONTRATO\nCláusula 1ª – do objeto\n\nLinha\nquebrada\n\nitem & outro\n\nFaixa Valor\n"
	if got := htmlParaTexto(doc); got != want {
		t.Errorf("htmlParaTexto:\n%q\nquero:\n%q", got, want)
	}
}

func TestHTMLParaTextoSemControle(t *testing.T) {
	got := htmlParaTexto("<p>Termo\x1b[31m vermelho\x07</p><p>linha\tdois</p>")
	if got != "Termo[31m vermelho\nlinha dois\n" {
		t.Errorf("htmlParaTexto = %q", got)
	}
}
