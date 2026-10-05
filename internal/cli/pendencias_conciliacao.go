package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// tiposConciliacao liga os valores de --listar aos tipos de pendência da API.
var tiposConciliacao = map[string]string{
	"notas":        api.PendenciaNotaSemRecebimento,
	"recebimentos": api.PendenciaRecebimentoSemNota,
}

func newPendenciasConciliacaoCmd() *cobra.Command {
	var listar string
	var failOnPendencias bool
	cmd := &cobra.Command{
		Use:   "conciliacao",
		Short: "Mostra as pendências de conciliação fiscal (notas e recebimentos)",
		Long: `Mostra quantas notas fiscais estão sem recebimento e quantos recebimentos do extrato
estão sem nota, como a tela de conciliação fiscal do painel, e a competência de referência.

Com --listar notas ou --listar recebimentos, lista os itens pendentes dos últimos 12 meses
(o mesmo período do painel). O formato desses itens não pôde ser verificado (a conta usada
no desenvolvimento não tinha pendências), por isso os campos saem como a API os devolve.

Com --fail-on-pendencias, o comando termina com código 4 quando há algo a conciliar.`,
		Example: `  ctbz pendencias conciliacao
  ctbz pendencias conciliacao --listar notas -o json
  ctbz pendencias conciliacao --fail-on-pendencias`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			if listar != "" {
				tipo, ok := tiposConciliacao[listar]
				if !ok {
					return usageError{fmt.Errorf("--listar deve ser notas ou recebimentos: %q", listar)}
				}
				ate := now()
				itens, err := api.BuscarPendenciasConciliacao(cmd.Context(), sessionGetter{s}, tipo, ate.AddDate(-1, 0, 0), ate)
				if err != nil {
					return err
				}
				raw, err := json.Marshal(itens) // []json.RawMessage vira uma lista com os itens intactos
				if err != nil {
					return err
				}
				data, err := output.FromJSON(raw)
				if err != nil {
					return err
				}
				if err := output.Write(s.out, f, data); err != nil {
					return err
				}
				if failOnPendencias && len(itens) > 0 {
					return errAttention
				}
				return nil
			}
			c, err := api.BuscarConciliacaoResumo(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			if err := output.Write(s.out, f, conciliacaoRecord(c)); err != nil {
				return err
			}
			if failOnPendencias && c.QtdNotasFiscaisPendentes+c.QtdRecebimentosPendentes > 0 {
				return errAttention
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&listar, "listar", "", "lista os itens pendentes: notas ou recebimentos")
	cmd.Flags().BoolVar(&failOnPendencias, "fail-on-pendencias", false, "termina com código 4 se houver algo a conciliar")
	cmd.AddCommand(newConciliacaoMotivosCmd(), newConciliacaoCandidatosCmd(), newConciliacaoDetalhesCmd(), newConciliacaoResolverCmd())
	return cmd
}

func conciliacaoRecord(c *api.ConciliacaoResumo) *output.Record {
	rec := &output.Record{}
	rec.Add("competencia_referencia", "Competência de referência", competenciaDeAnoMes(c.CompetenciaMesAnterior))
	rec.Add("notas_fiscais_pendentes", "Notas fiscais sem recebimento", c.QtdNotasFiscaisPendentes)
	rec.Add("recebimentos_pendentes", "Recebimentos sem nota fiscal", c.QtdRecebimentosPendentes)
	rec.Add("conciliacoes_automaticas", "Conciliações automáticas no mês", c.QtdConciliacoesAutomaticasMesAnterior)
	return rec
}

// competenciaDeAnoMes converte "2026-09" em "09/2026"; outro texto volta como veio.
func competenciaDeAnoMes(s string) any {
	ano, mes, ok := strings.Cut(s, "-")
	if !ok || len(ano) != 4 || len(mes) != 2 {
		return nilIfEmpty(s)
	}
	return mes + "/" + ano
}
