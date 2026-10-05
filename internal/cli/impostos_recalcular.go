package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// permiteRecalculo diz se o painel oferece o recálculo para a guia (vencida e marcada como
// não paga): alguma ação de recálculo em acoesBotoes.
func permiteRecalculo(g *api.GuiaDetalhe) bool {
	for _, a := range g.AcoesBotoes {
		if strings.Contains(strings.ToUpper(a), "RECALC") {
			return true
		}
	}
	return false
}

func recalculoRecord(id int64, r *api.RecalculoInit) *output.Record {
	recomendada, _ := output.ParseDate(r.DataRecomendada)
	venc, _ := output.ParseDate(r.Vencimento)
	rec := &output.Record{}
	return rec.Add("id", "ID", id).
		Add("descricao", "Guia", nilIfEmpty(r.Descricao)).
		Add("vencimento", "Vencimento original", venc).
		Add("valor_original", "Valor original", moneyOrNil(r.ValorOriginal)).
		Add("valor_recalculo", "Valor recalculado (estimado)", moneyOrNil(r.ValorRecalculo)).
		Add("data_recomendada", "Data recomendada", recomendada).
		Add("datas_indisponiveis", "Datas indisponíveis", nonNil(r.DatasIndisponiveis)).
		Add("cobrar_recalculo", "Cobrado na mensalidade", r.CobrarRecalculo).
		Add("motivo_sem_cobranca", "Motivo sem cobrança", nilIfEmpty(r.MotivoSemCobranca))
}

func newImpostosRecalcularCmd() *cobra.Command {
	var vencimento string
	cmd := &cobra.Command{
		Use:   "recalcular ID",
		Short: "Mostra ou pede o recálculo de uma guia vencida",
		Long: `Sem --vencimento, mostra os dados do recálculo de uma guia vencida: valor original,
valor recalculado estimado, data recomendada, datas indisponíveis e se o serviço será cobrado.
Nada é enviado.

Com --vencimento AAAA-MM-DD, pede uma nova guia com juros e multa para essa data (risco alto:
o recálculo é um serviço adicional, em geral cobrado na próxima mensalidade, e não dá para
cancelar o pedido). A guia precisa estar vencida e marcada como não paga
(ctbz impostos confirmar ID --nao-paguei); datas indisponíveis e passadas são recusadas, e
pendências críticas na Central de Rotinas bloqueiam o pedido, como no painel.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz impostos recalcular 1000000000000001
  ctbz impostos recalcular 1000000000000001 --vencimento 2026-10-20`,
		Args: exactArgs(1, "o ID da guia"),
		RunE: func(cmd *cobra.Command, args []string) error {
			var dia time.Time
			if vencimento != "" {
				d, err := time.Parse("2006-01-02", vencimento)
				if err != nil {
					return usageError{fmt.Errorf("vencimento inválido %q: use AAAA-MM-DD", vencimento)}
				}
				if d.Before(api.DiaEmBrasilia(now())) {
					return usageError{fmt.Errorf("o vencimento %s já passou", d.Format("02/01/2006"))}
				}
				dia = d
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			guia, err := api.BuscarGuia(ctx, g, id)
			if err != nil {
				return err
			}
			ini, err := api.BuscarRecalculo(ctx, g, id, origemGuia(guia), guia.Tipo)
			if err != nil {
				return err
			}
			if vencimento == "" {
				if ini.DataRecomendada != "" {
					fmt.Fprintf(s.err, "Para pedir: ctbz impostos recalcular %d --vencimento %s\n", id, ini.DataRecomendada)
				}
				return output.Write(s.out, f, recalculoRecord(id, ini))
			}
			for _, d := range ini.DatasIndisponiveis {
				if strings.HasPrefix(d, vencimento) {
					return usageError{fmt.Errorf("o vencimento %s está indisponível; a data recomendada é %s", dia.Format("02/01/2006"), ini.DataRecomendada)}
				}
			}
			if !permiteRecalculo(guia) {
				return errors.New("o painel não oferece recálculo para esta guia agora: ela precisa estar vencida e marcada como não paga (ctbz impostos confirmar ID --nao-paguei)")
			}
			if err := semPendenciasCriticas(ctx, g); err != nil {
				return err
			}
			consequencia := "O recálculo é um serviço adicional, cobrado na sua próxima mensalidade. Não dá para cancelar o pedido."
			if !ini.CobrarRecalculo {
				consequencia = "Não dá para cancelar o pedido."
				if ini.MotivoSemCobranca != "" {
					consequencia = "Sem cobrança: " + ini.MotivoSemCobranca + ". " + consequencia
				}
			}
			if hoje := api.DiaEmBrasilia(now()); dia.Year() > hoje.Year() || dia.Month() != hoje.Month() {
				consequencia += " Com vencimento no mês seguinte, os juros e a multa ficam maiores."
			}
			valor := int64(0)
			if ini.ValorRecalculo != nil {
				valor = centavos(*ini.ValorRecalculo)
			}
			req := api.PedidoRecalculo{Tipo: guia.Tipo, Origem: origemGuia(guia), DataVencimento: dia.Format("02/01/2006")}
			op := operacao{Risco: riscoAlto, ID: args[0], Consequencia: consequencia, Resumo: fmt.Sprintf(
				"Pedir o recálculo de %s com vencimento em %s (valor estimado %s)", nomeGuia(guia), dia.Format("02/01/2006"), formatarCentavos(valor))}
			var resp *api.RespostaRecalculo
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				var err error
				resp, err = api.Recalcular(ctx, snd, id, req)
				return err
			})
			if err != nil || !enviado {
				return err
			}
			novo := req.DataVencimento
			if resp != nil && resp.Guia.DataVencimento != "" {
				novo = resp.Guia.DataVencimento
			}
			fmt.Fprintln(s.err, "A guia será recalculada em até 3 dias úteis; a Contabilizei avisa por e-mail quando o novo valor estiver disponível.")
			d, _ := output.ParseDate(novo)
			rec := resultadoEscrita("recalcular", "solicitado", id).
				Add("guia", "Guia", nomeGuia(guia)).
				Add("vencimento", "Novo vencimento", d).
				Add("cobrado", "Cobrado na mensalidade", ini.CobrarRecalculo)
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().StringVar(&vencimento, "vencimento", "", "novo vencimento AAAA-MM-DD (sem ele, só mostra os dados do recálculo)")
	addWriteFlags(cmd)
	return cmd
}
