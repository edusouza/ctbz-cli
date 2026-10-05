package cli

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// Responsáveis por uma rotina.
const (
	responsavelEmpresa      = "empresa"
	responsavelContabilizei = "contabilizei"
)

// statusConcluidos são os status de rotina que não geram alerta de prazo.
var statusConcluidos = map[string]bool{"REALIZADA": true, "ENTREGUE": true, "CONCLUIDA": true, "FINALIZADA": true}

func newRotinasCmd() *cobra.Command {
	var mes string
	var failOnVencidas bool
	cmd := &cobra.Command{
		Use:   "rotinas",
		Short: "Lista as rotinas e obrigações do mês (da empresa e da Contabilizei)",
		Long: `Lista as rotinas do mês, como a Central de Rotinas do painel: as de responsabilidade da
empresa (importar extrato, pagar impostos e a mensalidade) e as obrigações entregues pela
Contabilizei (eSocial, DCTFWeb, EFD-Reinf…), com prazo, status e valor.

O mês é o do prazo (padrão: o mês atual). A API só devolve o mês anterior, o atual e o
próximo; outros meses saem vazios.

A coluna alerta marca prazos vencidos ("vencida") e que vencem em até 7 dias ("próxima").
A coluna pendencias traz os ids das pendências da rotina (ex.: lançamentos para reclassificar
com ctbz rotinas reclassificar).
Com --fail-on-vencidas, o comando termina com código 4 quando há rotina da empresa vencida.`,
		Example: `  ctbz rotinas
  ctbz rotinas --mes 2026-11
  ctbz rotinas --fail-on-vencidas -o json`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			ref := now()
			if mes != "" {
				if ref, err = time.Parse("2006-01", mes); err != nil {
					return usageError{fmt.Errorf("--mes deve ser AAAA-MM: %q", mes)}
				}
			}
			s := streamsOf(cmd)
			c, err := api.BuscarCentralRotinas(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			l, vencidas := rotinasList(c, ref)
			if err := output.Write(s.out, f, l); err != nil {
				return err
			}
			if failOnVencidas && vencidas > 0 {
				return errAttention
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&mes, "mes", "", "mês do prazo, AAAA-MM (padrão: o mês atual)")
	cmd.Flags().BoolVar(&failOnVencidas, "fail-on-vencidas", false, "termina com código 4 se houver rotina da empresa vencida")
	cmd.AddCommand(newRotinasReclassificarCmd())
	return cmd
}

// rotinasList monta as rotinas com prazo no mês de ref e conta as da empresa vencidas.
func rotinasList(c *api.CentralRotinas, ref time.Time) (*output.List, int) {
	l := &output.List{Columns: []output.Column{
		{Key: "responsavel", Header: "Responsável"},
		{Key: "rotina", Header: "Rotina"},
		{Key: "prazo", Header: "Prazo"},
		{Key: "status", Header: "Status"},
		{Key: "valor", Header: "Valor"},
		{Key: "alerta", Header: "Alerta"},
		{Key: "pendencias", Header: "Pendências"},
	}}
	noMes := func(d output.Date) bool { return d.Year() == ref.Year() && d.Month() == ref.Month() }
	vencidas := 0
	for _, r := range c.Rotinas {
		prazo, _ := output.ParseDate(r.Prazo)
		if !noMes(prazo) {
			continue
		}
		alerta := alertaDePrazo(prazo, !statusConcluidos[r.Status])
		if alerta == alertaVencida {
			vencidas++
		}
		var valor any
		if r.Propriedades != nil {
			valor = moneyOrNil(r.Propriedades.ValorPagamento)
		}
		var pendencias any
		if ids := r.IDsPendentes(); len(ids) > 0 {
			pendencias = ids
		}
		l.Append(responsavelEmpresa, nomeDaRotina(r), prazo, r.Status, valor, alerta, pendencias)
	}
	for _, r := range c.RotinasContabilizei {
		prazo, _ := output.ParseDate(r.Prazo)
		if !noMes(prazo) {
			continue
		}
		l.Append(responsavelContabilizei, nomeDaObrigacao(r.Titulo), prazo, r.Status, nil,
			alertaDePrazo(prazo, !statusConcluidos[r.Status]), nil)
	}
	return l, vencidas
}

// nomeDaRotina descreve uma rotina da empresa a partir do tipo e das propriedades.
func nomeDaRotina(r api.RotinaEmpresa) string {
	var titulo, mesRef string
	if p := r.Propriedades; p != nil {
		titulo, mesRef = p.TituloModal, p.MesReferencia
	}
	switch {
	case r.Tipo == "IMPORTACAO_EXTRATO" && mesRef != "":
		return "Importar extrato bancário de " + mesRef
	case r.Tipo == "IMPORTACAO_EXTRATO":
		return "Importar extrato bancário"
	case r.Tipo == "VENCIMENTO_MENSALIDADE":
		return "Mensalidade da Contabilizei"
	case titulo != "":
		return titulo
	}
	nome := strings.ToLower(strings.ReplaceAll(r.Tipo, "_", " "))
	if nome == "" {
		return ""
	}
	return strings.ToUpper(nome[:1]) + nome[1:]
}

var reParenteses = regexp.MustCompile(`\s*\([^)]*\)`)

// nomeDaObrigacao tira a primeira explicação entre parênteses do título:
// "EFD-Reinf (Escrituração…) - R2099 (Previdenciário)" → "EFD-Reinf - R2099 (Previdenciário)".
func nomeDaObrigacao(titulo string) string {
	if loc := reParenteses.FindStringIndex(titulo); loc != nil {
		titulo = titulo[:loc[0]] + titulo[loc[1]:]
	}
	return strings.TrimSpace(titulo)
}
