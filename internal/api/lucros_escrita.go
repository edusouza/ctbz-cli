package api

import (
	"context"
	"fmt"
)

// Distribuição de lucros do exercício aberto. Ver docs/api/escrita/notas-e-prolabore.md,
// seção 5.
const PathSalvarDistribuicao = "informerendimento/salvarconfiguracaocliente"

// DistribuicaoSocio é um item do corpo: percentual e valor vão como texto com duas casas,
// como o front manda.
type DistribuicaoSocio struct {
	IDSocio    any    `json:"idSocio"`
	Percentual string `json:"percentual"`
	Valor      string `json:"valor"`
}

// NovaDistribuicaoSocio monta um item a partir do valor em centavos e do percentual em
// centésimos de ponto (6000 = 60,00%).
func NovaDistribuicaoSocio(idSocio any, centavos, centesimos int64) DistribuicaoSocio {
	return DistribuicaoSocio{IDSocio: idSocio, Percentual: duasCasas(centesimos), Valor: duasCasas(centavos)}
}

func duasCasas(v int64) string {
	sinal := ""
	if v < 0 {
		sinal, v = "-", -v
	}
	return fmt.Sprintf("%s%d.%02d", sinal, v/100, v%100)
}

// SalvarDistribuicaoLucros registra a distribuição do exercício aberto (o ano não vai no
// corpo: é o de BuscarDistribuicaoLucros). Um item por sócio; a soma deve ser o lucro total.
func SalvarDistribuicaoLucros(ctx context.Context, s Sender, itens []DistribuicaoSocio) error {
	return s.Send(ctx, "POST", PathSalvarDistribuicao, itens, nil)
}
