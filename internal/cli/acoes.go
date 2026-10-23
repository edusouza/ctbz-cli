package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newAcoesCmd() *cobra.Command {
	var desde string
	var limite int
	cmd := &cobra.Command{
		Use:   "acoes",
		Short: "Lista as ações de escrita enviadas à Contabilizei por esta CLI",
		Long: `Lista as escritas que esta CLI enviou à Contabilizei, da mais antiga para a mais recente:
data e hora, CNPJ da empresa, comando, método, caminho, status HTTP, resultado, id do
objeto (quando conhecido) e, nos envios de arquivo, o nome e o tamanho de cada arquivo.

O registro fica em $CTBZ_HOME/acoes.jsonl (permissão 0600), uma linha JSON por escrita
enviada, inclusive as que falharam. Simulações (--dry-run) não entram. Corpos de
requisição e de resposta nunca são gravados, nem o conteúdo dos arquivos.

Resultado: "enviada" (HTTP 2xx), "recusada" (resposta de erro) ou "sem resposta" (falha de
conexão: a escrita pode ou não ter sido aplicada).`,
		Example: `  ctbz acoes
  ctbz acoes --desde 2026-10-01
  ctbz acoes --limite 0 -o csv > acoes.csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			var inicio time.Time
			if desde != "" {
				if inicio, err = time.ParseInLocation("2006-01-02", desde, time.Local); err != nil {
					return usageError{fmt.Errorf("data inválida %q em --desde: use AAAA-MM-DD", desde)}
				}
			}
			if limite < 0 {
				return usageError{fmt.Errorf("--limite não pode ser negativo")}
			}
			s := streamsOf(cmd)
			store, err := ctbz.DefaultStore()
			if err != nil {
				return err
			}
			acoes, invalidas, err := store.LoadAcoes()
			if err != nil {
				return err
			}
			if invalidas > 0 {
				fmt.Fprintf(s.err, "aviso: %d linha(s) ilegível(is) em acoes.jsonl foram ignoradas\n", invalidas)
			}
			return output.Write(s.out, f, acoesList(filtrarAcoes(acoes, inicio, limite)))
		},
	}
	cmd.Flags().StringVar(&desde, "desde", "", "só as ações a partir desta data (AAAA-MM-DD)")
	cmd.Flags().IntVar(&limite, "limite", 50, "quantas ações mais recentes mostrar (0 mostra todas)")
	return cmd
}

// filtrarAcoes aplica --desde e mantém as limite mais recentes (0 = todas).
func filtrarAcoes(acoes []ctbz.Acao, desde time.Time, limite int) []ctbz.Acao {
	var out []ctbz.Acao
	for _, a := range acoes {
		if !a.Data.Before(desde) {
			out = append(out, a)
		}
	}
	if limite > 0 && len(out) > limite {
		out = out[len(out)-limite:]
	}
	return out
}

func acoesList(acoes []ctbz.Acao) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "data", Header: "Data"},
		{Key: "cnpj", Header: "CNPJ"},
		{Key: "comando", Header: "Comando"},
		{Key: "metodo", Header: "Método"},
		{Key: "caminho", Header: "Caminho"},
		{Key: "status", Header: "Status"},
		{Key: "resultado", Header: "Resultado"},
		{Key: "id", Header: "ID"},
		{Key: "arquivos", Header: "Arquivos"},
	}}
	for _, a := range acoes {
		var status any
		if a.Status != 0 {
			status = a.Status
		}
		l.Append(output.DateTime{Time: a.Data}, output.NewCNPJ(a.CNPJ), a.Comando, a.Metodo, a.Caminho,
			status, a.Resultado, nilIfEmpty(a.ID), nilIfEmpty(a.Arquivos))
	}
	return l
}
