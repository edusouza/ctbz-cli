package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newAPICmd() *cobra.Command {
	var method, data string
	var form []string
	var raw bool
	cmd := &cobra.Command{
		Use:   "api CAMINHO",
		Short: "Chama uma URL da plataforma com a sessão atual",
		Long: `Chama qualquer endpoint da plataforma com os cookies da sessão.

CAMINHO relativo é resolvido contra o BFF da plataforma (/api/plataforma/);
caminhos absolutos ("/api/legado/...") alcançam as demais APIs do app.

No Git Bash do Windows (e em agentes que rodam por ele), um argumento que começa com "/" vira
caminho do Windows ("C:/Program Files/Git/api/...") antes de chegar à CLI. Comece o caminho
com "//" ("//api/legado/..."), que funciona em qualquer shell, ou rode com MSYS_NO_PATHCONV=1.
A CLI recusa um caminho que já chegou convertido.

Com -X diferente de GET, a chamada é uma escrita: a CLI confere a sessão antes e envia uma
única vez, sem re-login automático nem retentativa depois do envio. Não há confirmação nem
--dry-run: ctbz api é a ferramenta de baixo nível (ADR-0018).

-F monta um corpo multipart/form-data, como o curl: -F campo=valor para texto e
-F campo=@arquivo para um arquivo (o tipo vem da extensão; ;type=tipo escolhe outro). Sem -X,
-F envia com POST. -F e -d não podem ser usados juntos. Cada arquivo tem até 30 MB, e o envio
tem o mesmo prazo dos uploads da CLI (3 minutos). O registro de ações (ctbz acoes) guarda o
nome e o tamanho dos arquivos, nunca o conteúdo.

A resposta sai em JSON formatado (CTBZ_OUTPUT é ignorado). Com -o table ou -o csv,
listas de objetos viram tabelas. --raw imprime o corpo exatamente como veio.
Respostas HTTP 4xx/5xx terminam com código de saída 1.`,
		Example: `  ctbz api dadosempresa/get
  ctbz api -o table menu/get
  ctbz api /api/legado/empresa/cnpj?cnpj=00000000000100
  ctbz api //api/legado/empresa/cnpj?cnpj=00000000000100   # no Git Bash do Windows
  ctbz api -X POST -d @corpo.json caminho/qualquer
  ctbz api -F ano=2026 -F mes=9 -F documento=@extrato.csv caminho/de/upload`,
		Args: exactArgs(1, "exatamente um CAMINHO"),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := outputFormat(cmd, string(output.FormatJSON))
			if err != nil {
				return err
			}
			caminho, err := caminhoAPI(args[0])
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			metodo := strings.ToUpper(method)
			var resp *ctbz.Response
			if len(form) > 0 {
				if data != "" {
					return usageError{errors.New("-F e -d não podem ser usados juntos: -F monta um corpo multipart, -d envia JSON")}
				}
				if !cmd.Flags().Changed("method") {
					metodo = http.MethodPost
				}
				if metodo == http.MethodGet {
					return usageError{errors.New("-F envia um corpo: use -X POST ou -X PUT")}
				}
				campos, err := camposFormulario(form)
				if err != nil {
					return err
				}
				corpo, contentType, err := api.EncodeMultipart(campos)
				if err != nil {
					return err
				}
				origem := origemEscrita{comando: "ctbz api", arquivos: descreverArquivos(campos)}
				if resp, err = sendOnce(cmd.Context(), s, origem, metodo, caminho, corpo, contentType, timeoutUpload); err != nil {
					return err
				}
			} else {
				body, err := readBody(data, s.in)
				if err != nil {
					return err
				}
				if resp, err = callAPI(cmd.Context(), s, metodo, caminho, body); err != nil {
					return err
				}
			}
			if parsed, perr := output.FromJSON(resp.Body); !raw && perr == nil {
				if err := output.Write(s.out, f, parsed); err != nil {
					return err
				}
			} else {
				s.out.Write(resp.Body)
				if !raw && len(resp.Body) > 0 && resp.Body[len(resp.Body)-1] != '\n' {
					io.WriteString(s.out, "\n")
				}
			}
			if resp.Status >= 400 {
				return fmt.Errorf("HTTP %d", resp.Status)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&method, "method", "X", "GET", "método HTTP")
	f.StringVarP(&data, "data", "d", "", "corpo JSON da requisição (@arquivo lê de um arquivo, @- da entrada padrão)")
	f.StringArrayVarP(&form, "form", "F", nil, "campo multipart/form-data: campo=valor ou campo=@arquivo[;type=tipo] (repita para mais campos)")
	f.BoolVar(&raw, "raw", false, "imprime o corpo da resposta como veio, sem formatar")
	return cmd
}

// callAPI usa a camada de leitura (com re-login) no GET e a de escrita nos demais métodos.
func callAPI(ctx context.Context, s streams, method, path string, body []byte) (*ctbz.Response, error) {
	if method == http.MethodGet {
		return authedAPI(ctx, s, method, path, body)
	}
	contentType := ""
	if body != nil {
		contentType = "application/json"
	}
	return sendOnce(ctx, s, origemEscrita{comando: "ctbz api"}, method, path, body, contentType, 0)
}

// caminhoConvertidoMSYS reconhece um caminho absoluto que o Git Bash (MSYS) do Windows trocou
// por um caminho do Windows antes de passá-lo à CLI: "/api/x" → "C:/Program Files/Git/api/x".
var caminhoConvertidoMSYS = regexp.MustCompile(`^[A-Za-z]:[/\\]`)

// caminhoAPI valida o CAMINHO de ctbz api. Recusa o caminho convertido pelo MSYS, que daria um
// 404 sem explicação, e aceita "//api/..." (que o Git Bash não converte) como "/api/...".
func caminhoAPI(p string) (string, error) {
	if caminhoConvertidoMSYS.MatchString(p) {
		return "", usageError{fmt.Errorf("o caminho %q parece ter sido convertido pelo Git Bash do Windows a partir de um caminho absoluto; "+
			"comece o caminho com // (ex.: ctbz api //api/legado/...) ou rode com MSYS_NO_PATHCONV=1", p)}
	}
	if strings.HasPrefix(p, "//") {
		p = "/" + strings.TrimLeft(p, "/")
	}
	return p, nil
}

// camposFormulario interpreta os -F: "campo=valor" é texto e "campo=@arquivo[;type=tipo]" é
// arquivo, lido do disco (até 30 MB, o limite dos uploads do painel).
func camposFormulario(form []string) ([]api.Campo, error) {
	campos := make([]api.Campo, 0, len(form))
	for _, f := range form {
		nome, valor, ok := strings.Cut(f, "=")
		if !ok || nome == "" {
			return nil, usageError{fmt.Errorf("-F %q: use campo=valor ou campo=@arquivo", f)}
		}
		if !strings.HasPrefix(valor, "@") {
			campos = append(campos, api.Campo{Nome: nome, Valor: valor})
			continue
		}
		caminho, tipo := valor[1:], ""
		if i := strings.LastIndex(caminho, ";type="); i >= 0 {
			caminho, tipo = caminho[:i], caminho[i+len(";type="):]
		}
		fi, err := os.Stat(caminho)
		if err != nil {
			return nil, usageError{fmt.Errorf("lendo o arquivo de -F %s: %w", nome, err)}
		}
		if fi.Size() > api.TamanhoMaximoExtrato {
			return nil, usageError{fmt.Errorf("o arquivo de -F %s passa do limite de 30 MB", nome)}
		}
		conteudo, err := os.ReadFile(caminho)
		if err != nil {
			return nil, usageError{fmt.Errorf("lendo o arquivo de -F %s: %w", nome, err)}
		}
		campos = append(campos, api.Campo{Nome: nome, Arquivo: filepath.Base(caminho), Conteudo: conteudo, Tipo: tipo})
	}
	return campos, nil
}

func readBody(arg string, stdin io.Reader) ([]byte, error) {
	switch {
	case arg == "":
		return nil, nil
	case arg == "@-":
		return io.ReadAll(stdin)
	case strings.HasPrefix(arg, "@"):
		return os.ReadFile(arg[1:])
	default:
		return []byte(arg), nil
	}
}
