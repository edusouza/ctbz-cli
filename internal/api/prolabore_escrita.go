package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

// Gestão do pró-labore de um sócio (central do sócio). Ver
// docs/api/escrita/notas-e-prolabore.md, seções 4.1 e 4.4.

// PathGestaoSocio é a gestão de pró-labore de um sócio (id de ProlaboreCentral.Socios):
// GET lê, PUT altera.
func PathGestaoSocio(idSocio int64) string {
	return "prolabore/central/gestao/" + strconv.FormatInt(idSocio, 10)
}

// Tipos de gerenciamento do pró-labore.
const (
	GerenciamentoInteligente   = "INTELIGENTE"
	GerenciamentoSalarioMinimo = "SALARIO_MINIMO"
	GerenciamentoTetoINSS      = "TETO_INSS"
	GerenciamentoPersonalizado = "PERSONALIZADO"
)

// GestaoSocio é a resposta do GET de PathGestaoSocio. Campos citados no front; sem captura
// real (ADR-0021).
type GestaoSocio struct {
	TipoGerenciamentoProlabore string   `json:"tipoGerenciamentoProlabore" contract:"optional"`
	ValorProlaboreMinimo       *float64 `json:"valorProlaboreMinimo" contract:"optional"`
	ValorMaximoProlabore       *float64 `json:"valorMaximoProlabore" contract:"optional"` // teto do INSS
	SalarioMinimo              *float64 `json:"salarioMinimo" contract:"optional"`
	QtdSocioGestaoInteligente  int      `json:"qtdSocioGestaoInteligente" contract:"optional"`
	ElegivelNoMotor            bool     `json:"elegivelNoMotor" contract:"optional"`
}

// BuscarGestaoSocio lê a gestão de pró-labore de um sócio.
func BuscarGestaoSocio(ctx context.Context, g Getter, idSocio int64) (*GestaoSocio, error) {
	return get[GestaoSocio](ctx, g, PathGestaoSocio(idSocio))
}

// AlteracaoGestao é o corpo do PUT de PathGestaoSocio. Os valores são em reais, como a API.
type AlteracaoGestao struct {
	TipoGerenciamento      string   `json:"tipoGerenciamento"`
	ValorProlabore         *float64 `json:"valorProlabore"`
	ValorProlaboreMinimo   *float64 `json:"valorProlaboreMinimo"`
	ModoEdicaoProlaboreMin bool     `json:"modoEdicaoProlaboreMin"`
	SairGestaoInteligente  bool     `json:"sairGestaoInteligente"`
}

// NovaAlteracaoGestao monta o corpo como o front: o valor vem do teto do INSS ou do salário
// mínimo da gestão atual; o personalizado precisa ser pelo menos o salário mínimo; o piso
// (minimo) só vale na gestão inteligente; e a empresa sai da gestão inteligente quando este
// era o último sócio nela.
func NovaAlteracaoGestao(g GestaoSocio, tipo string, valor, minimo *float64) (AlteracaoGestao, error) {
	a := AlteracaoGestao{
		TipoGerenciamento:      tipo,
		ModoEdicaoProlaboreMin: g.ValorProlaboreMinimo != nil,
		SairGestaoInteligente:  tipo != GerenciamentoInteligente && g.QtdSocioGestaoInteligente <= 1,
	}
	if valor != nil && tipo != GerenciamentoPersonalizado {
		return a, errors.New("o valor só é informado no pró-labore personalizado")
	}
	if minimo != nil && tipo != GerenciamentoInteligente {
		return a, errors.New("o valor mínimo só vale na gestão inteligente")
	}
	switch tipo {
	case GerenciamentoTetoINSS:
		if g.ValorMaximoProlabore == nil {
			return a, errors.New("a Contabilizei não informou o teto do INSS")
		}
		a.ValorProlabore = g.ValorMaximoProlabore
	case GerenciamentoSalarioMinimo:
		if g.SalarioMinimo == nil {
			return a, errors.New("a Contabilizei não informou o salário mínimo")
		}
		a.ValorProlabore = g.SalarioMinimo
	case GerenciamentoPersonalizado:
		if valor == nil {
			return a, errors.New("informe o valor do pró-labore personalizado")
		}
		if g.SalarioMinimo != nil && *valor < *g.SalarioMinimo {
			return a, fmt.Errorf("valor deve ser maior ou igual ao salário mínimo vigente: %.2f", *g.SalarioMinimo)
		}
		a.ValorProlabore = valor
	case GerenciamentoInteligente:
		if !g.ElegivelNoMotor {
			return a, errors.New("a empresa não é elegível à gestão inteligente")
		}
		if minimo != nil && *minimo <= 0 {
			return a, errors.New("o valor mínimo deve ser positivo")
		}
		a.ValorProlaboreMinimo = minimo
	default:
		return a, fmt.Errorf("tipo de gerenciamento desconhecido: %q", tipo)
	}
	return a, nil
}

// AlterarGestaoSocio envia a nova gestão do pró-labore de um sócio. O servidor decide a
// partir de qual competência vale (seção 4.4).
func AlterarGestaoSocio(ctx context.Context, s Sender, idSocio int64, a AlteracaoGestao) error {
	return s.Send(ctx, "PUT", PathGestaoSocio(idSocio), a, nil)
}
