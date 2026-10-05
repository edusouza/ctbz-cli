package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// avisoAuditoria é o aviso do painel sobre a confirmação de pagamento.
const avisoAuditoria = "Essa confirmação não serve como comprovante oficial: a Contabilizei faz auditorias periódicas e pode atualizar o status com dados oficiais do governo."

// semPendenciasCriticas recusa a escrita quando a Central de Rotinas tem pendência crítica,
// que no painel bloqueia confirmação de pagamento e recálculo ("Resolva as pendências").
func semPendenciasCriticas(ctx context.Context, g api.Getter) error {
	c, err := api.BuscarCentralRotinas(ctx, g)
	if err != nil {
		return err
	}
	var abertas []string
	for nome, p := range c.Pendencias.Criticas {
		if p.PossuiPendencia {
			abertas = append(abertas, nome)
		}
	}
	if len(abertas) == 0 {
		return nil
	}
	sort.Strings(abertas)
	return fmt.Errorf("há pendências críticas na Central de Rotinas (%s); resolva-as antes (veja ctbz resumo): o painel também bloqueia esta ação", strings.Join(abertas, ", "))
}

// guiaParaEscrita lê a guia e a versão da rota de confirmação.
func guiaParaEscrita(ctx context.Context, g api.Getter, idArg string) (*api.GuiaDetalhe, string, error) {
	id, err := parseID(idArg)
	if err != nil {
		return nil, "", err
	}
	guia, err := api.BuscarGuia(ctx, g, id)
	if err != nil {
		return nil, "", err
	}
	r, err := api.BuscarRollout(ctx, g)
	if err != nil {
		return nil, "", err
	}
	return guia, api.VersaoConfirmacao(r.Versao), nil
}

func origemGuia(g *api.GuiaDetalhe) string {
	if g.Origem != "" {
		return g.Origem
	}
	return api.OrigemGuiaPadrao
}

func nomeGuia(g *api.GuiaDetalhe) string {
	nome := g.Oraculo.Nome
	if nome == "" {
		nome = g.IdentificadorImposto
	}
	if g.Competencia != "" {
		nome += " de " + g.Competencia
	}
	return nome
}

// confirmarPagamento é o corpo comum de confirmar, --nao-paguei e desmarcar.
func confirmarPagamento(cmd *cobra.Command, idArg string, pago bool, acao, verbo string) error {
	f, err := outputFormat(cmd, "")
	if err != nil {
		return err
	}
	s := streamsOf(cmd)
	ctx := cmd.Context()
	g := sessionGetter{s}
	guia, versao, err := guiaParaEscrita(ctx, g, idArg)
	if err != nil {
		return err
	}
	if err := semPendenciasCriticas(ctx, g); err != nil {
		return err
	}
	req := api.ConfirmacaoPagamento{Tipo: guia.Tipo, Origem: origemGuia(guia), PagamentoConfirmado: pago}
	op := operacao{Risco: riscoMedio, ID: idArg, Resumo: fmt.Sprintf("%s: %s (%s), situação atual %s",
		verbo, nomeGuia(guia), formatarCentavos(centavosMontante(guia.ValorTotal)), strings.Join(nonNil(guia.Status), ", "))}
	fmt.Fprintln(s.err, avisoAuditoria)
	var resp *api.RespostaConfirmacao
	enviado, err := escrever(cmd, op, func(snd api.Sender) error {
		var err error
		resp, err = api.ConfirmarPagamento(ctx, snd, versao, guia.ID, req)
		return err
	})
	if err != nil || !enviado {
		return err
	}
	situacao := "pago"
	if !pago {
		situacao = "nao_pago"
	}
	var status any
	if nova, err := api.BuscarGuia(ctx, g, guia.ID); err != nil {
		fmt.Fprintln(s.err, "aviso: pedido enviado, mas não foi possível reler a guia:", err)
		situacao = "enviado"
	} else {
		status = nonNil(nova.Status)
	}
	var movida any
	if resp != nil {
		movida = nilIfEmpty(resp.MovidaPara)
	}
	rec := resultadoEscrita(acao, situacao, guia.ID).
		Add("guia", "Guia", nomeGuia(guia)).
		Add("status", "Status", status).
		Add("movida_para", "Movida para", movida)
	return output.Write(s.out, f, rec)
}

func centavosMontante(m *api.Montante) int64 {
	if m == nil || m.Valor == nil {
		return 0
	}
	return centavos(*m.Valor)
}

func newImpostosConfirmarCmd() *cobra.Command {
	var naoPaguei bool
	cmd := &cobra.Command{
		Use:   "confirmar ID",
		Short: "Informa que uma guia foi paga (ou, com --nao-paguei, que não foi)",
		Long: `Marca uma guia como paga ("Já paguei" no painel) para que ela não apareça mais como a
pagar. Com --nao-paguei, informa que a guia não foi paga: ela volta para contas a pagar como
disponível, e o recálculo (ctbz impostos recalcular) passa a ser possível.

Risco médio: é uma declaração, não um pagamento. Uma confirmação falsa esconde um imposto
não pago. ` + avisoAuditoria + ` Desfaça com ctbz impostos desmarcar.

A rota segue a tela de impostos da empresa (impostos/rollout: v5 ou v3). Pendências
críticas na Central de Rotinas bloqueiam a confirmação, como no painel.
Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz impostos confirmar 1000000000000001
  ctbz impostos confirmar 1000000000000001 --nao-paguei --yes`,
		Args: exactArgs(1, "o ID da guia"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if naoPaguei {
				return confirmarPagamento(cmd, args[0], false, "nao-paguei", "Informar que a guia NÃO foi paga")
			}
			return confirmarPagamento(cmd, args[0], true, "confirmar", "Marcar a guia como paga")
		},
	}
	cmd.Flags().BoolVar(&naoPaguei, "nao-paguei", false, "informa que a guia não foi paga")
	addWriteFlags(cmd)
	return cmd
}

func newImpostosDesmarcarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "desmarcar ID",
		Short: "Desfaz a confirmação de pagamento de uma guia",
		Long: `Desfaz uma confirmação de pagamento feita por engano ("Desmarcar" no painel): a guia volta
a aparecer como a pagar. É a mesma requisição de ctbz impostos confirmar --nao-paguei.

Risco médio. Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz impostos desmarcar 1000000000000001`,
		Args:    exactArgs(1, "o ID da guia"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return confirmarPagamento(cmd, args[0], false, "desmarcar", "Desmarcar o pagamento da guia")
		},
	}
	addWriteFlags(cmd)
	return cmd
}
