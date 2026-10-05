package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newImpostosParcelamentosCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "parcelamentos",
		Short: "Lista os parcelamentos de impostos (em andamento, ativos e encerrados)",
		Long: `Lista os parcelamentos de impostos da empresa, como a aba "Parcelamentos" do painel:
em negociação (em_andamento), ativos e encerrados (historico). Para o detalhe de um
parcelamento, use "ctbz impostos parcelamento ID"; para simular um novo, "ctbz impostos
parcelamento simular TIPO".

Para contratar: "ctbz impostos parcelamento contratar TIPO".`,
		Example: `  ctbz impostos parcelamentos`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			p, err := api.BuscarParcelamentos(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			return output.Write(s.out, f, parcelamentosList(p))
		},
	}
}

func parcelamentosList(p *api.ParcelamentosV3) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "grupo", Header: "Grupo"},
		{Key: "id", Header: "ID"},
		{Key: "titulo", Header: "Título"},
		{Key: "tipo", Header: "Tipo"},
		{Key: "situacao", Header: "Situação"},
		{Key: "parcela_atual", Header: "Parcela atual"},
	}}
	aba := p.AbaParcelamentos
	for _, grupo := range []struct {
		chave string
		itens []api.Parcelamento
	}{{"em_andamento", aba.EmAndamento}, {"ativo", aba.Ativos}, {"historico", aba.Historico}} {
		for _, it := range grupo.itens {
			var parcela any
			if it.ParcelaAtual != nil {
				parcela = it.ParcelaAtual.NumeroParcela
			}
			l.Append(grupo.chave, it.IDParcelamento, it.Titulo, strings.ToLower(it.TipoParcelamento), it.Status, parcela)
		}
	}
	return l
}

func newImpostosParcelamentoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parcelamento ID",
		Short: "Mostra o detalhe de um parcelamento de impostos",
		Long: `Mostra o detalhe de um parcelamento (descrição, período da dívida, saldo devedor,
parcelas). O ID vem de "ctbz impostos parcelamentos".

O formato desta resposta não pôde ser verificado (a conta usada no desenvolvimento não tinha
parcelamentos), por isso os campos são mostrados como a API os devolve.`,
		Example: `  ctbz impostos parcelamento 1000000000000001`,
		Args:    exactArgs(1, "o ID do parcelamento"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			resp, err := authedAPI(cmd.Context(), s, "GET", api.PathParcelamento(id), nil)
			if err != nil {
				return err
			}
			if resp.Status != 200 {
				return &ctbz.HTTPError{Step: "detalhe do parcelamento", Status: resp.Status, Body: resp.Body}
			}
			data, err := output.FromJSON(resp.Body)
			if err != nil {
				return err
			}
			return output.Write(s.out, f, data)
		},
	}
	cmd.AddCommand(newParcelamentoSimularCmd(), newParcelamentoContratarCmd())
	return cmd
}

func newImpostosDebitosCmd() *cobra.Command {
	var failOnDebitos bool
	cmd := &cobra.Command{
		Use:   "debitos",
		Short: "Indica se a empresa tem débitos federais em aberto",
		Long: `Indica se a Contabilizei identificou débitos federais em aberto para a empresa.
Com --fail-on-debitos, termina com código 4 quando há débitos.`,
		Example: `  ctbz impostos debitos
  ctbz impostos debitos --fail-on-debitos -o json`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			d, err := api.BuscarDebitosFederais(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			rec := &output.Record{}
			rec.Add("possui_debitos_federais", "Possui débitos federais", d.PossuiDebitosFederais)
			if err := output.Write(s.out, f, rec); err != nil {
				return err
			}
			if failOnDebitos && d.PossuiDebitosFederais {
				return errAttention
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&failOnDebitos, "fail-on-debitos", false, "termina com código 4 se houver débitos federais")
	return cmd
}
