package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// tiposGerenciamento são os tipos de --tipo, no vocabulário da CLI.
var tiposGerenciamento = map[string]string{
	"salario-minimo": api.GerenciamentoSalarioMinimo,
	"teto-inss":      api.GerenciamentoTetoINSS,
	"personalizado":  api.GerenciamentoPersonalizado,
	"inteligente":    api.GerenciamentoInteligente,
}

const tiposGerenciamentoTexto = "salario-minimo, teto-inss, personalizado ou inteligente"

func newProlaboreDefinirCmd() *cobra.Command {
	var socio int64
	var tipoFlag, valorFlag, minimoFlag string
	cmd := &cobra.Command{
		Use:   "definir",
		Short: "Muda como o pró-labore de um sócio é definido (salário mínimo, teto, valor ou automático)",
		Long: `Muda a gestão do pró-labore de um sócio, como "Editar gestão" na central do sócio:

  salario-minimo  o salário mínimo vigente
  teto-inss       o teto do INSS
  personalizado   o valor de --valor, de pelo menos um salário mínimo
  inteligente     a Contabilizei calcula o valor todo mês (motor do Fator R); --minimo
                  define um piso opcional

O ID do sócio está em ctbz prolabore -o json. A Contabilizei decide a partir de que mês a
mudança vale: até o fechamento (em geral dia 25) vale para o mês atual; depois, para o
seguinte. A CLI avisa quando o painel diz que o mês está fechado.

Risco alto: muda INSS, IRRF e o Fator R; dá para alterar de novo enquanto o mês estiver
aberto. Aceita --yes e --dry-run.`,
		Example: `  ctbz prolabore definir --socio 1000000000000001 --tipo salario-minimo
  ctbz prolabore definir --socio 1000000000000001 --tipo personalizado --valor 3000,00
  ctbz prolabore definir --socio 1000000000000001 --tipo inteligente --minimo 1518,00`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			tipo, ok := tiposGerenciamento[tipoFlag]
			switch {
			case socio == 0:
				return usageError{errors.New("informe --socio (o ID está em ctbz prolabore -o json)")}
			case !ok:
				return usageError{fmt.Errorf("--tipo deve ser %s: %q", tiposGerenciamentoTexto, tipoFlag)}
			case tipo == api.GerenciamentoPersonalizado && valorFlag == "":
				return usageError{errors.New("--tipo personalizado exige --valor")}
			case valorFlag != "" && tipo != api.GerenciamentoPersonalizado:
				return usageError{errors.New("--valor só vale com --tipo personalizado")}
			case minimoFlag != "" && tipo != api.GerenciamentoInteligente:
				return usageError{errors.New("--minimo só vale com --tipo inteligente")}
			}
			valor, err := reaisOuNil("--valor", valorFlag)
			if err != nil {
				return err
			}
			minimo, err := reaisOuNil("--minimo", minimoFlag)
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			antes, err := socioProlabore(cmd, g, socio)
			if err != nil {
				return err
			}
			gestao, err := api.BuscarGestaoSocio(cmd.Context(), g, socio)
			if err != nil {
				return err
			}
			dash, err := api.BuscarProlaboreDashboard(cmd.Context(), g)
			if err != nil {
				return err
			}
			alt, err := api.NovaAlteracaoGestao(*gestao, tipo, valor, minimo)
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoAlto, ID: fmt.Sprint(socio),
				Resumo: fmt.Sprintf("Mudar o pró-labore de %s de %s (%s) para %s", antes.Nome,
					gestaoTexto(gestao.TipoGerenciamentoProlabore, nil), formatBRLOrDash(antes.ValorProlabore), gestaoTexto(tipo, alt.ValorProlabore)),
				Consequencia: consequenciaGestao(*gestao, alt, dash)}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return api.AlterarGestaoSocio(cmd.Context(), snd, socio, alt)
			})
			if err != nil || !enviado {
				return err
			}
			depois, err := socioProlabore(cmd, g, socio)
			if err != nil {
				return err
			}
			relida, err := api.BuscarGestaoSocio(cmd.Context(), g, socio)
			if err != nil {
				return err
			}
			situacao := "enviado"
			if relida.TipoGerenciamentoProlabore == tipo {
				situacao = "alterado"
			}
			rec := resultadoEscrita("definir", situacao, fmt.Sprint(socio)).Add("socio", "Sócio", antes.Nome).
				Add("gestao_anterior", "Gestão anterior", nilIfEmpty(gestao.TipoGerenciamentoProlabore)).
				Add("valor_anterior", "Valor anterior", moneyOrNil(antes.ValorProlabore)).
				Add("gestao", "Gestão", nilIfEmpty(relida.TipoGerenciamentoProlabore)).
				Add("valor", "Valor", moneyOrNil(depois.ValorProlabore))
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().Int64Var(&socio, "socio", 0, "ID do sócio (de ctbz prolabore -o json)")
	cmd.Flags().StringVar(&tipoFlag, "tipo", "", tiposGerenciamentoTexto)
	cmd.Flags().StringVar(&valorFlag, "valor", "", "pró-labore personalizado, em reais (3000,00)")
	cmd.Flags().StringVar(&minimoFlag, "minimo", "", "piso da gestão inteligente, em reais (opcional)")
	addWriteFlags(cmd)
	return cmd
}

