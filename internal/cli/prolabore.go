package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newProlaboreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prolabore",
		Short: "Mostra o pró-labore vigente por sócio e o tipo de gerenciamento",
		Long: `Mostra o pró-labore da empresa como a central de pró-labore do painel: tipo de
gerenciamento (ex.: INTELIGENTE, quando a Contabilizei calcula o valor ideal), total, valor e
competências do card do painel e, para cada sócio, valor, se recebe pró-labore, se é o
responsável na Receita, data da última atualização e quantidade de dependentes.

O histórico mensal está em "ctbz prolabore historico"; para mudar a gestão de um sócio:
ctbz prolabore definir.`,
		Example: `  ctbz prolabore
  ctbz prolabore -o json | jq '.socios[] | {nome, valor}'`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			c, err := api.BuscarProlaboreCentral(cmd.Context(), g)
			if err != nil {
				return err
			}
			d, err := api.BuscarProlaboreDashboard(cmd.Context(), g)
			if err != nil {
				return err
			}
			return output.Write(s.out, f, prolaboreRecord(c, d))
		},
	}
	cmd.AddCommand(newProlaboreHistoricoCmd(), newProlaboreParametrosCmd(), newProlaboreFatorRCmd(), newProlaboreDefinirCmd(),
		newProlaboreZerarCmd(), newGestaoInteligenteCmd())
	return cmd
}

func prolaboreRecord(c *api.ProlaboreCentral, d *api.ProlaboreDashboard) *output.Record {
	socios := make([]output.Record, 0, len(c.Socios))
	for _, s := range c.Socios {
		atualizado, _ := output.ParseDate(s.DataUltimaAtualizacao)
		r := output.Record{}
		r.Add("nome", "Nome", s.Nome)
		r.Add("cpf", "CPF", output.NewCPF(s.CPF))
		r.Add("valor", "Valor", moneyOrNil(s.ValorProlabore))
		r.Add("possui_prolabore", "Recebe pró-labore", s.PossuiProlabore)
		r.Add("responsavel_receita", "Responsável na Receita", s.ResponsavelReceita)
		r.Add("gestao", "Gestão", nilIfEmpty(s.InsightGestao))
		r.Add("atualizado", "Atualizado em", atualizado)
		r.Add("dependentes", "Dependentes", len(s.NomesDependentes))
		r.Add("id", "ID", fmt.Sprint(s.ID))
		socios = append(socios, r)
	}
	rec := &output.Record{}
	rec.Add("gerenciamento", "Gerenciamento", nilIfEmpty(c.TipoGerenciamento))
	rec.Add("total", "Total", moneyOrNil(c.TotalProLabore))
	rec.Add("valor_card", "Valor no painel", moneyOrNil(d.ValorProlabore))
	rec.Add("competencia_atual", "Competência atual", competenciaDeValor(d.CompetenciaAtual))
	rec.Add("competencia_anterior", "Competência anterior", competenciaDeValor(d.CompetenciaAnterior))
	rec.Add("calculando", "Calculando", d.Calculando)
	rec.Add("indisponivel", "Indisponível", c.ProlaboreIndisponivel)
	rec.Add("socios", "Sócios", socios)
	return rec
}

// competenciaDeValor converte uma competência de formato não verificado: texto como em
// competenciaDeTexto; outros valores voltam como texto.
func competenciaDeValor(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		return competenciaDeTexto(x)
	}
	return fmt.Sprint(v)
}

func newProlaboreHistoricoCmd() *cobra.Command {
	var ano int
	var socio int64
	cmd := &cobra.Command{
		Use:   "historico",
		Short: "Lista o histórico mensal de pró-labore dos sócios",
		Long: `Lista, por competência, o pró-labore e os descontos (INSS e IRRF) de cada sócio, como a
tabela de histórico da central de pró-labore. A API devolve todo o histórico de um sócio de
uma vez; --ano filtra o resultado e --socio escolhe um sócio pelo ID (o padrão é todos).`,
		Example: `  ctbz prolabore historico
  ctbz prolabore historico --ano 2026 -o csv`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			ids := []int64{socio}
			if socio == 0 {
				c, err := api.BuscarProlaboreCentral(cmd.Context(), g)
				if err != nil {
					return err
				}
				ids = ids[:0]
				for _, so := range c.Socios {
					ids = append(ids, so.ID)
				}
			}
			var meses []api.ProlaboreMes
			for _, id := range ids {
				h, err := api.BuscarProlaboreHistorico(cmd.Context(), g, id)
				if err != nil {
					return err
				}
				meses = append(meses, h...)
			}
			return output.Write(s.out, f, historicoProlaboreList(meses, ano))
		},
	}
	cmd.Flags().IntVar(&ano, "ano", 0, "só as competências deste ano")
	cmd.Flags().Int64Var(&socio, "socio", 0, "ID do sócio (de \"ctbz prolabore -o json\"); padrão: todos")
	return cmd
}

func historicoProlaboreList(meses []api.ProlaboreMes, ano int) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "competencia", Header: "Competência"},
		{Key: "socio", Header: "Sócio"},
		{Key: "prolabore", Header: "Pró-labore"},
		{Key: "descontos", Header: "Descontos"},
	}}
	for _, m := range meses {
		comp := competenciaDeTexto(m.Competencia)
		if ano != 0 && !strings.HasSuffix(fmt.Sprint(comp), fmt.Sprint(ano)) {
			continue
		}
		l.Append(comp, m.Nome, brlOuTexto(m.Prolabore), brlOuTexto(m.Descontos))
	}
	return l
}

// brlOuTexto converte um valor já formatado em reais; outro texto volta como veio.
func brlOuTexto(s string) any {
	if v, ok := output.ParseBRL(s); ok {
		return v
	}
	return nilIfEmpty(s)
}

func newProlaboreParametrosCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "parametros",
		Short: "Mostra os valores usados no cálculo do pró-labore (INSS e IRRF)",
		Long: `Mostra os parâmetros que a Contabilizei usa para calcular o pró-labore: salário mínimo
(pró-labore mínimo), alíquota do INSS do sócio, contribuição máxima ao INSS, pró-labore a
partir do qual a contribuição atinge o teto e o valor a partir do qual incide IRRF.`,
		Example: `  ctbz prolabore parametros`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			p, err := api.BuscarProlaboreParametros(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			rec := &output.Record{}
			rec.Add("salario_minimo", "Salário mínimo", brlOuTexto(p.SalarioMinimo))
			rec.Add("inss_aliquota", "Alíquota do INSS (%)", p.PorcentagemInss)
			rec.Add("inss_contribuicao_maxima", "Contribuição máxima ao INSS", brlOuTexto(p.ValorMaximoContribuicaoInss))
			rec.Add("prolabore_teto_inss", "Pró-labore no teto do INSS", brlOuTexto(p.ValorMaximoProlabore))
			rec.Add("irrf_a_partir_de", "IRRF a partir de", brlOuTexto(p.ValorMinimoProlaboreParaIncidenciaIrrf))
			return output.Write(s.out, f, rec)
		},
	}
}
