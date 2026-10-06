package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// classificacaoNota é uma opção de classificação no vocabulário da CLI, com o nome que a
// API usa em cada nível (ADR-0022).
type classificacaoNota struct {
	nome  string                          // na CLI
	lote  string                          // no caminho do lote
	campo func(*api.Quantidades) *float64 // no produto
}

var classificacoesNota = []classificacaoNota{
	{"estoque", api.ClassificacaoEstoque, func(q *api.Quantidades) *float64 { return &q.Estoque }},
	{"insumo", api.ClassificacaoInsumo, func(q *api.Quantidades) *float64 { return &q.Insumo }},
	{"uso-consumo", api.ClassificacaoUsoConsumo, func(q *api.Quantidades) *float64 { return &q.Consumo }},
	{"ativo-imobilizado", api.ClassificacaoAtivoImobilizado, func(q *api.Quantidades) *float64 { return &q.Ativo }},
	{"prestacao-servico", api.ClassificacaoPrestacaoServico, func(q *api.Quantidades) *float64 { return &q.Prestacao }},
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
	var produtos []string
	cmd := &cobra.Command{
		Use:   "classificar ID...",
		Short: "Classifica notas de entrada (estoque, insumo, uso e consumo…), inteiras ou por produto",
		Long: `Classifica notas da lista "a classificar" do mês: estoque, insumo, uso-consumo,
ativo-imobilizado ou prestacao-servico.

--como classifica as notas inteiras numa só opção; vários IDs vão num único envio, como o lote
do painel. --produto distribui a quantidade de cada item de uma nota (um ID só), no formato
ITEM:opcao=quantidade,opcao=quantidade (quantidades com ponto decimal); os itens não citados
ficam como estão. A soma de cada item precisa ser igual à quantidade total, sem negativos. Os
itens e a distribuição atual estão em ctbz notas entrada produtos.

O prazo é o dia 05 do mês seguinte à emissão; depois dele a Contabilizei confirma a
pré-classificação sugerida. --mes diz em que mês procurar as notas (padrão: o mês atual).
Para mudar uma nota já classificada: ctbz notas entrada reclassificar.

Risco médio: é contábil e pode ser refeito. Aceita --yes e --dry-run.`,
		Example: `  ctbz notas entrada classificar 1000000000000701 --como uso-consumo
  ctbz notas entrada classificar 1000000000000701 1000000000000702 --como estoque --mes 2026-09
  ctbz notas entrada classificar 1000000000000701 --produto 11:estoque=2,uso-consumo=3 --produto 12:insumo=1`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return usageError{errors.New("informe pelo menos um ID de nota (ver ctbz notas entrada --lista a-classificar)")}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := comoOuProduto(como, produtos, args)
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
			brutas, notas, err := notasDaLista(cmd, g, api.ListaAClassificar, ref, args, "não está a classificar")
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoMedio, ID: strings.Join(args, ",")}
			var enviar func(api.Sender) error
			if c != nil {
				op.Resumo = fmt.Sprintf("Classificar %d nota(s) como %s: %s", len(notas), c.nome, descreverNotas(args, notas))
				enviar = func(snd api.Sender) error { return api.ClassificarNotasLote(cmd.Context(), snd, c.lote, brutas) }
			} else {
				corpo, resumo, err := distribuir(cmd, g, args[0], nil, produtos)
				if err != nil {
					return err
				}
				op.Resumo = fmt.Sprintf("Classificar por produto a nota %s: %s", descreverNotas(args, notas), resumo)
				enviar = func(snd api.Sender) error { return api.ClassificarProdutos(cmd.Context(), snd, corpo) }
			}
			enviado, err := escrever(cmd, op, enviar)
			if err != nil || !enviado {
				return err
			}
			return resultadoClassificacao(cmd, f, g, ref, "classificar", args, c)
		},
	}
	cmd.Flags().StringVar(&como, "como", "", "classifica as notas inteiras: "+nomesClassificacoes())
	cmd.Flags().StringArrayVar(&produtos, "produto", nil, "distribui um item, ITEM:opcao=qtd,opcao=qtd (repetível)")
	cmd.Flags().StringVar(&mes, "mes", "", "mês das notas, AAAA-MM (padrão: o mês atual)")
	addWriteFlags(cmd)
	return cmd
}

