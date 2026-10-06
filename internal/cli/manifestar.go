package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// situacoesManifestaveis são as situações que o front deixa manifestar.
var situacoesManifestaveis = map[string]bool{"PENDENTE": true, "CIENCIA": true}

const avisoManifestacao = "A Contabilizei comunica a manifestação à Receita Federal; em alguns minutos a nota estará manifestada."

// notasDoMes procura as notas do mês nas listas de manifestação.
func notasDoMes(cmd *cobra.Command, g api.Getter, mes time.Time) (map[string]api.NotaEntrada, error) {
	notas := map[string]api.NotaEntrada{}
	for _, lista := range []string{api.ListaAManifestar, api.ListaManifestadas} {
		ns, err := api.BuscarNotasEntrada(cmd.Context(), g, filtroEntrada(lista, mes))
		if err != nil {
			return nil, err
		}
		for _, n := range ns {
			notas[idTexto(n.ID)] = n
		}
	}
	return notas, nil
}

func situacaoNota(n api.NotaEntrada) string {
	if n.Situacao == nil {
		return ""
	}
	return n.Situacao.ID
}

func newNotasEntradaManifestarCmd() *cobra.Command {
	var ciencia, confirmar, desconhecer, naoRealizada bool
	var justificativa, mes string
	cmd := &cobra.Command{
		Use:   "manifestar ID...",
		Short: "Envia a manifestação do destinatário das notas de entrada à SEFAZ",
		Long: `Manifesta as NF-e de compra como destinatário: --ciencia (ciência da operação),
--confirmar ("recebi"), --desconhecer ("desconheço") ou --nao-realizada ("não recebi", exige
--justificativa de 15 a 255 caracteres, regra da SEFAZ).

Só notas PENDENTE ou CIENCIA são manifestadas, como no painel; as demais são listadas como
ignoradas. --mes diz em que mês procurar as notas (padrão: o mês atual).

Risco alto nos eventos conclusivos (confirmar, desconhecer, não realizada): são eventos
fiscais enviados à SEFAZ, sem desfazer, e a CLI pede "confirmo". A ciência é de risco médio.
No painel, baixar o XML ou o DANFE de uma nota pendente manifesta ciência automaticamente.

Aceita --yes e --dry-run.`,
		Example: `  ctbz notas entrada manifestar 1000000000000701 --confirmar --mes 2026-09
  ctbz notas entrada manifestar 1000000000000702 --nao-realizada --justificativa "Mercadoria extraviada antes da entrega"`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return usageError{errors.New("informe pelo menos um ID de nota (ver ctbz notas entrada)")}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			tipo, marcadas := "", 0
			for _, o := range []struct {
				marcado bool
				tipo    string
			}{{ciencia, api.ManifestacaoCiencia}, {confirmar, api.ManifestacaoConfirmacao}, {desconhecer, api.ManifestacaoDesconhecimento}, {naoRealizada, api.ManifestacaoNaoRealizada}} {
				if o.marcado {
					tipo, marcadas = o.tipo, marcadas+1
				}
			}
			if marcadas != 1 {
				return usageError{errors.New("escolha uma manifestação: --ciencia, --confirmar, --desconhecer ou --nao-realizada")}
			}
			justificativa = strings.TrimSpace(justificativa)
			if tipo == api.ManifestacaoNaoRealizada {
				if c := utf8.RuneCountInString(justificativa); c < 15 || c > 255 {
					return usageError{fmt.Errorf("--nao-realizada exige --justificativa de 15 a 255 caracteres (tem %d)", c)}
				}
			} else if justificativa != "" {
				return usageError{errors.New("--justificativa só vale com --nao-realizada")}
			}
			ref, err := mesDaFlag(mes)
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			notas, err := notasDoMes(cmd, sessionGetter{s}, ref)
			if err != nil {
				return err
			}
			var ids []any
			var nomes []string
			for _, id := range args {
				nota, ok := notas[id]
				switch {
				case !ok:
					return fmt.Errorf("nota %s não encontrada em %s; use --mes", id, ref.Format("01/2006"))
				case !situacoesManifestaveis[situacaoNota(nota)]:
					fmt.Fprintf(s.err, "Ignorada: nota %s está %s (só PENDENTE ou CIENCIA podem ser manifestadas).\n", id, situacaoNota(nota))
					continue
				}
				ids = append(ids, nota.ID)
				nomes = append(nomes, fmt.Sprintf("%s de %s (%s)", id, nota.RazaoSocial, output.FormatBRL(valorNota(nota))))
			}
			if len(ids) == 0 {
				return errors.New("nenhuma das notas pode ser manifestada")
			}
			op := operacao{Risco: riscoAlto, ID: strings.Join(args, ","), Resumo: fmt.Sprintf("Manifestar %s em %d nota(s): %s", tipo, len(ids), strings.Join(nomes, "; ")),
				Consequencia: "É um evento fiscal enviado à SEFAZ e não dá para desfazer."}
			if tipo == api.ManifestacaoCiencia {
				op = operacao{Risco: riscoMedio, ID: op.ID, Resumo: op.Resumo}
			}
			var resp []api.NotaManifestada
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				var err error
				resp, err = api.Manifestar(cmd.Context(), snd, api.Manifestacao{TipoManifestacao: tipo, IDNotas: ids, Justificativa: justificativa})
				return err
			})
			if err != nil || !enviado {
				return err
			}
			fmt.Fprintln(s.err, avisoManifestacao)
			var recs []output.Record
			for _, r := range resp {
				situacao := "enviado"
				if r.Situacao != nil && r.Situacao.ID != "" {
					situacao = r.Situacao.ID
				}
				recs = append(recs, *resultadoEscrita("manifestar", situacao, idTexto(r.ID)).Add("manifestacao", "Manifestação", tipo))
			}
			return output.Write(s.out, f, output.RecordsToList(recs))
		},
	}
	cmd.Flags().BoolVar(&ciencia, "ciencia", false, "ciência da operação (preliminar)")
	cmd.Flags().BoolVar(&confirmar, "confirmar", false, "confirmação da operação (recebi)")
	cmd.Flags().BoolVar(&desconhecer, "desconhecer", false, "desconhecimento da operação")
	cmd.Flags().BoolVar(&naoRealizada, "nao-realizada", false, "operação não realizada (não recebi); exige --justificativa")
	cmd.Flags().StringVar(&justificativa, "justificativa", "", "motivo da operação não realizada (15 a 255 caracteres)")
	cmd.Flags().StringVar(&mes, "mes", "", "mês das notas, AAAA-MM (padrão: o mês atual)")
	addWriteFlags(cmd)
	return cmd
}

func valorNota(n api.NotaEntrada) float64 {
	if n.Valor == nil {
		return 0
	}
	return *n.Valor
}
