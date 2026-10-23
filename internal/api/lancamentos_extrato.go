package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// RegistrosPorPaginaExtrato é o tamanho de página que o painel pede.
const RegistrosPorPaginaExtrato = 100

// PathLancamentosExtrato lista os lançamentos do extrato de uma conta bancária num mês
// (páginas a partir de 1). Fora do catálogo; ver docs/api/escrita/contabilidade-e-documentos.md, 1.3.
func PathLancamentosExtrato(idContaBancaria int64, ano, mes, pagina int) string {
	return fmt.Sprintf("movimentacao-financeira/lancamento-usuario/paginado?idContaBancaria=%d&ano=%d&mes=%d&registroPorPagina=%d&pagina=%d",
		idContaBancaria, ano, mes, RegistrosPorPaginaExtrato, pagina)
}

// LancamentoExtrato é um lançamento do extrato importado.
type LancamentoExtrato struct {
	ID        int64    `json:"idLancamentoUsuario"`
	Data      Data     `json:"data"`
	Descricao string   `json:"descricao"`
	Valor     *float64 `json:"valor"` // negativo nas saídas
	// IDContaUsuario é a classificação atual (nil se ainda não classificado).
	IDContaUsuario  *int64          `json:"idContaUsuario"`
	IDLancamentoPai *int64          `json:"idLancamentoPai" contract:"optional"` // parte de um desmembramento
	IDVinculo       json.RawMessage `json:"idVinculo" contract:"optional"`
	Situacao        string          `json:"situacao" contract:"optional"` // ex.: CONFIRMADO
	// Sugestoes são as classificações que a Contabilizei sugere, que o painel mostra primeiro.
	Sugestoes []SugestaoClassificacao `json:"predicts" contract:"optional"`
}

// SugestaoClassificacao é uma classificação sugerida para um lançamento do extrato.
type SugestaoClassificacao struct {
	IDContaUsuario int64   `json:"idContaUsuario"`
	Descricao      string  `json:"descricao"`
	Probabilidade  float64 `json:"probabilidade"` // de 0 a 1
}

// MelhorSugestao devolve a sugestão mais provável, se houver.
func (l LancamentoExtrato) MelhorSugestao() *SugestaoClassificacao {
	var melhor *SugestaoClassificacao
	for i := range l.Sugestoes {
		if melhor == nil || l.Sugestoes[i].Probabilidade > melhor.Probabilidade {
			melhor = &l.Sugestoes[i]
		}
	}
	return melhor
}

// maxPaginasExtrato protege contra paginação que nunca termina (10 mil lançamentos no mês).
const maxPaginasExtrato = 100

// BuscarLancamentosExtrato lê todas as páginas de lançamentos do extrato do mês. Cada página é
// um array (a resposta não traz total): a leitura para na primeira página incompleta ou vazia.
// Uma página cheia que só repete lançamentos já lidos (o servidor ignorou o número da página) e
// o limite de páginas são erro, para nunca devolver a lista truncada em silêncio.
func BuscarLancamentosExtrato(ctx context.Context, g Getter, idContaBancaria int64, ano, mes int) ([]LancamentoExtrato, error) {
	todos := []LancamentoExtrato{}
	vistos := map[int64]bool{}
	for pagina := 1; pagina <= maxPaginasExtrato; pagina++ {
		p, err := get[[]LancamentoExtrato](ctx, g, PathLancamentosExtrato(idContaBancaria, ano, mes, pagina))
		if err != nil {
			return nil, err
		}
		novos := 0
		for _, l := range *p {
			if !vistos[l.ID] {
				vistos[l.ID] = true
				todos = append(todos, l)
				novos++
			}
		}
		if len(*p) < RegistrosPorPaginaExtrato {
			return todos, nil
		}
		if novos == 0 {
			return nil, fmt.Errorf("a página %d de lançamentos do extrato de %02d/%d repete a anterior; a paginação não avança e a lista ficaria incompleta", pagina, mes, ano)
		}
	}
	return nil, fmt.Errorf("o extrato de %02d/%d tem mais de %d páginas de lançamentos; a paginação parece não terminar", mes, ano, maxPaginasExtrato)
}

// PathExtratoInfo descreve o extrato de uma conta num mês: se dá para editar os lançamentos,
// se dá para excluir o extrato e os sócios para as classificações de sócio. Parâmetros incertos.
func PathExtratoInfo(idContaBancaria int64, ano, mes int) string {
	return fmt.Sprintf("movimentacao-financeira/extrato/info?idContaBancaria=%d&ano=%d&mes=%d", idContaBancaria, ano, mes)
}

