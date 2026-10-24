package api

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// senderContador conta as escritas, sem enviar nada.
type senderContador struct{ n int }

func (s *senderContador) Send(context.Context, string, string, any, any) error {
	s.n++
	return nil
}

func (s *senderContador) SendMultipart(context.Context, string, string, []Campo, any) error {
	s.n++
	return nil
}

// Competência ausente ou inválida é erro, nunca {"mes":0,"ano":0} (ADR-0021, #628, #652).
func TestEnvioDocumentoCompetenciaInvalida(t *testing.T) {
	ctx := context.Background()
	for nome, c := range map[string]*CompetenciaDocumento{
		"ausente": nil, "zerada": {}, "sem ano": {Mes: 8}, "mês 13": {Mes: 13, Ano: 2026}, "mês zero": {Ano: 2026},
	} {
		d := exemploDocumento("1000000000000602", 9)
		d.Propriedades.Competencia = c
		s := &senderContador{}
		if _, _, err := d.MesAno(); err == nil || !strings.Contains(err.Error(), "1000000000000602") {
			t.Errorf("%s: MesAno: %v", nome, err)
		}
		if err := EnviarDocumento(ctx, s, d, "a.pdf", nil); err == nil || !strings.Contains(err.Error(), "1000000000000602") {
			t.Errorf("%s: EnviarDocumento: %v", nome, err)
		}
		if err := EnviarDocumentoConsolidado(ctx, s, []DocumentoPendente{exemploDocumento("1000000000000601", 8), d}, "a.pdf", nil); err == nil ||
			!strings.Contains(err.Error(), "1000000000000602") {
			t.Errorf("%s: EnviarDocumentoConsolidado: %v", nome, err)
		}
		mes, ano := 0, 0
		if c != nil {
			mes, ano = c.Mes, c.Ano
		}
		if _, err := DeclararSemAplicacao(ctx, s, DeclaracaoSemAplicacao{TipoDocumento: TipoExtratoAplicacao, TipoPendencia: TipoExtratoAplicacao,
			IDContaBancaria: json.RawMessage("7"),
			CompetenciasPendentes: []CompetenciaPendente{{ID: json.RawMessage(`"1000000000000601"`), Mes: 8, Ano: 2026},
				{ID: d.Propriedades.IDPendencia, Mes: mes, Ano: ano}}}); err == nil || !strings.Contains(err.Error(), "1000000000000602") {
			t.Errorf("%s: DeclararSemAplicacao: %v", nome, err)
		}
		if s.n != 0 {
			t.Errorf("%s: %d escrita(s) com competência inválida", nome, s.n)
		}
	}
	if mes, ano, err := exemploDocumento("1000000000000601", 8).MesAno(); err != nil || mes != 8 || ano != 2026 {
		t.Errorf("competência válida: %02d/%d, %v", mes, ano, err)
	}
}
