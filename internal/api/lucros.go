package api

import (
	"context"
	"fmt"
)

const PathDistribuicaoLucros = "informerendimento/recuperardadosdistribuicaocliente"

// PathRestricoesInforme são as restrições do informe de rendimentos de um ano (pendências
// documentais, débitos federais, reabertura do balanço).
func PathRestricoesInforme(ano int) string {
	return fmt.Sprintf("informerendimento/v2/%d/restricoes", ano)
}

// DistribuicaoLucros é a distribuição de lucros do exercício aberto. A API não recebe ano:
// o exercício vem em Ano (0 quando ainda não há exercício para distribuir).
type DistribuicaoLucros struct {
	Ano                               int      `json:"ano"`
	Saldo                             *float64 `json:"saldo"`
	TotalDistribuido                  *float64 `json:"totalDistribuido"`
	TotalAdiantamentos                *float64 `json:"totalAdiantamentos"`
	LimitePermitidoDistribuicaoLucros *float64 `json:"limitePermitidoDistribuicaoLucros"`
	ExercicioFechado                  bool     `json:"exercicioFechado"`
	PodeAlterar                       bool     `json:"podeAlterar"`
	MotivoNaoPodeAlterar              any      `json:"motivoNaoPodeAlterar"`
	DataLimite                        any      `json:"dataLimite"`
	// Um item por sócio; a conta verificada não tinha distribuição (ver LucroSocio).
	LucrosSocios []LucroSocio `json:"lucrosSocios"`
}

// LucroSocio é a parte de um sócio na distribuição. O front lê {socio, valor} na tela e
// {id, porcentagem, valor} ao salvar; os campos ficam opcionais até haver uma distribuição
// real (#147).
type LucroSocio struct {
	ID          any      `json:"id" contract:"optional"` // o idSocio de salvarconfiguracaocliente
	Socio       string   `json:"socio" contract:"optional"`
	Porcentagem any      `json:"porcentagem" contract:"optional"`
	Valor       *float64 `json:"valor" contract:"optional"`
}

// BuscarDistribuicaoLucros lê a distribuição de lucros.
func BuscarDistribuicaoLucros(ctx context.Context, g Getter) (*DistribuicaoLucros, error) {
	return get[DistribuicaoLucros](ctx, g, PathDistribuicaoLucros)
}

// RestricoesInforme indica o que impede o informe de rendimentos de um ano.
type RestricoesInforme struct {
	Restricoes struct {
		PendenciaDocumental struct {
			PossuiPendencia    bool   `json:"possuiPendencia"`
			FluxoRegularizacao string `json:"fluxoRegularizacao"`
		} `json:"pendenciaDocumental"`
		DebitosFederais struct {
			PossuiPendencia           bool `json:"possuiPendencia"`
			DivergenciaContabilFiscal bool `json:"divergenciaContabilFiscal"`
		} `json:"debitosFederais"`
	} `json:"restricoes"`
	ProcessoReabertura struct {
		Status string `json:"status"` // ex.: NENHUM, EM_ANDAMENTO
	} `json:"processoReabertura"`
}

// BuscarRestricoesInforme lê as restrições do informe de um ano.
func BuscarRestricoesInforme(ctx context.Context, g Getter, ano int) (*RestricoesInforme, error) {
	return get[RestricoesInforme](ctx, g, PathRestricoesInforme(ano))
}

// PathSociosInforme lista os sócios com informe de rendimentos no ano ([{id, nome}]).
func PathSociosInforme(ano int) string {
	return fmt.Sprintf("informerendimento/listSocioInformeRendimentos/%d", ano)
}

// PathComprovanteRendimentos são os valores do comprovante de rendimentos de um sócio. O
// painel monta o PDF no navegador com esses valores; a API não devolve PDF.
func PathComprovanteRendimentos(idSocio any, ano int) string {
	return fmt.Sprintf("informerendimento/getValoresComprovanteRendimento/%v/%d", idSocio, ano)
}

// SocioInforme é um sócio com informe de rendimentos. Campos lidos pelo front; a conta
// verificada não tinha informes.
type SocioInforme struct {
	ID   any    `json:"id" contract:"optional"` // número ou texto
	Nome string `json:"nome" contract:"optional"`
}

// BuscarSociosInforme lista os sócios com informe no ano.
func BuscarSociosInforme(ctx context.Context, g Getter, ano int) ([]SocioInforme, error) {
	s, err := get[[]SocioInforme](ctx, g, PathSociosInforme(ano))
	if err != nil {
		return nil, err
	}
	return *s, nil
}

// ComprovanteRendimentos são os valores do comprovante de rendimentos (DIRF) de um sócio.
// Formato lido do front, sem exemplo real: os valores ficam como any (número esperado).
type ComprovanteRendimentos struct {
	Nome               string `json:"nome"`
	CPF                string `json:"cpf"`
	Rendimentos        any    `json:"rendimentos"`        // pró-labore tributável
	Previdencia        any    `json:"previdencia"`        // INSS
	IRRFRetido         any    `json:"irrfRetido"`         // IRRF sobre o pró-labore
	DecimoTerceiro     any    `json:"decimoTerceiro"`     // 13º
	IRRFDecimoTerceiro any    `json:"irrfdecimoTerceiro"` // IRRF sobre o 13º
	Lucro              any    `json:"lucro"`              // lucros e dividendos (isentos)
}

// BuscarComprovanteRendimentos lê os valores do comprovante de um sócio.
func BuscarComprovanteRendimentos(ctx context.Context, g Getter, idSocio any, ano int) (*ComprovanteRendimentos, error) {
	return get[ComprovanteRendimentos](ctx, g, PathComprovanteRendimentos(idSocio, ano))
}
