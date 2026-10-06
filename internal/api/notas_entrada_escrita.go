package api

import "context"

// Manifestação do destinatário das NF-e de compra (base /api/emissor/). Ver
// docs/api/escrita/notas-e-prolabore.md, seção 3.1.
const PathManifestarNotas = "/api/emissor/notasentrada/manifestar/"

// Tipos de manifestação.
const (
	ManifestacaoCiencia         = "CIENCIA"
	ManifestacaoConfirmacao     = "CONFIRMACAO"
	ManifestacaoDesconhecimento = "DESCONHECIMENTO"
	ManifestacaoNaoRealizada    = "NAO_REALIZADO"
)

// Manifestacao é o corpo de PathManifestarNotas.
type Manifestacao struct {
	TipoManifestacao string `json:"tipoManifestacao"`
	Origem           string `json:"origem"` // PLATAFORMA
	IDNotas          []any  `json:"idNotas"`
	Justificativa    string `json:"justificativa,omitempty"` // só em NAO_REALIZADO
}

// NotaManifestada é um item da resposta: a nota e a situação nova.
type NotaManifestada struct {
	ID       any `json:"id"`
	Situacao *struct {
		ID string `json:"id"`
	} `json:"situacao"`
}

// Manifestar envia a manifestação à SEFAZ (pela Contabilizei). Ciência é preliminar; os
// demais eventos são conclusivos e não têm desfazer.
func Manifestar(ctx context.Context, s Sender, m Manifestacao) ([]NotaManifestada, error) {
	if m.Origem == "" {
		m.Origem = "PLATAFORMA"
	}
	var r []NotaManifestada
	if err := s.Send(ctx, "POST", PathManifestarNotas, m, &r); err != nil {
		return nil, err
	}
	return r, nil
}
