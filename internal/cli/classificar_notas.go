package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// classificacaoNota é uma opção de classificação no vocabulário da CLI, com o nome que a
// API usa em cada nível (ADR-0022).
type classificacaoNota struct {
	nome string // na CLI
	lote string // no caminho do lote
}

var classificacoesNota = []classificacaoNota{
	{"estoque", api.ClassificacaoEstoque},
	{"insumo", api.ClassificacaoInsumo},
	{"uso-consumo", api.ClassificacaoUsoConsumo},
	{"ativo-imobilizado", api.ClassificacaoAtivoImobilizado},
	{"prestacao-servico", api.ClassificacaoPrestacaoServico},
}

func nomesClassificacoes() string {
	nomes := make([]string, len(classificacoesNota))
	for i, c := range classificacoesNota {
		nomes[i] = c.nome
	}
	return strings.Join(nomes, ", ")
}

func classificacaoPorNome(nome string) (classificacaoNota, error) {
	for _, c := range classificacoesNota {
		if c.nome == nome {
			return c, nil
		}
	}
	return classificacaoNota{}, usageError{fmt.Errorf("--como deve ser %s: %q", nomesClassificacoes(), nome)}
}

func newNotasEntradaClassificarCmd() *cobra.Command {
	var como, mes string
	cmd := &cobra.Command{
		Use:   "classificar ID...",
		Short: "Classifica notas de entrada inteiras (estoque, insumo, uso e consumo…)",
		Long: `Classifica as notas da lista "a classificar" numa só opção: estoque, insumo, uso-consumo,
ativo-imobilizado ou prestacao-servico. Vários IDs vão num único envio, como o lote do painel.

O prazo é o dia 05 do mês seguinte à emissão; depois dele a Contabilizei confirma a
pré-classificação sugerida. --mes diz em que mês procurar as notas (padrão: o mês atual).
Os itens de uma nota estão em ctbz notas entrada produtos.

Risco médio: é contábil e pode ser refeito. Aceita --yes e --dry-run.`,
		Example: `  ctbz notas entrada classificar 1000000000000701 --como uso-consumo
  ctbz notas entrada classificar 1000000000000701 1000000000000702 --como estoque --mes 2026-09`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return usageError{errors.New("informe pelo menos um ID de nota (ver ctbz notas entrada --lista a-classificar)")}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if como == "" {
				return usageError{fmt.Errorf("informe --como: %s", nomesClassificacoes())}
			}
			c, err := classificacaoPorNome(como)
			if err != nil {
				return err
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
			g := sessionGetter{s}
			brutas, err := api.BuscarNotasEntradaBrutas(cmd.Context(), g, filtroEntrada(api.ListaAClassificar, ref))
			if err != nil {
				return err
			}
			porID := map[string]int{}
			notas := make([]api.NotaEntrada, len(brutas))
			for i, b := range brutas {
				if err := json.Unmarshal(b, &notas[i]); err != nil {
					return err
				}
				porID[idTexto(notas[i].ID)] = i
			}
			var corpo []json.RawMessage
			var nomes []string
			for _, id := range args {
				i, ok := porID[id]
				if !ok {
					return fmt.Errorf("nota %s não está a classificar em %s; use --mes ou veja ctbz notas entrada --lista a-classificar", id, ref.Format("01/2006"))
				}
				corpo = append(corpo, brutas[i])
				nomes = append(nomes, fmt.Sprintf("%s de %s (%s)", id, notas[i].RazaoSocial, output.FormatBRL(valorNota(notas[i]))))
			}
			op := operacao{Risco: riscoMedio, ID: strings.Join(args, ","),
				Resumo: fmt.Sprintf("Classificar %d nota(s) como %s: %s", len(corpo), c.nome, strings.Join(nomes, "; "))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return api.ClassificarNotasLote(cmd.Context(), snd, c.lote, corpo)
			})
			if err != nil || !enviado {
				return err
			}
			situacoes, err := classificacoesAtuais(cmd, g, ref)
			if err != nil {
				return err
			}
			var recs []output.Record
			for _, id := range args {
				recs = append(recs, *resultadoEscrita("classificar", firstNonEmpty(situacoes[id], "enviado"), id).Add("classificacao", "Classificação", c.nome))
			}
			return output.Write(s.out, f, output.RecordsToList(recs))
		},
	}
	cmd.Flags().StringVar(&como, "como", "", "classificação: "+nomesClassificacoes())
	cmd.Flags().StringVar(&mes, "mes", "", "mês das notas, AAAA-MM (padrão: o mês atual)")
	addWriteFlags(cmd)
	return cmd
}

// classificacoesAtuais relê as listas de classificação do mês e devolve o tipoClassificacao
// de cada nota (ESTOQUE…, ou PROCESSANDO enquanto a Contabilizei aplica).
func classificacoesAtuais(cmd *cobra.Command, g api.Getter, mes time.Time) (map[string]string, error) {
	out := map[string]string{}
	for _, lista := range []string{api.ListaAClassificar, api.ListaClassificadas} {
		ns, err := api.BuscarNotasEntrada(cmd.Context(), g, filtroEntrada(lista, mes))
		if err != nil {
			return nil, err
		}
		for _, n := range ns {
			if n.TipoClassificacao != "" {
				out[idTexto(n.ID)] = n.TipoClassificacao
			}
		}
	}
	return out, nil
}

func newNotasEntradaProdutosCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "produtos ID",
		Short: "Lista os itens de uma nota de entrada e como estão classificados",
		Long: `Lista os produtos de uma NF-e de compra, com a quantidade total e quanto dela está em
estoque, insumo, uso e consumo, ativo imobilizado e prestação de serviço.`,
		Example: `  ctbz notas entrada produtos 1000000000000701`,
		Args:    exactArgs(1, "o ID da nota"),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ps, err := api.ListarProdutosNota(cmd.Context(), sessionGetter{s}, args[0])
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{
				{Key: "id", Header: "ID"}, {Key: "descricao", Header: "Descrição"}, {Key: "ncm", Header: "NCM"},
				{Key: "valor", Header: "Valor"}, {Key: "quantidade", Header: "Quantidade"},
				{Key: "estoque", Header: "Estoque"}, {Key: "insumo", Header: "Insumo"},
				{Key: "uso_consumo", Header: "Uso e consumo"}, {Key: "ativo_imobilizado", Header: "Ativo imobilizado"},
				{Key: "prestacao_servico", Header: "Prestação de serviço"},
			}}
			for _, p := range ps {
				l.Append(nilIfEmpty(idTexto(p.ID)), p.Descricao, nilIfEmpty(p.NCM), moneyOrNil(p.Valor), p.QuantidadeTotal,
					p.QuantidadeEstoque, p.QuantidadeInsumo, p.QuantidadeConsumo, p.QuantidadeAtivo, p.QuantidadePrestacao)
			}
			return output.Write(s.out, f, l)
		},
	}
}
