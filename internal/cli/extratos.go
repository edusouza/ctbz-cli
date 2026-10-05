package cli

import (
	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newExtratosCmd() *cobra.Command {
	var ano int
	cmd := &cobra.Command{
		Use:   "extratos",
		Short: "Lista a situação dos extratos bancários por mês e conta",
		Long: `Lista, para cada conta bancária e mês, a situação do extrato na Contabilizei (ex.: ABERTO
enquanto o mês não foi fechado) e a integração com o banco. Os extratos são a base da
conciliação: meses sem extrato aparecem nas rotinas como "Importar extrato bancário".

--ano filtra pelo ano; a API devolve todos os meses disponíveis.`,
		Example: `  ctbz extratos
  ctbz extratos --ano 2026 -o csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			es, err := api.BuscarExtratos(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			return output.Write(s.out, f, extratosList(es, ano))
		},
	}
	cmd.Flags().IntVar(&ano, "ano", 0, "só os meses deste ano")
	cmd.AddCommand(newExtratosContasCmd(), newExtratosLancamentosCmd(), newExtratosClassificarCmd(),
		newExtratosDesmembrarCmd(), newExtratosDesfazerCmd())
	return cmd
}

func extratosList(es []api.Extrato, ano int) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "competencia", Header: "Competência"},
		{Key: "banco", Header: "Banco"},
		{Key: "agencia", Header: "Agência"},
		{Key: "conta", Header: "Conta"},
		{Key: "situacao", Header: "Situação"},
		{Key: "integracao", Header: "Integração"},
		{Key: "id_conta", Header: "ID da conta"},
	}}
	for _, e := range es {
		if ano != 0 && e.Ano != ano {
			continue
		}
		l.Append(competencia(e.Mes, e.Ano), e.Banco, nilIfEmpty(e.Agencia), nilIfEmpty(e.NumeroConta),
			nilIfEmpty(e.Situacao), textoDeValor(e.StatusIntegracao), e.IDContaBancaria)
	}
	return l
}

func newContasBancariasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contas-bancarias",
		Short: "Lista as contas bancárias cadastradas da empresa",
		Long: `Lista as contas bancárias cadastradas na Contabilizei: banco, código do banco, agência,
conta, saldo inicial e sua data, e a situação da integração (ex.: INTEGRADA com a conta PJ
da Contabilizei).`,
		Example: `  ctbz contas-bancarias`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			c, err := api.BuscarContasBancarias(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{
				{Key: "banco", Header: "Banco"},
				{Key: "codigo_banco", Header: "Código"},
				{Key: "agencia", Header: "Agência"},
				{Key: "conta", Header: "Conta"},
				{Key: "saldo_inicial", Header: "Saldo inicial"},
				{Key: "data_saldo_inicial", Header: "Data do saldo"},
				{Key: "integracao", Header: "Integração"},
				{Key: "fluxo_integracao", Header: "Fluxo"},
				{Key: "id", Header: "ID"},
			}}
			for _, cb := range c.ContasBancarias {
				l.Append(cb.NomeBanco, nilIfEmpty(cb.CodigoBanco), nilIfEmpty(cb.Agencia), nilIfEmpty(cb.ContaCorrente),
					moneyOrNil(cb.VlrSaldoInicial), dateFromMillis(cb.DataSaldoInicial), nilIfEmpty(cb.StatusIntegracao),
					nilIfEmpty(cb.FluxoIntegracao), cb.ID)
			}
			return output.Write(s.out, f, l)
		},
	}
	cmd.AddCommand(newContasBancariasBancosCmd(), newContasBancariasAdicionarCmd(), newContasBancariasEditarCmd())
	return cmd
}
