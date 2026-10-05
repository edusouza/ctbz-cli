package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// TipoEspecializado é o parcelamento feito por um especialista (negociação especializada).
const TipoEspecializado = "especializado"

// tiposAutomaticos são os parcelamentos de negociação automática (Receita e PGFN).
var tiposAutomaticos = map[string]bool{
	"simples-nacional":        true,
	"pgfn-simples-nacional":   true,
	"pgfn-previdenciario":     true,
	"pgfn-nao-previdenciario": true,
}

// NegociacoesEspecializadas são os valores aceitos em tipoNegociacao.
var NegociacoesEspecializadas = []string{"VENCIDOS", "DIVIDA_ATIVA", "DIVIDA_ATIVA_E_VENCIDOS"}

// TiposParcelamento lista os tipos aceitos, em ordem.
func TiposParcelamento() []string {
	var ts []string
	for t := range tiposAutomaticos {
		ts = append(ts, t)
	}
	sort.Strings(ts)
	return append(ts, TipoEspecializado)
}

// ValidarParcelamento confere o tipo e, no especializado, a negociação.
func ValidarParcelamento(tipo, negociacao string) error {
	switch {
	case tiposAutomaticos[tipo]:
		if negociacao != "" {
			return fmt.Errorf("--negociacao só vale para o tipo %s", TipoEspecializado)
		}
		return nil
	case tipo == TipoEspecializado:
		for _, n := range NegociacoesEspecializadas {
			if n == negociacao {
				return nil
			}
		}
		return fmt.Errorf("o tipo %s exige --negociacao %s", TipoEspecializado, strings.Join(NegociacoesEspecializadas, "|"))
	}
	return fmt.Errorf("tipo de parcelamento desconhecido %q: use %s", tipo, strings.Join(TiposParcelamento(), ", "))
}

// PathSimulacaoParcelamento é a simulação (GET, sem efeito) de um tipo já validado.
// Ver docs/api/escrita/impostos-e-pagamentos.md, seção 1.5.
func PathSimulacaoParcelamento(tipo, negociacao string) string {
	if tipo == TipoEspecializado {
		return "impostos/parcelamento/negociacao-especializada/init?tipoNegociacao=" + url.QueryEscape(negociacao)
	}
	return "impostos/parcelamento/negociacao-automatica/" + tipo + "/init"
}

// OpcaoParcelas é uma opção de quantidade de parcelas.
type OpcaoParcelas struct {
	Quantidade          int      `json:"quantidade"`
	ValorEntrada        *float64 `json:"valorEntrada"`
	ValorDemaisParcelas *float64 `json:"valorDemaisParcelas"`
}

// SimulacaoParcelamento é a resposta da simulação. O formato não foi verificado com uma
// conta com dívidas (ADR-0021): os campos são todos opcionais.
type SimulacaoParcelamento struct {
	Parcelas               []OpcaoParcelas `json:"parcelas" contract:"optional"`
	CobrarServicoAdicional bool            `json:"cobrarServicoAdicional" contract:"optional"`
	ValorServicoAdicional  *float64        `json:"valorServicoAdicional" contract:"optional"`
	ValorEmissaoGuia       *float64        `json:"valorEmissaoGuia" contract:"optional"`
	TaxaReparcelamento     json.RawMessage `json:"taxaReparcelamento" contract:"optional"`
	Negociacao             *struct {
		Impostos []json.RawMessage `json:"impostos"`
	} `json:"negociacao" contract:"optional"`
}

// SimularParcelamento lê a simulação. Sem dívidas, a API responde HTTP 560
// ("Empresa não possui guias para simulação").
func SimularParcelamento(ctx context.Context, g Getter, tipo, negociacao string) (*SimulacaoParcelamento, error) {
	return get[SimulacaoParcelamento](ctx, g, PathSimulacaoParcelamento(tipo, negociacao))
}
