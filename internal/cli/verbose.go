package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

// Modo verbose (#902): cada requisição vai para o stderr, com os segredos sempre mascarados.
// Nível 1: método, URL (valores da query e CPF/CNPJ do caminho mascarados), status, tempo,
// tamanho e cabeçalhos (valores de cookies e de cabeçalhos de credencial mascarados). Nível 2:
// também os corpos, truncados e com os campos de credencial e de pagamento mascarados.

// addVerboseFlag acrescenta --verbose às flags globais. Não há -v: ele já é --version.
func addVerboseFlag(root *cobra.Command) {
	root.PersistentFlags().Count("verbose", "mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)")
}

// nivelVerbose lê o nível pedido: --verbose (contado) ou CTBZ_DEBUG.
func nivelVerbose(cmd *cobra.Command) (int, error) {
	if n, _ := cmd.Flags().GetCount("verbose"); n > 0 {
		return min(n, 2), nil
	}
	switch v := os.Getenv("CTBZ_DEBUG"); v {
	case "", "0":
		return 0, nil
	case "1", "2":
		return int(v[0] - '0'), nil
	default:
		return 0, usageError{fmt.Errorf("CTBZ_DEBUG deve ser 0, 1 ou 2, não %q", v)}
	}
}

// ligarVerbose põe no contexto do comando um tracer que escreve as trocas em w.
func ligarVerbose(cmd *cobra.Command, w io.Writer) error {
	nivel, err := nivelVerbose(cmd)
	if err != nil || nivel == 0 {
		return err
	}
	var mu sync.Mutex // leituras em paralelo (ex.: ctbz resumo) não misturam as linhas
	cmd.SetContext(ctbz.WithTracer(cmd.Context(), func(e ctbz.Exchange) {
		var buf bytes.Buffer
		escreverTroca(&buf, nivel, e)
		mu.Lock()
		defer mu.Unlock()
		w.Write(buf.Bytes())
	}))
	return nil
}

// maxCorpoVerbose limita o corpo mostrado no nível 2.
const maxCorpoVerbose = 2000

