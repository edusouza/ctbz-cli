package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// addContaBancariaFlag acrescenta --conta-bancaria (obrigatória).
func addContaBancariaFlag(cmd *cobra.Command) {
	cmd.Flags().Int64("conta-bancaria", 0, "id da conta bancária (ver ctbz contas-bancarias) (obrigatória)")
}

// extratoAlvo identifica o extrato de uma conta numa competência.
type extratoAlvo struct {
	conta int64
	mes   time.Time
}

func extratoFlags(cmd *cobra.Command) (extratoAlvo, error) {
	mes, err := competenciaFlag(cmd)
	if err != nil {
		return extratoAlvo{}, err
	}
	conta, _ := cmd.Flags().GetInt64("conta-bancaria")
	if conta <= 0 {
		return extratoAlvo{}, usageError{errors.New("informe --conta-bancaria ID (ver ctbz contas-bancarias)")}
	}
	return extratoAlvo{conta: conta, mes: mes}, nil
}

func (e extratoAlvo) ano() int    { return e.mes.Year() }
func (e extratoAlvo) mesNum() int { return int(e.mes.Month()) }

func (e extratoAlvo) lancamentos(ctx context.Context, g api.Getter) ([]api.LancamentoExtrato, error) {
	return api.BuscarLancamentosExtrato(ctx, g, e.conta, e.ano(), e.mesNum())
}

// lancamento acha um lançamento do extrato pelo id.
func (e extratoAlvo) lancamento(ctx context.Context, g api.Getter, idArg string) (*api.LancamentoExtrato, error) {
	id, err := strconv.ParseInt(idArg, 10, 64)
	if err != nil {
		return nil, usageError{fmt.Errorf("id de lançamento inválido %q", idArg)}
	}
	ls, err := e.lancamentos(ctx, g)
	if err != nil {
		return nil, err
	}
	for i := range ls {
		if ls[i].ID == id {
			return &ls[i], nil
		}
	}
	return nil, fmt.Errorf("lançamento %d não encontrado no extrato da conta %d em %s", id, e.conta, e.mes.Format("01/2006"))
}

// editavel lê extrato/info e recusa quando o painel não deixa editar os lançamentos.
func (e extratoAlvo) editavel(ctx context.Context, g api.Getter) (*api.ExtratoInfo, error) {
	info, err := api.BuscarExtratoInfo(ctx, g, e.conta, e.ano(), e.mesNum())
	if err != nil {
		return nil, err
	}
	if !info.PermiteEditarLancamento {
		return nil, fmt.Errorf("os lançamentos do extrato de %s não podem mais ser alterados (o painel também não permite)", e.mes.Format("01/2006"))
	}
	return info, nil
}

func valorExtrato(l api.LancamentoExtrato) int64 {
	if l.Valor == nil {
		return 0
	}
	return centavos(*l.Valor)
}

func diaExtrato(l api.LancamentoExtrato) any {
	if l.Data.IsZero() {
		return nil
	}
	d, _ := output.ParseDate(api.DiaEmBrasilia(l.Data.Time).Format("2006-01-02"))
	return d
}

func lancamentosExtratoList(ls []api.LancamentoExtrato, nomes map[int64]string) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "id", Header: "ID"},
		{Key: "data", Header: "Data"},
		{Key: "descricao", Header: "Descrição"},
		{Key: "valor", Header: "Valor"},
		{Key: "classificacao", Header: "Classificação"},
		{Key: "id_classificacao", Header: "ID da classificação"},
		{Key: "id_lancamento_pai", Header: "Parte de"},
	}}
	for _, x := range ls {
		var classif, idClassif, pai any
		if x.IDContaUsuario != nil {
			idClassif = *x.IDContaUsuario
			classif = nilIfEmpty(nomes[*x.IDContaUsuario])
		}
		if x.IDLancamentoPai != nil {
			pai = *x.IDLancamentoPai
		}
		l.Append(x.ID, diaExtrato(x), output.Text(x.Descricao), moneyOrNil(x.Valor), classif, idClassif, pai)
	}
	return l
}