func newNotasEntradaReclassificarCmd() *cobra.Command {
	var como, mes string
	var produtos []string
	cmd := &cobra.Command{
		Use:   "reclassificar ID",
		Short: "Muda a classificação de uma nota de entrada já classificada",
		Long: `Reclassifica uma nota da lista "classificadas" do mês. --como põe toda a quantidade de
cada item numa só opção (estoque, insumo, uso-consumo, ativo-imobilizado ou
prestacao-servico); --produto distribui item a item, como em ctbz notas entrada classificar.

A nota fica PROCESSANDO até a Contabilizei aplicar a nova classificação.
Risco médio: é contábil e pode ser refeito. Aceita --yes e --dry-run.`,
		Example: `  ctbz notas entrada reclassificar 1000000000000701 --como estoque --mes 2026-09
  ctbz notas entrada reclassificar 1000000000000701 --produto 11:estoque=5`,
		Args: exactArgs(1, "o ID da nota"),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := comoOuProduto(como, produtos, args)
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
			_, notas, err := notasDaLista(cmd, g, api.ListaClassificadas, ref, args, "não está classificada")
			if err != nil {
				return err
			}
			corpo, resumo, err := distribuir(cmd, g, args[0], c, produtos)
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoMedio, ID: args[0], Resumo: fmt.Sprintf("Reclassificar a nota %s (hoje %s): %s",
				descreverNotas(args, notas), firstNonEmpty(notas[0].TipoClassificacao, "sem classificação"), resumo)}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return api.ReclassificarNota(cmd.Context(), snd, args[0], corpo)
			})
			if err != nil || !enviado {
				return err
			}
			return resultadoClassificacao(cmd, f, g, ref, "reclassificar", args, c)
		},
	}
	cmd.Flags().StringVar(&como, "como", "", "toda a quantidade numa opção: "+nomesClassificacoes())
	cmd.Flags().StringArrayVar(&produtos, "produto", nil, "distribui um item, ITEM:opcao=qtd,opcao=qtd (repetível)")
	cmd.Flags().StringVar(&mes, "mes", "", "mês da nota, AAAA-MM (padrão: o mês atual)")
	addWriteFlags(cmd)
	return cmd
}

// comoOuProduto confere que só um modo foi escolhido e devolve a classificação de --como
// (nil no modo --produto, que aceita uma nota só).
func comoOuProduto(como string, produtos []string, ids []string) (*classificacaoNota, error) {
	switch {
	case como == "" && len(produtos) == 0:
		return nil, usageError{fmt.Errorf("informe --como (%s) ou --produto", nomesClassificacoes())}
	case como != "" && len(produtos) > 0:
		return nil, usageError{errors.New("use --como ou --produto, não os dois")}
	case len(produtos) > 0 && len(ids) > 1:
		return nil, usageError{errors.New("--produto classifica uma nota por vez")}
	case len(produtos) > 0:
		return nil, nil
	}
	c, err := classificacaoPorNome(como)
	return &c, err
}

// notasDaLista acha as notas pelos IDs numa lista do mês; devolve-as como vieram e lidas.
func notasDaLista(cmd *cobra.Command, g api.Getter, lista string, mes time.Time, ids []string, falta string) ([]json.RawMessage, []api.NotaEntrada, error) {
	todas, err := api.BuscarNotasEntradaBrutas(cmd.Context(), g, filtroEntrada(lista, mes))
	if err != nil {
		return nil, nil, err
	}
	porID := map[string]api.NotaEntrada{}
	brutaPorID := map[string]json.RawMessage{}
	for _, b := range todas {
		var n api.NotaEntrada
		if err := json.Unmarshal(b, &n); err != nil {
			return nil, nil, err
		}
		porID[idTexto(n.ID)], brutaPorID[idTexto(n.ID)] = n, b
	}
	brutas := make([]json.RawMessage, len(ids))
	notas := make([]api.NotaEntrada, len(ids))
	for i, id := range ids {
		n, ok := porID[id]
		if !ok {
			return nil, nil, fmt.Errorf("nota %s %s em %s; use --mes ou veja ctbz notas entrada --lista %s", id, falta, mes.Format("01/2006"), lista)
		}
		brutas[i], notas[i] = brutaPorID[id], n
	}
	return brutas, notas, nil
}

func descreverNotas(ids []string, notas []api.NotaEntrada) string {
	nomes := make([]string, len(ids))
	for i, id := range ids {
		nomes[i] = fmt.Sprintf("%s de %s (%s)", id, notas[i].RazaoSocial, output.FormatBRL(valorNota(notas[i])))
	}
	return strings.Join(nomes, "; ")
}

// toleranciaQuantidade absorve o arredondamento de ponto flutuante na soma das quantidades.
const toleranciaQuantidade = 1e-6

