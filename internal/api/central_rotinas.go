package api

import "context"

// PathCentralRotinasInit é a Central de Rotinas usada pelos aceites e pela procuração do
// e-CAC (diferente de dashboard/v2/central-rotinas). Ver docs/api/escrita/pendencias-e-conta.md, 1.1.
const PathCentralRotinasInit = "central-rotinas/init"

// PendenciaCentral é uma pendência da Central de Rotinas. Só possuiPendencia é comum; os
// demais campos dependem da chave (ADR-0021).
type PendenciaCentral struct {
	PossuiPendencia               bool   `json:"possuiPendencia"`
	ConteudoCartaResponsabilidade string `json:"conteudoCartaResponsabilidade" contract:"optional"`
	ConteudoAceiteTermoDebito     string `json:"conteudoAceiteTermoDebito" contract:"optional"`
	PrazoAceiteTacito             any    `json:"prazoAceiteTacito" contract:"optional"` // dias
	ConteudoTermoAdesaoTotalPass  string `json:"conteudoTermoAdesaoTotalPass" contract:"optional"`
	TipoPendencia                 string `json:"tipoPendencia" contract:"optional"` // procuração: SEM_PROCURACAO, EM_EXPIRACAO…
	CnpjOutorgado                 string `json:"cnpjOutorgado" contract:"optional"`
}

// CentralRotinasInit é a resposta de PathCentralRotinasInit.
type CentralRotinasInit struct {
	Pendencias struct {
		Criticas map[string]PendenciaCentral `json:"pendenciasCriticas"`
		Outras   map[string]PendenciaCentral `json:"outrasPendencias"`
	} `json:"pendencias"`
}

// Pendencia procura uma pendência pela chave nas críticas e nas outras.
func (c *CentralRotinasInit) Pendencia(chave string) (PendenciaCentral, bool) {
	if p, ok := c.Pendencias.Criticas[chave]; ok {
		return p, true
	}
	p, ok := c.Pendencias.Outras[chave]
	return p, ok
}

// BuscarCentralRotinasInit lê as pendências da Central de Rotinas com o conteúdo dos termos.
func BuscarCentralRotinasInit(ctx context.Context, g Getter) (*CentralRotinasInit, error) {
	return get[CentralRotinasInit](ctx, g, PathCentralRotinasInit)
}

// Chaves das pendências com aceite ou declaração.
const (
	PendenciaCartaResponsabilidade = "pendenciaCartaResponsabilidade"
	PendenciaTermoDebitos          = "pendenciaAceiteTermoDebitos"
	PendenciaTermoTotalPass        = "pendenciaTermoAdesaoTotalPass"
	PendenciaProcuracaoEcac        = "pendenciaProcuracaoEcac"
)
