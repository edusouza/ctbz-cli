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

// lancamentoDoCaixa acha um lançamento pelo id na competência e recusa os lançados pelo
// sistema, que o painel não deixa editar nem excluir.
func lancamentoDoCaixa(ctx context.Context, g api.Getter, mes time.Time, idArg string) (*api.LancamentoCaixa, error) {
	id, err := strconv.ParseInt(idArg, 10, 64)
	if err != nil {
		return nil, usageError{fmt.Errorf("id de lançamento inválido %q", idArg)}
	}
	c, err := api.BuscarCaixa(ctx, g, mes.Year(), int(mes.Month()))
	if err != nil {
		return nil, err
	}
	for i := range c.List {
		if c.List[i].ID == id {
			if c.List[i].ConfirmadoViaSistema {
				return nil, fmt.Errorf("o lançamento %d foi feito pelo sistema e não pode ser alterado (o painel também não permite)", id)
			}
			return &c.List[i], nil
		}
	}
	return nil, fmt.Errorf("lançamento %d não encontrado no caixa de %s", id, mes.Format("01/2006"))
}

// mudanca descreve uma alteração para o resumo ("valor R$ 1,00 → R$ 2,00").
func mudanca(campo, antes, depois string) string {
	return fmt.Sprintf("%s %s → %s", campo, antes, depois)
}

func newCaixaEditarCmd() *cobra.Command {
	var f lancamentoFlags
	cmd := &cobra.Command{
		Use:   "editar ID",
		Short: "Altera um lançamento manual do caixa",
		Long: `Altera data, valor, tipo (recebimento ou pagamento), classificação, vínculo ou descrição de
um lançamento manual do caixa (risco médio, reversível editando de novo). Só as flags
informadas mudam; o resto é mantido.

O painel não tem um "editar" separado: reenvia o lançamento inteiro. A CLI lê o lançamento
atual na competência, aplica as mudanças com as mesmas validações de ctbz caixa adicionar e
mostra o antes e o depois no resumo. Lançamentos feitos pelo sistema não podem ser editados.

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz caixa editar 1000000000000002 --competencia 2026-09 --valor 900
  ctbz caixa editar 1000000000000002 --competencia 2026-09 --descricao "Venda balcão" --dry-run`,
		Args: exactArgs(1, "o ID do lançamento"),
		RunE: func(cmd *cobra.Command, args []string) error {
			mes, err := competenciaFlag(cmd)
			if err != nil {
				return err
			}
			if f.recebimento && f.pagamento {
				return usageError{errors.New("use --recebimento ou --pagamento, não os dois")}
			}
			alteracoes := []string{"data", "valor", "recebimento", "pagamento", "conta", "descricao", "guia", "sem-guia", "socio"}
			mudou := false
			for _, nome := range alteracoes {
				mudou = mudou || cmd.Flags().Changed(nome)
			}
			if !mudou {
				return usageError{errors.New("informe o que alterar: --data, --valor, --recebimento/--pagamento, --conta, --descricao, --guia/--sem-guia ou --socio")}
			}
			formato, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			atual, err := lancamentoDoCaixa(ctx, g, mes, args[0])
			if err != nil {
				return err
			}
			cc, err := lerCaixaCompetencia(ctx, g, mes)
			if err != nil {
				return err
			}
			req, conta, resumo, err := aplicarEdicao(cc, atual, f)
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoMedio, ID: args[0], Resumo: fmt.Sprintf("Editar o lançamento %d do caixa de %s: %s",
				atual.ID, mes.Format("01/2006"), strings.Join(resumo, "; "))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.SalvarLancamento(ctx, snd, req) })
			if err != nil || !enviado {
				return err
			}
			l, err := releLancamento(ctx, g, mes, func(lc api.LancamentoCaixa) bool { return lc.ID == atual.ID })
			if err != nil {
				fmt.Fprintln(s.err, "aviso: lançamento salvo, mas não foi possível reler o caixa:", err)
			}
			return output.Write(s.out, formato, lancamentoResultado("editar", "editado", l, req.LancamentoUsuario, conta))
		},
	}
	addCompetenciaFlag(cmd)
	f.add(cmd)
	addWriteFlags(cmd)
	return cmd
}

