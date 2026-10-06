package cli

import (
	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newDocumentosCmd() *cobra.Command {
	var tipo string
	cmd := &cobra.Command{
		Use:   "documentos",
		Short: "Lista os tipos de documento aceitos e os documentos enviados",
		Long: `Sem --tipo, lista os tipos de documento da central de documentos (extratos, contratos de
empréstimo, informes de investimentos…) com quantos já foram enviados e a última
modificação. Com --tipo, lista os documentos enviados daquele tipo: arquivo, competência,
data de envio, descrição e o link do arquivo.

Enviar documentos continua sendo feito pelo painel.`,
		Example: `  ctbz documentos
  ctbz documentos --tipo EXTRATO_BANCARIO_MOVIMENTACOES -o csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			if tipo == "" {
				ts, err := api.BuscarTiposDocumento(cmd.Context(), g, api.AreaDocumentosContabeis)
				if err != nil {
					return err
				}
				return output.Write(s.out, f, tiposDocumentoList(ts))
			}
			ds, err := api.BuscarDocumentosEnviados(cmd.Context(), g, tipo)
			if err != nil {
				return err
			}
			return output.Write(s.out, f, documentosList(ds))
		},
	}
	cmd.Flags().StringVar(&tipo, "tipo", "", "tipo de documento (veja a coluna tipo de \"ctbz documentos\")")
	cmd.AddCommand(newDocumentosPendentesCmd(), newDocumentosEnviarCmd(), newDocumentosSemArquivoCmd(), newDocumentosSemAplicacaoCmd())
	return cmd
}

func tiposDocumentoList(ts []api.TipoDocumento) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "tipo", Header: "Tipo"},
		{Key: "nome", Header: "Nome"},
		{Key: "quantidade", Header: "Enviados"},
		{Key: "modificado", Header: "Última modificação"},
	}}
	for _, t := range ts {
		l.Append(t.Tipo, t.Nome, t.Quantidade, dataOuTexto(t.DataModificacao))
	}
	return l
}

func documentosList(ds []api.DocumentoEnviado) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "competencia", Header: "Competência"},
		{Key: "arquivo", Header: "Arquivo"},
		{Key: "enviado", Header: "Enviado em"},
		{Key: "descricao", Header: "Descrição"},
		{Key: "valor", Header: "Valor"},
		{Key: "link", Header: "Link"},
	}}
	for _, d := range ds {
		var descricao, valor any
		if m := d.Metadados; m != nil {
			descricao, valor = nilIfEmpty(m.Nome), valorDeValor(m.Valor)
		}
		l.Append(competenciaDeDocumento(d.Competencia), nilIfEmpty(d.NomeArquivo), dataDeValor(d.DataEnvio),
			descricao, valor, nilIfEmpty(d.URLArquivo))
	}
	return l
}

// competenciaDeDocumento aceita a competência como texto ou como {mes, ano} (o formato
// usado no envio).
func competenciaDeDocumento(v any) any {
	if m, ok := v.(map[string]any); ok {
		mes, _ := m["mes"].(float64)
		ano, _ := m["ano"].(float64)
		return competencia(int(mes), int(ano))
	}
	return competenciaDeValor(v)
}
