package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// anoInforme é o --ano dos comandos do informe: o padrão é o ano anterior (o da declaração).
func anoInforme(ano int) int {
	if ano == 0 {
		return now().Year() - 1
	}
	return ano
}

func newInformeRestricoesCmd() *cobra.Command {
	var ano int
	cmd := &cobra.Command{
		Use:   "restricoes",
		Short: "Mostra o que bloqueia o informe de rendimentos de um ano e o que falta aceitar",
		Long: `Mostra as restrições do informe de rendimentos de um ano-calendário: débitos federais,
pendência documental (e como regularizá-la), reabertura do balanço e se a carta de
responsabilidade ainda precisa ser aceita. O padrão de --ano é o ano anterior.

Para aceitar: ctbz lucros informe aceitar; para decidir sobre as restrições:
ctbz lucros informe decidir.`,
		Example: `  ctbz lucros informe restricoes --ano 2025`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			ano := anoInforme(ano)
			s := streamsOf(cmd)
			g := sessionGetter{s}
			r, err := api.BuscarRestricoesInforme(cmd.Context(), g, ano)
			if err != nil {
				return err
			}
			c, err := api.BuscarCartaResponsabilidadeInforme(cmd.Context(), g, ano)
			if err != nil {
				return err
			}
			pd, df := r.Restricoes.PendenciaDocumental, r.Restricoes.DebitosFederais
			rec := &output.Record{}
			rec.Add("ano", "Ano", ano).
				Add("carta_pendente", "Carta de responsabilidade a aceitar", c.DeveAssinarCartaResponsabilidade).
				Add("debitos_federais", "Débitos federais", df.PossuiPendencia).
				Add("debitos", "Débitos apontados", len(df.Debitos)).
				Add("divergencia_contabil_fiscal", "Divergência contábil e fiscal", df.DivergenciaContabilFiscal).
				Add("pendencia_documental", "Pendência documental", pd.PossuiPendencia).
				Add("pendencias", "Pendências apontadas", len(pd.Pendencias)).
				Add("fluxo_regularizacao", "Como regularizar", nilIfEmpty(pd.FluxoRegularizacao)).
				Add("valor_servico_adicional", "Custo do serviço adicional", moneyOrNil(pd.ValorServicoAdicional)).
				Add("reabertura_balanco", "Reabertura do balanço", nilIfEmpty(r.ProcessoReabertura.Status))
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().IntVar(&ano, "ano", 0, "ano-calendário (padrão: o ano anterior)")
	return cmd
}

func newInformeCartaCmd() *cobra.Command {
	var ano int
	cmd := &cobra.Command{
		Use:     "carta",
		Short:   "Mostra a carta de responsabilidade do informe de rendimentos de um ano",
		Example: `  ctbz lucros informe carta --ano 2025`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			ano := anoInforme(ano)
			s := streamsOf(cmd)
			c, err := api.BuscarCartaResponsabilidadeInforme(cmd.Context(), sessionGetter{s}, ano)
			if err != nil {
				return err
			}
			if !c.DeveAssinarCartaResponsabilidade {
				fmt.Fprintf(s.err, "A carta de %d não precisa ser aceita.\n", ano)
			}
			_, err = fmt.Fprint(s.out, htmlParaTexto(c.HTML))
			return err
		},
	}
	cmd.Flags().IntVar(&ano, "ano", 0, "ano-calendário (padrão: o ano anterior)")
	return cmd
}

