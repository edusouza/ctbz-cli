package cli

import (
	"errors"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// addCompetenciaFlag acrescenta --competencia AAAA-MM (obrigatória) a um comando.
func addCompetenciaFlag(cmd *cobra.Command) {
	cmd.Flags().String("competencia", "", "competência AAAA-MM (obrigatória)")
}

// competenciaFlag lê --competencia; ausente ou inválida é erro de uso.
func competenciaFlag(cmd *cobra.Command) (time.Time, error) {
	v, _ := cmd.Flags().GetString("competencia")
	if v == "" {
		return time.Time{}, usageError{errors.New("informe --competencia AAAA-MM")}
	}
	return parseMes(v)
}

// exibidas indexa as classificações com exibir=true numa competência.
func exibidas(cs []api.ContaUsuarioCompetencia) map[int64]api.ContaUsuarioCompetencia {
	m := map[int64]api.ContaUsuarioCompetencia{}
	for _, c := range cs {
		if c.Exibir {
			m[c.ID] = c
		}
	}
	return m
}

// classificacoesCaixa aplica o filtro do modal de lançamento do caixa: CONTA_USUARIO exibida
// na competência, do lado pedido (entrada = recebimento) e sem as contas vetadas para
// pagamento. lado nil traz os dois lados.
func classificacoesCaixa(cats []api.CategoriaVinculo, competencia []api.ContaUsuarioCompetencia, entrada *bool) []api.CategoriaVinculo {
	ok := exibidas(competencia)
	var out []api.CategoriaVinculo
	for _, c := range cats {
		id, err := strconv.ParseInt(c.IDTexto(), 10, 64)
		if err != nil || c.Tipo != api.VinculoContaUsuario {
			continue
		}
		if _, exibir := ok[id]; !exibir {
			continue
		}
		if entrada != nil && c.Entrada != *entrada {
			continue
		}
		if !c.Entrada && id == api.IDContaNaoPermitidaPagamento {
			continue
		}
		out = append(out, c)
	}
	return out
}

// vinculosCaixa filtra as guias e os sócios que servem de vínculo.
func vinculosCaixa(cats []api.CategoriaVinculo, tipo string) []api.CategoriaVinculo {
	var out []api.CategoriaVinculo
	for _, c := range cats {
		if c.Tipo == tipo {
			out = append(out, c)
		}
	}
	return out
}

func ladoCaixa(entrada bool) string {
	if entrada {
		return "recebimento"
	}
	return "pagamento"
}

func nomeVinculo(tipo string) any {
	switch tipo {
	case api.VinculoGuiaImposto:
		return "guia"
	case api.VinculoSocio:
		return "sócio"
	}
	return nil
}

func classificacoesCaixaList(cs []api.CategoriaVinculo) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "id", Header: "ID"},
		{Key: "descricao", Header: "Classificação"},
		{Key: "tipo", Header: "Tipo"},
		{Key: "vinculo", Header: "Exige"},
	}}
	for _, c := range cs {
		l.Append(c.IDTexto(), c.Descricao, ladoCaixa(c.Entrada), nomeVinculo(api.VinculoExigido(c.Descricao)))
	}
	return l
}

func vinculosCaixaList(cats []api.CategoriaVinculo) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "id", Header: "ID"},
		{Key: "descricao", Header: "Descrição"},
		{Key: "tipo", Header: "Tipo"},
	}}
	l.Append("0", "SEM GUIA", "guia")
	for _, tipo := range []string{api.VinculoGuiaImposto, api.VinculoSocio} {
		for _, c := range vinculosCaixa(cats, tipo) {
			l.Append(c.IDTexto(), c.Descricao, nomeVinculo(tipo))
		}
	}
	return l
}

