package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// tomadorFlags são os dados de um tomador informados por flag; vazio = não informado.
type tomadorFlags struct {
	documento, nome, im, email, telefone                       string
	cep, logradouro, numero, complemento, bairro, cidade, pais string
	exterior                                                   bool
}

func (f *tomadorFlags) add(cmd *cobra.Command, comDocumento bool) {
	fl := cmd.Flags()
	if comDocumento {
		fl.StringVar(&f.documento, "documento", "", "CPF ou CNPJ do tomador nacional")
		fl.BoolVar(&f.exterior, "exterior", false, "tomador do exterior (sem documento brasileiro)")
	}
	fl.StringVar(&f.nome, "nome", "", "razão social ou nome (CNPJ: padrão é o da Receita)")
	fl.StringVar(&f.im, "im", "", "inscrição municipal (só pessoa jurídica)")
	fl.StringVar(&f.email, "email", "", "e-mail")
	fl.StringVar(&f.telefone, "telefone", "", "telefone (só nacional)")
	fl.StringVar(&f.cep, "cep", "", "CEP (nacional)")
	fl.StringVar(&f.logradouro, "logradouro", "", "logradouro (padrão: o do CEP)")
	fl.StringVar(&f.numero, "numero", "", "número")
	fl.StringVar(&f.complemento, "complemento", "", "complemento")
	fl.StringVar(&f.bairro, "bairro", "", "bairro (nacional; padrão: o do CEP)")
	fl.StringVar(&f.cidade, "cidade", "", "cidade (exterior)")
	fl.StringVar(&f.pais, "pais", "", "país: sigla (ex.: US), nome ou código (exterior)")
}

func valorOu(v, padrao string) string {
	if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return padrao
}

// aplicarCEP consulta o CEP e preenche código IBGE, UF e, se vazios, logradouro e bairro.
// CEP não encontrado marca cepInvalido, como o front.
func aplicarCEP(ctx context.Context, g api.Getter, e *api.EnderecoNacional, cep string) error {
	e.CEP = strings.NewReplacer("-", "", ".", "", " ", "").Replace(cep)
	end, err := api.BuscarCEP(ctx, g, e.CEP)
	var he *ctbz.HTTPError
	if errors.As(err, &he) {
		e.CEPInvalido, e.CodIBGE = true, ""
		return nil
	}
	if err != nil {
		return err
	}
	e.CEPInvalido, e.CodIBGE = false, end.CodIBGE
	e.Estado = valorOu(end.UF, e.Estado)
	e.Logradouro = valorOu(e.Logradouro, end.Logradouro)
	e.Bairro = valorOu(e.Bairro, end.Bairro)
	return nil
}

// tomadorNacional monta o cadastro de um tomador nacional novo: CNPJ parte da Receita, CPF
// exige nome e endereço.
func tomadorNacional(ctx context.Context, g api.Getter, f tomadorFlags) (api.ClienteNacional, error) {
	doc, err := validarDocumento(f.documento)
	if err != nil {
		return api.ClienteNacional{}, usageError{err}
	}
	c := api.ClienteNacional{CPFCNPJ: doc}
	pj := len(doc) == 14
	if pj {
		r, err := api.BuscarConsultaCNPJ(ctx, g, doc)
		if err != nil {
			return c, fmt.Errorf("consultando o CNPJ na Receita: %w", err)
		}
		c.RazaoSocialOuNome, c.Email, c.Telefone = r.RazaoSocial, r.Email, r.Telefone
		c.Endereco = api.EnderecoNacional{Logradouro: r.Logradouro, Numero: r.Numero, Complemento: r.Complemento, Bairro: r.Bairro, CEP: r.CEP, Estado: r.UF}
	} else if f.im != "" {
		return c, usageError{errors.New("pessoa física não tem inscrição municipal")}
	}
	return aplicarFlagsNacional(ctx, g, c, f)
}

// aplicarFlagsNacional aplica as flags informadas sobre um cadastro e confere o obrigatório.
func aplicarFlagsNacional(ctx context.Context, g api.Getter, c api.ClienteNacional, f tomadorFlags) (api.ClienteNacional, error) {
	c.RazaoSocialOuNome = valorOu(f.nome, c.RazaoSocialOuNome)
	c.Email = valorOu(f.email, c.Email)
	c.Telefone = valorOu(f.telefone, c.Telefone)
	if f.im != "" {
		im := strings.TrimSpace(f.im)
		c.InscricaoMunicipal = &im
	}
	e := &c.Endereco
	e.Logradouro = valorOu(f.logradouro, e.Logradouro)
	e.Numero = valorOu(f.numero, e.Numero)
	e.Complemento = valorOu(f.complemento, e.Complemento)
	e.Bairro = valorOu(f.bairro, e.Bairro)
	if f.cep != "" || (e.CEP != "" && e.CodIBGE == "") {
		if err := aplicarCEP(ctx, g, e, valorOu(f.cep, e.CEP)); err != nil {
			return c, err
		}
	}
	var faltam []string
	for nome, v := range map[string]string{"--nome": c.RazaoSocialOuNome, "--cep": e.CEP, "--logradouro": e.Logradouro, "--numero": e.Numero, "--bairro": e.Bairro} {
		if strings.TrimSpace(v) == "" {
			faltam = append(faltam, nome)
		}
	}
	if len(faltam) > 0 {
		return c, usageError{fmt.Errorf("faltam dados do tomador: informe %s", strings.Join(sortedCopy(faltam), ", "))}
	}
	return c, nil
}