// reaisOuNil lê um valor em reais de uma flag; vazio é nil.
func reaisOuNil(flag, v string) (*float64, error) {
	if v == "" {
		return nil, nil
	}
	c, err := parseValor(v)
	if err != nil {
		return nil, usageError{fmt.Errorf("%s: %w", flag, err)}
	}
	r := float64(c) / 100
	return &r, nil
}

// socioProlabore acha o sócio na central de pró-labore.
func socioProlabore(cmd *cobra.Command, g api.Getter, id int64) (*api.SocioProlabore, error) {
	c, err := api.BuscarProlaboreCentral(cmd.Context(), g)
	if err != nil {
		return nil, err
	}
	for i := range c.Socios {
		if c.Socios[i].ID == id {
			return &c.Socios[i], nil
		}
	}
	return nil, fmt.Errorf("sócio %d não encontrado na central de pró-labore (veja ctbz prolabore -o json)", id)
}

func gestaoTexto(tipo string, valor *float64) string {
	if tipo == "" {
		tipo = "gestão não informada"
	}
	if valor == nil {
		return tipo
	}
	return tipo + " de " + output.FormatBRL(*valor)
}

func formatBRLOrDash(v *float64) string {
	if v == nil {
		return "sem valor"
	}
	return output.FormatBRL(*v)
}

// consequenciaGestao explica o efeito da mudança: impostos, gestão inteligente da empresa e
// a partir de quando vale.
func consequenciaGestao(atual api.GestaoSocio, alt api.AlteracaoGestao, dash *api.ProlaboreDashboard) string {
	linhas := []string{"Muda o INSS, o IRRF e o Fator R do sócio; dá para alterar de novo enquanto o mês estiver aberto."}
	switch {
	case alt.SairGestaoInteligente && atual.TipoGerenciamentoProlabore == api.GerenciamentoInteligente:
		linhas = append(linhas, "A empresa sai da gestão inteligente: era o último sócio nela.")
	case alt.TipoGerenciamento != api.GerenciamentoInteligente && atual.QtdSocioGestaoInteligente > 1:
		linhas = append(linhas, "A empresa permanece na gestão inteligente para os outros sócios.")
	}
	if dash.PodeAlterar != nil && !*dash.PodeAlterar {
		linhas = append(linhas, "O pró-labore deste mês está fechado: a mudança vale a partir do próximo mês.")
	}
	return strings.Join(linhas, "\n")
}
