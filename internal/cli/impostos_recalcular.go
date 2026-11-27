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

// boolOrNil devolve nil (null no JSON) quando a API não informou o valor.
func boolOrNil(b *bool) any {
	if b == nil {
		return nil
	}
	return *b
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
		Add("cobrar_recalculo", "Cobrado na mensalidade", boolOrNil(r.CobrarRecalculo)).
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
			switch {
			case ini.CobrarRecalculo == nil:
				consequencia = "O painel não informou se será cobrado: pode entrar na sua próxima mensalidade. Não dá para cancelar o pedido."
			case !*ini.CobrarRecalculo:
				consequencia = "Não dá para cancelar o pedido."
				if ini.MotivoSemCobranca != "" {
					consequencia = "Sem cobrança: " + ini.MotivoSemCobranca + ". " + consequencia
				}
			}
			if hoje := api.DiaEmBrasilia(now()); dia.Year() > hoje.Year() || dia.Month() != hoje.Month() {
				consequencia += " Com vencimento no mês seguinte, os juros e a multa ficam maiores."
			}
			estimado := "indisponível"
			if ini.ValorRecalculo != nil {
				estimado = formatarCentavos(centavos(*ini.ValorRecalculo))
			}
			req := api.PedidoRecalculo{Tipo: guia.Tipo, Origem: origemGuia(guia), DataVencimento: dia.Format("02/01/2006")}
			op := operacao{Risco: riscoAlto, ID: args[0], Consequencia: consequencia, Resumo: fmt.Sprintf(
				"Pedir o recálculo de %s com vencimento em %s (valor estimado %s)", nomeGuia(guia), dia.Format("02/01/2006"), estimado)}
			var resp *api.RespostaRecalculo
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				var err error
				resp, err = api.Recalcular(ctx, snd, id, req)
				return err
			})
			if err != nil || !enviado {
				return err
			}
			// O que o painel confirmou (resposta do PUT ou, na falta dela, a guia relida se o
			// vencimento mudou) fica separado do que foi pedido.
			confirmado := ""
			if resp != nil {
				confirmado = resp.Guia.DataVencimento
			}
			situacao := "solicitado"
			status := []string{}
			avisos := []string{}
			avisar := func(msg string) {
				fmt.Fprintln(s.err, "aviso: "+msg)
				avisos = append(avisos, msg)
			}
			if nova, err := api.BuscarGuia(ctx, g, id); err != nil {
				avisar(fmt.Sprintf("pedido enviado, mas não foi possível reler a guia: %v", err))
				situacao = "enviado"
			} else {
				status = nonNil(nova.Status)
				if confirmado == "" && nova.Vencimento != guia.Vencimento {
					confirmado = nova.Vencimento
				}
			}
			// "vencimento" mantém o contrato anterior (sempre uma data: a confirmada ou, na falta
			// dela, a pedida); "vencimento_confirmado" diz qual das duas é.
			novoVenc := confirmado
			if novoVenc == "" {
				novoVenc = req.DataVencimento
				avisar(fmt.Sprintf("o painel não confirmou o novo vencimento; o pedido foi para %s e a guia segue com o vencimento anterior até o recálculo.", req.DataVencimento))
			}
			novo, _ := output.ParseDate(novoVenc)
			pedido, _ := output.ParseDate(req.DataVencimento)
			fmt.Fprintln(s.err, "A guia será recalculada em até 3 dias úteis; a Contabilizei avisa por e-mail quando o novo valor estiver disponível.")
			rec := resultadoEscrita("recalcular", situacao, id).
				Add("guia", "Guia", nomeGuia(guia)).
				Add("status", "Status", status).
				Add("vencimento_pedido", "Vencimento pedido", pedido).
				Add("vencimento", "Novo vencimento", novo).
				Add("vencimento_confirmado", "Vencimento confirmado", confirmado != "").
				Add("cobrado", "Cobrado na mensalidade", boolOrNil(ini.CobrarRecalculo)).
				Add("avisos", "Avisos", avisos)
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().StringVar(&vencimento, "vencimento", "", "novo vencimento AAAA-MM-DD (sem ele, só mostra os dados do recálculo)")
	addWriteFlags(cmd)
	return cmd
}
