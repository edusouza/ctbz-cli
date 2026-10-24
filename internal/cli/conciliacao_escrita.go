package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newConciliacaoMotivosCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "motivos",
		Short: "Lista os motivos aceitos para justificar uma pendência de conciliação",
		Long: `Lista os motivos (tipoResolucaoPendencia) que ctbz pendencias conciliacao resolver --motivo
aceita, com o rótulo em português e se o motivo exige --socio. "Outro" não existe na API: no
painel, ele abre um atendimento.`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{
				{Key: "motivo", Header: "Motivo"},
				{Key: "rotulo", Header: "Descrição"},
				{Key: "exige_socio", Header: "Exige sócio"},
			}}
			for _, m := range api.MotivosConciliacao {
				l.Append(m.Codigo, m.Rotulo, m.ExigeSocio)
			}
			return output.Write(streamsOf(cmd).out, f, l)
		},
	}
}

// escreverJSONCru imprime uma resposta JSON como a API devolveu (formato não verificado).
func escreverJSONCru(cmd *cobra.Command, s streams, body []byte) error {
	f, err := outputFormat(cmd, "")
	if err != nil {
		return err
	}
	data, err := output.FromJSON(body)
	if err != nil {
		return fmt.Errorf("resposta inesperada: %w", err)
	}
	return output.Write(s.out, f, data)
}

func newConciliacaoCandidatosCmd() *cobra.Command {
	var notas, recebimentos bool
	cmd := &cobra.Command{
		Use:   "candidatos",
		Short: "Lista as notas ou os recebimentos que podem ser vinculados numa conciliação",
		Long: `Lista os itens que podem ser vinculados a uma pendência: --notas (para pendências de
recebimento sem nota) ou --recebimentos (para notas sem recebimento). Use os ids em
ctbz pendencias conciliacao resolver --vincular.

O formato desta resposta não pôde ser verificado (a conta usada no desenvolvimento não tinha
pendências), por isso os campos saem como a API os devolve.`,
		Example: `  ctbz pendencias conciliacao candidatos --notas -o json`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if notas == recebimentos {
				return usageError{errors.New("informe --notas ou --recebimentos")}
			}
			path := api.PathCandidatosRecebimentos
			if notas {
				path = api.PathCandidatosNotas
			}
			s := streamsOf(cmd)
			resp, err := authedAPI(cmd.Context(), s, "GET", path, nil)
			if err != nil {
				return err
			}
			if !sucesso(resp.Status) {
				return &ctbz.HTTPError{Step: path, Status: resp.Status, Body: resp.Body}
			}
			return escreverJSONCru(cmd, s, resp.Body)
		},
	}
	cmd.Flags().BoolVar(&notas, "notas", false, "notas fiscais para vincular a recebimentos")
	cmd.Flags().BoolVar(&recebimentos, "recebimentos", false, "recebimentos para vincular a notas")
	return cmd
}

