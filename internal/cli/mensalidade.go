package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newMensalidadeCmd() *cobra.Command {
	var failOnAtraso bool
	cmd := &cobra.Command{
		Use:   "mensalidade",
		Short: "Mostra a mensalidade atual da Contabilizei (valor, vencimento e situação)",
		Long: `Mostra a fatura atual da Contabilizei: competência, valor, vencimento, situação e
observações (ex.: serviços adicionais), e se há mensalidade de competência anterior em atraso.
Traz também os botões do card e as ações deles (botao_acao, botao_adicional_acao): com a
mensalidade atrasada, o botão adicional gera a segunda via (ctbz mensalidade segunda-via).

Com --fail-on-atraso, o comando termina com código 4 quando há competência anterior em atraso.`,
		Example: `  ctbz mensalidade
  ctbz mensalidade -o json | jq .valor
  ctbz mensalidade --fail-on-atraso`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			fat, err := api.BuscarFatura(cmd.Context(), g)
			if err != nil {
				return err
			}
			m, err := api.BuscarMensalidade(cmd.Context(), g)
			if err != nil {
				return err
			}
			if err := output.Write(s.out, f, mensalidadeRecord(fat, m)); err != nil {
				return err
			}
			if failOnAtraso && m.CompetenciaAnteriorAtrasada {
				return errAttention
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&failOnAtraso, "fail-on-atraso", false, "termina com código 4 se houver competência anterior em atraso")
	cmd.AddCommand(newMensalidadeSituacaoCmd())
	cmd.AddCommand(newMensalidadeHistoricoCmd())
	cmd.AddCommand(newMensalidadeSegundaViaCmd())
	return cmd
}

func newMensalidadeSituacaoCmd() *cobra.Command {
	var failOnInadimplencia bool
	cmd := &cobra.Command{
		Use:   "situacao",
		Short: "Indica se a empresa está em dia com a Contabilizei",
		Long: `Indica se a empresa está em dia com a Contabilizei (consulta de inadimplência do painel).
A API responde "OK" para empresas em dia; qualquer outra resposta é mostrada como veio e
tratada como "não está em dia".

Com --fail-on-inadimplencia, o comando termina com código 4 quando a empresa não está em dia.`,
		Example: `  ctbz mensalidade situacao
  ctbz mensalidade situacao --fail-on-inadimplencia || echo "Mensalidade pendente"`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			situacao, err := api.BuscarSituacaoMensalidade(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			emDia := situacao == api.SituacaoEmDia
			rec := &output.Record{}
			rec.Add("em_dia", "Em dia", emDia)
			rec.Add("situacao", "Situação", nilIfEmpty(situacao))
			if err := output.Write(s.out, f, rec); err != nil {
				return err
			}
			if failOnInadimplencia && !emDia {
				return errAttention
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&failOnInadimplencia, "fail-on-inadimplencia", false, "termina com código 4 se a empresa não estiver em dia")
	return cmd
}

func mensalidadeRecord(fat *api.Fatura, m *api.Mensalidade) *output.Record {
	valor := moneyOrNil(fat.Total)
	if valor == nil {
		valor = moneyOrNil(m.Valor)
	}
	vencimento := dataDeValor(fat.Vencimento)
	if vencimento == nil {
		vencimento = dataDeValor(m.DataVencimento)
	}
	rec := &output.Record{}
	rec.Add("competencia", "Competência", competenciaDeTexto(fat.Competencia))
	rec.Add("valor", "Valor", valor)
	rec.Add("vencimento", "Vencimento", vencimento)
	rec.Add("situacao", "Situação", nilIfEmpty(fat.Status))
	rec.Add("observacao", "Observação", nilIfEmpty(fat.MensagemAdicional))
	rec.Add("competencia_anterior_atrasada", "Competência anterior em atraso", m.CompetenciaAnteriorAtrasada)
	rec.Add("botao_titulo", "Botão do card", nilIfEmpty(fat.BotaoTitulo))
	rec.Add("botao_acao", "Ação do botão", nilIfEmpty(fat.BotaoAcao))
	rec.Add("botao_adicional_titulo", "Botão adicional", nilIfEmpty(fat.BotaoAdicionalTitulo))
	rec.Add("botao_adicional_acao", "Ação do botão adicional", nilIfEmpty(fat.BotaoAdicionalAcao))
	return rec
}

// dataDeValor interpreta uma data de formato não verificado: texto ("dd/mm/aaaa",
// "aaaa-mm-dd"…) ou epoch em milissegundos. Outros valores voltam como texto.
func dataDeValor(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		return dataOuTexto(x)
	case float64:
		return dateFromMillis(int64(x))
	}
	return fmt.Sprint(v)
}
