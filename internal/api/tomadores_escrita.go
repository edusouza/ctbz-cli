package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Cadastro de tomadores (clientes) do emissor. Ver docs/api/escrita/notas-e-prolabore.md, seção 2.

// PathCEP consulta um CEP (logradouro, bairro, UF e código IBGE).
func PathCEP(cep string) string { return "novo-emissor/cep/logradouro?cep=" + url.QueryEscape(cep) }

// EnderecoCEP é a resposta de PathCEP (nomes dos campos incertos; ADR-0021).
type EnderecoCEP struct {
	Logradouro string `json:"logradouro" contract:"optional"`
	Bairro     string `json:"bairro" contract:"optional"`
	Cidade     string `json:"cidade" contract:"optional"`
	UF         string `json:"uf" contract:"optional"`
	CodIBGE    string `json:"codIbge"`
}

// BuscarCEP consulta um CEP (só dígitos).
func BuscarCEP(ctx context.Context, g Getter, cep string) (*EnderecoCEP, error) {
	return get[EnderecoCEP](ctx, g, PathCEP(cep))
}

// PathPaisesEmissao são os países aceitos para tomadores do exterior.
const PathPaisesEmissao = "notafiscal/emitir/buscarPaisesParaEmissao/"

// PaisEmissao é um país aceito (nomes dos campos incertos; ADR-0021).
type PaisEmissao struct {
	Codigo    json.RawMessage `json:"codigo"`
	Descricao string          `json:"descricao"`
	Simbolo   string          `json:"simbolo" contract:"optional"` // ex.: US
}

// BuscarPaisesEmissao lê os países aceitos.
func BuscarPaisesEmissao(ctx context.Context, g Getter) ([]PaisEmissao, error) {
	p, err := get[[]PaisEmissao](ctx, g, PathPaisesEmissao)
	if err != nil {
		return nil, err
	}
	return *p, nil
}

// EnderecoNacional é o endereço de um tomador nacional.
type EnderecoNacional struct {
	Bairro      string `json:"bairro"`
	CEP         string `json:"cep"` // só dígitos
	CodIBGE     string `json:"codIbge"`
	Complemento string `json:"complemento"`
	Logradouro  string `json:"logradouro"`
	Estado      string `json:"estado"` // UF
	Numero      string `json:"numero"`
	CEPInvalido bool   `json:"cepInvalido"`
}

// ClienteNacional é o corpo de PathSalvarClienteNacional (e o DTO do cadastro).
type ClienteNacional struct {
	CPFCNPJ            string           `json:"cpfCnpj"` // só dígitos; CNPJ alfanumérico em maiúsculas
	RazaoSocialOuNome  string           `json:"razaoSocialOuNome"`
	Telefone           string           `json:"telefone"`
	Email              string           `json:"email"`
	InscricaoMunicipal *string          `json:"inscricaoMunicipal"` // null para pessoa física
	Endereco           EnderecoNacional `json:"endereco"`
}

// EnderecoExterior é o endereço de um tomador do exterior.
type EnderecoExterior struct {
	Complemento   string          `json:"complemento"`
	Logradouro    string          `json:"logradouro"`
	Numero        string          `json:"numero"`
	Cidade        string          `json:"cidade"`
	Pais          json.RawMessage `json:"pais"` // código, como em PaisEmissao
	DescricaoPais string          `json:"descricaoPais"`
	SimboloPais   string          `json:"simboloPais"`
}

// ClienteExterior é o corpo de PathSalvarClienteExterior (e o DTO do cadastro).
type ClienteExterior struct {
	RazaoSocialOuNome string           `json:"razaoSocialOuNome"`
	Email             string           `json:"email"`
	Endereco          EnderecoExterior `json:"endereco"`
	ID                json.RawMessage  `json:"id,omitempty"` // só na edição
}

// Cadastro de tomadores: o nacional é um upsert pelo documento (sem id).
const (
	PathSalvarClienteNacional = "novo-emissor/clientes/salvar-cliente-nacional"
	PathSalvarClienteExterior = "novo-emissor/clientes/salvar-cliente-exterior"
)

// SalvarClienteNacional cria ou atualiza um tomador nacional.
func SalvarClienteNacional(ctx context.Context, s Sender, c ClienteNacional) error {
	return s.Send(ctx, "POST", PathSalvarClienteNacional, c, nil)
}

// SalvarClienteExterior cria (sem ID) ou atualiza um tomador do exterior.
func SalvarClienteExterior(ctx context.Context, s Sender, c ClienteExterior) error {
	return s.Send(ctx, "POST", PathSalvarClienteExterior, c, nil)
}

// PathCadastroCliente é o cadastro de um tomador para edição (HTTP 204 = não encontrado).
func PathCadastroCliente(id string) string {
	return fmt.Sprintf("novo-emissor/cadastro-clientes/%s", url.PathEscape(id))
}

// CadastroCliente é a resposta de PathCadastroCliente (ADR-0021).
type CadastroCliente struct {
	TipoCliente        string           `json:"tipoCliente"` // NACIONAL ou EXTERIOR
	ClienteNacionalDTO *ClienteNacional `json:"clienteNacionalDTO" contract:"optional"`
	ClienteExteriorDTO *ClienteExterior `json:"clienteExteriorDTO" contract:"optional"`
}

// BuscarCadastroCliente lê o cadastro de um tomador.
func BuscarCadastroCliente(ctx context.Context, g Getter, id string) (*CadastroCliente, error) {
	return get[CadastroCliente](ctx, g, PathCadastroCliente(id))
}
