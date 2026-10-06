package cli

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

var listasNotasEntrada = []string{api.ListaAManifestar, api.ListaManifestadas, api.ListaAClassificar, api.ListaClassificadas}

func newNotasEntradaCmd() *cobra.Command {
	var mes, lista, emitente string
	cmd := &cobra.Command{
		Use:   "entrada",
		Short: "Lista as notas fiscais de entrada (NF-e recebidas pela empresa)",
		Long: `Lista as NF-e de compra recebidas pela empresa, como as abas da tela "Notas fiscais de
entrada": a manifestar (padrão), manifestadas, a classificar e classificadas. Mostra emissão,
emitente, CNPJ do emitente, valor, situação, chave de acesso e o ID.

--mes escolhe o mês (AAAA-MM; padrão: o mês atual) e --emitente filtra pela razão social do
emitente. Para manifestar: ctbz notas entrada manifestar; para classificar: ctbz notas
entrada classificar (itens em ctbz notas entrada produtos).`,
		Example: `  ctbz notas entrada
  ctbz notas entrada --lista manifestadas --mes 2026-09 -o csv
  ctbz notas entrada --lista a-classificar --emitente "ACME"`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			if !slices.Contains(listasNotasEntrada, lista) {
				return usageError{fmt.Errorf("--lista deve ser %s: %q", strings.Join(listasNotasEntrada, ", "), lista)}
			}
			ref, err := mesDaFlag(mes)
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			filtro := filtroEntrada(lista, ref)
			filtro.Emitente = emitente
			ns, err := api.BuscarNotasEntrada(cmd.Context(), sessionGetter{s}, filtro)
			if err != nil {
				return err
			}
			return output.Write(s.out, f, notasEntradaList(ns))
		},
	}
	cmd.Flags().StringVar(&mes, "mes", "", "mês, AAAA-MM (padrão: o mês atual)")
	cmd.Flags().StringVar(&lista, "lista", api.ListaAManifestar, "a-manifestar, manifestadas, a-classificar ou classificadas")
	cmd.Flags().StringVar(&emitente, "emitente", "", "filtra pela razão social do emitente")
	cmd.AddCommand(newNotasEntradaManifestarCmd(), newNotasEntradaClassificarCmd(), newNotasEntradaProdutosCmd())
	return cmd
}

func notasEntradaList(ns []api.NotaEntrada) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "emissao", Header: "Emissão"},
		{Key: "emitente", Header: "Emitente"},
		{Key: "cnpj_emitente", Header: "CNPJ do emitente"},
		{Key: "valor", Header: "Valor"},
		{Key: "situacao", Header: "Situação"},
		{Key: "chave", Header: "Chave de acesso"},
		{Key: "id", Header: "ID"},
		{Key: "classificacao", Header: "Classificação"},
	}}
	for _, n := range ns {
		var situacao any
		if n.Situacao != nil {
			situacao = nilIfEmpty(firstNonEmpty(n.Situacao.Descricao, n.Situacao.ID))
		}
		l.Append(dataDeValor(n.DataEmissao), n.RazaoSocial, documentoTomador(n.CNPJEmitente), moneyOrNil(n.Valor),
			situacao, nilIfEmpty(n.Chave), nilIfEmpty(idTexto(n.ID)), nilIfEmpty(n.TipoClassificacao))
	}
	return l
}

// mesDaFlag lê --mes (AAAA-MM); vazio é o mês atual.
func mesDaFlag(mes string) (time.Time, error) {
	if mes == "" {
		return now(), nil
	}
	t, err := time.Parse("2006-01", mes)
	if err != nil {
		return time.Time{}, usageError{fmt.Errorf("--mes deve ser AAAA-MM: %q", mes)}
	}
	return t, nil
}

func filtroEntrada(lista string, mes time.Time) api.FiltroNotasEntrada {
	return api.FiltroNotasEntrada{Lista: lista, Ano: mes.Year(), Mes: int(mes.Month())}
}
