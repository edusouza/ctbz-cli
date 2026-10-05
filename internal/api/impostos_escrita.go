package api

import (
	"context"
	"fmt"
)

// PathRollout diz qual geração da tela de impostos a empresa usa (v3, v4 ou v5).
const PathRollout = "impostos/rollout"

// Rollout é a resposta de PathRollout.
type Rollout struct {
	Versao string `json:"versao"`
}

// BuscarRollout lê a versão da tela de impostos.
func BuscarRollout(ctx context.Context, g Getter) (*Rollout, error) {
	return get[Rollout](ctx, g, PathRollout)
}

// VersaoConfirmacao escolhe a rota de confirmação de pagamento como o front: a v5 só com
// rollout v5; as demais telas usam a v3.
func VersaoConfirmacao(rollout string) string {
	if rollout == "v5" {
		return "v5"
	}
	return "v3"
}

// PathConfirmarPagamento confirma ou nega o pagamento de uma guia (PUT). versao é "v5" ou "v3"
// (VersaoConfirmacao). Ver docs/api/escrita/impostos-e-pagamentos.md, seção 1.1.
func PathConfirmarPagamento(versao string, idGuia int64) string {
	return fmt.Sprintf("impostos/%s/impostos-a-pagar/guia/%d/confirmar-pagamento", versao, idGuia)
}

// ConfirmacaoPagamento é o corpo de PathConfirmarPagamento. Tipo e Origem vêm da guia.
type ConfirmacaoPagamento struct {
	Tipo                string `json:"tipo"`
	Origem              string `json:"origem"`
	PagamentoConfirmado bool   `json:"pagamentoConfirmado"`
}

// OrigemGuiaPadrao é a origem que o drawer v5 usa quando a guia não informa.
const OrigemGuiaPadrao = "GUIAS"

// RespostaConfirmacao junta o que as versões devolvem: a v5 só diz para onde a guia foi.
type RespostaConfirmacao struct {
	MovidaPara string `json:"movidaPara"`
}

// ConfirmarPagamento declara que a guia foi paga (true) ou não (false). Não paga nada.
func ConfirmarPagamento(ctx context.Context, s Sender, versao string, idGuia int64, req ConfirmacaoPagamento) (*RespostaConfirmacao, error) {
	var r RespostaConfirmacao
	if err := s.Send(ctx, "PUT", PathConfirmarPagamento(versao, idGuia), req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
