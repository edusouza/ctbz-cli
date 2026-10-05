package cli

import (
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
			if resp.Status != 200 {
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
			if resp.Status != 200 {
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
			if len(vincular) == 0 && motivo == "" {
				return usageError{errors.New("informe --vincular ORIGEM:ID ou --motivo MOTIVO")}
			}
			req := api.ResolucaoConciliacao{IDPendencia: ids}
			var resumo []string
			for _, v := range vincular {
				c, err := parseContraparte(v)
				if err != nil {
					return usageError{err}
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
			op := operacao{Risco: riscoAlto, ID: strings.Join(args, ","),
				Resumo:       fmt.Sprintf("Resolver %d pendência(s) de conciliação fiscal (%s): %s", len(ids), strings.Join(args, ", "), strings.Join(resumo, "; ")),
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
