package api

import (
	"strings"
	"testing"
)

func TestNovaAlteracaoGestao(t *testing.T) {
	atual := GestaoSocio{TipoGerenciamentoProlabore: GerenciamentoInteligente, ValorMaximoProlabore: ptr(8157.41), SalarioMinimo: ptr(1518.0), QtdSocioGestaoInteligente: ptr(2), ElegivelNoMotor: true}
	a, err := NovaAlteracaoGestao(atual, GerenciamentoTetoINSS, nil, nil)
	if err != nil || *a.ValorProlabore != 8157.41 || a.SairGestaoInteligente || a.ModoEdicaoProlaboreMin {
		t.Errorf("teto com outro sócio na GI: %+v, %v", a, err)
	}
	atual.QtdSocioGestaoInteligente, atual.ValorProlaboreMinimo = ptr(1), ptr(2000.0)
	a, err = NovaAlteracaoGestao(atual, GerenciamentoSalarioMinimo, nil, nil)
	if err != nil || *a.ValorProlabore != 1518 || !a.SairGestaoInteligente || !a.ModoEdicaoProlaboreMin {
		t.Errorf("salário mínimo, último na GI: %+v, %v", a, err)
	}
	a, err = NovaAlteracaoGestao(atual, GerenciamentoInteligente, nil, nil)
	if err != nil || a.ValorProlabore != nil || a.ValorProlaboreMinimo != nil || a.SairGestaoInteligente {
		t.Errorf("inteligente sem piso: %+v, %v", a, err)
	}
	// #995: quem não está na gestão inteligente nunca tira a empresa dela, qualquer que seja a
	// quantidade (o único nela pode ser outro sócio).
	for _, qtd := range []int{0, 1} {
		fora := atual
		fora.TipoGerenciamentoProlabore, fora.QtdSocioGestaoInteligente = GerenciamentoSalarioMinimo, ptr(qtd)
		if a, err = NovaAlteracaoGestao(fora, GerenciamentoTetoINSS, nil, nil); err != nil || a.SairGestaoInteligente {
			t.Errorf("sócio fora da GI, qtd %d: %+v, %v", qtd, a, err)
		}
	}
	// Na gestão inteligente sairGestaoInteligente é sempre false: a quantidade não é usada.
	semQtd := atual
	semQtd.QtdSocioGestaoInteligente = nil
	if a, err = NovaAlteracaoGestao(semQtd, GerenciamentoInteligente, nil, nil); err != nil || a.SairGestaoInteligente {
		t.Errorf("inteligente sem quantidade na GI: %+v, %v", a, err)
	}

	for _, tc := range []struct {
		tipo          string
		valor, minimo *float64
		muda          func(*GestaoSocio)
		want          string
	}{
		{GerenciamentoPersonalizado, ptr(1000.0), nil, nil, "maior ou igual ao salário mínimo vigente: 1518.00"},
		{GerenciamentoPersonalizado, nil, nil, nil, "informe o valor"},
		{GerenciamentoTetoINSS, ptr(3000.0), nil, nil, "só é informado no pró-labore personalizado"},
		{GerenciamentoSalarioMinimo, nil, ptr(2000.0), nil, "só vale na gestão inteligente"},
		{GerenciamentoInteligente, nil, ptr(-1.0), nil, "deve ser positivo"},
		{GerenciamentoInteligente, nil, nil, func(g *GestaoSocio) { g.ElegivelNoMotor = false }, "não é elegível"},
		{GerenciamentoTetoINSS, nil, nil, func(g *GestaoSocio) { g.ValorMaximoProlabore = nil }, "teto do INSS"},
		{GerenciamentoTetoINSS, nil, nil, func(g *GestaoSocio) { g.QtdSocioGestaoInteligente = nil }, "quantos sócios estão na gestão inteligente"},
		{GerenciamentoSalarioMinimo, nil, nil, func(g *GestaoSocio) { g.QtdSocioGestaoInteligente = nil }, "quantos sócios estão na gestão inteligente"},
		{GerenciamentoPersonalizado, ptr(3000.0), nil, func(g *GestaoSocio) { g.QtdSocioGestaoInteligente = nil }, "quantos sócios estão na gestão inteligente"},
		// #995: sem o tipo atual não dá para saber se o sócio está na gestão inteligente.
		{GerenciamentoTetoINSS, nil, nil, func(g *GestaoSocio) { g.TipoGerenciamentoProlabore = "" }, "não informou o tipo de gestão atual"},
		{"OUTRO", nil, nil, nil, "desconhecido"},
	} {
		g := atual
		if tc.muda != nil {
			tc.muda(&g)
		}
		if _, err := NovaAlteracaoGestao(g, tc.tipo, tc.valor, tc.minimo); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: erro %v, quero %q", tc.tipo, err, tc.want)
		}
	}
}
