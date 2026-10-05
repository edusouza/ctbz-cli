package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// avisoAbertura é o aviso do painel sobre a data de abertura da conta.
const avisoAbertura = "A data de abertura define desde quando a Contabilizei vai pedir os extratos desta conta."

func newContasBancariasBancosCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bancos",
		Short: "Lista os bancos aceitos no cadastro de contas bancárias",
		Long:  `Lista os bancos aceitos no cadastro de contas (id, código e nome); use o id ou o código em --banco.`,
		Args:  exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			c, err := api.BuscarContasBancarias(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{{Key: "id", Header: "ID"}, {Key: "codigo", Header: "Código"}, {Key: "nome", Header: "Banco"}}}
			for _, b := range c.Bancos {
				l.Append(b.ID, nilIfEmpty(b.Codigo), nilIfEmpty(b.Nome))
			}
			return output.Write(s.out, f, l)
		},
	}
}

// acharBanco resolve --banco pelo id ou pelo código (com ou sem dígito, ex.: 341 ou 341-7).
func acharBanco(bancos []api.Banco, v string) (api.Banco, error) {
	for _, b := range bancos {
		cod, _, _ := strings.Cut(b.Codigo, "-")
		if strconv.FormatInt(b.ID, 10) == v || (b.Codigo != "" && (b.Codigo == v || cod == v)) {
			return b, nil
		}
	}
	return api.Banco{}, usageError{fmt.Errorf("banco %q não encontrado; veja ctbz contas-bancarias bancos", v)}
}

// normalizarConta tira hífen, pontos e espaços: o painel envia conta e dígito juntos.
func normalizarConta(v string) string {
	return strings.NewReplacer("-", "", ".", "", " ", "").Replace(strings.TrimSpace(v))
}

// contaBancariaFlags são as flags de adicionar e editar.
type contaBancariaFlags struct {
	banco, agencia, conta, abertura, saldo string
}

func (f *contaBancariaFlags) add(cmd *cobra.Command) {
	fl := cmd.Flags()
	fl.StringVar(&f.banco, "banco", "", "banco: id ou código (ver ctbz contas-bancarias bancos)")
	fl.StringVar(&f.agencia, "agencia", "", "agência, sem dígito")
	fl.StringVar(&f.conta, "conta", "", "conta com dígito (ex.: 12345-6)")
	fl.StringVar(&f.abertura, "abertura", "", "data de abertura da conta AAAA-MM-DD (define desde quando os extratos são pedidos)")
	fl.StringVar(&f.saldo, "saldo-inicial", "", "saldo na abertura, em reais (padrão: 0)")
}

func parseAbertura(v string) (string, error) {
	d, err := time.Parse("2006-01-02", v)
	if err != nil {
		return "", usageError{fmt.Errorf("data de abertura inválida %q: use AAAA-MM-DD", v)}
	}
	if d.After(api.DiaEmBrasilia(now())) {
		return "", usageError{fmt.Errorf("a data de abertura %s é futura", d.Format("02/01/2006"))}
	}
	return v, nil
}

func parseSaldo(v string) (float64, error) {
	if v == "" {
		return 0, nil
	}
	c, err := parseValor(v)
	if err != nil {
		return 0, usageError{err}
	}
	return reais(c), nil
}

func descreverConta(nome, agencia, conta string) string {
	return fmt.Sprintf("%s agência %s conta %s", nome, agencia, conta)
}

// releConta relê a lista e devolve a conta pelo id, ou a que bate com banco, agência e conta.
func releConta(ctx context.Context, g api.Getter, id int64, agencia, conta string) (*api.ContaBancaria, error) {
	c, err := api.BuscarContasBancarias(ctx, g)
	if err != nil {
		return nil, err
	}
	for i, cb := range c.ContasBancarias {
		if (id != 0 && cb.ID == id) || (id == 0 && cb.Agencia == agencia && normalizarConta(cb.ContaCorrente) == conta) {
			return &c.ContasBancarias[i], nil
		}
	}
	return nil, nil
}

