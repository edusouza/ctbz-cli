package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Origens das classificações em contas-usuario/{ano}/{mes}.
const (
	OrigemCaixa   = "CAIXA"
	OrigemExtrato = "EXTRATO"
)

// PathContasUsuarioCompetencia são as classificações aceitas numa competência para uma origem
// (CAIXA ou EXTRATO). Fonte: docs/api/escrita/contabilidade-e-documentos.md, seções 1.1 e 1.3.
func PathContasUsuarioCompetencia(ano, mes int, origem string) string {
	return fmt.Sprintf("movimentacao-financeira/contas-usuario/%d/%d?origem=%s", ano, mes, origem)
}

// ContaUsuarioCompetencia diz se uma classificação (ContaUsuario) aparece numa competência.
type ContaUsuarioCompetencia struct {
	ID            int64  `json:"id"`
	Exibir        bool   `json:"exibir"`
	Situacao      string `json:"situacao"`      // ATIVO ou INATIVO
	Classificacao string `json:"classificacao"` // RECEITA, DESPESA…
	Descricao     string `json:"descricao" contract:"optional"`
}

// BuscarContasUsuarioCompetencia lê as classificações de uma competência e origem.
func BuscarContasUsuarioCompetencia(ctx context.Context, g Getter, ano, mes int, origem string) ([]ContaUsuarioCompetencia, error) {
	c, err := get[[]ContaUsuarioCompetencia](ctx, g, PathContasUsuarioCompetencia(ano, mes, origem))
	if err != nil {
		return nil, err
	}
	return *c, nil
}

// Tipos de CategoriaVinculo.
const (
	VinculoContaUsuario = "CONTA_USUARIO"
	VinculoGuiaImposto  = "GUIA_IMPOSTO"
	VinculoSocio        = "SOCIO"
)

// CategoriaVinculo é uma opção do modal de lançamento do caixa: uma classificação
// (CONTA_USUARIO) ou um vínculo exigido por ela (GUIA_IMPOSTO, SOCIO).
type CategoriaVinculo struct {
	// ID é guardado como veio (número ou texto), para ser reenviado igual no idVinculo.
	ID        json.RawMessage `json:"id"`
	Descricao string          `json:"descricao"`
	Tipo      string          `json:"tipo"`
	Entrada   bool            `json:"entrada" contract:"optional"` // só em CONTA_USUARIO
}

// IDTexto devolve o id sem aspas, para comparar com o que o usuário digitou.
func (c CategoriaVinculo) IDTexto() string {
	return strings.Trim(string(c.ID), `"`)
}

// CaixaCategorias é a parte de caixa/listpaginada com as opções do modal de lançamento.
type CaixaCategorias struct {
	SerializedList struct {
		CategoriasVinculoDTO []CategoriaVinculo `json:"categoriasVinculoDTO"`
	} `json:"serializedList"`
}

// BuscarCaixaCategorias lê as classificações e vínculos do caixa de uma competência.
func BuscarCaixaCategorias(ctx context.Context, g Getter, ano, mes int) ([]CategoriaVinculo, error) {
	c, err := get[CaixaCategorias](ctx, g, PathCaixa(ano, mes))
	if err != nil {
		return nil, err
	}
	return c.SerializedList.CategoriasVinculoDTO, nil
}

// IDContaNaoPermitidaPagamento é a classificação que o front tira das opções de pagamento
// do caixa (categoriasIdNaoPermitidas).
const IDContaNaoPermitidaPagamento int64 = 5981343255101440

// IDsContasDeSocio são as classificações do extrato desdobradas por sócio: exigem idSocio.
var IDsContasDeSocio = map[int64]bool{
	6199733752168448: true,
	5491304964816896: true,
	4975757608615936: true,
	5605376636485632: true,
}

// DescricaoDistribuicaoAntecipada é a classificação do caixa que exige um sócio no idVinculo.
const DescricaoDistribuicaoAntecipada = "Sócios - Distribuição de Lucros Antecipados"

// impostosComGuia são as classificações do caixa que exigem uma guia (ou SEM GUIA) no idVinculo.
var impostosComGuia = []string{"COFINS", "CSLL", "FGTS", "INSS", "IRPJ", "IRRF", "ISS", "Simples Nacional"}

// VinculoExigido diz que vínculo uma classificação do caixa exige pela descrição:
// VinculoGuiaImposto, VinculoSocio ou "" (nenhum).
func VinculoExigido(descricao string) string {
	if descricao == DescricaoDistribuicaoAntecipada {
		return VinculoSocio
	}
	for _, imposto := range impostosComGuia {
		if descricao == "Impostos - "+imposto {
			return VinculoGuiaImposto
		}
	}
	return ""
}
