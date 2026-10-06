package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// errSemEscritorioVirtual explica o 560 das leituras do escritório virtual.
var errSemEscritorioVirtual = errors.New("a empresa não contratou o escritório virtual")

// semEscritorioVirtual traduz o 560 de "serviço não contratado".
func semEscritorioVirtual(err error) error {
	var he *ctbz.HTTPError
	if errors.As(err, &he) && he.Status == api.StatusServicoNaoContratado {
		return errSemEscritorioVirtual
	}
	return err
}

func newCorrespondenciasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "correspondencias",
		Short: "Lista as correspondências recebidas no escritório virtual",
		Long: `Lista as correspondências recebidas no escritório virtual: ID, data, remetente, descrição,
situação e valor do envio. Sem o serviço contratado, avisa e termina com código 0.

Para baixar: ctbz correspondencias baixar ID; endereço de envio e autorização:
ctbz correspondencias endereco e autorizar.`,
		Example: `  ctbz correspondencias
  ctbz correspondencias -o json`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			cs, err := api.BuscarCorrespondencias(cmd.Context(), sessionGetter{s})
			if err = semEscritorioVirtual(err); errors.Is(err, errSemEscritorioVirtual) {
				fmt.Fprintln(s.err, "A empresa não contratou o escritório virtual: não há correspondências.")
				cs, err = nil, nil
			}
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{
				{Key: "id", Header: "ID"}, {Key: "data", Header: "Data"}, {Key: "remetente", Header: "Remetente"},
				{Key: "descricao", Header: "Descrição"}, {Key: "situacao", Header: "Situação"}, {Key: "valor_envio", Header: "Valor do envio"},
			}}
			for _, c := range cs {
				l.Append(nilIfEmpty(idTexto(c.ID)), dataDeValor(c.Data), nilIfEmpty(semControle(c.Remetente)),
					nilIfEmpty(semControle(c.Descricao)), nilIfEmpty(c.Status), moneyOrNil(c.Taxa))
			}
			return output.Write(s.out, f, l)
		},
	}
	cmd.AddCommand(newCorrespondenciasBaixarCmd(), newCorrespondenciasEnderecoCmd(), newCorrespondenciasAutorizarCmd())
	return cmd
}

func newCorrespondenciasBaixarCmd() *cobra.Command {
	var dir string
	var force bool
	cmd := &cobra.Command{
		Use:     "baixar ID",
		Short:   "Baixa uma correspondência do escritório virtual",
		Long:    `Baixa uma correspondência (o ID está em ctbz correspondencias) como correspondencia-ID.pdf.`,
		Example: `  ctbz correspondencias baixar 501 --dir ~/Documentos`,
		Args:    exactArgs(1, "o ID da correspondência"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := streamsOf(cmd)
			path := filepath.Join(dir, "correspondencia-"+slug(args[0])+".pdf")
			if _, err := os.Stat(path); err == nil && !force {
				return fmt.Errorf("%s já existe (use --force para sobrescrever)", path)
			}
			link, err := api.LinkCorrespondencia(cmd.Context(), sessionGetter{s}, args[0])
			if err != nil {
				return semEscritorioVirtual(err)
			}
			if err := download(cmd.Context(), link.URL, path); err != nil {
				return err
			}
			fmt.Fprintln(s.out, path)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".", "diretório onde salvar")
	cmd.Flags().BoolVar(&force, "force", false, "sobrescreve o arquivo se existir")
	return cmd
}

func newCorrespondenciasEnderecoCmd() *cobra.Command {
	var cep, numero, complemento string
	cmd := &cobra.Command{
		Use:   "endereco",
		Short: "Mostra ou troca o endereço de envio das correspondências",
		Long: `Sem flags, mostra o endereço para onde as correspondências são enviadas. Com --cep e
--numero (e --complemento, opcional), troca o endereço: logradouro, bairro e cidade vêm do CEP.
A mudança vale só para o envio das correspondências.

Risco médio: muda para onde vão documentos oficiais; reversível regravando o endereço
anterior. Aceita --yes e --dry-run.`,
		Example: `  ctbz correspondencias endereco
  ctbz correspondencias endereco --cep 01001-000 --numero 100 --complemento "sala 1"`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			alterar := cep != "" || numero != "" || cmd.Flags().Changed("complemento")
			if alterar {
				if cep = ctbz.OnlyDigits(cep); len(cep) != 8 || strings.TrimSpace(numero) == "" {
					return usageError{errors.New("para trocar o endereço, informe --cep (8 dígitos) e --numero")}
				}
			}
			atual, err := api.BuscarEnderecoEntrega(cmd.Context(), g)
			if err != nil {
				return semEscritorioVirtual(err)
			}
			if !alterar {
				return output.Write(s.out, f, enderecoEntregaRecord(atual))
			}
			novo, err := api.PesquisarEnderecoCorrespondencia(cmd.Context(), g, cep)
			if err != nil {
				return err
			}
			novo.CEP, novo.Numero, novo.Complemento = cep, strings.TrimSpace(numero), strings.TrimSpace(complemento)
			op := operacao{Risco: riscoMedio, Resumo: fmt.Sprintf("Trocar o endereço de envio das correspondências de %s para %s", enderecoTexto(atual), enderecoTexto(novo))}
			var salvo *api.EnderecoEntrega
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				var err error
				salvo, err = api.SalvarEnderecoCorrespondencia(cmd.Context(), snd, novo.CEP, novo.Numero, novo.Complemento)
				return err
			})
			if err != nil || !enviado {
				return err
			}
			relido, err := api.BuscarEnderecoEntrega(cmd.Context(), g)
			if err != nil {
				relido = salvo
			}
			situacao := "enviado"
			if ctbz.OnlyDigits(relido.CEP) == cep && relido.Numero == novo.Numero {
				situacao = "alterado"
			}
			rec := resultadoEscrita("endereco", situacao, nil)
			for _, c := range enderecoEntregaRecord(relido).Fields {
				rec.Add(c.Key, c.Label, c.Value)
			}
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().StringVar(&cep, "cep", "", "CEP do novo endereço")
	cmd.Flags().StringVar(&numero, "numero", "", "número do novo endereço")
	cmd.Flags().StringVar(&complemento, "complemento", "", "complemento do novo endereço")
	addWriteFlags(cmd)
	return cmd
}

