package api

import (
	"context"
	"encoding/json"
	"strings"
)

// PathChecklistInit é o checklist "Primeiros passos" do painel.
// Ver docs/api/escrita/pendencias-e-conta.md, seção 1.5.
const PathChecklistInit = "checklist-onboarding/aside/init"

// EtapaChecklist é uma tarefa do checklist.
type EtapaChecklist struct {
	Nome     string `json:"nome"`
	Situacao string `json:"situacao"`
}

// Concluida diz se a situação indica tarefa feita (os valores exatos não foram verificados).
func (e EtapaChecklist) Concluida() bool {
	s := strings.ToUpper(e.Situacao)
	return strings.HasPrefix(s, "CONCLU") || strings.HasPrefix(s, "FINALIZ") || s == "REALIZADA"
}

// ChecklistOnboarding é a resposta de PathChecklistInit (ADR-0021).
type ChecklistOnboarding struct {
	Etapas    []EtapaChecklist `json:"etapas"`
	Progresso json.RawMessage  `json:"progresso" contract:"optional"`
}

// Etapa procura uma etapa pelo nome.
func (c *ChecklistOnboarding) Etapa(nome string) (EtapaChecklist, bool) {
	for _, e := range c.Etapas {
		if e.Nome == nome {
			return e, true
		}
	}
	return EtapaChecklist{}, false
}

// BuscarChecklist lê as etapas do checklist de primeiros passos.
func BuscarChecklist(ctx context.Context, g Getter) (*ChecklistOnboarding, error) {
	return get[ChecklistOnboarding](ctx, g, PathChecklistInit)
}

// EtapaGerarProcuracao é a etapa do checklist que pede a procuração do e-CAC.
const EtapaGerarProcuracao = "GERAR_PROCURACAO_VIRTUAL"

// Declarações de procuração do e-CAC criada ("Já criei a procuração"), POST sem corpo:
// uma pela Central de Rotinas, outra pelo checklist de primeiros passos.
const (
	PathResolverProcuracaoEcac = "central-rotinas/resolver-pendencia-procuracao-ecac"
	PathCompartilharProcuracao = "checklist-onboarding/procuracoes/compartilhamentos"
)

// DeclararProcuracao envia a declaração pelo caminho da origem da pendência.
func DeclararProcuracao(ctx context.Context, s Sender, path string) error {
	return s.Send(ctx, "POST", path, nil, nil)
}
