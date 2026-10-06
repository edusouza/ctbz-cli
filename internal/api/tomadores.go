package api

import (
	"context"
	"encoding/json"
)

const PathTomadores = "novo-emissor/tomadores/init"

// PathConsultaCNPJ consulta os dados cadastrais (Receita Federal) de um CNPJ, como o
// emissor faz ao cadastrar um tomador. Recebe só os dígitos.
func PathConsultaCNPJ(cnpj string) string { return "novo-emissor/clientes/consulta/" + cnpj }

// Tomador é um cliente cadastrado no emissor de notas. Os campos são os lidos pelo front;
// a conta verificada não tinha tomadores, por isso todos são opcionais.
type Tomador struct {
	ID                 json.RawMessage `json:"id" contract:"optional"` // número ou texto
	Nome               string          `json:"nome" contract:"optional"`
	CPFCNPJ            string          `json:"cpfCnpj" contract:"optional"`
	Email              string          `json:"email" contract:"optional"`
	Telefone           string          `json:"telefone" contract:"optional"`
	InscricaoMunicipal string          `json:"inscricaoMunicipal" contract:"optional"`
	Estrangeiro        bool            `json:"estrangeiro" contract:"optional"`
	Endereco           *struct {
		Municipio any `json:"municipio" contract:"optional"` // texto ou objeto
		UF        any `json:"uf" contract:"optional"`
	} `json:"endereco" contract:"optional"`
}

// Tomadores é a resposta de novo-emissor/tomadores/init.
type Tomadores struct {
	Tomadores              []Tomador `json:"tomadores"`
	EmissaoSemTomador      bool      `json:"emissaoSemTomador"`
	PermiteEmissaoExterior bool      `json:"permiteEmissaoExterior"`
}

// BuscarTomadores lê os tomadores cadastrados.
func BuscarTomadores(ctx context.Context, g Getter) (*Tomadores, error) {
	return get[Tomadores](ctx, g, PathTomadores)
}

// ConsultaCNPJ é o cadastro de um CNPJ na Receita, como o emissor o devolve.
type ConsultaCNPJ struct {
	CNPJ               string `json:"cnpj"`
	RazaoSocial        string `json:"razaoSocial"`
	NomeFantasia       string `json:"nomeFantasia"`
	DataAbertura       string `json:"dataAbertura"` // dd/mm/aaaa
	AtividadePrincipal *struct {
		Codigo    string `json:"codigo"`
		Descricao string `json:"descricao"`
	} `json:"atividadePrincipal"`
	NaturezaJuridica *struct {
		Descricao string `json:"descricao"`
	} `json:"naturezaJuridica"`
	Logradouro  string `json:"logradouro"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	CEP         string `json:"cep"`
	Municipio   string `json:"municipio"`
	UF          string `json:"uf"`
	Email       string `json:"email"`
	Telefone    string `json:"telefone"`
	// Situação na Receita (ex.: ATIVA) e opção pelo Simples ("SIM"/"NÃO").
	SituacaoCadastral string `json:"situacaoCadastral"`
	OptanteSimples    string `json:"optanteSimples"`
}

// BuscarConsultaCNPJ consulta um CNPJ (só dígitos).
func BuscarConsultaCNPJ(ctx context.Context, g Getter, cnpj string) (*ConsultaCNPJ, error) {
	return get[ConsultaCNPJ](ctx, g, PathConsultaCNPJ(cnpj))
}
