package api

import (
	"context"
	"fmt"
)

const (
	PathExtratos        = "movimentacao-financeira/v2/extratos"
	PathContasBancarias = "contabancaria/list"
)

// Extrato é a situação do extrato de uma conta bancária em um mês.
type Extrato struct {
	Ano              int    `json:"ano"`
	Mes              int    `json:"mes"`
	IDContaBancaria  int64  `json:"idContaBancaria"`
	Banco            string `json:"banco"`
	Agencia          string `json:"agencia"`
	NumeroConta      string `json:"numeroConta"`
	Situacao         string `json:"situacao"`         // ex.: ABERTO
	StatusIntegracao any    `json:"statusIntegracao"` // formato não verificado (vinha null)
}

// BuscarExtratos lista os extratos de todas as contas (a API não tem filtro).
func BuscarExtratos(ctx context.Context, g Getter) ([]Extrato, error) {
	e, err := get[[]Extrato](ctx, g, PathExtratos)
	if err != nil {
		return nil, err
	}
	return *e, nil
}

// ContaBancaria é uma conta bancária cadastrada da empresa.
type ContaBancaria struct {
	ID               int64    `json:"id"`
	NomeBanco        string   `json:"nomeBanco"`
	CodigoBanco      string   `json:"codigoBanco"`
	Agencia          string   `json:"agencia"`
	ContaCorrente    string   `json:"contaCorrente"`
	VlrSaldoInicial  *float64 `json:"vlrSaldoInicial"`
	DataSaldoInicial int64    `json:"dataSaldoInicial"`            // epoch em ms
	StatusIntegracao string   `json:"statusIntegracao"`            // ex.: INTEGRADA
	FluxoIntegracao  string   `json:"fluxoIntegracao"`             // ex.: CONTABILIZEI_BANK
	BancoID          *int64   `json:"bancoId" contract:"optional"` // usado na edição
}

// Banco é um banco aceito no cadastro de contas (nomes dos campos incertos; ADR-0021).
type Banco struct {
	ID     int64  `json:"id"`
	Nome   string `json:"nome" contract:"optional"`
	Codigo string `json:"codigo" contract:"optional"`
}

// ContasBancarias é a resposta de contabancaria/list (que também traz a lista de bancos).
type ContasBancarias struct {
	ContasBancarias []ContaBancaria `json:"contasBancarias"`
	Bancos          []Banco         `json:"bancos" contract:"optional"`
}

// BuscarContasBancarias lista as contas bancárias da empresa.
func BuscarContasBancarias(ctx context.Context, g Getter) (*ContasBancarias, error) {
	return get[ContasBancarias](ctx, g, PathContasBancarias)
}

// PathDetalheContaBancaria diz se a conta pode ser editada ou excluída, e por quê.
func PathDetalheContaBancaria(id int64) string {
	return fmt.Sprintf("contabancaria/detalhes-da-conta/init/%d", id)
}

// DetalheContaBancaria é a resposta de PathDetalheContaBancaria (ADR-0021).
type DetalheContaBancaria struct {
	PermiteEditar                   bool   `json:"permiteEditar"`
	PermiteExcluir                  bool   `json:"permiteExcluir"`
	MotivoPermissoesExclusaoEEdicao string `json:"motivoPermissoesExclusaoEEdicao" contract:"optional"`
}

// BuscarDetalheContaBancaria lê as permissões de edição e exclusão de uma conta.
func BuscarDetalheContaBancaria(ctx context.Context, g Getter, id int64) (*DetalheContaBancaria, error) {
	return get[DetalheContaBancaria](ctx, g, PathDetalheContaBancaria(id))
}

// PathSalvarContaBancaria cadastra (sem ID) ou edita (com ID) uma conta bancária.
// Ver docs/api/escrita/contabilidade-e-documentos.md, seção 2.4.
const PathSalvarContaBancaria = "contabancaria/salvar"

// ContaBancariaSalvar é o corpo de PathSalvarContaBancaria.
type ContaBancariaSalvar struct {
	ID               *int64  `json:"id,omitempty"`
	BancoID          int64   `json:"bancoId"`
	Agencia          string  `json:"agencia"`
	ContaCorrente    string  `json:"contaCorrente"`    // conta e dígito, sem hífen
	DataSaldoInicial string  `json:"dataSaldoInicial"` // abertura da conta, AAAA-MM-DD
	VlrSaldoInicial  float64 `json:"vlrSaldoInicial"`
}

// SalvarContaBancaria cadastra ou edita a conta.
func SalvarContaBancaria(ctx context.Context, s Sender, c ContaBancariaSalvar) error {
	return s.Send(ctx, "POST", PathSalvarContaBancaria, c, nil)
}

// PathExcluirContaBancaria exclui uma conta bancária (DELETE, sem corpo). Só com permiteExcluir.
func PathExcluirContaBancaria(id int64) string {
	return fmt.Sprintf("contabancaria/excluir/%d", id)
}

// ExcluirContaBancaria exclui a conta permanentemente; os vínculos não voltam com um recadastro.
func ExcluirContaBancaria(ctx context.Context, s Sender, id int64) error {
	return s.Send(ctx, "DELETE", PathExcluirContaBancaria(id), nil, nil)
}