// aplicarFlagsExterior aplica as flags sobre um tomador do exterior e confere o obrigatório.
func aplicarFlagsExterior(ctx context.Context, g api.Getter, c api.ClienteExterior, f tomadorFlags) (api.ClienteExterior, error) {
	if f.im != "" || f.telefone != "" || f.cep != "" || f.bairro != "" {
		return c, usageError{errors.New("tomador do exterior não tem inscrição municipal, telefone, CEP nem bairro")}
	}
	c.RazaoSocialOuNome = valorOu(f.nome, c.RazaoSocialOuNome)
	c.Email = valorOu(f.email, c.Email)
	e := &c.Endereco
	e.Logradouro = valorOu(f.logradouro, e.Logradouro)
	e.Numero = valorOu(f.numero, e.Numero)
	e.Complemento = valorOu(f.complemento, e.Complemento)
	e.Cidade = valorOu(f.cidade, e.Cidade)
	if f.pais != "" {
		paises, err := api.BuscarPaisesEmissao(ctx, g)
		if err != nil {
			return c, err
		}
		achou := false
		for _, p := range paises {
			if strings.EqualFold(p.Simbolo, f.pais) || strings.EqualFold(p.Descricao, f.pais) || rawTexto(p.Codigo) == f.pais {
				e.Pais, e.DescricaoPais, e.SimboloPais, achou = p.Codigo, p.Descricao, p.Simbolo, true
				break
			}
		}
		if !achou {
			return c, usageError{fmt.Errorf("país %q não encontrado entre os aceitos pelo emissor", f.pais)}
		}
	}
	var faltam []string
	for nome, v := range map[string]string{"--nome": c.RazaoSocialOuNome, "--pais": e.DescricaoPais, "--cidade": e.Cidade, "--logradouro": e.Logradouro, "--numero": e.Numero} {
		if strings.TrimSpace(v) == "" {
			faltam = append(faltam, nome)
		}
	}
	if len(faltam) > 0 {
		return c, usageError{fmt.Errorf("faltam dados do tomador: informe %s", strings.Join(sortedCopy(faltam), ", "))}
	}
	return c, nil
}

func resumoNacional(acao string, c api.ClienteNacional) string {
	e := c.Endereco
	return fmt.Sprintf("%s o tomador %s (%s): %s, %s, %s - %s, CEP %s", acao, c.RazaoSocialOuNome, c.CPFCNPJ, e.Logradouro, e.Numero, e.Bairro, e.Estado, e.CEP)
}

func resumoExterior(acao string, c api.ClienteExterior) string {
	e := c.Endereco
	return fmt.Sprintf("%s o tomador do exterior %s: %s, %s, %s, %s", acao, c.RazaoSocialOuNome, e.Logradouro, e.Numero, e.Cidade, e.DescricaoPais)
}

// tomadorRelido procura o tomador na lista pelo documento ou pelo nome.
func tomadorRelido(ctx context.Context, g api.Getter, documento, nome string) (*api.Tomador, error) {
	t, err := api.BuscarTomadores(ctx, g)
	if err != nil {
		return nil, err
	}
	for i, x := range t.Tomadores {
		if (documento != "" && normalizarDocumento(x.CPFCNPJ) == documento) || (documento == "" && strings.EqualFold(x.Nome, nome)) {
			return &t.Tomadores[i], nil
		}
	}
	return nil, nil
}

func tomadorResultado(acao, situacao string, relido *api.Tomador, nome, documento string) *output.Record {
	var id any
	if relido == nil {
		situacao = "enviado"
	} else {
		id = rawTexto(relido.ID)
	}
	return resultadoEscrita(acao, situacao, id).Add("nome", "Nome", nome).Add("documento", "Documento", nilIfEmpty(documento))
}

// salvarTomador confirma, envia e relê um cadastro nacional ou do exterior.
func salvarTomador(cmd *cobra.Command, acao, situacao, verbo string, nacional *api.ClienteNacional, exterior *api.ClienteExterior, id string) error {
	formato, err := outputFormat(cmd, "")
	if err != nil {
		return err
	}
	s := streamsOf(cmd)
	ctx := cmd.Context()
	var op operacao
	var enviar func(api.Sender) error
	nome, documento := "", ""
	if nacional != nil {
		op = operacao{Risco: riscoMedio, ID: nacional.CPFCNPJ, Resumo: resumoNacional(verbo, *nacional)}
		enviar = func(snd api.Sender) error { return api.SalvarClienteNacional(ctx, snd, *nacional) }
		nome, documento = nacional.RazaoSocialOuNome, nacional.CPFCNPJ
	} else {
		op = operacao{Risco: riscoMedio, ID: id, Resumo: resumoExterior(verbo, *exterior)}
		enviar = func(snd api.Sender) error { return api.SalvarClienteExterior(ctx, snd, *exterior) }
		nome = exterior.RazaoSocialOuNome
	}
	enviado, err := escrever(cmd, op, enviar)
	if err != nil || !enviado {
		return err
	}
	relido, err := tomadorRelido(ctx, sessionGetter{s}, documento, nome)
	if err != nil {
		fmt.Fprintln(s.err, "aviso: tomador salvo, mas não foi possível reler a lista:", err)
	}
	return output.Write(s.out, formato, tomadorResultado(acao, situacao, relido, nome, documento))
}

