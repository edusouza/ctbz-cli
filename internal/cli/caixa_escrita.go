package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// lancamentoFlags são as flags que descrevem um lançamento do caixa (adicionar e editar).
type lancamentoFlags struct {
	data, valor, conta, descricao, guia, socio string
	recebimento, pagamento, semGuia            bool
}

func (f *lancamentoFlags) add(cmd *cobra.Command) {
	fl := cmd.Flags()
	fl.StringVar(&f.data, "data", "", "data do lançamento AAAA-MM-DD (não pode ser futura)")
	fl.StringVar(&f.valor, "valor", "", "valor em reais, sempre positivo (ex.: 150,25)")
	fl.BoolVar(&f.recebimento, "recebimento", false, "entrada de dinheiro no caixa")
	fl.BoolVar(&f.pagamento, "pagamento", false, "saída de dinheiro do caixa")
	fl.StringVar(&f.conta, "conta", "", "classificação: id ou descrição exata (ver ctbz caixa contas)")
	fl.StringVar(&f.descricao, "descricao", "", "descrição do lançamento")
	fl.StringVar(&f.guia, "guia", "", "guia de imposto vinculada (contas \"Impostos - …\"; ver --vinculos)")
	fl.BoolVar(&f.semGuia, "sem-guia", false, "classificação de imposto sem guia vinculada")
	fl.StringVar(&f.socio, "socio", "", "sócio vinculado (Distribuição de Lucros Antecipados; ver --vinculos)")
}

// caixaCompetencia reúne o que o modal de lançamento do caixa lê de uma competência.
type caixaCompetencia struct {
	mes  time.Time
	cats []api.CategoriaVinculo
	comp []api.ContaUsuarioCompetencia
}

func lerCaixaCompetencia(ctx context.Context, g api.Getter, mes time.Time) (*caixaCompetencia, error) {
	cats, err := api.BuscarCaixaCategorias(ctx, g, mes.Year(), int(mes.Month()))
	if err != nil {
		return nil, err
	}
	comp, err := api.BuscarContasUsuarioCompetencia(ctx, g, mes.Year(), int(mes.Month()), api.OrigemCaixa)
	if err != nil {
		return nil, err
	}
	return &caixaCompetencia{mes: mes, cats: cats, comp: comp}, nil
}

// conta resolve --conta (id ou descrição) entre as classificações aceitas do lado pedido.
func (c *caixaCompetencia) conta(valor string, entrada bool) (api.CategoriaVinculo, error) {
	for _, cat := range classificacoesCaixa(c.cats, c.comp, &entrada) {
		if cat.IDTexto() == valor || strings.EqualFold(cat.Descricao, valor) {
			return cat, nil
		}
	}
	return api.CategoriaVinculo{}, usageError{fmt.Errorf("classificação %q não aceita para %s em %s; veja ctbz caixa contas --competencia %s --%s",
		valor, ladoCaixa(entrada), c.mes.Format("01/2006"), c.mes.Format("2006-01"), ladoCaixa(entrada))}
}

