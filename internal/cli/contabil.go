package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newBalanceteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balancete AAAA-MM",
		Short: "Mostra o balancete de verificação de um mês",
		Long: `Mostra o balancete do mês: cada conta da árvore contábil com saldo anterior, débitos,
créditos e saldo do exercício. Na tabela, as contas aparecem recuadas por nível; em CSV e
JSON, a descrição vem sem recuo e o nível fica na coluna "nivel" (bom para planilhas).

O padrão do painel é dezembro do ano anterior; aqui o mês é obrigatório.`,
		Example: `  ctbz balancete 2026-08
  ctbz balancete 2026-08 -o csv > balancete-2026-08.csv`,
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
			contas, err := api.BuscarRelatorio(cmd.Context(), sessionGetter{s}, api.PathBalancete(mes.Year(), int(mes.Month())))
			if err != nil {
				return err
			}
			return output.Write(s.out, f, balanceteList(contas))
		},
	}
}

func balanceteList(contas []api.ContaRelatorio) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "conta", Header: "Conta"},
		{Key: "descricao", Header: "Descrição"},
		{Key: "nivel", Header: "Nível"},
		{Key: "saldo_anterior", Header: "Saldo anterior"},
		{Key: "debitos", Header: "Débitos"},
		{Key: "creditos", Header: "Créditos"},
		{Key: "saldo", Header: "Saldo"},
	}}
	for _, c := range contas {
		l.Append(c.ID, output.Indent{Level: c.Nivel, Text: c.Descricao}, c.Nivel, moneyOrNil(c.SaldoAnterior),
			moneyOrNil(c.TotalDebito), moneyOrNil(c.TotalCredito), moneyOrNil(c.SaldoExercicio))
	}
	return l
}

// parseMes lê um mês AAAA-MM como argumento.
func parseMes(s string) (time.Time, error) {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return t, usageError{fmt.Errorf("mês inválido %q: use AAAA-MM", s)}
	}
	return t, nil
}

func newBalancoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balanco AAAA[-MM]",
		Short: "Mostra o balanço patrimonial (ativo, passivo e patrimônio líquido)",
		Long: `Mostra o balanço patrimonial: as contas de ativo, passivo e patrimônio líquido com o saldo
do exercício e o do exercício anterior, recuadas por nível na tabela. As contas de resultado
ficam de fora, como no painel.

Só com o ano, o balanço é o de dezembro (fechamento do exercício). Para reabrir o
exercício e regularizar a pendência documental do informe: ctbz balanco reabrir.`,
		Example: `  ctbz balanco 2025
  ctbz balanco 2026-09 -o csv`,
		Args: exactArgs(1, "o ano (AAAA) ou o mês (AAAA-MM)"),
		RunE: func(cmd *cobra.Command, args []string) error {
			mes, err := parseMes(args[0])
			if err != nil {
				if mes, err = time.Parse("2006", args[0]); err != nil {
					return usageError{fmt.Errorf("período inválido %q: use AAAA ou AAAA-MM", args[0])}
				}
				mes = time.Date(mes.Year(), time.December, 1, 0, 0, 0, 0, time.UTC)
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			contas, err := api.BuscarRelatorio(cmd.Context(), sessionGetter{s}, api.PathBalanco(mes.Year(), int(mes.Month())))
			if err != nil {
				return err
			}
			return output.Write(s.out, f, balancoList(contas))
		},
	}
	cmd.AddCommand(newBalancoReabrirCmd())
	return cmd
}

// classificacaoResultado são as contas de resultado, que o balanço não mostra.
const classificacaoResultado = "RESULTADO"

func balancoList(contas []api.ContaRelatorio) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "conta", Header: "Conta"},
		{Key: "descricao", Header: "Descrição"},
		{Key: "nivel", Header: "Nível"},
		{Key: "grupo", Header: "Grupo"},
		{Key: "saldo", Header: "Saldo"},
		{Key: "saldo_exercicio_anterior", Header: "Exercício anterior"},
	}}
	for _, c := range contas {
		if c.ClassificacaoConta == classificacaoResultado {
			continue
		}
		l.Append(c.ID, output.Indent{Level: c.Nivel, Text: c.Descricao}, c.Nivel, nilIfEmpty(c.ClassificacaoConta),
			moneyOrNil(c.SaldoExercicio), moneyOrNil(c.SaldoExercicioAnterior))
	}
	return l
}

func newRazaoCmd() *cobra.Command {
	var de, ate, conta string
	cmd := &cobra.Command{
		Use:   "razao",
		Short: "Lista os lançamentos do razão contábil por conta",
		Long: `Lista os lançamentos do razão de cada conta no período: data, conta, histórico,
contrapartida, débito, crédito e o saldo acumulado no exercício depois do lançamento.

O período vai de --de até --ate (meses AAAA-MM; padrão: o mês atual), no máximo 24 meses;
a API devolve um mês por vez. --conta filtra pelo código da conta (ex.: 1.01.01.01.00) ou
por um prefixo dele (ex.: 1.01 para todo o ativo circulante).`,
		Example: `  ctbz razao --de 2026-01 --ate 2026-09
  ctbz razao --conta 1.01.01.01.00 --de 2026-07 -o csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			meses, err := mesesDoPeriodo(de, ate)
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			var contas []api.ContaRelatorio
			for _, m := range meses {
				cs, err := api.BuscarRelatorio(cmd.Context(), sessionGetter{s}, api.PathRazao(m.Year(), int(m.Month())))
				if err != nil {
					return err
				}
				contas = append(contas, cs...)
			}
			return output.Write(s.out, f, razaoList(contas, conta))
		},
	}
	cmd.Flags().StringVar(&de, "de", "", "primeiro mês, AAAA-MM (padrão: o mês atual)")
	cmd.Flags().StringVar(&ate, "ate", "", "último mês, AAAA-MM (padrão: --de ou o mês atual)")
	cmd.Flags().StringVar(&conta, "conta", "", "código da conta ou prefixo (ex.: 1.01)")
	return cmd
}

func razaoList(contas []api.ContaRelatorio, prefixo string) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "data", Header: "Data"},
		{Key: "conta", Header: "Conta"},
		{Key: "conta_descricao", Header: "Descrição da conta"},
		{Key: "historico", Header: "Histórico"},
		{Key: "contrapartida", Header: "Contrapartida"},
		{Key: "debito", Header: "Débito"},
		{Key: "credito", Header: "Crédito"},
		{Key: "saldo", Header: "Saldo"},
	}}
	for _, c := range contas {
		if prefixo != "" && !strings.HasPrefix(c.ID, prefixo) {
			continue
		}
		for _, lc := range c.ListaLancamento {
			var contrapartida any
			if cp := lc.ContaContrapartida; cp != nil {
				contrapartida = nilIfEmpty(strings.TrimSpace(cp.ID + " " + cp.Descricao))
			}
			l.Append(dateFromMillis(lc.Data), c.ID, c.Descricao, output.Text(lc.Descricao), contrapartida,
				moneyOrNil(lc.Debito), moneyOrNil(lc.Credito), moneyOrNil(lc.SaldoExercicio))
		}
	}
	return l
}