func enderecoEntregaRecord(e *api.EnderecoEntrega) *output.Record {
	rec := &output.Record{}
	return rec.Add("cep", "CEP", nilIfEmpty(e.CEP)).Add("logradouro", "Logradouro", nilIfEmpty(semControle(e.Logradouro))).
		Add("numero", "Número", nilIfEmpty(semControle(e.Numero))).Add("complemento", "Complemento", nilIfEmpty(semControle(e.Complemento))).
		Add("bairro", "Bairro", nilIfEmpty(semControle(e.Bairro))).Add("cidade", "Cidade", nilIfEmpty(semControle(e.Cidade))).
		Add("uf", "UF", nilIfEmpty(e.UF))
}

func enderecoTexto(e *api.EnderecoEntrega) string {
	if e.CEP == "" && e.Logradouro == "" {
		return "(sem endereço)"
	}
	partes := []string{strings.TrimSpace(e.Logradouro + ", " + e.Numero)}
	if e.Complemento != "" {
		partes = append(partes, e.Complemento)
	}
	partes = append(partes, strings.Trim(e.Cidade+"/"+e.UF, "/"), "CEP "+e.CEP)
	return semControle(strings.Join(partes, " - "))
}

func newCorrespondenciasAutorizarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autorizar on|off",
		Short: "Autoriza ou não o recebimento de correspondências no endereço informado",
		Long: `Liga (on) ou desliga (off) "Autorizo o recebimento de correspondências e documentos
entregues no endereço informado". O envio de correspondências pode ser cobrado (a lista mostra
o valor do envio).

Risco médio e reversível. Aceita --yes e --dry-run.`,
		Example:   `  ctbz correspondencias autorizar on`,
		Args:      exactArgs(1, "on ou off"),
		ValidArgs: []string{"on", "off"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "on" && args[0] != "off" {
				return usageError{fmt.Errorf("use on ou off: %q", args[0])}
			}
			autoriza := args[0] == "on"
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			atual, err := api.BuscarAutorizacaoRecebimento(cmd.Context(), g)
			if err != nil {
				return semEscritorioVirtual(err)
			}
			estado := map[bool]string{true: "autorizado", false: "não autorizado"}
			if atual.Autoriza == autoriza {
				fmt.Fprintf(s.err, "O recebimento já está %s; nada foi enviado.\n", estado[autoriza])
				return output.Write(s.out, f, resultadoEscrita("autorizar", "sem mudança", nil).Add("autoriza", "Autoriza", autoriza))
			}
			op := operacao{Risco: riscoMedio, Resumo: "Deixar de autorizar o recebimento de correspondências no endereço informado"}
			if autoriza {
				op.Resumo = "Autorizar o recebimento de correspondências no endereço informado; o envio pode ser cobrado"
			}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.AutorizarRecebimento(cmd.Context(), snd, autoriza) })
			if err != nil || !enviado {
				return err
			}
			relida, err := api.BuscarAutorizacaoRecebimento(cmd.Context(), g)
			situacao := "enviado"
			if err == nil && relida.Autoriza == autoriza {
				situacao = estado[autoriza]
			}
			return output.Write(s.out, f, resultadoEscrita("autorizar", situacao, nil).Add("autoriza", "Autoriza", autoriza))
		},
	}
	addWriteFlags(cmd)
	return cmd
}
