package cli

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// alertaCritica marca uma pendência crítica do painel, que não tem prazo (ADR-0012).
const alertaCritica = "crítica"

// guiaPaga diz se a situação é de uma guia já paga, que o painel mantém entre as do mês. A
// API usa "PAGA" nas guias e "PAGO" no histórico; "PAGAR" não conta.
func guiaPaga(situacao string) bool {
	return strings.EqualFold(situacao, "PAGA") || strings.EqualFold(situacao, "PAGO")
}

func newResumoCmd() *cobra.Command {
	var failOnAtencao bool
	cmd := &cobra.Command{
		Use:   "resumo",
		Short: "Mostra numa lista só o que precisa de atenção",
		Long: `Junta numa lista só o que precisa de atenção: impostos em atraso e do mês, pendências
abertas, rotinas do mês ainda não realizadas, a mensalidade do mês e as pendências que o
painel marca como críticas. As consultas são feitas em paralelo.

A coluna secao indica a origem (impostos, pendencias, rotinas, mensalidade, painel) e a
coluna alerta segue a regra dos outros comandos: "vencida", "próxima" (até 7 dias) ou
"crítica" (pendência crítica do painel). Com --fail-on-atencao, o comando termina com
código 4 quando há algo vencido ou crítico, para alertas em cron.`,
		Example: `  ctbz resumo
  ctbz resumo -o json | jq '.[] | select(.alerta == "vencida")'
  ctbz resumo --fail-on-atencao -o csv > /dev/null || notify-send "Contabilizei precisa de atenção"`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			d, err := buscarResumo(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			l, atencao := resumoList(d)
			if err := output.Write(s.out, f, l); err != nil {
				return err
			}
			if f == output.FormatTable && len(l.Rows) == 0 {
				fmt.Fprintln(s.err, "Nada precisa de atenção.")
			}
			if failOnAtencao && atencao {
				return errAttention
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&failOnAtencao, "fail-on-atencao", false, "termina com código 4 se houver algo vencido ou crítico")
	return cmd
}

// dadosResumo reúne as respostas usadas pelo resumo.
type dadosResumo struct {
	guias      *api.GuiasAPagar
	pendencias []api.PendenciaEmpresa
	central    *api.CentralRotinas
}

// buscarResumo faz as três consultas em paralelo e falha se qualquer uma falhar.
func buscarResumo(ctx context.Context, g api.Getter) (*dadosResumo, error) {
	var d dadosResumo
	errs := make([]error, 3)
	var wg sync.WaitGroup
	run := func(i int, f func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = f()
		}()
	}
	run(0, func() (err error) { d.guias, err = api.BuscarGuiasAPagar(ctx, g); return })
	run(1, func() (err error) { d.pendencias, err = api.BuscarPendenciasEmpresa(ctx, g); return })
	run(2, func() (err error) { d.central, err = api.BuscarCentralRotinas(ctx, g); return })
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return &d, nil
}

// resumoList monta a lista e indica se há algo vencido ou crítico.
func resumoList(d *dadosResumo) (*output.List, bool) {
	l := &output.List{Columns: []output.Column{
		{Key: "secao", Header: "Seção"},
		{Key: "item", Header: "Item"},
		{Key: "prazo", Header: "Prazo"},
		{Key: "valor", Header: "Valor"},
		{Key: "situacao", Header: "Situação"},
		{Key: "alerta", Header: "Alerta"},
	}}
	atencao := false
	add := func(secao, item string, prazo output.Date, valor any, situacao string, alerta any) {
		if alerta == alertaVencida || alerta == alertaCritica {
			atencao = true
		}
		l.Append(secao, item, prazo, valor, nilIfEmpty(situacao), alerta)
	}

	for _, gr := range gruposDeGuias(d.guias, false)[:2] { // em atraso e deste mês
		for _, g := range gr.guias {
			if guiaPaga(g.Status.Label) {
				continue // continua na lista do mês, mas não pede ação
			}
			prazo, _ := output.ParseDate(g.VencimentoOriginal)
			alerta := alertaDePrazo(prazo, true)
			if gr.chave == "em_atraso" {
				alerta = alertaVencida
			}
			item := strings.TrimSpace(fmt.Sprint(g.Nome.Label, " ", competenciaOuVazio(g.Competencia)))
			add("impostos", item, prazo, moneyOrNil(g.Valor.Label), g.Status.Label, alerta)
		}
	}

	for _, p := range d.pendencias {
		if p.SituacaoPendencia.ID == situacaoFinalizada {
			continue
		}
		prazo, _ := output.ParseDate(p.DataLimite)
		add("pendencias", p.TipoPendencia.Titulo, prazo, nil, p.SituacaoPendencia.Descricao, alertaDePrazo(prazo, true))
	}

	hoje := now()
	for _, r := range d.central.Rotinas {
		prazo, _ := output.ParseDate(r.Prazo)
		if prazo.Year() != hoje.Year() || prazo.Month() != hoje.Month() || r.Tipo == "IMPOSTO" {
			continue // impostos já aparecem pelas guias
		}
		aberta := !statusConcluidos[r.Status]
		var valor any
		if r.Propriedades != nil {
			valor = moneyOrNil(r.Propriedades.ValorPagamento)
		}
		switch {
		case r.Tipo == "VENCIMENTO_MENSALIDADE":
			add("mensalidade", nomeDaRotina(r), prazo, valor, r.Status, alertaDePrazo(prazo, aberta))
		case aberta:
			add("rotinas", nomeDaRotina(r), prazo, valor, r.Status, alertaDePrazo(prazo, true))
		}
	}

	for _, grupo := range []struct {
		indicadores map[string]api.IndicadorPendencia
		situacao    string
		alerta      any
	}{{d.central.Pendencias.Criticas, "crítica", alertaCritica}, {d.central.Pendencias.Outras, "aviso", nil}} {
		for _, chave := range slices.Sorted(maps.Keys(grupo.indicadores)) {
			if grupo.indicadores[chave].PossuiPendencia {
				add("painel", nomeDoIndicador(chave), output.Date{}, nil, grupo.situacao, grupo.alerta)
			}
		}
	}
	return l, atencao
}

// competenciaOuVazio formata a competência de uma guia para compor o nome do item.
func competenciaOuVazio(s string) string {
	if c := competenciaDeTexto(s); c != nil {
		return fmt.Sprint(c)
	}
	return ""
}

// nomesDosIndicadores traduz as chaves de pendenciasCriticas/outrasPendencias.
var nomesDosIndicadores = map[string]string{
	"pendenciaAceiteTermoDebitos":               "Aceitar o termo de débitos",
	"pendenciaAquisicaoAtivoImobilizado":        "Informar aquisição de ativo imobilizado",
	"pendenciaCadastroContaBancaria":            "Cadastrar conta bancária",
	"pendenciaCadastroProlabore":                "Cadastrar pró-labore",
	"pendenciaCartaResponsabilidade":            "Assinar a carta de responsabilidade",
	"pendenciaCertificadoDigital":               "Certificado digital",
	"pendenciaConciliacaoFiscal":                "Conciliação fiscal",
	"pendenciaContratoAFAC":                     "Enviar contrato de AFAC",
	"pendenciaContratoDeEmprestimo":             "Enviar contrato de empréstimo",
	"pendenciaContratoDeFinanciamento":          "Enviar contrato de financiamento",
	"pendenciaContratoDeInvestimentoAnjo":       "Enviar contrato de investimento-anjo",
	"pendenciaContratoPrestacaoServico":         "Enviar contrato de prestação de serviço",
	"pendenciaControleDeIntermediacoes":         "Controle de intermediações",
	"pendenciaCredencialPrefeitura":             "Credencial da prefeitura",
	"pendenciaEstoque":                          "Informar estoque",
	"pendenciaExigibilidadeDocumental":          "Exigibilidade documental",
	"pendenciaExtratoAplicacaoFinanceira":       "Enviar extrato de aplicação financeira",
	"pendenciaImportacaoExtrato":                "Importar extrato bancário",
	"pendenciaImposto":                          "Impostos",
	"pendenciaInformeDeRendimentoInvestimentos": "Enviar informe de rendimentos de investimentos",
	"pendenciaIntegracaoAExpirar":               "Integração bancária a expirar",
	"pendenciaIntegracaoExpirada":               "Integração bancária expirada",
	"pendenciaMensalidade":                      "Mensalidade",
	"pendenciaProcuracaoEcac":                   "Procuração no e-CAC",
	"pendenciaTermoAdesaoTotalPass":             "Termo de adesão TotalPass",
}

// nomeDoIndicador traduz uma chave conhecida ou separa o camelCase de uma nova
// ("pendenciaNovaCoisa" → "Nova coisa").
func nomeDoIndicador(chave string) string {
	if nome, ok := nomesDosIndicadores[chave]; ok {
		return nome
	}
	var b strings.Builder
	for i, r := range strings.TrimPrefix(chave, "pendencia") {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteRune(' ')
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}
	return b.String()
}
