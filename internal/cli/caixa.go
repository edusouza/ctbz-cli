package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newCaixaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "caixa AAAA-MM",
		Short: "Lista os lançamentos do caixa de um mês",
		Long: `Lista os lançamentos do caixa do mês (entradas e saídas classificadas): data, descrição,
conta de classificação, tipo (receita, despesa…), valor (negativo nas saídas), situação e se
foi lançado pelo sistema. Na tabela, o total do mês vai para o stderr.

O painel pede até 1000 lançamentos por mês; se houver mais, a CLI avisa no stderr.`,
		Example: `  ctbz caixa 2026-09
  ctbz caixa 2026-09 -o csv > caixa-2026-09.csv`,
		Args: exactArgs(1, "o mês (AAAA-MM)"),
		RunE: func(cmd *cobra.Command, args []string) error {
			mes, err := parseMes(args[0])
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			c, err := api.BuscarCaixa(cmd.Context(), g, mes.Year(), int(mes.Month()))
			if err != nil {
				return err
			}
			contas, err := api.BuscarContasUsuario(cmd.Context(), g)
			if err != nil {
				return err
			}
			if c.Total > len(c.List) {
				fmt.Fprintf(s.err, "aviso: o mês tem %d lançamentos; a Contabilizei devolveu %d\n", c.Total, len(c.List))
			}
			l, total := caixaList(c.List, contas)
			if err := output.Write(s.out, f, l); err != nil {
				return err
			}
			if f == output.FormatTable {
				fmt.Fprintf(s.err, "Total do mês: %s em %d lançamento(s)\n", output.FormatBRL(total), len(c.List))
			}
			return nil
		},
	}
	cmd.AddCommand(newCaixaContasCmd(), newCaixaAdicionarCmd())
	return cmd
}

func caixaList(ls []api.LancamentoCaixa, contas []api.ContaUsuario) (*output.List, float64) {
	porID := make(map[int64]api.ContaUsuario, len(contas))
	for _, c := range contas {
		porID[c.ID] = c
	}
	l := &output.List{Columns: []output.Column{
		{Key: "data", Header: "Data"},
		{Key: "descricao", Header: "Descrição"},
		{Key: "conta", Header: "Conta"},
		{Key: "classificacao", Header: "Classificação"},
		{Key: "valor", Header: "Valor"},
		{Key: "situacao", Header: "Situação"},
		{Key: "automatico", Header: "Automático"},
	}}
	var total float64
	for _, lc := range ls {
		var conta, classificacao any
		if lc.IDContaUsuario != nil {
			if c, ok := porID[*lc.IDContaUsuario]; ok {
				conta, classificacao = nilIfEmpty(c.Descricao), nilIfEmpty(c.Classificacao)
			}
		}
		if lc.Valor != nil {
			total += *lc.Valor
		}
		l.Append(dateFromMillis(lc.Data), output.Text(lc.Descricao), conta, classificacao, moneyOrNil(lc.Valor),
			nilIfEmpty(lc.Situacao), lc.ConfirmadoViaSistema)
	}
	return l, total
}