func parseIDs(vs []string, nome string) ([]int64, error) {
	ids := make([]int64, 0, len(vs))
	for _, v := range vs {
		id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil || id <= 0 {
			return nil, usageError{fmt.Errorf("%s inválido %q", nome, v)}
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func newConciliacaoDetalhesCmd() *cobra.Command {
	var recebimentos, notas []string
	cmd := &cobra.Command{
		Use:   "detalhes",
		Short: "Mostra os detalhes de uma pendência ou conciliação (justificativa, vínculos)",
		Long: `Mostra os detalhes de pendências ou conciliações de recebimentos (--recebimento) e de notas
(--nota). A API usa POST, mas é só leitura: nada muda e nada é registrado em ctbz acoes.
Os campos saem como a API os devolve (formato não verificado).`,
		Example: `  ctbz pendencias conciliacao detalhes --recebimento 123`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(recebimentos)+len(notas) == 0 {
				return usageError{errors.New("informe --recebimento ID ou --nota ID")}
			}
			req := api.DetalhesConciliacao{}
			var err error
			if req.IDsRecebimento, err = parseIDs(recebimentos, "id de recebimento"); err != nil {
				return err
			}
			if req.IDsReceita, err = parseIDs(notas, "id de nota"); err != nil {
				return err
			}
			body, err := json.Marshal(req)
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			resp, err := authedAPI(cmd.Context(), s, "POST", api.PathDetalhesConciliacao, body)
			if err != nil {
				return err
			}
			if !sucesso(resp.Status) {
				return &ctbz.HTTPError{Step: api.PathDetalhesConciliacao, Status: resp.Status, Body: resp.Body}
			}
			return escreverJSONCru(cmd, s, resp.Body)
		},
	}
	cmd.Flags().StringSliceVar(&recebimentos, "recebimento", nil, "id de recebimento (repita ou separe por vírgula)")
	cmd.Flags().StringSliceVar(&notas, "nota", nil, "id de nota fiscal (repita ou separe por vírgula)")
	return cmd
}

// parseContraparte lê "ORIGEM:ID" (ORIGEM: NOTAFISCAL ou MOVIMENTACAO; aceita nota e recebimento).
func parseContraparte(s string) (api.Contraparte, error) {
	origem, id, ok := strings.Cut(s, ":")
	if !ok {
		return api.Contraparte{}, fmt.Errorf("--vincular inválido %q: use NOTAFISCAL:ID ou MOVIMENTACAO:ID", s)
	}
	switch strings.ToUpper(strings.TrimSpace(origem)) {
	case api.OrigemNotaFiscal, "NOTA":
		origem = api.OrigemNotaFiscal
	case api.OrigemMovimentacao, "RECEBIMENTO":
		origem = api.OrigemMovimentacao
	default:
		return api.Contraparte{}, fmt.Errorf("--vincular inválido %q: a origem é NOTAFISCAL ou MOVIMENTACAO", s)
	}
	n, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if err != nil || n <= 0 {
		return api.Contraparte{}, fmt.Errorf("--vincular inválido %q: id deve ser um número", s)
	}
	return api.Contraparte{Origem: origem, ID: n}, nil
}

func newConciliacaoResolverCmd() *cobra.Command {
	var vincular []string
	var motivo, socio, notaAtivacao string
	cmd := &cobra.Command{
		Use:   "resolver ID_PENDENCIA...",
		Short: "Concilia ou justifica pendências de conciliação fiscal",
		Long: `Resolve pendências de conciliação fiscal (risco alto: a resolução define a natureza fiscal
do recebimento, como receita tributável, empréstimo ou capital; pode ser refeita, mas não
apagada). Três formas, como no painel:

  --vincular ORIGEM:ID…      concilia com notas (NOTAFISCAL, para pendências de recebimento)
                             ou movimentações (MOVIMENTACAO, para pendências de nota)
  --motivo MOTIVO            justifica sem contraparte (ctbz pendencias conciliacao motivos)
  os dois                    concilia com diferença justificada

Motivos que envolvem sócio exigem --socio ID. --nota-ativacao informa o número da nota no
fluxo de ativação contábil. Os ids das pendências vêm de ctbz pendencias conciliacao --listar.

Antes de pedir a confirmação, a CLI lê as pendências abertas dos últimos 12 meses (o mesmo
período de --listar) e recusa ids que não estejam entre elas, ids ou contrapartes repetidos,
NOTAFISCAL e MOVIMENTACAO no mesmo pedido e pendências de tipos diferentes juntas. Com
--vincular, os ids precisam ser do tipo que a origem concilia. O resumo mostra os campos de
cada pendência como a API os devolve (data, valor…).

No painel, algumas opções só valem para itens abaixo de R$ 1.000,00; a CLI não conhece a lista
exata, e a Contabilizei recusa o pedido se a regra não for atendida.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz pendencias conciliacao resolver 123 --vincular NOTAFISCAL:987
  ctbz pendencias conciliacao resolver 123 124 --motivo CAIXA_OU_PESSOA_FISICA
  ctbz pendencias conciliacao resolver 123 --motivo EMPRESTIMO_DO_SOCIO_A_EMPRESA --socio 55 --dry-run`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return usageError{errors.New("informe pelo menos um ID de pendência")}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args, "id de pendência")
			if err != nil {
				return err
			}
			if r := repetidos(ids); len(r) > 0 {
				return usageError{fmt.Errorf("id de pendência repetido: %s; informe cada pendência uma vez", juntarIDs(r))}
			}
			if len(vincular) == 0 && motivo == "" {
				return usageError{errors.New("informe --vincular ORIGEM:ID ou --motivo MOTIVO")}
			}
			req := api.ResolucaoConciliacao{IDPendencia: ids}
			var resumo []string
			vistas := map[api.Contraparte]bool{}
			for _, v := range vincular {
				c, err := parseContraparte(v)
				if err != nil {
					return usageError{err}
				}
				if vistas[c] {
					return usageError{fmt.Errorf("contraparte repetida: %s:%d; informe cada uma uma vez", c.Origem, c.ID)}
				}
				vistas[c] = true
				if len(req.Contraparte) > 0 && req.Contraparte[0].Origem != c.Origem {
					return usageError{errors.New("não misture NOTAFISCAL e MOVIMENTACAO no mesmo pedido: notas conciliam " +
						"recebimentos sem nota fiscal, e movimentações conciliam notas sem recebimento")}
				}
				req.Contraparte = append(req.Contraparte, c)
				resumo = append(resumo, fmt.Sprintf("vincular %s %d", c.Origem, c.ID))
			}
			if motivo != "" {
				m, ok := api.MotivoPorCodigo(strings.ToUpper(motivo))
				if !ok {
					return usageError{fmt.Errorf("motivo desconhecido %q; veja ctbz pendencias conciliacao motivos", motivo)}
				}
				req.TipoResolucaoPendencia = &m.Codigo
				resumo = append(resumo, fmt.Sprintf("motivo %q", m.Rotulo))
				switch {
				case m.ExigeSocio && socio == "":
					return usageError{fmt.Errorf("o motivo %s exige --socio ID", m.Codigo)}
				case !m.ExigeSocio && socio != "":
					return usageError{fmt.Errorf("o motivo %s não usa sócio", m.Codigo)}
				case socio != "":
					id, err := strconv.ParseInt(socio, 10, 64)
					if err != nil {
						return usageError{fmt.Errorf("--socio inválido %q", socio)}
					}
					req.IDVinculo = &id
					resumo = append(resumo, "sócio "+socio)
				}
				req.NumeroNotaAtivacaoContabil = strings.TrimSpace(notaAtivacao)
			} else if socio != "" || notaAtivacao != "" {
				return usageError{errors.New("--socio e --nota-ativacao só valem com --motivo")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			antes, err := api.BuscarConciliacaoResumo(ctx, g)
			if err != nil {
				return err
			}
			// Notas conciliam recebimentos e movimentações conciliam notas; só com --motivo,
			// o tipo vem de onde os ids estão.
			tipos := []string{api.PendenciaNotaSemRecebimento, api.PendenciaRecebimentoSemNota}
			if len(req.Contraparte) > 0 {
				tipos = []string{tipoPorOrigem[req.Contraparte[0].Origem]}
			}
			abertas, err := pendenciasConciliacaoAbertas(ctx, g, tipos)
			if err != nil {
				return err
			}
			tipo, itens, err := conferirPendencias(ids, abertas, tipos)
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoAlto, ID: strings.Join(args, ","),
				Resumo: fmt.Sprintf("Resolver %d pendência(s) de conciliação fiscal (%s) — %s: %s",
					len(ids), rotulosTipoConciliacao[tipo], strings.Join(itens, "; "), strings.Join(resumo, "; ")),
				Consequencia: "A resolução define a natureza fiscal do recebimento (receita tributável, empréstimo, capital…). Pode ser refeita, mas não apagada."}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.ResolverConciliacao(ctx, snd, req) })
			if err != nil || !enviado {
				return err
			}
			situacao := "resolvido"
			var pendentes any
			if depois, err := api.BuscarConciliacaoResumo(ctx, g); err != nil {
				fmt.Fprintln(s.err, "aviso: pedido enviado, mas não foi possível reler a conciliação:", err)
				situacao = "enviado"
			} else {
				total := depois.QtdNotasFiscaisPendentes + depois.QtdRecebimentosPendentes
				pendentes = total
				if total >= antes.QtdNotasFiscaisPendentes+antes.QtdRecebimentosPendentes {
					fmt.Fprintln(s.err, "aviso: o total de pendências não diminuiu; confira com ctbz pendencias conciliacao")
					situacao = "enviado"
				}
			}
			rec := resultadoEscrita("resolver-conciliacao", situacao, strings.Join(args, ",")).
				Add("pendencias_restantes", "Pendências restantes", pendentes)
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().StringArrayVar(&vincular, "vincular", nil, "contraparte ORIGEM:ID (NOTAFISCAL ou MOVIMENTACAO; repita para várias)")
	cmd.Flags().StringVar(&motivo, "motivo", "", "motivo da resolução (ver ctbz pendencias conciliacao motivos)")
	cmd.Flags().StringVar(&socio, "socio", "", "sócio, nos motivos que exigem")
	cmd.Flags().StringVar(&notaAtivacao, "nota-ativacao", "", "número da nota no fluxo de ativação contábil")
	addWriteFlags(cmd)
	return cmd
}

// tipoPorOrigem liga a origem das contrapartes ao tipo de pendência que ela concilia.
var tipoPorOrigem = map[string]string{
	api.OrigemNotaFiscal:   api.PendenciaRecebimentoSemNota,
	api.OrigemMovimentacao: api.PendenciaNotaSemRecebimento,
}

// rotulosTipoConciliacao nomeia os tipos de pendência nas mensagens.
var rotulosTipoConciliacao = map[string]string{
	api.PendenciaNotaSemRecebimento: "notas sem recebimento",
	api.PendenciaRecebimentoSemNota: "recebimentos sem nota fiscal",
}

// pendenciaConciliacao é uma pendência lida de BuscarPendenciasConciliacao, com os campos que a
// API devolveu resumidos para a confirmação.
type pendenciaConciliacao struct {
	tipo, descricao string
}

// pendenciasConciliacaoAbertas lê as pendências abertas dos tipos pedidos, na mesma janela de
// ctbz pendencias conciliacao --listar, indexadas pelo id. O formato dos itens não foi
// verificado: o id é o campo "id" (o que --listar mostra) e, sem ele, a leitura falha em vez
// de deixar passar um id que não dá para conferir.
func pendenciasConciliacaoAbertas(ctx context.Context, g api.Getter, tipos []string) (map[int64][]pendenciaConciliacao, error) {
	de, ate := janelaConciliacao()
	abertas := map[int64][]pendenciaConciliacao{}
	for _, tipo := range tipos {
		itens, err := api.BuscarPendenciasConciliacao(ctx, g, tipo, de, ate)
		if err != nil {
			return nil, fmt.Errorf("lendo as pendências de conciliação: %w", err)
		}
		for _, item := range itens {
			id, descricao, err := lerPendencia(item)
			if err != nil {
				return nil, err
			}
			abertas[id] = append(abertas[id], pendenciaConciliacao{tipo: tipo, descricao: descricao})
		}
	}
	return abertas, nil
}

// lerPendencia tira o id de um item de pendência e descreve os outros campos simples (data,
// valor, descrição…) na ordem em que a API os devolveu, sem caracteres de controle.
func lerPendencia(item json.RawMessage) (int64, string, error) {
	errSemID := errors.New("pendência de conciliação sem id numérico na resposta: o formato não é o esperado; resolva pelo painel")
	v, err := output.FromJSON(item)
	if err != nil {
		return 0, "", fmt.Errorf("lendo uma pendência de conciliação: %w", err)
	}
	rec, ok := v.(*output.Record)
	if !ok {
		return 0, "", errSemID
	}
	var id int64
	var campos []string
	for _, f := range rec.Fields {
		var texto string
		switch x := f.Value.(type) {
		case json.Number:
			texto = x.String()
		case string:
			texto = x
		case bool:
			texto = strconv.FormatBool(x)
		default: // nulo, objeto ou lista
			continue
		}
		if f.Key == "id" {
			if id, err = strconv.ParseInt(texto, 10, 64); err != nil || id <= 0 {
				return 0, "", errSemID
			}
			continue
		}
		if texto = strings.Join(strings.Fields(semControle(texto)), " "); texto != "" {
			campos = append(campos, semControle(f.Key)+" "+texto)
		}
	}
	if id == 0 {
		return 0, "", errSemID
	}
	return id, strings.Join(campos, ", "), nil
}

// conferirPendencias recusa ids que não estão abertos, que aparecem em mais de um tipo ou
// que misturam tipos, e devolve o tipo e a descrição de cada pendência, na ordem dos ids.
func conferirPendencias(ids []int64, abertas map[int64][]pendenciaConciliacao, tipos []string) (string, []string, error) {
	var ausentes []int64
	for _, id := range ids {
		switch ps := abertas[id]; {
		case len(ps) == 0:
			ausentes = append(ausentes, id)
		case len(ps) > 1:
			return "", nil, usageError{fmt.Errorf("a pendência %d aparece entre as %s e entre os %s; use --vincular para indicar o tipo ou resolva pelo painel",
				id, rotulosTipoConciliacao[api.PendenciaNotaSemRecebimento], rotulosTipoConciliacao[api.PendenciaRecebimentoSemNota])}
		}
	}
	if len(ausentes) > 0 {
		var rotulos []string
		for _, t := range tipos {
			rotulos = append(rotulos, rotulosTipoConciliacao[t])
		}
		listar := "notas|recebimentos"
		if len(tipos) == 1 && tipos[0] == api.PendenciaNotaSemRecebimento {
			listar = "notas"
		} else if len(tipos) == 1 {
			listar = "recebimentos"
		}
		verbo := "não está"
		if len(ausentes) > 1 {
			verbo = "não estão"
		}
		return "", nil, usageError{fmt.Errorf("%s %s entre as pendências abertas (%s) dos últimos 12 meses; confira os ids com ctbz pendencias conciliacao --listar %s",
			juntarIDs(ausentes), verbo, strings.Join(rotulos, " ou "), listar)}
	}
	tipo := abertas[ids[0]][0].tipo
	itens := make([]string, 0, len(ids))
	for _, id := range ids {
		p := abertas[id][0]
		if p.tipo != tipo {
			return "", nil, usageError{fmt.Errorf("as pendências misturam %s (%d) e %s (%d); resolva cada tipo num pedido separado",
				rotulosTipoConciliacao[tipo], ids[0], rotulosTipoConciliacao[p.tipo], id)}
		}
		item := strconv.FormatInt(id, 10)
		if p.descricao != "" {
			item += " [" + p.descricao + "]"
		}
		itens = append(itens, item)
	}
	return tipo, itens, nil
}

// repetidos devolve os ids que aparecem mais de uma vez, na ordem da primeira repetição.
func repetidos(ids []int64) []int64 {
	vistos := map[int64]int{}
	var out []int64
	for _, id := range ids {
		if vistos[id]++; vistos[id] == 2 {
			out = append(out, id)
		}
	}
	return out
}

func juntarIDs(ids []int64) string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(s, ", ")
}