func newInformeAceitarCmd() *cobra.Command {
	var ano int
	cmd := &cobra.Command{
		Use:   "aceitar carta-responsabilidade|termo-debitos",
		Short: "Aceita a carta de responsabilidade ou o termo de débitos federais do informe",
		Long: `Dá os aceites que o painel pede para liberar o informe de rendimentos de um ano:

  carta-responsabilidade  a carta exigida para ver o informe (o texto aparece antes da
                          confirmação; também em ctbz lucros informe carta)
  termo-debitos           a ciência dos débitos federais da empresa, quando há débitos

Risco alto: são declarações sem desfazer, e a CLI pede "confirmo". O padrão de --ano é o ano
anterior. Aceita --yes e --dry-run.`,
		Example: `  ctbz lucros informe aceitar carta-responsabilidade --ano 2025
  ctbz lucros informe aceitar termo-debitos --ano 2025`,
		Args:      exactArgs(1, "carta-responsabilidade ou termo-debitos"),
		ValidArgs: []string{"carta-responsabilidade", "termo-debitos"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "carta-responsabilidade" && args[0] != "termo-debitos" {
				return usageError{fmt.Errorf("aceite deve ser carta-responsabilidade ou termo-debitos: %q", args[0])}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			ano := anoInforme(ano)
			s := streamsOf(cmd)
			g := sessionGetter{s}
			acao := "aceitar " + args[0]
			if args[0] == "carta-responsabilidade" {
				return aceitarCartaInforme(cmd, f, g, ano, acao)
			}
			r, err := api.BuscarRestricoesInforme(cmd.Context(), g, ano)
			if err != nil {
				return err
			}
			df := r.Restricoes.DebitosFederais
			if !df.PossuiPendencia {
				fmt.Fprintf(s.err, "O informe de %d não aponta débitos federais; não há termo a aceitar.\n", ano)
				return output.Write(s.out, f, resultadoEscrita(acao, "sem mudança", fmt.Sprint(ano)))
			}
			fmt.Fprintf(s.err, "Débitos federais apontados no informe de %d: %d\n", ano, len(df.Debitos))
			for _, d := range df.Debitos {
				fmt.Fprintln(s.err, "  "+semControle(string(d)))
			}
			op := operacao{Risco: riscoAlto, ID: fmt.Sprint(ano), Resumo: fmt.Sprintf("Declarar ciência dos débitos federais no informe de %d", ano),
				Consequencia: "Registra que você está ciente dos débitos federais da empresa; não dá para desfazer."}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroAdmin(api.AceitarTermoDebitos(cmd.Context(), snd, ano, api.AceiteTermoPrimeiraVez))
			})
			if err != nil || !enviado {
				return err
			}
			return output.Write(s.out, f, resultadoEscrita(acao, "enviado", fmt.Sprint(ano)))
		},
	}
	cmd.Flags().IntVar(&ano, "ano", 0, "ano-calendário (padrão: o ano anterior)")
	addWriteFlags(cmd)
	return cmd
}

// aceitarCartaInforme mostra a carta, pede confirmação e relê se ela ainda precisa de aceite.
func aceitarCartaInforme(cmd *cobra.Command, f output.Format, g api.Getter, ano int, acao string) error {
	s := streamsOf(cmd)
	c, err := api.BuscarCartaResponsabilidadeInforme(cmd.Context(), g, ano)
	if err != nil {
		return err
	}
	if !c.DeveAssinarCartaResponsabilidade {
		fmt.Fprintf(s.err, "A carta de %d não precisa ser aceita; nada foi enviado.\n", ano)
		return output.Write(s.out, f, resultadoEscrita(acao, "sem mudança", fmt.Sprint(ano)))
	}
	fmt.Fprint(s.err, htmlParaTexto(c.HTML))
	op := operacao{Risco: riscoAlto, ID: fmt.Sprint(ano), Resumo: fmt.Sprintf("Aceitar a carta de responsabilidade do informe de %d (texto acima)", ano),
		Consequencia: "É uma declaração legal; não dá para desfazer."}
	enviado, err := escrever(cmd, op, func(snd api.Sender) error {
		return erroAdmin(api.AceitarCartaResponsabilidadeInforme(cmd.Context(), snd, ano))
	})
	if err != nil || !enviado {
		return err
	}
	relida, err := api.BuscarCartaResponsabilidadeInforme(cmd.Context(), g, ano)
	if err != nil {
		return err
	}
	situacao := "enviado"
	if !relida.DeveAssinarCartaResponsabilidade {
		situacao = "aceita"
	}
	return output.Write(s.out, f, resultadoEscrita(acao, situacao, fmt.Sprint(ano)))
}

