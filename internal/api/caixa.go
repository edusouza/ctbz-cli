package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// PathCaixa é a lista de lançamentos do caixa de um mês (1–12). O painel pede 1000 registros
// e manda a página literalmente como "null"; a CLI faz o mesmo.
func PathCaixa(ano, mes int) string {
	return fmt.Sprintf("caixa/listpaginada/%d/%d/%d/null", ano, mes, TamanhoPaginaCaixa)
}

// TamanhoPaginaCaixa é o número de registros pedido por mês, como no painel.
const TamanhoPaginaCaixa = 1000

const PathContasUsuario = "movimentacao-financeira/contasUsuario"

// LancamentoCaixa é um lançamento do caixa (entrada ou saída classificada).
type LancamentoCaixa struct {
	ID                   int64    `json:"id"`
	Data                 int64    `json:"data"` // epoch em ms
	Descricao            string   `json:"descricao"`
	Valor                *float64 `json:"valor"` // negativo nas saídas
	Situacao             string   `json:"situacao"`
	Tipo                 string   `json:"tipo"`
	IDContaUsuario       *int64   `json:"idContaUsuario"` // classificação (ContaUsuario)
	ConfirmadoViaSistema bool     `json:"confirmadoViaSistema"`
	// IDVinculo é a guia ou o sócio vinculado (incerto se a listagem traz; ADR-0021).
	IDVinculo json.RawMessage `json:"idVinculo" contract:"optional"`
}

// Caixa é a resposta de caixa/listpaginada.
type Caixa struct {
	List  []LancamentoCaixa `json:"list"`
	Total int               `json:"total"`
}

// BuscarCaixa lê os lançamentos do caixa de um mês.
func BuscarCaixa(ctx context.Context, g Getter, ano, mes int) (*Caixa, error) {
	return get[Caixa](ctx, g, PathCaixa(ano, mes))
}

// ContaUsuario é uma conta do plano de contas simplificado usado para classificar os
// lançamentos (ex.: "Pagamento de Fornecedores"), ligada a uma conta contábil.
type ContaUsuario struct {
	ID                     int64  `json:"id"`
	Descricao              string `json:"descricao"`
	DescricaoContaContabil string `json:"descricaoContaContabil"`
	Classificacao          string `json:"classificacao"` // ex.: RECEITA, DESPESA
	Situacao               string `json:"situacao"`      // ATIVO ou INATIVO
}

// BuscarContasUsuario lê o plano de contas usado nas classificações.
func BuscarContasUsuario(ctx context.Context, g Getter) ([]ContaUsuario, error) {
	c, err := get[[]ContaUsuario](ctx, g, PathContasUsuario)
	if err != nil {
		return nil, err
	}
	return *c, nil
}

// PathSalvarLancamentoCaixa cria e edita lançamentos do caixa: com LancamentoUsuario.ID é
// uma edição (não existe PUT). Ver docs/api/escrita/contabilidade-e-documentos.md, seção 1.1.
const PathSalvarLancamentoCaixa = "caixa/lancamentousuario/novo/"

// SalvarLancamentoCaixa é o corpo de PathSalvarLancamentoCaixa.
type SalvarLancamentoCaixa struct {
	Ano               string                 `json:"ano"` // string, como o front envia
	Mes               int                    `json:"mes"`
	LancamentoUsuario LancamentoCaixaUsuario `json:"lancamentoUsuario"`
}

// LancamentoCaixaUsuario é o lançamento enviado ao salvar. Não há campo de tipo: o sinal de
// Valor é o tipo (positivo = recebimento, negativo = pagamento).
type LancamentoCaixaUsuario struct {
	Data           string `json:"data"` // ISO, meia-noite de Brasília (DataISOBrasilia)
	Descricao      string `json:"descricao"`
	ID             *int64 `json:"id,omitempty"` // só na edição
	IDContaUsuario int64  `json:"idContaUsuario"`
	// IDVinculo é a guia (ou "0", SEM GUIA) ou o sócio exigido pela classificação, como veio
	// em CategoriaVinculo.ID; nil é enviado como null.
	IDVinculo json.RawMessage `json:"idVinculo"`
	Valor     float64         `json:"valor"`
}

// SemGuia é o idVinculo de "SEM GUIA" nas classificações de imposto.
var SemGuia = json.RawMessage(`"0"`)

// SalvarLancamento cria (ou edita, com ID) um lançamento do caixa.
func SalvarLancamento(ctx context.Context, s Sender, req SalvarLancamentoCaixa) error {
	return s.Send(ctx, "POST", PathSalvarLancamentoCaixa, req, nil)
}

// brasilia é o fuso que o painel usa nas datas (sem horário de verão desde 2019).
var brasilia = time.FixedZone("BRT", -3*60*60)

// DataISOBrasilia formata o dia como o front (new Date de meia-noite em Brasília, em UTC).
func DataISOBrasilia(dia time.Time) string {
	d := time.Date(dia.Year(), dia.Month(), dia.Day(), 0, 0, 0, 0, brasilia)
	return d.UTC().Format("2006-01-02T15:04:05.000Z")
}

// DiaEmBrasilia devolve o dia (meia-noite UTC) que o instante t é em Brasília. Para epoch em
// milissegundos das respostas, use DiaEmBrasilia(time.UnixMilli(ms)).
func DiaEmBrasilia(t time.Time) time.Time {
	t = t.In(brasilia)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// PathRemoverLancamentoCaixa exclui um lançamento manual do caixa (DELETE, sem corpo).
// Ver docs/api/escrita/contabilidade-e-documentos.md, seção 1.2.
func PathRemoverLancamentoCaixa(ano, mes int, id int64) string {
	return fmt.Sprintf("caixa/lancamentousuario/remover/%d/%d/%d", ano, mes, id)
}

// RemoverLancamento exclui um lançamento do caixa. Não há como desfazer pela API.
func RemoverLancamento(ctx context.Context, s Sender, ano, mes int, id int64) error {
	return s.Send(ctx, "DELETE", PathRemoverLancamentoCaixa(ano, mes, id), nil, nil)
}
