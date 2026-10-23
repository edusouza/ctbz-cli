package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newAPICmd() *cobra.Command {
	var method, data string
	var raw bool
	cmd := &cobra.Command{
		Use:   "api CAMINHO",
		Short: "Chama uma URL da plataforma com a sessão atual",
		Long: `Chama qualquer endpoint da plataforma com os cookies da sessão.

CAMINHO relativo é resolvido contra o BFF da plataforma (/api/plataforma/);
caminhos absolutos ("/api/legado/...") alcançam as demais APIs do app.

Com -X diferente de GET, a chamada é uma escrita: a CLI confere a sessão antes e envia uma
única vez, sem re-login automático nem retentativa depois do envio. Não há confirmação nem
--dry-run: ctbz api é a ferramenta de baixo nível (ADR-0018).

A resposta sai em JSON formatado (CTBZ_OUTPUT é ignorado). Com -o table ou -o csv,
listas de objetos viram tabelas. --raw imprime o corpo exatamente como veio.
Respostas HTTP 4xx/5xx terminam com código de saída 1.`,
		Example: `  ctbz api dadosempresa/get
  ctbz api -o table menu/get
  ctbz api /api/legado/empresa/cnpj?cnpj=00000000000100
  ctbz api -X POST -d @corpo.json caminho/qualquer`,
		Args: exactArgs(1, "exatamente um CAMINHO"),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := outputFormat(cmd, string(output.FormatJSON))
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			body, err := readBody(data, s.in)
			if err != nil {
				return err
			}
			resp, err := callAPI(cmd.Context(), s, strings.ToUpper(method), args[0], body)
			if err != nil {
				return err
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