func newCaixaContasCmd() *cobra.Command {
	var recebimento, pagamento, vinculos bool
	cmd := &cobra.Command{
		Use:   "contas",
		Short: "Lista as classificações aceitas nos lançamentos do caixa de uma competência",
		Long: `Lista as classificações (contas) que o caixa aceita na competência, com o mesmo filtro do
painel: só as exibidas no mês, do lado escolhido (recebimento ou pagamento). A coluna "Exige"
indica as classificações que pedem um vínculo: uma guia de imposto (contas "Impostos - …")
ou um sócio ("Sócios - Distribuição de Lucros Antecipados").

--vinculos lista as guias e os sócios aceitos como vínculo (use o id em --guia ou --socio
de ctbz caixa adicionar); "0" é SEM GUIA.`,
		Example: `  ctbz caixa contas --competencia 2026-09
  ctbz caixa contas --competencia 2026-09 --pagamento
  ctbz caixa contas --competencia 2026-09 --vinculos`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			mes, err := competenciaFlag(cmd)
			if err != nil {
				return err
			}
			if recebimento && pagamento {
				return usageError{errors.New("use --recebimento ou --pagamento, não os dois")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			ano, m := mes.Year(), int(mes.Month())
			cats, err := api.BuscarCaixaCategorias(cmd.Context(), g, ano, m)
			if err != nil {
				return err
			}
			if vinculos {
				return output.Write(s.out, f, vinculosCaixaList(cats))
			}
			comp, err := api.BuscarContasUsuarioCompetencia(cmd.Context(), g, ano, m, api.OrigemCaixa)
			if err != nil {
				return err
			}
			var entrada *bool
			if recebimento || pagamento {
				entrada = &recebimento
			}
			return output.Write(s.out, f, classificacoesCaixaList(classificacoesCaixa(cats, comp, entrada)))
		},
	}
	addCompetenciaFlag(cmd)
	cmd.Flags().BoolVar(&recebimento, "recebimento", false, "só as classificações de recebimento")
	cmd.Flags().BoolVar(&pagamento, "pagamento", false, "só as classificações de pagamento")
	cmd.Flags().BoolVar(&vinculos, "vinculos", false, "lista as guias e os sócios aceitos como vínculo")
	return cmd
}

// classificacoesExtrato aplica o filtro do modal de classificação do extrato: exibida, ativa e
// do tipo pedido (RECEITA para valor positivo, DESPESA para negativo); vazio traz os dois.
func classificacoesExtrato(cs []api.ContaUsuarioCompetencia, classificacao string) []api.ContaUsuarioCompetencia {
	var out []api.ContaUsuarioCompetencia
	for _, c := range cs {
		if !c.Exibir || c.Situacao != "ATIVO" {
			continue
		}
		if c.Classificacao != "RECEITA" && c.Classificacao != "DESPESA" {
			continue
		}
		if classificacao != "" && c.Classificacao != classificacao {
			continue
		}
		out = append(out, c)
	}
	return out
}

func classificacoesExtratoList(cs []api.ContaUsuarioCompetencia, nomes []api.ContaUsuario) *output.List {
	desc := map[int64]string{}
	for _, n := range nomes {
		desc[n.ID] = n.Descricao
	}
	l := &output.List{Columns: []output.Column{
		{Key: "id", Header: "ID"},
		{Key: "descricao", Header: "Classificação"},
		{Key: "classificacao", Header: "Tipo"},
		{Key: "exige_socio", Header: "Exige sócio"},
	}}
	for _, c := range cs {
		d := desc[c.ID]
		if d == "" {
			d = c.Descricao
		}
		l.Append(strconv.FormatInt(c.ID, 10), nilIfEmpty(d), c.Classificacao, api.IDsContasDeSocio[c.ID])
	}
	return l
}

func newExtratosContasCmd() *cobra.Command {
	var receita, despesa bool
	cmd := &cobra.Command{
		Use:   "contas",
		Short: "Lista as classificações aceitas nos lançamentos do extrato de uma competência",
		Long: `Lista as classificações que os lançamentos do extrato aceitam na competência, com o filtro
do painel: exibidas no mês, ativas, RECEITA (para entradas) ou DESPESA (para saídas). As
classificações com "Exige sócio" pedem --socio em ctbz extratos classificar.`,
		Example: `  ctbz extratos contas --competencia 2026-09
  ctbz extratos contas --competencia 2026-09 --despesa -o csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			mes, err := competenciaFlag(cmd)
			if err != nil {
				return err
			}
			if receita && despesa {
				return usageError{errors.New("use --receita ou --despesa, não os dois")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			comp, err := api.BuscarContasUsuarioCompetencia(cmd.Context(), g, mes.Year(), int(mes.Month()), api.OrigemExtrato)
			if err != nil {
				return err
			}
			nomes, err := api.BuscarContasUsuario(cmd.Context(), g)
			if err != nil {
				return err
			}
			tipo := ""
			if receita {
				tipo = "RECEITA"
			} else if despesa {
				tipo = "DESPESA"
			}
			return output.Write(s.out, f, classificacoesExtratoList(classificacoesExtrato(comp, tipo), nomes))
		},
	}
	addCompetenciaFlag(cmd)
	cmd.Flags().BoolVar(&receita, "receita", false, "só as classificações de entrada (RECEITA)")
	cmd.Flags().BoolVar(&despesa, "despesa", false, "só as classificações de saída (DESPESA)")
	return cmd
}