// SocioExtrato é um sócio oferecido nas classificações de sócio do extrato.
type SocioExtrato struct {
	ID   int64  `json:"id"`
	Nome string `json:"nome"`
}

// ExtratoInfo é a resposta de PathExtratoInfo (ADR-0021).
type ExtratoInfo struct {
	PermiteEditarLancamento bool           `json:"permiteEditarLancamento"`
	PermiteExclusaoExtrato  bool           `json:"permiteExclusaoExtrato" contract:"optional"`
	Socios                  []SocioExtrato `json:"socios" contract:"optional"`
}

// BuscarExtratoInfo lê as permissões e os sócios do extrato do mês.
func BuscarExtratoInfo(ctx context.Context, g Getter, idContaBancaria int64, ano, mes int) (*ExtratoInfo, error) {
	return get[ExtratoInfo](ctx, g, PathExtratoInfo(idContaBancaria, ano, mes))
}

// PathClassificarLancamento troca a classificação de um lançamento do extrato (PUT, corpo em
// string JSON). O front escreve "/movimentacao-financeira/…", relativo à base do BFF.
const PathClassificarLancamento = "movimentacao-financeira/classificar"

// ClassificarLancamento é o corpo de PathClassificarLancamento.
type ClassificarLancamento struct {
	IDLancamentoUsuario int64  `json:"idLancamentoUsuario"`
	IDContaUsuario      int64  `json:"idContaUsuario"`
	IDSocio             *int64 `json:"idSocio"` // null fora das contas de sócio
}

// Classificar troca a classificação de um lançamento do extrato. Um 406
// (exception/erro-negocial-406) indica que os contadores já classificaram o período.
func Classificar(ctx context.Context, s Sender, req ClassificarLancamento) error {
	return s.Send(ctx, "PUT", PathClassificarLancamento, JSONString{req}, nil)
}

// Data é uma data de resposta que pode vir como epoch em milissegundos ou como texto ISO.
type Data struct{ time.Time }

// UnmarshalJSON aceita número (epoch em ms), "AAAA-MM-DD…" ou null.
func (d *Data) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	if ms, err := strconv.ParseInt(string(b), 10, 64); err == nil {
		d.Time = time.UnixMilli(ms)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("data inesperada %s", b)
	}
	if len(s) >= 10 {
		if t, err := time.Parse("2006-01-02", s[:10]); err == nil {
			d.Time = t.Add(3 * time.Hour) // meia-noite em Brasília
			return nil
		}
	}
	return fmt.Errorf("data inesperada %q", s)
}

// PathDesmembrar divide um lançamento do extrato em partes (PUT, corpo em string JSON).
// Ver docs/api/escrita/contabilidade-e-documentos.md, seção 1.4.
const PathDesmembrar = "movimentacao-financeira/desmembrar"

// Desmembramento é o corpo de PathDesmembrar.
type Desmembramento struct {
	IDLancamentoPai  int64             `json:"idLancamentoPai"`
	LancamentosFilho []LancamentoFilho `json:"lancamentosFilho"`
}

// LancamentoFilho é uma parte do desmembramento, com o mesmo sinal do lançamento original.
type LancamentoFilho struct {
	Descricao      string  `json:"descricao"`
	Valor          float64 `json:"valor"`
	IDContaUsuario int64   `json:"idContaUsuario"`
	IDVinculo      *int64  `json:"idVinculo"` // sócio, nas classificações de sócio
}

// Desmembrar divide um lançamento do extrato. Desfaz-se com DesfazerDesmembramento.
func Desmembrar(ctx context.Context, s Sender, req Desmembramento) error {
	return s.Send(ctx, "PUT", PathDesmembrar, JSONString{req}, nil)
}

// PathDesfazerDesmembramento volta um lançamento desmembrado ao original. O front repete o id
// na query. Fora do catálogo; ver a seção 1.5.
func PathDesfazerDesmembramento(idLancamentoPai int64) string {
	return fmt.Sprintf("movimentacao-financeira/desmembrar/desfazer/%d?idLancamentoPai=%d", idLancamentoPai, idLancamentoPai)
}

// DesfazerDesmembramento apaga as partes e restaura o lançamento original.
func DesfazerDesmembramento(ctx context.Context, s Sender, idLancamentoPai int64) error {
	return s.Send(ctx, "DELETE", PathDesfazerDesmembramento(idLancamentoPai), nil, nil)
}
