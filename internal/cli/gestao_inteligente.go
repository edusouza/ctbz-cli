package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// avisoSairMotor é o aviso do painel ao sair do cálculo automático.
const avisoSairMotor = "A partir de agora, os sócios devem ajustar o valor do pró-labore todos os meses, até o penúltimo dia do mês, para valer no mês seguinte."

func newGestaoInteligenteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gestao-inteligente",
		Short: "Põe a empresa no cálculo automático do pró-labore ou a tira dele",
		Long: `Gestão inteligente é o cálculo automático do pró-labore: a Contabilizei define o valor
todo mês para manter o Fator R no melhor ponto (ver ctbz prolabore fator-r).

  ativar  põe a empresa na gestão inteligente (só se ela for elegível)
  sair    tira a empresa: os sócios passam a ajustar o pró-labore todo mês

Risco alto: muda quem decide o valor do pró-labore. Aceitam --yes e --dry-run.`,
		Args: exactArgs(0, "ativar ou sair"),
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newGestaoInteligenteAcaoCmd(true), newGestaoInteligenteAcaoCmd(false))
	return cmd
}

func newGestaoInteligenteAcaoCmd(ativar bool) *cobra.Command {
	use, short, acao := "sair", "Tira a empresa do cálculo automático do pró-labore", "gestao-inteligente sair"
	if ativar {
		use, short, acao = "ativar", "Põe a empresa no cálculo automático do pró-labore", "gestao-inteligente ativar"
	}
	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Example: "  ctbz prolabore gestao-inteligente " + use,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			c, ativa, err := estadoGestaoInteligente(cmd, g)
			if err != nil {
				return err
			}
			if ativa == ativar {
				fmt.Fprintf(s.err, "A empresa já está %s da gestão inteligente; nada foi enviado.\n", map[bool]string{true: "dentro", false: "fora"}[ativa])
				return output.Write(s.out, f, resultadoEscrita(acao, "sem mudança", nil).Add("gestao_inteligente", "Gestão inteligente", ativa))
			}
			op := operacao{Risco: riscoAlto, Resumo: "Sair da gestão inteligente (cálculo automático do pró-labore)",
				Consequencia: avisoSairMotor + " Para voltar: ctbz prolabore gestao-inteligente ativar."}
			enviar := api.SairGestaoInteligente
			if ativar {
				if !c.ElegivelNoMotor {
					return errors.New("a empresa não é elegível à gestão inteligente (veja ctbz prolabore fator-r)")
				}
				op = operacao{Risco: riscoAlto, Resumo: "Ativar a gestão inteligente (cálculo automático do pró-labore)",
					Consequencia: "A Contabilizei passa a definir o pró-labore de todos os sócios todo mês. Para voltar: ctbz prolabore gestao-inteligente sair."}
				enviar = api.AtivarGestaoInteligente
			}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return enviar(cmd.Context(), snd) })
			if err != nil || !enviado {
				return err
			}
			_, agora, err := estadoGestaoInteligente(cmd, g)
			if err != nil {
				return err
			}
			situacao := "enviado"
			if agora == ativar {
				situacao = map[bool]string{true: "ativada", false: "desativada"}[ativar]
			}
			return output.Write(s.out, f, resultadoEscrita(acao, situacao, nil).Add("gestao_inteligente", "Gestão inteligente", agora))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

// estadoGestaoInteligente lê a central e os parâmetros: a empresa está na gestão inteligente
// quando a central diz INTELIGENTE ou os parâmetros dizem que ela está no motor do Fator R.
func estadoGestaoInteligente(cmd *cobra.Command, g api.Getter) (*api.ProlaboreCentral, bool, error) {
	c, err := api.BuscarProlaboreCentral(cmd.Context(), g)
	if err != nil {
		return nil, false, err
	}
	p, err := api.BuscarProlaboreParametros(cmd.Context(), g)
	if err != nil {
		return nil, false, err
	}
	return c, c.TipoGerenciamento == api.GerenciamentoInteligente || p.IsMotorFatorR, nil
}