func nomesContas(ctx context.Context, g api.Getter) (map[int64]string, error) {
	cs, err := api.BuscarContasUsuario(ctx, g)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]string, len(cs))
	for _, c := range cs {
		m[c.ID] = c.Descricao
	}
	return m, nil
}

func newExtratosLancamentosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lancamentos",
		Short: "Lista os lançamentos do extrato de uma conta bancária num mês",
		Long: `Lista os lançamentos do extrato importado de uma conta bancária na competência: id, data,
descrição, valor (negativo nas saídas), classificação atual e, nas partes de um lançamento
desmembrado, o id do lançamento original ("Parte de").

Os ids são usados por ctbz extratos classificar e desmembrar.`,
		Example: `  ctbz extratos lancamentos --conta-bancaria 1000000000000001 --competencia 2026-09
  ctbz extratos lancamentos --conta-bancaria 1000000000000001 --competencia 2026-09 -o csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			alvo, err := extratoFlags(cmd)
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			ls, err := alvo.lancamentos(cmd.Context(), g)
			if err != nil {
				return err
			}
			nomes, err := nomesContas(cmd.Context(), g)
			if err != nil {
				return err
			}
			return output.Write(s.out, f, lancamentosExtratoList(ls, nomes))
		},
	}
	addContaBancariaFlag(cmd)
	addCompetenciaFlag(cmd)
	return cmd
}

// contaExtrato resolve --conta (id ou descrição) entre as classificações aceitas para o sinal
// do lançamento, com o sócio quando a conta exige.
func contaExtrato(comp []api.ContaUsuarioCompetencia, nomes map[int64]string, info *api.ExtratoInfo, valor int64, conta, socio string, mes time.Time) (int64, *int64, string, error) {
	tipo := "DESPESA"
	if valor > 0 {
		tipo = "RECEITA"
	}
	var escolhida *api.ContaUsuarioCompetencia
	for _, c := range classificacoesExtrato(comp, tipo) {
		desc := nomes[c.ID]
		if desc == "" {
			desc = c.Descricao
		}
		if strconv.FormatInt(c.ID, 10) == conta || (desc != "" && strings.EqualFold(desc, conta)) {
			c := c
			escolhida = &c
			break
		}
	}
	if escolhida == nil {
		return 0, nil, "", usageError{fmt.Errorf("classificação %q não aceita para %s em %s; veja ctbz extratos contas --competencia %s --%s",
			conta, strings.ToLower(tipo), mes.Format("01/2006"), mes.Format("2006-01"), strings.ToLower(tipo))}
	}
	nome := nomes[escolhida.ID]
	if nome == "" {
		nome = strconv.FormatInt(escolhida.ID, 10)
	}
	if !api.IDsContasDeSocio[escolhida.ID] {
		if socio != "" {
			return 0, nil, "", usageError{fmt.Errorf("a classificação %q não usa sócio", nome)}
		}
		return escolhida.ID, nil, nome, nil
	}
	if socio == "" {
		return 0, nil, "", usageError{fmt.Errorf("a classificação %q exige --socio ID", nome)}
	}
	for _, sc := range info.Socios {
		if strconv.FormatInt(sc.ID, 10) == socio {
			id := sc.ID
			return escolhida.ID, &id, nome + " / " + sc.Nome, nil
		}
	}
	return 0, nil, "", usageError{fmt.Errorf("sócio %q não encontrado no extrato", socio)}
}

// erroPeriodoClassificado traduz o 406 do classificar para a mensagem do painel.
func erroPeriodoClassificado(err error) error {
	var we *ctbz.WriteError
	if errors.As(err, &we) && we.Status == http.StatusNotAcceptable {
		return errors.New("os contadores já classificaram as movimentações deste período para fazer a contabilidade; para correções, fale com o suporte da Contabilizei")
	}
	return err
}

func newExtratosClassificarCmd() *cobra.Command {
	var conta, socio string
	cmd := &cobra.Command{
		Use:   "classificar ID",
		Short: "Troca a classificação de um lançamento do extrato",
		Long: `Troca a classificação de um lançamento do extrato importado (risco médio, reversível
classificando de novo enquanto o período estiver aberto).