// aplicarEdicao aplica as flags informadas ao lançamento atual e devolve a requisição, a
// classificação final e a lista de mudanças para o resumo.
func aplicarEdicao(cc *caixaCompetencia, atual *api.LancamentoCaixa, f lancamentoFlags) (api.SalvarLancamentoCaixa, string, []string, error) {
	var req api.SalvarLancamentoCaixa
	var resumo []string
	valorAtual := int64(0)
	if atual.Valor != nil {
		valorAtual = centavos(*atual.Valor)
	}
	entrada := valorAtual > 0
	if f.recebimento || f.pagamento {
		if f.recebimento != entrada {
			resumo = append(resumo, mudanca("tipo", ladoCaixa(entrada), ladoCaixa(f.recebimento)))
		}
		entrada = f.recebimento
	}
	valor := abs(valorAtual)
	if f.valor != "" {
		v, err := parseValor(f.valor)
		if err != nil {
			return req, "", nil, usageError{err}
		}
		if v <= 0 {
			return req, "", nil, usageError{errors.New("o valor deve ser maior que zero; o sinal vem de --recebimento ou --pagamento")}
		}
		if v != valor {
			resumo = append(resumo, mudanca("valor", formatarCentavos(valor), formatarCentavos(v)))
		}
		valor = v
	}
	if !entrada {
		valor = -valor
	}
	dia := api.DiaEmBrasilia(time.UnixMilli(atual.Data))
	if f.data != "" {
		d, err := parseDiaLancamento(f.data)
		if err != nil {
			return req, "", nil, err
		}
		if !d.Equal(dia) {
			resumo = append(resumo, mudanca("data", dia.Format("02/01/2006"), d.Format("02/01/2006")))
		}
		dia = d
	}
	idConta := ""
	if atual.IDContaUsuario != nil {
		idConta = strconv.FormatInt(*atual.IDContaUsuario, 10)
	}
	vinculoAtual := atual.IDVinculo
	if string(vinculoAtual) == "null" {
		vinculoAtual = nil
	}
	if f.conta != "" {
		vinculoAtual = nil // nova classificação: o vínculo antigo não vale
	}
	alvo := idConta
	if f.conta != "" {
		alvo = f.conta
	}
	if alvo == "" {
		return req, "", nil, usageError{errors.New("o lançamento não tem classificação; informe --conta")}
	}
	conta, err := cc.conta(alvo, entrada)
	if err != nil {
		return req, "", nil, err
	}
	if conta.IDTexto() != idConta {
		antes := idConta
		for _, c := range cc.cats {
			if c.IDTexto() == idConta {
				antes = c.Descricao
			}
		}
		resumo = append(resumo, mudanca("conta", antes, conta.Descricao))
	}
	vinculo, nomeVinc, err := cc.vinculo(conta, f, vinculoAtual)
	if err != nil {
		return req, "", nil, err
	}
	if nomeVinc != "" {
		resumo = append(resumo, "vínculo: "+nomeVinc)
	}
	descricao := atual.Descricao
	if strings.TrimSpace(f.descricao) != "" && strings.TrimSpace(f.descricao) != descricao {
		resumo = append(resumo, mudanca("descrição", strconv.Quote(descricao), strconv.Quote(strings.TrimSpace(f.descricao))))
		descricao = strings.TrimSpace(f.descricao)
	}
	if len(resumo) == 0 {
		return req, "", nil, usageError{errors.New("nada muda: os valores informados são iguais aos atuais")}
	}
	id := atual.ID
	idNum, _ := strconv.ParseInt(conta.IDTexto(), 10, 64)
	req = api.SalvarLancamentoCaixa{Ano: strconv.Itoa(cc.mes.Year()), Mes: int(cc.mes.Month()), LancamentoUsuario: api.LancamentoCaixaUsuario{
		Data: api.DataISOBrasilia(dia), Descricao: descricao, ID: &id, IDContaUsuario: idNum, IDVinculo: vinculo, Valor: reais(valor),
	}}
	return req, conta.Descricao, resumo, nil
}

func newCaixaRemoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remover ID",
		Short: "Exclui permanentemente um lançamento manual do caixa",
		Long: `Exclui um lançamento manual do caixa da competência (risco médio). A API não tem como
desfazer: para voltar atrás, é preciso recriar o lançamento. Por isso o resumo mostra
descrição, data e valor antes da confirmação, e a saída traz os dados do lançamento removido
e, no stderr, o comando ctbz caixa adicionar que o recria.

Lançamentos feitos pelo sistema não podem ser removidos, como no painel.
Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz caixa remover 1000000000000002 --competencia 2026-09
  ctbz caixa remover 1000000000000002 --competencia 2026-09 --yes -o json`,
		Args: exactArgs(1, "o ID do lançamento"),
		RunE: func(cmd *cobra.Command, args []string) error {
			mes, err := competenciaFlag(cmd)
			if err != nil {
				return err
			}
			formato, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			l, err := lancamentoDoCaixa(ctx, g, mes, args[0])
			if err != nil {
				return err
			}
			dia := api.DiaEmBrasilia(time.UnixMilli(l.Data))
			valor := int64(0)
			if l.Valor != nil {
				valor = centavos(*l.Valor)
			}
			op := operacao{Risco: riscoMedio, ID: args[0], Resumo: fmt.Sprintf("Excluir permanentemente o lançamento %d do caixa de %s: %q de %s no valor de %s",
				l.ID, mes.Format("01/2006"), l.Descricao, dia.Format("02/01/2006"), formatarCentavos(valor))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return api.RemoverLancamento(ctx, snd, mes.Year(), int(mes.Month()), l.ID)
			})
			if err != nil || !enviado {
				return err
			}
			situacao := "removido"
			if ainda, err := releLancamento(ctx, g, mes, func(lc api.LancamentoCaixa) bool { return lc.ID == l.ID }); err != nil {
				fmt.Fprintln(s.err, "aviso: lançamento removido, mas não foi possível reler o caixa:", err)
				situacao = "enviado"
			} else if ainda != nil {
				fmt.Fprintln(s.err, "aviso: o lançamento ainda aparece no caixa; confira com ctbz caixa", mes.Format("2006-01"))
				situacao = "enviado"
			}
			var idConta any
			if l.IDContaUsuario != nil {
				idConta = *l.IDContaUsuario
				lado := "recebimento"
				if valor < 0 {
					lado = "pagamento"
				}
				fmt.Fprintf(s.err, "Para recriar: ctbz caixa adicionar --competencia %s --data %s --%s --valor %s --conta %d --descricao %q\n",
					mes.Format("2006-01"), dia.Format("2006-01-02"), lado, strings.TrimPrefix(formatarCentavos(abs(valor)), "R$ "), *l.IDContaUsuario, l.Descricao)
			}
			d, _ := output.ParseDate(dia.Format("2006-01-02"))
			rec := resultadoEscrita("remover", situacao, l.ID).
				Add("data", "Data", d).
				Add("descricao", "Descrição", l.Descricao).
				Add("id_conta", "ID da conta", idConta).
				Add("valor", "Valor", output.Money(reais(valor)))
			return output.Write(s.out, formato, rec)
		},
	}
	addCompetenciaFlag(cmd)
	addWriteFlags(cmd)
	return cmd
}