func contaResultado(acao, situacao string, relida *api.ContaBancaria, enviada api.ContaBancariaSalvar, banco string) *output.Record {
	var id any
	if enviada.ID != nil {
		id = *enviada.ID
	}
	if relida == nil {
		situacao = "enviado"
	} else {
		id = relida.ID
	}
	abertura, _ := output.ParseDate(enviada.DataSaldoInicial)
	return resultadoEscrita(acao, situacao, id).
		Add("banco", "Banco", banco).
		Add("agencia", "Agência", enviada.Agencia).
		Add("conta", "Conta", enviada.ContaCorrente).
		Add("abertura", "Abertura", abertura).
		Add("saldo_inicial", "Saldo inicial", output.Money(enviada.VlrSaldoInicial))
}

func newContasBancariasAdicionarCmd() *cobra.Command {
	var f contaBancariaFlags
	var declaro bool
	cmd := &cobra.Command{
		Use:   "adicionar",
		Short: "Cadastra uma conta bancária PJ",
		Long: `Cadastra uma conta bancária PJ da empresa, para poder importar os extratos dela (risco
médio: dá para editar depois). ` + avisoAbertura + `

--declaro-conta-pj é obrigatório, como o checkbox do painel: "a conta é uma Conta PJ e suas
movimentações refletem exclusivamente este CNPJ".

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz contas-bancarias adicionar --banco 341 --agencia 1234 --conta 12345-6 \
    --abertura 2024-01-10 --declaro-conta-pj`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !declaro {
				return usageError{errors.New("informe --declaro-conta-pj: a conta precisa ser PJ e movimentar só este CNPJ")}
			}
			if f.banco == "" || f.agencia == "" || f.conta == "" || f.abertura == "" {
				return usageError{errors.New("informe --banco, --agencia, --conta e --abertura")}
			}
			abertura, err := parseAbertura(f.abertura)
			if err != nil {
				return err
			}
			saldo, err := parseSaldo(f.saldo)
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
			lista, err := api.BuscarContasBancarias(ctx, g)
			if err != nil {
				return err
			}
			banco, err := acharBanco(lista.Bancos, f.banco)
			if err != nil {
				return err
			}
			req := api.ContaBancariaSalvar{BancoID: banco.ID, Agencia: strings.TrimSpace(f.agencia), ContaCorrente: normalizarConta(f.conta),
				DataSaldoInicial: abertura, VlrSaldoInicial: saldo}
			fmt.Fprintln(s.err, avisoAbertura)
			op := operacao{Risco: riscoMedio, Resumo: fmt.Sprintf("Cadastrar a conta %s, aberta em %s com saldo de %s",
				descreverConta(banco.Nome, req.Agencia, req.ContaCorrente), abertura, formatarCentavos(centavos(saldo)))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.SalvarContaBancaria(ctx, snd, req) })
			if err != nil || !enviado {
				return err
			}
			relida, err := releConta(ctx, g, 0, req.Agencia, req.ContaCorrente)
			if err != nil {
				fmt.Fprintln(s.err, "aviso: conta salva, mas não foi possível reler a lista:", err)
			}
			return output.Write(s.out, formato, contaResultado("adicionar", "cadastrado", relida, req, banco.Nome))
		},
	}
	f.add(cmd)
	cmd.Flags().BoolVar(&declaro, "declaro-conta-pj", false, "declara que é uma conta PJ que movimenta só este CNPJ (obrigatório)")
	addWriteFlags(cmd)
	return cmd
}

// contaComPermissao acha a conta na lista e lê as permissões do painel.
func contaComPermissao(ctx context.Context, g api.Getter, idArg string) (*api.ContasBancarias, *api.ContaBancaria, *api.DetalheContaBancaria, error) {
	id, err := parseID(idArg)
	if err != nil {
		return nil, nil, nil, err
	}
	lista, err := api.BuscarContasBancarias(ctx, g)
	if err != nil {
		return nil, nil, nil, err
	}
	var conta *api.ContaBancaria
	for i := range lista.ContasBancarias {
		if lista.ContasBancarias[i].ID == id {
			conta = &lista.ContasBancarias[i]
		}
	}
	if conta == nil {
		return nil, nil, nil, fmt.Errorf("conta bancária %d não encontrada; veja ctbz contas-bancarias", id)
	}
	det, err := api.BuscarDetalheContaBancaria(ctx, g, id)
	if err != nil {
		return nil, nil, nil, err
	}
	return lista, conta, det, nil
}