// distribuir lê os itens da nota e aplica --como (toda a quantidade de cada item numa opção)
// ou --produto (os itens citados; os demais ficam como estão). Confere as regras do front em
// todos os itens e devolve o corpo da escrita e um resumo da distribuição.
func distribuir(cmd *cobra.Command, g api.Getter, idNota string, como *classificacaoNota, especs []string) ([]json.RawMessage, string, error) {
	itens, err := api.ListarProdutosNota(cmd.Context(), g, idNota)
	if err != nil {
		return nil, "", err
	}
	if len(itens) == 0 {
		return nil, "", fmt.Errorf("a nota %s não tem itens para classificar", idNota)
	}
	novas, err := lerEspecsProduto(especs)
	if err != nil {
		return nil, "", err
	}
	corpo := make([]json.RawMessage, len(itens))
	resumos := make([]string, len(itens))
	for i, p := range itens {
		id := idTexto(p.ID)
		q, citado := novas[id]
		switch {
		case como != nil:
			q = api.Quantidades{}
			*como.campo(&q) = p.QuantidadeTotal
		case !citado:
			q = p.Quantidades()
		}
		delete(novas, id)
		if q.Negativa() {
			return nil, "", usageError{fmt.Errorf("item %s (%s): a classificação não pode ser negativa", id, p.Descricao)}
		}
		if d := q.Soma() - p.QuantidadeTotal; d > toleranciaQuantidade || d < -toleranciaQuantidade {
			return nil, "", usageError{fmt.Errorf("item %s (%s): distribua toda a quantidade (%s de %s); use --produto %s:opcao=qtd",
				id, p.Descricao, quantidadeTexto(q.Soma()), quantidadeTexto(p.QuantidadeTotal), id)}
		}
		if corpo[i], err = p.ComQuantidades(q); err != nil {
			return nil, "", err
		}
		resumos[i] = fmt.Sprintf("item %s %s: %s", id, p.Descricao, descreverQuantidades(q))
	}
	if len(novas) > 0 {
		return nil, "", usageError{fmt.Errorf("a nota %s não tem o item %s (ver ctbz notas entrada produtos %s)",
			idNota, strings.Join(slices.Sorted(maps.Keys(novas)), ", "), idNota)}
	}
	return corpo, strings.Join(resumos, "; "), nil
}

// lerEspecsProduto lê os --produto ITEM:opcao=qtd,opcao=qtd.
func lerEspecsProduto(especs []string) (map[string]api.Quantidades, error) {
	out := map[string]api.Quantidades{}
	for _, e := range especs {
		id, dist, ok := strings.Cut(e, ":")
		id = strings.TrimSpace(id)
		if !ok || id == "" || strings.TrimSpace(dist) == "" {
			return nil, usageError{fmt.Errorf("--produto deve ser ITEM:opcao=qtd,opcao=qtd: %q", e)}
		}
		if _, dup := out[id]; dup {
			return nil, usageError{fmt.Errorf("item %s repetido em --produto", id)}
		}
		var q api.Quantidades
		for _, par := range strings.Split(dist, ",") {
			nome, qtd, ok := strings.Cut(par, "=")
			if !ok {
				return nil, usageError{fmt.Errorf("--produto %s: esperado opcao=qtd em %q (quantidades usam ponto decimal: 2.5)", id, par)}
			}
			c, err := classificacaoPorNome(strings.TrimSpace(nome))
			if err != nil {
				return nil, usageError{fmt.Errorf("--produto %s: opção deve ser %s: %q", id, nomesClassificacoes(), strings.TrimSpace(nome))}
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(qtd), 64)
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, usageError{fmt.Errorf("--produto %s: quantidade inválida %q (use ponto decimal: 2.5)", id, qtd)}
			}
			*c.campo(&q) += v
		}
		out[id] = q
	}
	return out, nil
}

func descreverQuantidades(q api.Quantidades) string {
	var partes []string
	for _, c := range classificacoesNota {
		if v := *c.campo(&q); v != 0 {
			partes = append(partes, c.nome+" "+quantidadeTexto(v))
		}
	}
	if len(partes) == 0 {
		return "nada"
	}
	return strings.Join(partes, ", ")
}

func quantidadeTexto(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// resultadoClassificacao relê as listas do mês e mostra a classificação de cada nota.
func resultadoClassificacao(cmd *cobra.Command, f output.Format, g api.Getter, mes time.Time, acao string, ids []string, c *classificacaoNota) error {
	situacoes, err := classificacoesAtuais(cmd, g, mes)
	if err != nil {
		return err
	}
	pedida := "por produto"
	if c != nil {
		pedida = c.nome
	}
	recs := make([]output.Record, len(ids))
	for i, id := range ids {
		recs[i] = *resultadoEscrita(acao, firstNonEmpty(situacoes[id], "enviado"), id).Add("classificacao", "Classificação", pedida)
	}
	return output.Write(streamsOf(cmd).out, f, output.RecordsToList(recs))
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
