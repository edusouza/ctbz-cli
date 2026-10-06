package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newNotasTomadoresCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tomadores",
		Short: "Lista os tomadores (clientes) cadastrados no emissor de notas",
		Long: `Lista os tomadores cadastrados no emissor de notas: nome, documento, e-mail, telefone,
inscrição municipal, município, UF e se é do exterior.

Para os dados cadastrais de um CNPJ qualquer na Receita (como o emissor faz ao cadastrar um
tomador), use "ctbz notas tomadores consulta CNPJ".`,
		Example: `  ctbz notas tomadores
  ctbz notas tomadores consulta 00.000.000/0001-91`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			t, err := api.BuscarTomadores(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			return output.Write(s.out, f, tomadoresList(t.Tomadores))
		},
	}
	cmd.AddCommand(newNotasTomadoresConsultaCmd(), newTomadoresAdicionarCmd(), newTomadoresEditarCmd(), newTomadoresRemoverCmd())
	return cmd
}

func tomadoresList(ts []api.Tomador) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "nome", Header: "Nome"},
		{Key: "documento", Header: "Documento"},
		{Key: "email", Header: "E-mail"},
		{Key: "telefone", Header: "Telefone"},
		{Key: "inscricao_municipal", Header: "Inscrição municipal"},
		{Key: "municipio", Header: "Município"},
		{Key: "uf", Header: "UF"},
		{Key: "exterior", Header: "Exterior"},
		{Key: "id", Header: "ID"},
	}}
	for _, t := range ts {
		var municipio, uf any
		if t.Endereco != nil {
			municipio, uf = nomeDeValor(t.Endereco.Municipio), nomeDeValor(t.Endereco.UF)
		}
		l.Append(t.Nome, documentoTomador(t.CPFCNPJ), nilIfEmpty(t.Email), nilIfEmpty(t.Telefone),
			nilIfEmpty(t.InscricaoMunicipal), municipio, uf, t.Estrangeiro, rawTexto(t.ID))
	}
	return l
}

// nomeDeValor lê um campo que pode vir como texto ou como objeto com "nome" ou "id"
// (ex.: município e UF nos endereços do emissor).
func nomeDeValor(v any) any {
	switch x := v.(type) {
	case string:
		return nilIfEmpty(x)
	case map[string]any:
		for _, k := range []string{"nome", "id"} {
			if s, ok := x[k].(string); ok && s != "" {
				return s
			}
		}
	}
	return nil
}

func newNotasTomadoresConsultaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "consulta CNPJ",
		Short: "Consulta os dados cadastrais de um CNPJ na Receita",
		Long: `Consulta um CNPJ como o emissor de notas faz ao cadastrar um tomador: razão social, nome
fantasia, abertura, atividade principal, natureza jurídica, situação cadastral, opção pelo
Simples, endereço e contatos.`,
		Example: `  ctbz notas tomadores consulta 00.000.000/0001-91 -o json`,
		Args:    exactArgs(1, "o CNPJ"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cnpj := output.NewCNPJ(args[0])
			if len(cnpj) != 14 {
				return usageError{fmt.Errorf("CNPJ inválido: %q", args[0])}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			c, err := api.BuscarConsultaCNPJ(cmd.Context(), sessionGetter{s}, string(cnpj))
			if err != nil {
				return err
			}
			return output.Write(s.out, f, consultaCNPJRecord(c))
		},
	}
}

func consultaCNPJRecord(c *api.ConsultaCNPJ) *output.Record {
	var atividade, natureza any
	if a := c.AtividadePrincipal; a != nil {
		atividade = nilIfEmpty(strings.TrimSpace(formatCNAE(a.Codigo) + " " + a.Descricao))
	}
	if n := c.NaturezaJuridica; n != nil {
		natureza = nilIfEmpty(n.Descricao)
	}
	abertura, _ := output.ParseDate(c.DataAbertura)
	endereco := strings.Join(naoVazios(c.Logradouro, c.Numero, c.Complemento, c.Bairro), ", ")
	rec := &output.Record{}
	rec.Add("cnpj", "CNPJ", output.NewCNPJ(c.CNPJ))
	rec.Add("razao_social", "Razão social", nilIfEmpty(c.RazaoSocial))
	rec.Add("nome_fantasia", "Nome fantasia", nilIfEmpty(c.NomeFantasia))
	rec.Add("abertura", "Abertura", abertura)
	rec.Add("atividade_principal", "Atividade principal", atividade)
	rec.Add("natureza_juridica", "Natureza jurídica", natureza)
	rec.Add("situacao_cadastral", "Situação cadastral", nilIfEmpty(c.SituacaoCadastral))
	rec.Add("optante_simples", "Optante do Simples", nilIfEmpty(c.OptanteSimples))
	rec.Add("endereco", "Endereço", nilIfEmpty(endereco))
	rec.Add("cep", "CEP", nilIfEmpty(c.CEP))
	rec.Add("municipio", "Município", nilIfEmpty(c.Municipio))
	rec.Add("uf", "UF", nilIfEmpty(c.UF))
	rec.Add("email", "E-mail", nilIfEmpty(c.Email))
	rec.Add("telefone", "Telefone", nilIfEmpty(c.Telefone))
	return rec
}

// naoVazios devolve os textos não vazios, sem espaços nas pontas.
func naoVazios(ss ...string) []string {
	var out []string
	for _, s := range ss {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
