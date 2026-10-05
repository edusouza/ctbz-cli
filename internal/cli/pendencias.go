package cli

import (
	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// situacaoFinalizada é a situação de uma pendência resolvida.
const situacaoFinalizada = "FINALIZADO"

func newPendenciasCmd() *cobra.Command {
	var todas, failOnVencidas bool
	cmd := &cobra.Command{
		Use:   "pendencias",
		Short: "Lista as pendências da empresa (tipo, detalhe, prazo e situação)",
		Long: `Lista as pendências cadastradas para a empresa, como o card de pendências da tela
inicial do painel: tipo, detalhe, data de criação, prazo e situação.

Por padrão mostra só as abertas; --todas inclui as finalizadas. A coluna alerta marca
prazos vencidos ("vencida") e que vencem em até 7 dias ("próxima"). Com --fail-on-vencidas,
o comando termina com código 4 quando há pendência vencida.

Na tabela, o detalhe é cortado; use -o json para o texto completo.`,
		Example: `  ctbz pendencias
  ctbz pendencias --todas -o json
  ctbz pendencias --fail-on-vencidas || notify-send "Há pendências vencidas"`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ps, err := api.BuscarPendenciasEmpresa(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			l, vencidas := pendenciasList(ps, todas)
			if err := output.Write(s.out, f, l); err != nil {
				return err
			}
			if failOnVencidas && vencidas > 0 {
				return errAttention
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&todas, "todas", false, "inclui as pendências finalizadas")
	cmd.Flags().BoolVar(&failOnVencidas, "fail-on-vencidas", false, "termina com código 4 se houver pendência vencida")
	cmd.AddCommand(newPendenciasConciliacaoCmd(), newPendenciasTermosCmd(), newPendenciasTermoCmd(), newPendenciasAceitarCmd())
	return cmd
}

// pendenciasList monta a lista e conta as pendências vencidas.
func pendenciasList(ps []api.PendenciaEmpresa, todas bool) (*output.List, int) {
	l := &output.List{Columns: []output.Column{
		{Key: "id", Header: "ID"},
		{Key: "tipo", Header: "Tipo"},
		{Key: "detalhe", Header: "Detalhe"},
		{Key: "criada", Header: "Criada em"},
		{Key: "prazo", Header: "Prazo"},
		{Key: "situacao", Header: "Situação"},
		{Key: "alerta", Header: "Alerta"},
	}}
	vencidas := 0
	for _, p := range ps {
		aberta := p.SituacaoPendencia.ID != situacaoFinalizada
		if !aberta && !todas {
			continue
		}
		prazo, _ := output.ParseDate(p.DataLimite)
		alerta := alertaDePrazo(prazo, aberta)
		if alerta == alertaVencida {
			vencidas++
		}
		l.Append(p.ID, p.TipoPendencia.Titulo, textoSimples(p.Detalhe), dateFromMillis(p.DataCriacao),
			prazo, p.SituacaoPendencia.Descricao, alerta)
	}
	return l, vencidas
}
