package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// risco de uma escrita (ADR-0018 e docs/api/escrita).
type risco int

const (
	riscoBaixo risco = iota // estado de tela, reversível
	riscoMedio              // cadastro ou contabilidade, reversível ou declaração simples
	riscoAlto               // efeito fiscal, legal ou financeiro, em geral sem desfazer
)

func (r risco) String() string {
	return [...]string{"baixo", "médio", "alto"}[r]
}

// palavraConfirmo é o que o usuário digita para confirmar uma escrita de risco alto.
const palavraConfirmo = "confirmo"

// errCancelada sinaliza que o usuário recusou a confirmação (código 1).
var errCancelada = errors.New("operação cancelada")

// operacao descreve uma escrita para a confirmação.
type operacao struct {
	Risco risco
	// Resumo diz em uma linha o que vai acontecer ("Excluir o lançamento 123 de 15/09/2026").
	Resumo string
	// Consequencia, no risco alto, diz o efeito e se dá para desfazer.
	Consequencia string
	// ID identifica o objeto alterado no registro de ações, quando conhecido antes do envio.
	ID string
}

// addWriteFlags acrescenta --yes e --dry-run a um comando de escrita.
func addWriteFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("yes", "y", false, "envia sem pedir confirmação (obrigatório sem terminal)")
	cmd.Flags().Bool("dry-run", false, "mostra a requisição que seria enviada, sem enviar")
}

// escrever é o caminho único dos comandos de escrita (ADR-0018). Com --dry-run, enviar recebe
// um Sender que só imprime as requisições, e escrever devolve false. Sem --dry-run, mostra o
// resumo, pede confirmação (ou exige --yes sem terminal) e chama enviar com a sessão.
// Devolve true quando a escrita foi enviada; quem chama então relê o estado e mostra o resultado.
func escrever(cmd *cobra.Command, op operacao, enviar func(api.Sender) error) (bool, error) {
	s := streamsOf(cmd)
	if dry, _ := cmd.Flags().GetBool("dry-run"); dry {
		f, err := outputFormat(cmd, "")
		if err != nil {
			return false, err
		}
		var reqs []requisicaoSimulada
		if err := enviar(dryRunSender{&reqs}); err != nil {
			return false, err
		}
		fmt.Fprintf(s.err, "Simulação (--dry-run), nada foi enviado: %s (risco %s)\n", op.Resumo, op.Risco)
		return false, escreverSimulacao(s.out, f, reqs)
	}
	yes, _ := cmd.Flags().GetBool("yes")
	if err := confirmar(s, op, yes); err != nil {
		return false, err
	}
	return true, enviar(sessionSender{s: s, origem: origemEscrita{comando: cmd.CommandPath(), id: op.ID}})
}

// confirmar mostra o resumo e pede a confirmação adequada ao risco.
func confirmar(s streams, op operacao, yes bool) error {
	fmt.Fprintf(s.err, "%s (risco %s)\n", op.Resumo, op.Risco)
	if op.Risco == riscoAlto && op.Consequencia != "" {
		fmt.Fprintln(s.err, op.Consequencia)
	}
	if yes {
		return nil
	}
	if !s.interactive {
		return usageError{errors.New("confirmação necessária: rode num terminal, use --yes para enviar ou --dry-run para simular")}
	}
	if op.Risco == riscoAlto {
		fmt.Fprintf(s.err, "Digite %q para continuar: ", palavraConfirmo)
	} else {
		fmt.Fprint(s.err, "Confirma? [s/N] ")
	}
	resp := lerResposta(s)
	if op.Risco == riscoAlto && resp == palavraConfirmo {
		return nil
	}
	if op.Risco != riscoAlto && respostaSim(resp) {
		return nil
	}
	return errCancelada
}

// lerResposta lê uma linha byte a byte, sem buffer, para que perguntas seguidas na mesma
// entrada não percam as respostas seguintes.
func lerResposta(s streams) string {
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := s.in.Read(buf)
		if n == 1 {
			if buf[0] == '\n' {
				break
			}
			b.WriteByte(buf[0])
		}
		if err != nil {
			break
		}
	}
	return strings.ToLower(strings.TrimSpace(b.String()))
}

func respostaSim(r string) bool { return r == "s" || r == "sim" || r == "y" || r == "yes" }

// perguntar faz uma pergunta de sim ou não no terminal (no meio de uma escrita em etapas).
func perguntar(s streams, pergunta string) bool {
	fmt.Fprintf(s.err, "%s [s/N] ", pergunta)
	return respostaSim(lerResposta(s))
}

// resultadoEscrita começa o registro de saída de uma escrita com as chaves estáveis do
// contrato público (ADR-0017): acao, situacao e id. O comando acrescenta os demais campos.
func resultadoEscrita(acao, situacao string, id any) *output.Record {
	rec := &output.Record{}
	return rec.Add("acao", "Ação", acao).Add("situacao", "Situação", situacao).Add("id", "ID", id)
}