--conta aceita o id ou a descrição exata de uma classificação aceita na competência: RECEITA
para entradas e DESPESA para saídas (ctbz extratos contas). As classificações de sócio exigem
--socio. Se os contadores já fecharam o período, o painel não deixa alterar, e a CLI também
não.

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz extratos classificar 1000000000000102 --conta-bancaria 1000000000000001 \
    --competencia 2026-09 --conta "Pagamento de Fornecedores"`,
		Args: exactArgs(1, "o ID do lançamento"),
		RunE: func(cmd *cobra.Command, args []string) error {
			alvo, err := extratoFlags(cmd)
			if err != nil {
				return err
			}
			if conta == "" {
				return usageError{errors.New("informe --conta (ver ctbz extratos contas)")}
			}
			formato, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			info, err := alvo.editavel(ctx, g)
			if err != nil {
				return err
			}
			l, err := alvo.lancamento(ctx, g, args[0])
			if err != nil {
				return err
			}
			valor := valorExtrato(*l)
			if valor == 0 {
				return errors.New("o lançamento tem valor zero e não pode ser classificado")
			}
			comp, err := api.BuscarContasUsuarioCompetencia(ctx, g, alvo.ano(), alvo.mesNum(), api.OrigemExtrato)
			if err != nil {
				return err
			}
			nomes, err := nomesContas(ctx, g)
			if err != nil {
				return err
			}
			idConta, idSocio, nome, err := contaExtrato(comp, nomes, info, valor, conta, socio, alvo.mes)
			if err != nil {
				return err
			}
			antes := "sem classificação"
			if l.IDContaUsuario != nil {
				if *l.IDContaUsuario == idConta && idSocio == nil {
					return usageError{fmt.Errorf("o lançamento já está classificado como %q", nome)}
				}
				antes = nomes[*l.IDContaUsuario]
			}
			req := api.ClassificarLancamento{IDLancamentoUsuario: l.ID, IDContaUsuario: idConta, IDSocio: idSocio}
			op := operacao{Risco: riscoMedio, ID: args[0], Resumo: fmt.Sprintf("Classificar o lançamento %d (%q, %s): %s → %s",
				l.ID, l.Descricao, formatarCentavos(valor), antes, nome)}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroPeriodoClassificado(api.Classificar(ctx, snd, req))
			})
			if err != nil || !enviado {
				return err
			}
			situacao := "classificado"
			if novo, err := alvo.lancamento(ctx, g, args[0]); err != nil {
				fmt.Fprintln(s.err, "aviso: classificação enviada, mas não foi possível reler o extrato:", err)
				situacao = "enviado"
			} else if novo.IDContaUsuario == nil || *novo.IDContaUsuario != idConta {
				fmt.Fprintln(s.err, "aviso: o extrato ainda mostra a classificação antiga; confira com ctbz extratos lancamentos")
				situacao = "enviado"
			}
			rec := resultadoEscrita("classificar", situacao, l.ID).
				Add("descricao", "Descrição", l.Descricao).
				Add("valor", "Valor", output.Money(reais(valor))).
				Add("classificacao", "Classificação", nome).
				Add("id_classificacao", "ID da classificação", idConta)
			return output.Write(s.out, formato, rec)
		},
	}
	addContaBancariaFlag(cmd)
	addCompetenciaFlag(cmd)
	cmd.Flags().StringVar(&conta, "conta", "", "classificação: id ou descrição exata (ver ctbz extratos contas)")
	cmd.Flags().StringVar(&socio, "socio", "", "sócio, nas classificações de sócio")
	addWriteFlags(cmd)
	return cmd
}
