package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// parteFlag é uma --parte "descrição:valor:conta[:sócio]" já lida.
type parteFlag struct {
	descricao, conta, socio string
	valor                   int64 // centavos, positivo
}

func parseParte(s string) (parteFlag, error) {
	campos := strings.Split(s, ":")
	if len(campos) != 3 && len(campos) != 4 {
		return parteFlag{}, fmt.Errorf("parte inválida %q: use \"descrição:valor:conta[:sócio]\" (a descrição não pode ter \":\")", s)
	}
	p := parteFlag{descricao: strings.TrimSpace(campos[0]), conta: strings.TrimSpace(campos[2])}
	if len(campos) == 4 {
		p.socio = strings.TrimSpace(campos[3])
	}
	if p.descricao == "" || p.conta == "" {
		return parteFlag{}, fmt.Errorf("parte inválida %q: descrição e conta são obrigatórias", s)
	}
	v, err := parseValor(campos[1])
	if err != nil {
		return parteFlag{}, fmt.Errorf("parte %q: %w", s, err)
	}
	if v <= 0 {
		return parteFlag{}, fmt.Errorf("parte %q: o valor deve ser maior que zero (o sinal é o do lançamento original)", s)
	}
	p.valor = v
	return p, nil
}

func newExtratosDesmembrarCmd() *cobra.Command {
	var partesArg []string
	cmd := &cobra.Command{
		Use:   "desmembrar ID",
		Short: "Divide um lançamento do extrato em partes com classificações diferentes",
		Long: `Divide um lançamento do extrato em duas ou mais partes, cada uma com descrição, valor e
classificação próprios (risco médio, reversível com ctbz extratos desfazer-desmembramento).

Cada --parte é "descrição:valor:conta[:sócio]": o valor é positivo (o sinal é o do lançamento
original), a conta é o id ou a descrição exata de uma classificação aceita (ctbz extratos
contas) e o sócio é obrigatório nas classificações de sócio.

Regras do painel, conferidas antes de enviar: pelo menos 2 partes, nenhuma com valor zero, a
soma igual ao valor do lançamento (em centavos), e o lançamento não pode ser parte de outro
desmembramento nem ter vínculo.

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz extratos desmembrar 1000000000000102 --conta-bancaria 1000000000000001 --competencia 2026-09 \
    --parte "Aluguel:1.000,00:Pagamento de Fornecedores" --parte "Condomínio:234,56:1000000000000012"`,
		Args: exactArgs(1, "o ID do lançamento"),
		RunE: func(cmd *cobra.Command, args []string) error {
			alvo, err := extratoFlags(cmd)
			if err != nil {
				return err
			}
			if len(partesArg) < 2 {
				return usageError{errors.New("informe pelo menos 2 --parte \"descrição:valor:conta[:sócio]\"")}
			}
			partes := make([]parteFlag, len(partesArg))
			soma := int64(0)
			for i, p := range partesArg {
				if partes[i], err = parseParte(p); err != nil {
					return usageError{err}
				}
				soma += partes[i].valor
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
			switch {
			case valor == 0:
				return errors.New("lançamento com valor zero não pode ser desmembrado")
			case l.IDLancamentoPai != nil:
				return fmt.Errorf("o lançamento %d já é parte do desmembramento de %d; desfaça aquele primeiro", l.ID, *l.IDLancamentoPai)
			case len(l.IDVinculo) > 0 && string(l.IDVinculo) != "null":
				return fmt.Errorf("o lançamento %d tem vínculo e não pode ser desmembrado (o painel também não permite)", l.ID)
			case soma != abs(valor):
				return usageError{fmt.Errorf("a soma das partes (%s) é diferente do valor do lançamento (%s)", formatarCentavos(soma), formatarCentavos(abs(valor)))}
			}
			comp, err := api.BuscarContasUsuarioCompetencia(ctx, g, alvo.ano(), alvo.mesNum(), api.OrigemExtrato)
			if err != nil {
				return err
			}
			nomes, err := nomesContas(ctx, g)
			if err != nil {
				return err
			}
			req := api.Desmembramento{IDLancamentoPai: l.ID}
			var resumo []string
			for _, p := range partes {
				idConta, idSocio, nome, err := contaExtrato(comp, nomes, info, valor, p.conta, p.socio, alvo.mes)
				if err != nil {
					return err
				}
				v := p.valor
				if valor < 0 {
					v = -v
				}
				req.LancamentosFilho = append(req.LancamentosFilho, api.LancamentoFilho{Descricao: p.descricao, Valor: reais(v), IDContaUsuario: idConta, IDVinculo: idSocio})
				resumo = append(resumo, fmt.Sprintf("%q %s (%s)", p.descricao, formatarCentavos(v), nome))
			}
			op := operacao{Risco: riscoMedio, ID: args[0], Resumo: fmt.Sprintf("Desmembrar o lançamento %d (%q, %s) em %d partes: %s",
				l.ID, l.Descricao, formatarCentavos(valor), len(partes), strings.Join(resumo, "; "))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroPeriodoClassificado(api.Desmembrar(ctx, snd, req))
			})
			if err != nil || !enviado {
				return err
			}
			situacao := "desmembrado"
			var filhos []map[string]any
			ls, err := alvo.lancamentos(ctx, g)
			if err != nil {
				fmt.Fprintln(s.err, "aviso: desmembramento enviado, mas não foi possível reler o extrato:", err)
			}
			for _, x := range ls {
				if x.IDLancamentoPai != nil && *x.IDLancamentoPai == l.ID {
					var classif any
					if x.IDContaUsuario != nil {
						classif = nomes[*x.IDContaUsuario]
					}
					filhos = append(filhos, map[string]any{"id": x.ID, "descricao": x.Descricao, "valor": moneyOrNil(x.Valor), "classificacao": classif})
				}
			}
			if len(filhos) == 0 {
				if err == nil {
					fmt.Fprintln(s.err, "aviso: as partes ainda não aparecem no extrato; confira com ctbz extratos lancamentos")
				}
				situacao = "enviado"
			}
			rec := resultadoEscrita("desmembrar", situacao, l.ID).
				Add("descricao", "Descrição", l.Descricao).
				Add("valor", "Valor", output.Money(reais(valor))).
				Add("partes", "Partes", filhos)
			return output.Write(s.out, formato, rec)
		},
	}
	addContaBancariaFlag(cmd)
	addCompetenciaFlag(cmd)
	cmd.Flags().StringArrayVar(&partesArg, "parte", nil, "parte \"descrição:valor:conta[:sócio]\" (repita para cada parte)")
	addWriteFlags(cmd)
	return cmd
}

func newExtratosDesfazerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "desfazer-desmembramento ID",
		Short: "Volta um lançamento desmembrado do extrato ao original",
		Long: `Apaga as partes de um lançamento desmembrado e restaura o lançamento original (risco
médio; dá para desmembrar de novo). ID pode ser o do lançamento original ou o de qualquer uma
das partes (coluna "Parte de" em ctbz extratos lancamentos).

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz extratos desfazer-desmembramento 1000000000000102 --conta-bancaria 1000000000000001 --competencia 2026-09`,
		Args:    exactArgs(1, "o ID do lançamento original ou de uma parte"),
		RunE: func(cmd *cobra.Command, args []string) error {
			alvo, err := extratoFlags(cmd)
			if err != nil {
				return err
			}
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return usageError{fmt.Errorf("id de lançamento inválido %q", args[0])}
			}
			formato, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			if _, err := alvo.editavel(ctx, g); err != nil {
				return err
			}
			ls, err := alvo.lancamentos(ctx, g)
			if err != nil {
				return err
			}
			pai, partes := paiEPartes(ls, id)
			if len(partes) == 0 {
				return fmt.Errorf("o lançamento %d não é um lançamento desmembrado nem parte de um", id)
			}
			op := operacao{Risco: riscoMedio, ID: strconv.FormatInt(pai, 10), Resumo: fmt.Sprintf(
				"Desfazer o desmembramento do lançamento %d (%d parte(s) voltam a ser um lançamento só)", pai, len(partes))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroPeriodoClassificado(api.DesfazerDesmembramento(ctx, snd, pai))
			})
			if err != nil || !enviado {
				return err
			}
			situacao := "desfeito"
			if ls, err := alvo.lancamentos(ctx, g); err != nil {
				fmt.Fprintln(s.err, "aviso: pedido enviado, mas não foi possível reler o extrato:", err)
				situacao = "enviado"
			} else if _, restantes := paiEPartes(ls, pai); len(restantes) > 0 {
				fmt.Fprintln(s.err, "aviso: as partes ainda aparecem no extrato; confira com ctbz extratos lancamentos")
				situacao = "enviado"
			}
			rec := resultadoEscrita("desfazer-desmembramento", situacao, pai).Add("partes_removidas", "Partes removidas", len(partes))
			return output.Write(s.out, formato, rec)
		},
	}
	addContaBancariaFlag(cmd)
	addCompetenciaFlag(cmd)
	addWriteFlags(cmd)
	return cmd
}

// paiEPartes devolve o id do lançamento original e suas partes, dado o id do original ou de
// uma parte.
func paiEPartes(ls []api.LancamentoExtrato, id int64) (int64, []api.LancamentoExtrato) {
	pai := id
	for _, x := range ls {
		if x.ID == id && x.IDLancamentoPai != nil {
			pai = *x.IDLancamentoPai
		}
	}
	var partes []api.LancamentoExtrato
	for _, x := range ls {
		if x.IDLancamentoPai != nil && *x.IDLancamentoPai == pai {
			partes = append(partes, x)
		}
	}
	return pai, partes
}
