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
		Long: `Lista os bancos aceitos no cadastro de contas (id, código e nome); use o id ou o código em --banco.

Alguns códigos se repetem (ex.: dois bancos com 341) ou têm uma versão internacional (077 e
077-1). --banco escolhe o código exato; se mais de um banco tiver o código, a CLI lista os
candidatos e pede o id.`,
		Args: exactArgs(0, "nenhum argumento"),
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

// acharBanco resolve --banco pelo id; depois pelo código exato (ex.: 077 ou 341-7); e, só sem
// código exato, pelo código sem o sufixo (341 → 341-7). Nunca escolhe entre candidatos: há
// bancos com o mesmo código (ex.: 341) e pares nacional/internacional (077 e 077-1).
func acharBanco(bancos []api.Banco, v string) (api.Banco, error) {
	v = strings.TrimSpace(v)
	for _, b := range bancos {
		if strconv.FormatInt(b.ID, 10) == v {
			return b, nil
		}
	}
	var exatos, semSufixo []api.Banco
	for _, b := range bancos {
		cod, _, _ := strings.Cut(b.Codigo, "-")
		switch {
		case b.Codigo == "":
		case b.Codigo == v:
			exatos = append(exatos, b)
		case cod == v:
			semSufixo = append(semSufixo, b)
		}
	}
	candidatos := exatos
	if len(candidatos) == 0 {
		candidatos = semSufixo
	}
	switch len(candidatos) {
	case 0:
		return api.Banco{}, usageError{fmt.Errorf("banco %q não encontrado; veja ctbz contas-bancarias bancos", v)}
	case 1:
		return candidatos[0], nil
	}
	linhas := make([]string, len(candidatos))
	for i, b := range candidatos {
		linhas[i] = fmt.Sprintf("  %d  %s  %s", b.ID, b.Codigo, b.Nome)
	}
	return api.Banco{}, usageError{fmt.Errorf("o código %q corresponde a %d bancos; informe o id em --banco:\n%s",
		v, len(candidatos), strings.Join(linhas, "\n"))}
}

// bancoAtual acha o banco de uma conta que não traz o id do banco, pelo código e, se o
// código for ambíguo, também pelo nome.
func bancoAtual(bancos []api.Banco, codigo, nome string) (api.Banco, bool) {
	if b, err := acharBanco(bancos, codigo); err == nil {
		return b, true
	}
	for _, b := range bancos {
		if codigo != "" && b.Codigo == codigo && strings.EqualFold(b.Nome, nome) {
			return b, true
		}
	}
	return api.Banco{}, false
}