// vinculo resolve o idVinculo exigido pela classificação (guia, SEM GUIA ou sócio).
func (c *caixaCompetencia) vinculo(conta api.CategoriaVinculo, f lancamentoFlags, atual json.RawMessage) (json.RawMessage, string, error) {
	exigido := api.VinculoExigido(conta.Descricao)
	escolher := func(tipo, id string) (json.RawMessage, string, error) {
		for _, v := range vinculosCaixa(c.cats, tipo) {
			if v.IDTexto() == id {
				return v.ID, v.Descricao, nil
			}
		}
		return nil, "", usageError{fmt.Errorf("%s %q não encontrado na competência; veja ctbz caixa contas --competencia %s --vinculos",
			nomeVinculo(tipo), id, c.mes.Format("2006-01"))}
	}
	switch exigido {
	case api.VinculoGuiaImposto:
		if f.socio != "" {
			return nil, "", usageError{fmt.Errorf("%q exige guia, não sócio", conta.Descricao)}
		}
		switch {
		case f.semGuia && f.guia != "":
			return nil, "", usageError{errors.New("use --guia ou --sem-guia, não os dois")}
		case f.semGuia:
			return api.SemGuia, "SEM GUIA", nil
		case f.guia != "":
			return escolher(api.VinculoGuiaImposto, f.guia)
		case atual != nil:
			return atual, "", nil
		}
		return nil, "", usageError{fmt.Errorf("%q exige --guia ID ou --sem-guia", conta.Descricao)}
	case api.VinculoSocio:
		if f.guia != "" || f.semGuia {
			return nil, "", usageError{fmt.Errorf("%q exige sócio, não guia", conta.Descricao)}
		}
		if f.socio != "" {
			return escolher(api.VinculoSocio, f.socio)
		}
		if atual != nil {
			return atual, "", nil
		}
		return nil, "", usageError{fmt.Errorf("%q exige --socio ID", conta.Descricao)}
	}
	if f.guia != "" || f.semGuia || f.socio != "" {
		return nil, "", usageError{fmt.Errorf("%q não aceita guia nem sócio", conta.Descricao)}
	}
	return nil, "", nil
}

// parseDiaLancamento lê --data e recusa datas futuras (em Brasília), como o front.
func parseDiaLancamento(s string) (time.Time, error) {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		return d, usageError{fmt.Errorf("data inválida %q: use AAAA-MM-DD", s)}
	}
	if d.After(api.DiaEmBrasilia(now())) {
		return d, usageError{fmt.Errorf("a data %s é futura", d.Format("02/01/2006"))}
	}
	return d, nil
}

// diaPadrao é o dia que o front sugere: hoje na competência atual, senão o dia 1.
func diaPadrao(mes time.Time) time.Time {
	hoje := api.DiaEmBrasilia(now())
	if hoje.Year() == mes.Year() && hoje.Month() == mes.Month() {
		return hoje
	}
	return time.Date(mes.Year(), mes.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func newCaixaAdicionarCmd() *cobra.Command {
	var f lancamentoFlags
	cmd := &cobra.Command{
		Use:   "adicionar",
		Short: "Adiciona um recebimento ou pagamento no caixa de uma competência",
		Long: `Registra um recebimento ou pagamento em dinheiro no caixa da competência (risco médio:
entra na contabilidade do mês; dá para editar ou remover depois).

O valor é sempre positivo: --recebimento ou --pagamento define o sinal. A classificação
(--conta) precisa estar entre as aceitas na competência (ctbz caixa contas). Contas
"Impostos - …" exigem --guia ou --sem-guia, e "Sócios - Distribuição de Lucros
Antecipados" exige --socio (ids em ctbz caixa contas --vinculos). Sem --data, usa o dia
que o painel sugere: hoje na competência atual, senão o dia 1.

Pede confirmação (--yes em scripts) e aceita --dry-run (ver docs/guia/escrita.md).`,
		Example: `  ctbz caixa adicionar --competencia 2026-09 --data 2026-09-15 --pagamento \
    --valor 150,25 --conta 1000000000000012 --descricao "Material de escritório"
  ctbz caixa adicionar --competencia 2026-09 --pagamento --valor 80 \
    --conta "Impostos - Simples Nacional" --sem-guia --descricao "DAS" --dry-run`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			mes, err := competenciaFlag(cmd)
			if err != nil {
				return err
			}
			if f.recebimento == f.pagamento {
				return usageError{errors.New("informe --recebimento ou --pagamento")}
			}
			if f.valor == "" || f.conta == "" || strings.TrimSpace(f.descricao) == "" {
				return usageError{errors.New("informe --valor, --conta e --descricao")}
			}
			valor, err := parseValor(f.valor)
			if err != nil {
				return usageError{err}
			}
			if valor <= 0 {
				return usageError{errors.New("o valor deve ser maior que zero; o sinal vem de --recebimento ou --pagamento")}
			}
			dia := diaPadrao(mes)
			if f.data != "" {
				if dia, err = parseDiaLancamento(f.data); err != nil {
					return err
				}
			}
			formato, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			cc, err := lerCaixaCompetencia(ctx, sessionGetter{s}, mes)
			if err != nil {
				return err
			}
			conta, err := cc.conta(f.conta, f.recebimento)
			if err != nil {
				return err
			}
			vinculo, nomeVinc, err := cc.vinculo(conta, f, nil)
			if err != nil {
				return err
			}
			if f.pagamento {
				valor = -valor
			}
			idConta, _ := strconv.ParseInt(conta.IDTexto(), 10, 64)
			req := api.SalvarLancamentoCaixa{Ano: strconv.Itoa(mes.Year()), Mes: int(mes.Month()), LancamentoUsuario: api.LancamentoCaixaUsuario{
				Data: api.DataISOBrasilia(dia), Descricao: strings.TrimSpace(f.descricao), IDContaUsuario: idConta, IDVinculo: vinculo, Valor: reais(valor),
			}}
			op := operacao{Risco: riscoMedio, Resumo: fmt.Sprintf("Adicionar %s de %s em %s no caixa de %s: %q (%s%s)",
				ladoCaixa(f.recebimento), formatarCentavos(abs(valor)), dia.Format("02/01/2006"), mes.Format("01/2006"),
				req.LancamentoUsuario.Descricao, conta.Descricao, sufixoVinculo(nomeVinc))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.SalvarLancamento(ctx, snd, req) })
			if err != nil || !enviado {
				return err
			}
			l, err := releLancamento(ctx, sessionGetter{s}, mes, func(lc api.LancamentoCaixa) bool {
				return lc.Descricao == req.LancamentoUsuario.Descricao && lc.Valor != nil && centavos(*lc.Valor) == valor &&
					api.DiaEmBrasilia(time.UnixMilli(lc.Data)).Equal(dia)
			})
			if err != nil {
				fmt.Fprintln(s.err, "aviso: lançamento salvo, mas não foi possível reler o caixa:", err)
			}
			return output.Write(s.out, formato, lancamentoResultado("adicionar", "adicionado", l, req.LancamentoUsuario, conta.Descricao))
		},
	}
	addCompetenciaFlag(cmd)
	f.add(cmd)
	addWriteFlags(cmd)
	return cmd
}

