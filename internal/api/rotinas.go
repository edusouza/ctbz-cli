package api

import (
	"context"
	"encoding/json"
	"strings"
)

const PathCentralRotinas = "dashboard/v2/central-rotinas"

// IndicadorPendencia é um item de pendenciasCriticas/outrasPendencias: só o indicador é
// comum a todos; os demais campos variam por tipo de pendência.
type IndicadorPendencia struct {
	PossuiPendencia bool `json:"possuiPendencia"`
}

// RotinaEmpresa é uma rotina mensal de responsabilidade da empresa.
type RotinaEmpresa struct {
	Tipo         string `json:"tipo"`  // ex.: IMPORTACAO_EXTRATO, IMPOSTO, VENCIMENTO_MENSALIDADE
	Prazo        string `json:"prazo"` // AAAA-MM-DD
	Status       string `json:"status"`
	Automatica   bool   `json:"automatica"`
	Propriedades *struct {
		TituloModal    string   `json:"tituloModal" contract:"optional"`
		MesReferencia  string   `json:"mesReferencia" contract:"optional"`
		ValorPagamento *float64 `json:"valorPagamento" contract:"optional"`
		// DocumentosPendentes são as pendências que a rotina pede para resolver (ex.: as de
		// reclassificação); formato dos itens incerto (ADR-0021).
		DocumentosPendentes []json.RawMessage `json:"documentosPendentes" contract:"optional"`
	} `json:"propriedades"`
}

// RotinaContabilizei é uma obrigação entregue pela Contabilizei (eSocial, DCTFWeb…).
type RotinaContabilizei struct {
	Titulo   string `json:"titulo"`
	Prazo    string `json:"prazo"` // AAAA-MM-DD
	Status   string `json:"status"`
	Conteudo struct {
		Sigla string `json:"sigla"`
	} `json:"conteudo"`
}

// CentralRotinas é a resposta de dashboard/v2/central-rotinas.
type CentralRotinas struct {
	Pendencias struct {
		Criticas map[string]IndicadorPendencia `json:"pendenciasCriticas"`
		Outras   map[string]IndicadorPendencia `json:"outrasPendencias"`
	} `json:"pendencias"`
	Rotinas             []RotinaEmpresa      `json:"rotinas"`
	RotinasContabilizei []RotinaContabilizei `json:"rotinasContabilizei"`
}

// BuscarCentralRotinas lê pendências críticas e rotinas do painel.
func BuscarCentralRotinas(ctx context.Context, g Getter) (*CentralRotinas, error) {
	return get[CentralRotinas](ctx, g, PathCentralRotinas)
}

// IDsPendentes extrai os ids de DocumentosPendentes: itens que são o próprio id (texto ou
// número) ou objetos com "id" ou "idPendencia".
func (r RotinaEmpresa) IDsPendentes() []string {
	if r.Propriedades == nil {
		return nil
	}
	var ids []string
	for _, raw := range r.Propriedades.DocumentosPendentes {
		var obj struct {
			ID          json.RawMessage `json:"id"`
			IDPendencia json.RawMessage `json:"idPendencia"`
		}
		if json.Unmarshal(raw, &obj) == nil {
			if len(obj.IDPendencia) > 0 {
				raw = obj.IDPendencia
			} else if len(obj.ID) > 0 {
				raw = obj.ID
			}
		}
		if id := strings.Trim(string(raw), `"`); id != "" && id != "null" && !strings.HasPrefix(id, "{") {
			ids = append(ids, id)
		}
	}
	return ids
}