// nomeBanco mostra o nome do banco com o código, para não confundir bancos de mesmo nome.
func nomeBanco(b api.Banco) string {
	if b.Codigo == "" {
		return b.Nome
	}
	return fmt.Sprintf("%s (%s)", b.Nome, b.Codigo)
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
	fl.StringVar(&f.banco, "banco", "", "banco: id ou código exato (ver ctbz contas-bancarias bancos)")
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
		Example: `  ctbz contas-bancarias adicionar --banco 237 --agencia 1234 --conta 12345-6 \
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
				descreverConta(nomeBanco(banco), req.Agencia, req.ContaCorrente), abertura, formatarCentavos(centavos(saldo)))}
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

// aberturaAtual devolve a abertura da conta na lista (AAAA-MM-DD), ou "" se a data não for
// plausível: 0, null ou ausente viram 1969-12-31, e nenhuma conta PJ do painel é de antes de 1990.
func aberturaAtual(ms int64) string {
	d := api.DiaEmBrasilia(time.UnixMilli(ms))
	if d.Year() < 1990 {
		return ""
	}
	return d.Format("2006-01-02")
}

// exigirAberturaESaldo recusa a edição quando a lista não traz a abertura ou o saldo inicial e
// a flag correspondente não foi informada: a edição reenvia os dois, e a CLI não inventa um valor.
func exigirAberturaESaldo(temAbertura, temSaldo bool) error {
	var faltam, flags []string
	if !temAbertura {
		faltam, flags = append(faltam, "uma data de abertura válida"), append(flags, "--abertura AAAA-MM-DD")
	}
	if !temSaldo {
		faltam, flags = append(faltam, "o saldo inicial"), append(flags, "--saldo-inicial")
	}
	if len(faltam) == 0 {
		return nil
	}
	return usageError{fmt.Errorf("a Contabilizei não informou %s desta conta e a edição reenvia todos os campos; informe %s",
		strings.Join(faltam, " nem "), strings.Join(flags, " e "))}
}

func newContasBancariasEditarCmd() *cobra.Command {
	var f contaBancariaFlags
	cmd := &cobra.Command{
		Use:   "editar ID",
		Short: "Corrige os dados de uma conta bancária cadastrada",
		Long: `Corrige banco, agência, conta, data de abertura ou saldo inicial de uma conta (risco médio:
dá para editar de novo). Só as flags informadas mudam; como a edição reenvia todos os campos,
se a Contabilizei não informar a abertura ou o saldo inicial da conta, informe --abertura ou
--saldo-inicial. O painel pode bloquear a edição (por exemplo, com extratos já classificados);
nesse caso a CLI mostra o motivo.
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
				DataSaldoInicial: aberturaAtual(atual.DataSaldoInicial)}
			if atual.VlrSaldoInicial != nil {
				req.VlrSaldoInicial = *atual.VlrSaldoInicial
			}
			if err := exigirAberturaESaldo(req.DataSaldoInicial != "" || f.abertura != "", atual.VlrSaldoInicial != nil || f.saldo != ""); err != nil {
				return err
			}
			bancoAntes := nomeBanco(api.Banco{Nome: atual.NomeBanco, Codigo: atual.CodigoBanco})
			bancoDepois, nomeResultado := bancoAntes, atual.NomeBanco
			if atual.BancoID != nil {
				req.BancoID = *atual.BancoID
			} else if b, ok := bancoAtual(lista.Bancos, atual.CodigoBanco, atual.NomeBanco); ok {
				req.BancoID = b.ID
			}
			if f.banco != "" {
				b, err := acharBanco(lista.Bancos, f.banco)
				if err != nil {
					return err
				}
				req.BancoID, bancoDepois, nomeResultado = b.ID, nomeBanco(b), b.Nome
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
				atual.ID, descreverConta(bancoAntes, atual.Agencia, atual.ContaCorrente), descreverConta(bancoDepois, req.Agencia, req.ContaCorrente),
				req.DataSaldoInicial, formatarCentavos(centavos(req.VlrSaldoInicial)))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.SalvarContaBancaria(ctx, snd, req) })
			if err != nil || !enviado {
				return err
			}
			relida, err := releConta(ctx, g, atual.ID, "", "")
			if err != nil {
				fmt.Fprintln(s.err, "aviso: conta salva, mas não foi possível reler a lista:", err)
			}
			return output.Write(s.out, formato, contaResultado("editar", "editado", relida, req, nomeResultado))
		},
	}
	f.add(cmd)
	addWriteFlags(cmd)
	return cmd
}

func newContasBancariasRemoverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remover ID",
		Short: "Exclui permanentemente uma conta bancária",
		Long: `Exclui uma conta bancária cadastrada por engano ou encerrada (risco alto: "Excluir
permanentemente" no painel; os vínculos da conta se perdem e um novo cadastro não os
recupera). O painel pode bloquear a exclusão; nesse caso a CLI mostra o motivo.

Mostra banco, agência e conta e pede "confirmo" no terminal (--yes em scripts); aceita --dry-run.`,
		Example: `  ctbz contas-bancarias remover 1000000000000001`,
		Args:    exactArgs(1, "o ID da conta"),
		RunE: func(cmd *cobra.Command, args []string) error {
			formato, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			_, conta, det, err := contaComPermissao(ctx, g, args[0])
			if err != nil {
				return err
			}
			if !det.PermiteExcluir {
				return fmt.Errorf("o painel não permite excluir esta conta%s", motivoBloqueio(det))
			}
			desc := descreverConta(conta.NomeBanco, conta.Agencia, conta.ContaCorrente)
			op := operacao{Risco: riscoAlto, ID: args[0], Resumo: "Excluir permanentemente a conta bancária " + desc,
				Consequencia: "Os vínculos da conta (extratos, classificações) se perdem, e um novo cadastro não os recupera."}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.ExcluirContaBancaria(ctx, snd, conta.ID) })
			if err != nil || !enviado {
				return err
			}
			situacao := "removido"
			if ainda, err := releConta(ctx, g, conta.ID, "", ""); err != nil {
				fmt.Fprintln(s.err, "aviso: pedido enviado, mas não foi possível reler a lista:", err)
				situacao = "enviado"
			} else if ainda != nil {
				fmt.Fprintln(s.err, "aviso: a conta ainda aparece; confira com ctbz contas-bancarias")
				situacao = "enviado"
			}
			return output.Write(s.out, formato, resultadoEscrita("remover", situacao, conta.ID).Add("conta", "Conta", desc))
		},
	}
	addWriteFlags(cmd)
	return cmd
}