// decisaoInforme é uma escolha diante das restrições do informe.
type decisaoInforme struct {
	flag, aceite, resumo, consequencia string
	exigeDebitos                       bool // senão, exige pendência documental
}

var decisoesInforme = []decisaoInforme{
	{"regularizar-pendencia", api.AceiteRegularizarPendencia, "Regularizar a pendência documental",
		"A Contabilizei passa a esperar os documentos, enviados pelo formulário de atendimento; não dá para desfazer.", false},
	{"nao-regularizar-pendencia", api.AceiteNaoRegularizarPendencia, "Não regularizar a pendência documental",
		"Sem regularizar a pendência, não haverá distribuição de lucros no informe; não dá para desfazer.", false},
	{"nao-distribuir-lucros", api.AceiteNaoDistribuirLucros, "Não distribuir lucros",
		"O informe fica sem distribuição de lucros isenta para os sócios; não dá para desfazer.", true},
}

func newInformeDecidirCmd() *cobra.Command {
	var ano int
	marcadas := make([]bool, len(decisoesInforme))
	cmd := &cobra.Command{
		Use:   "decidir",
		Short: "Decide o que fazer com a pendência documental ou os débitos do informe",
		Long: `Registra a decisão que o painel pede diante das restrições do informe de rendimentos:

  --regularizar-pendencia      regularizar a pendência documental (a CLI mostra o formulário
                               de atendimento onde enviar os documentos)
  --nao-regularizar-pendencia  não regularizar: não haverá distribuição de lucros
  --nao-distribuir-lucros      não distribuir lucros no ano (diante de débitos federais)

As duas primeiras exigem pendência documental e a última, débitos federais (ver ctbz lucros
informe restricoes). Risco alto: decide se haverá distribuição de lucros isenta e não dá para
desfazer. O padrão de --ano é o ano anterior. Aceita --yes e --dry-run.`,
		Example: `  ctbz lucros informe decidir --ano 2025 --regularizar-pendencia`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if contar(marcadas) != 1 {
				return usageError{errors.New("escolha uma decisão: --regularizar-pendencia, --nao-regularizar-pendencia ou --nao-distribuir-lucros")}
			}
			var d decisaoInforme
			for i, m := range marcadas {
				if m {
					d = decisoesInforme[i]
				}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			ano := anoInforme(ano)
			s := streamsOf(cmd)
			r, err := api.BuscarRestricoesInforme(cmd.Context(), sessionGetter{s}, ano)
			if err != nil {
				return err
			}
			if d.exigeDebitos && !r.Restricoes.DebitosFederais.PossuiPendencia {
				return fmt.Errorf("o informe de %d não aponta débitos federais (veja ctbz lucros informe restricoes)", ano)
			}
			if !d.exigeDebitos && !r.Restricoes.PendenciaDocumental.PossuiPendencia {
				return fmt.Errorf("o informe de %d não tem pendência documental (veja ctbz lucros informe restricoes)", ano)
			}
			op := operacao{Risco: riscoAlto, ID: fmt.Sprint(ano), Resumo: fmt.Sprintf("%s no informe de %d", d.resumo, ano), Consequencia: d.consequencia}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroAdmin(api.RegistrarAceiteInforme(cmd.Context(), snd, ano, d.aceite))
			})
			if err != nil || !enviado {
				return err
			}
			if d.aceite == api.AceiteRegularizarPendencia {
				fmt.Fprintln(s.err, "Envie os documentos pelo formulário de atendimento: "+api.FormularioRegularizarPendencias)
			}
			return output.Write(s.out, f, resultadoEscrita("decidir", "enviado", fmt.Sprint(ano)).Add("decisao", "Decisão", d.flag))
		},
	}
	cmd.Flags().IntVar(&ano, "ano", 0, "ano-calendário (padrão: o ano anterior)")
	for i, d := range decisoesInforme {
		cmd.Flags().BoolVar(&marcadas[i], d.flag, false, strings.ToLower(d.resumo[:1])+d.resumo[1:])
	}
	addWriteFlags(cmd)
	return cmd
}

func contar(bs []bool) int {
	n := 0
	for _, b := range bs {
		if b {
			n++
		}
	}
	return n
}