// uuid reconhece identificadores que funcionam como credencial (ex.: o da fatura pública).
var uuid = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// jwt reconhece um JSON Web Token pelo formato, em qualquer chave ou cabeçalho.
var jwt = regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]*`)

// mascararIdentificadores troca UUIDs e JWTs por ***.
func mascararIdentificadores(s string) string {
	return jwt.ReplaceAllString(uuid.ReplaceAllString(s, "***"), "***")
}

// escreverTroca escreve uma troca no formato do modo verbose ("> " enviado, "< " recebido).
func escreverTroca(w io.Writer, nivel int, e ctbz.Exchange) {
	var b strings.Builder
	fmt.Fprintf(&b, "> %s %s\n", e.Method, mascararURL(e.URL, nivel))
	escreverCabecalhos(&b, "> ", e.Header, nivel)
	if nivel >= 2 && len(e.Body) > 0 {
		escreverCorpo(&b, "> ", e.Header.Get("Content-Type"), e.Body)
	}
	ms := e.Elapsed.Milliseconds()
	if e.Err != nil && e.Status == 0 {
		fmt.Fprintf(&b, "< sem resposta (%d ms): %s\n", ms, erroMascarado(e.Err, nivel))
	} else {
		fmt.Fprintf(&b, "< HTTP %d (%d ms, %s)\n", e.Status, ms, tamanhoArquivo(len(e.RespBody)))
		escreverCabecalhos(&b, "< ", e.RespHeader, nivel)
		if nivel >= 2 && len(e.RespBody) > 0 {
			escreverCorpo(&b, "< ", e.RespHeader.Get("Content-Type"), e.RespBody)
		}
	}
	io.WriteString(w, mascararIdentificadores(b.String()))
}

// erroMascarado descreve um erro de conexão. O *url.Error do net/http traz a URL inteira, que
// passa pela mesma máscara da linha da requisição.
func erroMascarado(err error, nivel int) string {
	var ue *url.Error
	if errors.As(err, &ue) {
		return fmt.Sprintf("%s %q: %v", ue.Op, mascararURL(ue.URL, nivel), ue.Err)
	}
	return err.Error()
}

// trechosSecretos identificam, no nome de um cabeçalho, campo ou parâmetro (em minúsculas e
// sem "-" e "_"), valores mascarados no modo verbose além dos do --dry-run: credenciais e os
// dados de pagamento das respostas (chave do gateway, Pix, linha digitável, código de barras).
var trechosSecretos = []string{"auth", "bearer", "jwt", "cookie", "session", "csrf", "xsrf", "apikey", "refresh",
	"credencial", "chave", "hash", "pix", "copiaecola", "qrcode", "digitavel", "codigobarras", "barras", "boleto"}

// secretaVerbose diz se o valor de um cabeçalho, campo ou parâmetro deve sair como *** no modo
// verbose. Cabeçalhos e corpos usam o mesmo critério. Inclui o do --dry-run (secreta).
func secretaVerbose(nome string) bool {
	n := strings.NewReplacer("-", "", "_", "").Replace(strings.ToLower(nome))
	if secreta(n) || strings.HasSuffix(n, "pin") || strings.HasSuffix(n, "key") {
		return true
	}
	for _, s := range trechosSecretos {
		if strings.Contains(n, s) {
			return true
		}
	}
	return false
}

func escreverCabecalhos(b *strings.Builder, prefixo string, h http.Header, nivel int) {
	nomes := make([]string, 0, len(h))
	for k := range h {
		nomes = append(nomes, k)
	}
	sort.Strings(nomes)
	for _, k := range nomes {
		for _, v := range h[k] {
			switch canon := http.CanonicalHeaderKey(k); {
			case canon == "Cookie":
				v = mascararCookies(v)
			case canon == "Set-Cookie":
				nome, _, _ := strings.Cut(v, "=")
				v = strings.TrimSpace(nome) + "=***"
			case canon == "Location":
				v = mascararURL(v, nivel)
			case secretaVerbose(canon):
				v = "***"
			}
			fmt.Fprintf(b, "%s%s: %s\n", prefixo, http.CanonicalHeaderKey(k), v)
		}
	}
}

// mascararCookies troca o valor de cada cookie por ***: "a=1; b=2" → "a=***; b=***".
func mascararCookies(v string) string {
	partes := strings.Split(v, ";")
	for i, p := range partes {
		nome, _, _ := strings.Cut(strings.TrimSpace(p), "=")
		partes[i] = nome + "=***"
	}
	return strings.Join(partes, "; ")
}

// documentoPontuado reconhece CPF (000.000.000-00) e CNPJ (00.000.000/0000-00, com a barra
// escapada ou não); sequencia, as sequências de dígitos.
var (
	documentoPontuado = regexp.MustCompile(`\d{3}\.\d{3}\.\d{3}-\d{2}|\d{2}\.\d{3}\.\d{3}(?:/|%2[Ff])\d{4}-\d{2}`)
	sequencia         = regexp.MustCompile(`\d+`)
)

// mascararDocumentos troca por *** CPFs e CNPJs: pontuados ou sequências de exatamente 11 ou 14
// dígitos (ex.: /clientes/consulta/00000000000191 ou o nome de arquivo 00000000000191_2026_9…).
// Ids, anos e meses (outros tamanhos) ficam como estão.
func mascararDocumentos(s string) string {
	s = documentoPontuado.ReplaceAllString(s, "***")
	return sequencia.ReplaceAllStringFunc(s, func(d string) string {
		if len(d) == 11 || len(d) == 14 {
			return "***"
		}
		return d
	})
}

// mascararURL mascara CPFs e CNPJs do caminho e os valores da query: todos no nível 1 (podem
// ter CNPJ e outros dados pessoais) e só os secretos no nível 2. Pares chave=valor num
// fragmento depois da query seguem a mesma regra.
func mascararURL(raw string, nivel int) string {
	antes, query, ok := strings.Cut(raw, "?")
	antes = mascararDocumentos(antes)
	if !ok {
		return antes
	}
	query, fragmento, temFragmento := strings.Cut(query, "#")
	saida := antes + "?" + mascararPares(query, nivel)
	if temFragmento {
		saida += "#" + mascararPares(fragmento, nivel)
	}
	return saida
}

// mascararPares mascara os valores de "a=1&b=2"; partes sem "=" ficam como estão.
func mascararPares(s string, nivel int) string {
	partes := strings.Split(s, "&")
	for i, p := range partes {
		nome, _, temValor := strings.Cut(p, "=")
		if !temValor {
			continue
		}
		chave, _ := url.QueryUnescape(nome)
		if nivel < 2 || secretaVerbose(chave) {
			partes[i] = nome + "=***"
		}
	}
	return strings.Join(partes, "&")
}

// escreverCorpo mostra um corpo JSON ou de formulário com os campos secretos mascarados; de
// multipart, só os campos de texto e o nome e o tamanho dos arquivos. Os demais (HTML, PDF…)
// são omitidos: podem ter tokens embutidos.
func escreverCorpo(b *strings.Builder, prefixo, contentType string, corpo []byte) {
	tipo, params, _ := mime.ParseMediaType(contentType)
	switch {
	case tipo == "application/x-www-form-urlencoded":
		fmt.Fprintf(b, "%s%s\n", prefixo, truncar(mascararFormulario(string(corpo)), len(corpo)))
	case tipo == "multipart/form-data":
		escreverMultipart(b, prefixo, params["boundary"], corpo)
	case (tipo == "application/json" || strings.HasSuffix(tipo, "+json") || tipo == "") && json.Valid(corpo):
		fmt.Fprintf(b, "%s%s\n", prefixo, truncar(jsonMascarado(corpo), len(corpo)))
	default:
		if tipo == "" {
			tipo = "sem tipo"
		}
		fmt.Fprintf(b, "%s(corpo %s de %s omitido)\n", prefixo, tipo, tamanhoArquivo(len(corpo)))
	}
}

// jsonMascarado reescreve um JSON válido com os campos secretos mascarados.
func jsonMascarado(corpo []byte) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if enc.Encode(mascararCom(decodificar(corpo), secretaVerbose)) != nil {
		return "(JSON ilegível omitido)"
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

func mascararFormulario(corpo string) string {
	vals, err := url.ParseQuery(corpo)
	if err != nil {
		return "(formulário ilegível omitido)"
	}
	chaves := make([]string, 0, len(vals))
	for k := range vals {
		chaves = append(chaves, k)
	}
	sort.Strings(chaves)
	var partes []string
	for _, k := range chaves {
		for _, v := range vals[k] {
			if secretaVerbose(k) {
				v = "***"
			} else {
				v = url.QueryEscape(v)
			}
			partes = append(partes, url.QueryEscape(k)+"="+v)
		}
	}
	return strings.Join(partes, "&")
}

func escreverMultipart(b *strings.Builder, prefixo, boundary string, corpo []byte) {
	r := multipart.NewReader(bytes.NewReader(corpo), boundary)
	for {
		p, err := r.NextPart()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			fmt.Fprintf(b, "%s(multipart ilegível: %v)\n", prefixo, err)
			return
		}
		dados, _ := io.ReadAll(p)
		switch {
		case p.FileName() != "":
			fmt.Fprintf(b, "%s%s: arquivo %s (%s)\n", prefixo, p.FormName(), p.FileName(), tamanhoArquivo(len(dados)))
		case secretaVerbose(p.FormName()):
			fmt.Fprintf(b, "%s%s=***\n", prefixo, p.FormName())
		default:
			fmt.Fprintf(b, "%s%s=%s\n", prefixo, p.FormName(), truncar(string(dados), len(dados)))
		}
	}
}

// truncar corta o texto em maxCorpoVerbose, avisando o tamanho original. Os UUIDs e JWTs são
// mascarados antes, para o corte não deixar um pedaço deles.
func truncar(s string, total int) string {
	s = mascararIdentificadores(s)
	if len(s) <= maxCorpoVerbose {
		return s
	}
	corte := maxCorpoVerbose
	for corte > 0 && !utf8.RuneStart(s[corte]) {
		corte--
	}
	return fmt.Sprintf("%s… (truncado; %d bytes)", s[:corte], total)
}
