package cli

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// etapasChecklist dá o rótulo de cada etapa e se o painel deixa concluí-la à mão (inferido
// dos botões marcarConcluido/concluirTarefa/agendarEConcluir; as demais o sistema conclui).
var etapasChecklist = map[string]struct {
	rotulo string
	manual bool
}{
	"CADASTRO_CONTA_PJ":             {"Abrir a conta PJ", false},
	"CADASTRE_PRO_LABORE":           {"Cadastrar o pró-labore", false},
	"GERAR_PROCURACAO_VIRTUAL":      {"Gerar a procuração eletrônica (ctbz pendencias procuracao)", false},
	"ACEITE_TERMOS_MULTIBENEFICIOS": {"Aceitar os termos dos benefícios", false},
	"REUNIAO_BOAS_VINDAS":           {"Reunião de boas-vindas", true},
	"LIVE_EMISSAO_NOTAS_FISCAIS":    {"Live: emissão de notas fiscais", true},
	"LIVE_VIDA_COM_CNPJ":            {"Live: a vida com CNPJ", true},
}

func etapasManuais() string {
	var ms []string
	for nome, e := range etapasChecklist {
		if e.manual {
			ms = append(ms, nome)
		}
	}
	sort.Strings(ms)
	return strings.Join(ms, ", ")
}

func newPrimeirosPassosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "primeiros-passos",
		Short: "Lista as tarefas do checklist de primeiros passos do painel",
		Long: `Lista as tarefas do checklist "Primeiros passos" do painel com a situação de cada uma e se
dá para concluí-la pela CLI (ctbz primeiros-passos concluir ETAPA). As demais são concluídas
pelo sistema quando a ação é feita (abrir a conta PJ, cadastrar o pró-labore…).`,
		Example: `  ctbz primeiros-passos
  ctbz primeiros-passos concluir REUNIAO_BOAS_VINDAS`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			c, err := api.BuscarChecklist(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{
				{Key: "etapa", Header: "Etapa"},
				{Key: "tarefa", Header: "Tarefa"},
				{Key: "situacao", Header: "Situação"},
				{Key: "concluida", Header: "Concluída"},
				{Key: "conclusao_manual", Header: "Concluir pela CLI"},
			}}
			for _, e := range c.Etapas {
				def := etapasChecklist[e.Nome]
				l.Append(e.Nome, nilIfEmpty(def.rotulo), e.Situacao, e.Concluida(), def.manual && !e.Concluida())
			}
			return output.Write(s.out, f, l)
		},
	}
	cmd.AddCommand(newPrimeirosPassosConcluirCmd(), newPrimeirosPassosAlternarCmd(true), newPrimeirosPassosAlternarCmd(false))
	return cmd
}

// jaConcluida reconhece o 304 de concluir-etapa: a etapa já estava concluída.
func jaConcluida(err error) bool {
	var we *ctbz.WriteError
	return errors.As(err, &we) && we.Status == http.StatusNotModified
}

func newPrimeirosPassosConcluirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "concluir ETAPA",
		Short: "Marca uma tarefa dos primeiros passos como concluída",
		Long: `Marca como concluída uma tarefa do checklist que o painel deixa concluir à mão
(` + etapasManuais() + `). Risco baixo, mas não há como desconcluir.
Se a tarefa já estava concluída, a CLI avisa e termina com código 0.

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz primeiros-passos concluir LIVE_VIDA_COM_CNPJ`,
		Args:    exactArgs(1, "a ETAPA"),
		RunE: func(cmd *cobra.Command, args []string) error {
			etapa := strings.ToUpper(args[0])
			def, ok := etapasChecklist[etapa]
			if !ok {
				return usageError{fmt.Errorf("etapa desconhecida %q; veja ctbz primeiros-passos", args[0])}
			}
			if !def.manual {
				return usageError{fmt.Errorf("a etapa %s é concluída pelo sistema, não à mão (manuais: %s)", etapa, etapasManuais())}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			situacao := "concluido"
			op := operacao{Risco: riscoBaixo, ID: etapa, Resumo: fmt.Sprintf("Concluir a tarefa %q dos primeiros passos", def.rotulo)}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				err := api.ConcluirEtapa(ctx, snd, etapa)
				if jaConcluida(err) {
					situacao = "ja_concluido"
					fmt.Fprintln(s.err, "A tarefa já estava concluída.")
					return nil
				}
				return err
			})
			if err != nil || !enviado {
				return err
			}
			return output.Write(s.out, f, resultadoEscrita("concluir", situacao, etapa).Add("tarefa", "Tarefa", def.rotulo))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

// newPrimeirosPassosAlternarCmd cria "dispensar" (true) ou "reativar" (false).
func newPrimeirosPassosAlternarCmd(dispensar bool) *cobra.Command {
	use, short, resumo, acao, situacao := "reativar", "Volta a mostrar as tarefas dos primeiros passos",
		"Reativar as tarefas dos primeiros passos", "reativar", "reativado"
	enviar := api.ReativarChecklist
	if dispensar {
		use, short, resumo, acao, situacao = "dispensar", "Oculta o card de primeiros passos do painel",
			"Dispensar o card de primeiros passos", "dispensar", "dispensado"
		enviar = api.DispensarChecklist
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long: short + ` (risco baixo: dispensar e reativar desfazem um ao outro).
Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			enviado, err := escrever(cmd, operacao{Risco: riscoBaixo, Resumo: resumo}, func(snd api.Sender) error { return enviar(ctx, snd) })
			if err != nil || !enviado {
				return err
			}
			return output.Write(streamsOf(cmd).out, f, resultadoEscrita(acao, situacao, nil))
		},
	}
	addWriteFlags(cmd)
	return cmd
}
