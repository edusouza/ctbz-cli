package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newEmpresaCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "empresa",
		Short: "Mostra os dados da empresa selecionada",
		Long: `Mostra razão social, CNPJ, situação, regime tributário, plano, certificado digital
e as outras empresas do usuário.

A resposta crua da API está em "ctbz api dadosempresa/get".`,
		Example: `  ctbz empresa
  ctbz empresa -o json | jq -r .certificado_validade`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			fallback := ""
			if asJSON {
				fallback = string(output.FormatJSON)
			}
			f, err := outputFormat(cmd, fallback)
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			data, err := api.BuscarDadosEmpresa(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			cadastro, err := empresaDaSessao()
			if err != nil {
				fmt.Fprintln(s.err, "aviso:", err)
				cadastro = &api.EmpresaSessao{}
			}
			return output.Write(s.out, f, empresaRecord(data, cadastro))
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "atalho para -o json")
	cmd.AddCommand(newEmpresaUsarCmd(), newEmpresaCertificadoCmd(), newEmpresaSociosCmd(), newEmpresaAtividadesCmd(),
		newEmpresaCredenciaisCmd())
	return cmd
}

// empresaDaSessao lê os dados cadastrais gravados no login.
func empresaDaSessao() (*api.EmpresaSessao, error) {
	store, err := ctbz.DefaultStore()
	if err != nil {
		return nil, err
	}
	sess, err := store.LoadSession()
	if err != nil {
		return nil, err
	}
	return api.EmpresaDaSessao(sess.Storage)
}

// empresaRecord monta a saída de "ctbz empresa": dados atuais da API e cadastro do login.
func empresaRecord(d *api.DadosEmpresa, c *api.EmpresaSessao) *output.Record {
	e := d.EmpresaAtual
	var certSituacao any
	var certValidade output.Date
	if e.Certificado != nil {
		certSituacao = e.Certificado.Status.Descricao
		certValidade, _ = output.ParseDate(e.Certificado.DataValidade)
	}
	outras := []output.Record{}
	for _, o := range d.Empresas {
		if ctbz.OnlyDigits(o.CNPJ) == ctbz.OnlyDigits(e.CNPJ) {
			continue
		}
		r := output.Record{}
		r.Add("cnpj", "CNPJ", output.NewCNPJ(o.CNPJ)).
			Add("razao_social", "Razão social", o.RazaoSocial).
			Add("situacao", "Situação", o.StatusEmpresa)
		outras = append(outras, r)
	}
	rec := &output.Record{}
	rec.Add("razao_social", "Razão social", e.RazaoSocial).
		Add("cnpj", "CNPJ", output.NewCNPJ(e.CNPJ)).
		Add("situacao", "Situação", e.StatusEmpresa).
		Add("regime_tributario", "Regime tributário", e.RegimeTributario).
		Add("nome_fantasia", "Nome fantasia", nilIfEmpty(c.NomeFantasia)).
		Add("natureza_juridica", "Natureza jurídica", naturezaJuridica(c)).
		Add("abertura", "Abertura", dateFromMillis(c.DataAbertura)).
		Add("inicio_contabilidade", "Início na Contabilizei", competencia(c.ResponsabilidadeInicial.Mes, c.ResponsabilidadeInicial.Ano)).
		Add("inscricao_municipal", "Inscrição municipal", nilIfEmpty(e.InscricaoMunicipal)).
		Add("inscricao_estadual", "Inscrição estadual", nilIfEmpty(c.InscricaoEstadual)).
		Add("endereco", "Endereço", nilIfEmpty(endereco(c))).
		Add("ramos_atividade", "Ramos de atividade", nonNil(e.RamosAtividade)).
		Add("plano", "Plano", nilIfEmpty(e.Plano)).
		Add("certificado_situacao", "Certificado digital", certSituacao).
		Add("certificado_validade", "Validade do certificado", certValidade).
		Add("outras_empresas", "Outras empresas", outras)
	return rec
}

func naturezaJuridica(c *api.EmpresaSessao) any {
	n := c.NaturezaJuridica
	if n.Descricao == "" {
		return nil
	}
	return n.Descricao + " (" + n.CodReceita + ")"
}

// endereco junta o endereço numa linha: "Rua X, 10, Apto 1 - Bairro, Cidade/UF, CEP 00000-000".
func endereco(c *api.EmpresaSessao) string {
	e := c.Endereco
	var parts []string
	rua := strings.TrimSpace(strings.Join(nonEmpty(e.Logradouro, e.Numero, e.Complemento), ", "))
	if rua != "" && e.Bairro != "" {
		rua += " - " + e.Bairro
	}
	parts = nonEmpty(rua)
	if e.Municipio.Nome != "" {
		parts = append(parts, e.Municipio.Nome+"/"+e.Municipio.UF.ID)
	}
	if cep := ctbz.OnlyDigits(e.CEP); len(cep) == 8 {
		parts = append(parts, "CEP "+cep[:5]+"-"+cep[5:])
	}
	return strings.Join(parts, ", ")
}

func nonEmpty(ss ...string) []string {
	var out []string
	for _, s := range ss {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
