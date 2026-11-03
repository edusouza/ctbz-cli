package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// maxMesesNotas limita o período de uma consulta (uma chamada por mês e página).
const maxMesesNotas = 24

func newNotasCmd() *cobra.Command {
	var de, ate, tomador, numero string
	cmd := &cobra.Command{
		Use:   "notas",
		Short: "Lista as notas fiscais de serviço (NFS-e) emitidas",
		Long: `Lista as NFS-e emitidas pelo emissor da Contabilizei: número, data de emissão, tomador,
documento, valor, status e situação. Na tabela, o total do período vai para o stderr.

O período vai de --de até --ate (meses AAAA-MM; padrão: o mês atual), no máximo 24 meses.
--tomador filtra pelo nome do tomador ou, se for um CPF/CNPJ, pelo documento; --numero
filtra pelo número da nota.

O PDF e o XML da nota não estão disponíveis pela API: a Contabilizei os envia por e-mail
quando a prefeitura autoriza a nota.`,
		Example: `  ctbz notas
  ctbz notas --de 2026-01 --ate 2026-09 -o csv > notas.csv
  ctbz notas --tomador 11.222.333/0001-81
  ctbz notas --tomador "ACME" --de 2026-07`,
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
			if tomador != "" && numero != "" {
				return usageError{fmt.Errorf("use --tomador ou --numero, não os dois")}
			}
			filtro := filtroDeNotas(tomador, numero)
			s := streamsOf(cmd)
			var notas []api.NotaEmitida
			for _, m := range meses {
				filtro.Ano, filtro.Mes = m.Year(), int(m.Month())
				ns, err := api.BuscarNotasEmitidas(cmd.Context(), sessionGetter{s}, filtro)
				if err != nil {
					return err
				}
				notas = append(notas, ns...)
			}
			l, total := notasList(notas)
			if err := output.Write(s.out, f, l); err != nil {
				return err
			}
			if f == output.FormatTable {
				fmt.Fprintf(s.err, "Total: %s em %d nota(s)\n", output.FormatBRL(total), len(notas))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&de, "de", "", "primeiro mês, AAAA-MM (padrão: o mês atual)")
	cmd.Flags().StringVar(&ate, "ate", "", "último mês, AAAA-MM (padrão: --de ou o mês atual)")
	cmd.Flags().StringVar(&tomador, "tomador", "", "nome do tomador ou CPF/CNPJ")
	cmd.Flags().StringVar(&numero, "numero", "", "número da nota")
	cmd.AddCommand(newNotasTomadoresCmd(), newNotasConfigCmd(), newNotasAliquotasCmd(), newNotasEntradaCmd())
	return cmd
}

// mesesDoPeriodo devolve o primeiro dia de cada mês entre de e ate (AAAA-MM).
func mesesDoPeriodo(de, ate string) ([]time.Time, error) {
	atual := time.Date(now().Year(), now().Month(), 1, 0, 0, 0, 0, time.UTC)
	parse := func(flag, s string, padrao time.Time) (time.Time, error) {
		if s == "" {
			return padrao, nil
		}
		t, err := time.Parse("2006-01", s)
		if err != nil {
			return t, usageError{fmt.Errorf("--%s deve ser AAAA-MM: %q", flag, s)}
		}
		return t, nil
	}
	ini, err := parse("de", de, atual)
	if err != nil {
		return nil, err
	}
	fim, err := parse("ate", ate, ini)
	if err != nil {
		return nil, err
	}
	if de == "" && ate != "" {
		ini = fim
	}
	if fim.Before(ini) {
		return nil, usageError{fmt.Errorf("--ate (%s) é anterior a --de (%s)", fim.Format("2006-01"), ini.Format("2006-01"))}
	}
	var meses []time.Time
	for m := ini; !m.After(fim); m = m.AddDate(0, 1, 0) {
		if len(meses) == maxMesesNotas {
			return nil, usageError{fmt.Errorf("período maior que %d meses", maxMesesNotas)}
		}
		meses = append(meses, m)
	}
	return meses, nil
}

// filtroDeNotas escolhe o campo de filtro: CPF/CNPJ vai por documento, texto por nome.
func filtroDeNotas(tomador, numero string) api.FiltroNotas {
	switch doc := string(output.NewCNPJ(tomador)); {
	case numero != "":
		return api.FiltroNotas{Campo: api.FiltroNumeroNota, Termo: numero}
	case tomador == "":
		return api.FiltroNotas{}
	case (len(doc) == 11 || len(doc) == 14) && len(doc) == len(somenteDocumento(tomador)):
		return api.FiltroNotas{Campo: api.FiltroDocumento, Termo: doc}
	default:
		return api.FiltroNotas{Campo: api.FiltroNomeTomador, Termo: tomador}
	}
}

// somenteDocumento tira a pontuação comum de CPF/CNPJ, para saber se o texto era só isso.
func somenteDocumento(s string) string {
	var b []rune
	for _, r := range s {
		switch r {
		case '.', '-', '/', ' ':
			continue
		}
		b = append(b, r)
	}
	return string(b)
}

func notasList(ns []api.NotaEmitida) (*output.List, float64) {
	l := &output.List{Columns: []output.Column{
		{Key: "numero", Header: "Número"},
		{Key: "emissao", Header: "Emissão"},
		{Key: "tomador", Header: "Tomador"},
		{Key: "documento", Header: "Documento"},
		{Key: "valor", Header: "Valor"},
		{Key: "status", Header: "Status"},
		{Key: "situacao", Header: "Situação"},
	}}
	var total float64
	for _, n := range ns {
		if n.ValorServico != nil {
			total += *n.ValorServico
		}
		emissao := dataDeValor(n.DataEmissao)
		if emissao == nil {
			emissao = dataOuTexto(n.DataEmissaoFormatada)
		}
		l.Append(idTexto(n.Numero), emissao, n.NomeRazaoTomador, documentoTomador(n.CPFCNPJTomador),
			moneyOrNil(n.ValorServico), nilIfEmpty(n.StatusNotaFiscal.Rotulo()), nilIfEmpty(n.SituacaoNota))
	}
	return l, total
}

// documentoTomador tipa o documento como CPF ou CNPJ pelo número de dígitos.
func documentoTomador(s string) any {
	switch d := output.NewCNPJ(s); len(d) {
	case 14:
		return d
	case 11:
		return output.CPF(d)
	case 0:
		return nilIfEmpty(s)
	}
	return s
}