func newTomadoresAdicionarCmd() *cobra.Command {
	var f tomadorFlags
	cmd := &cobra.Command{
		Use:   "adicionar",
		Short: "Cadastra um tomador (cliente) nacional ou do exterior",
		Long: `Cadastra um tomador para as notas (risco médio: dá para editar depois).

Nacional (--documento): com CNPJ, a razão social e o endereço vêm da consulta à Receita, e as
flags só completam ou corrigem; com CPF, informe --nome, --cep e --numero (logradouro e bairro
vêm do CEP). O documento é validado pelos dígitos verificadores, inclusive o CNPJ
alfanumérico. Se o documento já estiver cadastrado, o cadastro é atualizado (o emissor salva
pelo documento).

Exterior (--exterior): informe --nome, --pais, --cidade, --logradouro e --numero.

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz notas tomadores adicionar --documento 00.000.000/0001-91 --email financeiro@exemplo.com
  ctbz notas tomadores adicionar --documento 529.982.247-25 --nome "Fulano de Tal" --cep 01001-000 --numero 10
  ctbz notas tomadores adicionar --exterior --nome "Example Inc" --pais US --cidade Springfield \
    --logradouro "Main St" --numero 100`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if f.exterior == (f.documento != "") {
				return usageError{errors.New("informe --documento (nacional) ou --exterior")}
			}
			ctx := cmd.Context()
			g := sessionGetter{streamsOf(cmd)}
			if f.exterior {
				c, err := aplicarFlagsExterior(ctx, g, api.ClienteExterior{}, f)
				if err != nil {
					return err
				}
				return salvarTomador(cmd, "adicionar", "cadastrado", "Cadastrar", nil, &c, "")
			}
			c, err := tomadorNacional(ctx, g, f)
			if err != nil {
				return err
			}
			return salvarTomador(cmd, "adicionar", "cadastrado", "Cadastrar", &c, nil, "")
		},
	}
	f.add(cmd, true)
	addWriteFlags(cmd)
	return cmd
}

func newTomadoresEditarCmd() *cobra.Command {
	var f tomadorFlags
	cmd := &cobra.Command{
		Use:   "editar DOCUMENTO|ID",
		Short: "Corrige os dados de um tomador cadastrado",
		Long: `Corrige os dados de um tomador (risco médio: dá para editar de novo). A CLI lê o cadastro
atual e aplica só as flags informadas. Nacionais são identificados pelo documento; tomadores
do exterior, pelo id (coluna id de ctbz notas tomadores). O documento não muda.

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz notas tomadores editar 00000000000191 --email novo@exemplo.com
  ctbz notas tomadores editar 1000000000000002 --cidade Boston`,
		Args: exactArgs(1, "o DOCUMENTO ou ID do tomador"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			g := sessionGetter{streamsOf(cmd)}
			id := normalizarDocumento(args[0])
			cad, err := api.BuscarCadastroCliente(ctx, g, id)
			var he *ctbz.HTTPError
			if errors.As(err, &he) && he.Status == 204 {
				return fmt.Errorf("tomador %s não encontrado", args[0])
			}
			if err != nil {
				return err
			}
			switch {
			case cad.TipoCliente == "EXTERIOR" && cad.ClienteExteriorDTO != nil:
				c := *cad.ClienteExteriorDTO
				if len(c.ID) == 0 {
					c.ID = []byte(fmt.Sprintf("%q", id))
					if soDigitos(id) {
						c.ID = []byte(id)
					}
				}
				if c, err = aplicarFlagsExterior(ctx, g, c, f); err != nil {
					return err
				}
				return salvarTomador(cmd, "editar", "editado", "Atualizar", nil, &c, id)
			case cad.ClienteNacionalDTO != nil:
				c := *cad.ClienteNacionalDTO
				if f.im != "" && len(c.CPFCNPJ) == 11 {
					return usageError{errors.New("pessoa física não tem inscrição municipal")}
				}
				if c, err = aplicarFlagsNacional(ctx, g, c, f); err != nil {
					return err
				}
				return salvarTomador(cmd, "editar", "editado", "Atualizar", &c, nil, "")
			}
			return fmt.Errorf("cadastro do tomador %s em formato inesperado (%s)", args[0], cad.TipoCliente)
		},
	}
	f.add(cmd, false)
	addWriteFlags(cmd)
	return cmd
}

func sortedCopy(ss []string) []string {
	sort.Strings(ss)
	return ss
}