func motivoBloqueio(d *api.DetalheContaBancaria) string {
	if d.MotivoPermissoesExclusaoEEdicao != "" {
		return ": " + d.MotivoPermissoesExclusaoEEdicao
	}
	return ""
}

func newContasBancariasEditarCmd() *cobra.Command {
	var f contaBancariaFlags
	cmd := &cobra.Command{
		Use:   "editar ID",
		Short: "Corrige os dados de uma conta bancária cadastrada",
		Long: `Corrige banco, agência, conta, data de abertura ou saldo inicial de uma conta (risco médio:
dá para editar de novo). Só as flags informadas mudam. O painel pode bloquear a edição (por
exemplo, com extratos já classificados); nesse caso a CLI mostra o motivo.
` + avisoAbertura + `

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz contas-bancarias editar 1000000000000001 --abertura 2024-02-01`,
		Args:    exactArgs(1, "o ID da conta"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.banco == "" && f.agencia == "" && f.conta == "" && f.abertura == "" && f.saldo == "" {
				return usageError{errors.New("informe o que alterar: --banco, --agencia, --conta, --abertura ou --saldo-inicial")}
			}
			formato, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			lista, atual, det, err := contaComPermissao(ctx, g, args[0])
			if err != nil {
				return err
			}
			if !det.PermiteEditar {
				return fmt.Errorf("o painel não permite editar esta conta%s", motivoBloqueio(det))
			}
			req := api.ContaBancariaSalvar{ID: &atual.ID, Agencia: atual.Agencia, ContaCorrente: normalizarConta(atual.ContaCorrente),
				DataSaldoInicial: api.DiaEmBrasilia(time.UnixMilli(atual.DataSaldoInicial)).Format("2006-01-02")}
			if atual.VlrSaldoInicial != nil {
				req.VlrSaldoInicial = *atual.VlrSaldoInicial
			}
			nomeBanco := atual.NomeBanco
			if atual.BancoID != nil {
				req.BancoID = *atual.BancoID
			} else if b, err := acharBanco(lista.Bancos, atual.CodigoBanco); err == nil {
				req.BancoID = b.ID
			}
			if f.banco != "" {
				b, err := acharBanco(lista.Bancos, f.banco)
				if err != nil {
					return err
				}
				req.BancoID, nomeBanco = b.ID, b.Nome
			}
			if req.BancoID == 0 {
				return usageError{errors.New("não foi possível identificar o banco atual da conta; informe --banco")}
			}
			if f.agencia != "" {
				req.Agencia = strings.TrimSpace(f.agencia)
			}
			if f.conta != "" {
				req.ContaCorrente = normalizarConta(f.conta)
			}
			if f.abertura != "" {
				if req.DataSaldoInicial, err = parseAbertura(f.abertura); err != nil {
					return err
				}
				fmt.Fprintln(s.err, avisoAbertura)
			}
			if f.saldo != "" {
				if req.VlrSaldoInicial, err = parseSaldo(f.saldo); err != nil {
					return err
				}
			}
			op := operacao{Risco: riscoMedio, ID: args[0], Resumo: fmt.Sprintf("Editar a conta %d: %s → %s, abertura %s, saldo inicial %s",
				atual.ID, descreverConta(atual.NomeBanco, atual.Agencia, atual.ContaCorrente), descreverConta(nomeBanco, req.Agencia, req.ContaCorrente),
				req.DataSaldoInicial, formatarCentavos(centavos(req.VlrSaldoInicial)))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.SalvarContaBancaria(ctx, snd, req) })
			if err != nil || !enviado {
				return err
			}
			relida, err := releConta(ctx, g, atual.ID, "", "")
			if err != nil {
				fmt.Fprintln(s.err, "aviso: conta salva, mas não foi possível reler a lista:", err)
			}
			return output.Write(s.out, formato, contaResultado("editar", "editado", relida, req, nomeBanco))
		},
	}
	f.add(cmd)
	addWriteFlags(cmd)
	return cmd
}
