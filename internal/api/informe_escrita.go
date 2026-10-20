package api

import (
	"context"
	"fmt"
)

// Aceites do informe de rendimentos de um ano-calendário. Ver
// docs/api/escrita/pendencias-e-conta.md, seção 1.2.

// PathCartaResponsabilidadeInforme é a carta de responsabilidade do informe de um ano.
func PathCartaResponsabilidadeInforme(ano int) string {
	return fmt.Sprintf("informerendimento/carta-responsabilidade/%d", ano)
}

// CartaResponsabilidadeInforme diz se a carta precisa ser aceita e traz o texto em HTML.
// Campos citados no front; sem captura real (ADR-0021).
type CartaResponsabilidadeInforme struct {
	DeveAssinarCartaResponsabilidade bool   `json:"deveAssinarCartaResponsabilidade" contract:"optional"`
	HTML                             string `json:"html" contract:"optional"`
}

// BuscarCartaResponsabilidadeInforme lê a carta de responsabilidade do informe de um ano.
func BuscarCartaResponsabilidadeInforme(ctx context.Context, g Getter, ano int) (*CartaResponsabilidadeInforme, error) {
	return get[CartaResponsabilidadeInforme](ctx, g, PathCartaResponsabilidadeInforme(ano))
}

// AceitarCartaResponsabilidadeInforme aceita a carta do ano, exigida para ver o informe.
func AceitarCartaResponsabilidadeInforme(ctx context.Context, s Sender, ano int) error {
	return s.Send(ctx, "POST", fmt.Sprintf("informerendimento/aceitar-carta-responsabilidade/%d", ano), nil, nil)
}

// Em que apresentação do termo de débitos federais o aceite aconteceu.
const (
	AceiteTermoPrimeiraVez = "PRIMEIRA_VEZ"
	AceiteTermoDepoisNegar = "NEGOU_DA_PRIMEIRA_VEZ"
)

// AceitarTermoDebitos registra a ciência do termo de débitos federais do ano.
func AceitarTermoDebitos(ctx context.Context, s Sender, ano int, aceite string) error {
	return s.Send(ctx, "POST", fmt.Sprintf("informerendimento/aceitar-termo-debitos/%d", ano), struct {
		Aceite string `json:"aceite"`
	}{aceite}, nil)
}

// Decisões diante das restrições do informe.
const (
	AceiteNaoDistribuirLucros       = "IR_NAO_DISTRIBUIR_LUCROS"
	AceiteRegularizarPendencia      = "REGULARIZAR_PENDENCIA_DOCUMENTAL"
	AceiteNaoRegularizarPendencia   = "NAO_REGULARIZAR_PENDENCIA_DOCUMENTAL"
	FormularioRegularizarPendencias = "https://suporte.contabilizei.com.br/hc/pt-br/requests/new?ticket_form_id=360000562139"
)

// RegistrarAceiteInforme registra uma decisão (Aceite*) para o informe do ano.
func RegistrarAceiteInforme(ctx context.Context, s Sender, ano int, tipoAceite string) error {
	return s.Send(ctx, "POST", fmt.Sprintf("informerendimento/aceite/%d", ano), struct {
		TipoAceite string `json:"tipoAceite"`
	}{tipoAceite}, nil)
}

// Fluxos de regularização da pendência documental do informe.
const (
	FluxoReaberturaBalanco         = "REABERTURA_BALANCO"
	FluxoContratarServicoAdicional = "CONTRATAR_SERVICO_ADICIONAL"
)

// ReabrirBalanco reabre o exercício contábil do ano para regularizar as pendências
// documentais do informe; o informe fica indisponível até a análise. Sem desfazer pela API;
// 403 para usuário administrador. Ver docs/api/escrita/contabilidade-e-documentos.md, 3.1.
func ReabrirBalanco(ctx context.Context, s Sender, ano int) error {
	return s.Send(ctx, "POST", fmt.Sprintf("informerendimento/reabrir-balanco/%d", ano), nil, nil)
}