func abs(c int64) int64 {
	if c < 0 {
		return -c
	}
	return c
}

func sufixoVinculo(nome string) string {
	if nome == "" {
		return ""
	}
	return "; vínculo: " + nome
}

// releLancamento relê o caixa da competência e devolve o lançamento mais recente (maior id)
// que satisfaz achar, ou nil.
func releLancamento(ctx context.Context, g api.Getter, mes time.Time, achar func(api.LancamentoCaixa) bool) (*api.LancamentoCaixa, error) {
	c, err := api.BuscarCaixa(ctx, g, mes.Year(), int(mes.Month()))
	if err != nil {
		return nil, err
	}
	var achado *api.LancamentoCaixa
	for i, lc := range c.List {
		if achar(lc) && (achado == nil || lc.ID > achado.ID) {
			achado = &c.List[i]
		}
	}
	return achado, nil
}

// lancamentoResultado mostra o lançamento salvo: o relido quando encontrado, senão o enviado.
func lancamentoResultado(acao, situacao string, relido *api.LancamentoCaixa, enviado api.LancamentoCaixaUsuario, conta string) *output.Record {
	var id any
	if enviado.ID != nil {
		id = *enviado.ID
	}
	data := enviado.Data[:10]
	valor := output.Money(enviado.Valor)
	descricao := enviado.Descricao
	if relido != nil {
		id = relido.ID
		data = api.DiaEmBrasilia(time.UnixMilli(relido.Data)).Format("2006-01-02")
		if relido.Valor != nil {
			valor = output.Money(*relido.Valor)
		}
		descricao = relido.Descricao
	} else {
		situacao = "enviado"
	}
	d, _ := output.ParseDate(data)
	return resultadoEscrita(acao, situacao, id).
		Add("data", "Data", d).
		Add("descricao", "Descrição", descricao).
		Add("conta", "Conta", conta).
		Add("valor", "Valor", valor)
}
