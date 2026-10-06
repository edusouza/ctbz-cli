package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/otp"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newContaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conta",
		Short: "Mostra os dados de login do usuário (e-mail, telefone e método do 2FA)",
		Long: `Mostra o e-mail e o telefone de login, mascarados como no painel, e o método do segundo
fator (EMAIL, SMS ou APP).

Para trocar: ctbz conta alterar (e-mail e telefone) e ctbz conta senha.`,
		Example: `  ctbz conta`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			d, err := api.BuscarDadosLogin(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			return output.Write(s.out, f, dadosLoginRecord(d))
		},
	}
	cmd.AddCommand(newContaAlterarCmd(), newContaSenhaCmd())
	return cmd
}

func dadosLoginRecord(d *api.DadosLogin) *output.Record {
	rec := &output.Record{}
	return rec.Add("email", "E-mail", nilIfEmpty(semControle(d.Email))).
		Add("telefone", "Telefone", nilIfEmpty(semControle(d.Telefone))).
		Add("metodo_2fa", "Método do 2FA", nilIfEmpty(d.Metodo))
}

// otpOpts dizem como pedir e obter o código das escritas da conta.
type otpOpts struct {
	via     string
	cmd     string
	timeout time.Duration
}

func addOTPFlags(cmd *cobra.Command, o *otpOpts) {
	cmd.Flags().StringVar(&o.via, "via", "", "por onde receber o código: email ou sms (padrão: o método do 2FA)")
	cmd.Flags().StringVar(&o.cmd, "otp-cmd", "", "comando de shell que imprime o código (env CTBZ_OTP_CMD)")
	cmd.Flags().DurationVar(&o.timeout, "otp-timeout", defaultOTPTimeout, "tempo máximo aguardando o --otp-cmd (env CTBZ_OTP_TIMEOUT)")
}

// metodo escolhe o método do código: --via, ou o do 2FA (APP usa o app autenticador).
func (o *otpOpts) metodo(cmd *cobra.Command, atual string) (string, error) {
	if o.cmd == "" {
		o.cmd = os.Getenv("CTBZ_OTP_CMD")
	}
	if !cmd.Flags().Changed("otp-timeout") {
		o.timeout = envDuration("CTBZ_OTP_TIMEOUT", defaultOTPTimeout)
	}
	switch o.via {
	case "email":
		return api.MetodoOTPEmail, nil
	case "sms":
		return api.MetodoOTPSMS, nil
	case "":
		return firstNonEmpty(atual, api.MetodoOTPEmail), nil
	}
	return "", usageError{fmt.Errorf("--via deve ser email ou sms: %q", o.via)}
}

// escreverComOTP confirma a escrita e só então pede o código (exceto no app autenticador),
// obtém o código pelo --otp-cmd ou no terminal e chama enviar com ele. Sem terminal, exige
// --otp-cmd antes de qualquer envio.
func escreverComOTP(cmd *cobra.Command, op operacao, metodo string, o otpOpts, enviar func(snd api.Sender, codigo string) error) (bool, error) {
	s := streamsOf(cmd)
	dry, _ := cmd.Flags().GetBool("dry-run")
	if !dry && !s.interactive && o.cmd == "" {
		return false, usageError{errors.New("sem terminal, informe --otp-cmd (ou CTBZ_OTP_CMD) para ler o código")}
	}
	return escrever(cmd, op, func(snd api.Sender) error {
		desde := now()
		if metodo != api.MetodoOTPApp {
			if _, err := api.EnviarOTPConta(cmd.Context(), snd, metodo); err != nil {
				return erroEnvioOTP(err)
			}
		}
		codigo := "000000" // no --dry-run não há código; aparece como ***
		if !dry {
			var err error
			if codigo, err = obterCodigo(cmd, s, o, metodo, desde); err != nil {
				return err
			}
		}
		if err := enviar(snd, codigo); err != nil {
			var we *ctbz.WriteError
			if errors.As(err, &we) && we.Status == http.StatusNotFound {
				return errors.New("o código está incorreto ou inválido; nada foi alterado")
			}
			return err
		}
		return nil
	})
}

// erroEnvioOTP traduz o 422 do envio do código: o número é quanto falta para pedir outro.
func erroEnvioOTP(err error) error {
	var we *ctbz.WriteError
	if errors.As(err, &we) && we.Status == http.StatusUnprocessableEntity {
		var v struct {
			Error json.Number `json:"error"`
		}
		if json.Unmarshal(we.Body, &v) == nil && v.Error != "" {
			return fmt.Errorf("um código foi pedido há pouco; aguarde %s segundos e repita", v.Error)
		}
	}
	return err
}

func obterCodigo(cmd *cobra.Command, s streams, o otpOpts, metodo string, desde time.Time) (string, error) {
	destino := map[string]string{api.MetodoOTPEmail: "por e-mail", api.MetodoOTPSMS: "por SMS", api.MetodoOTPApp: "no app autenticador"}[metodo]
	if o.cmd != "" {
		fmt.Fprintf(s.err, "Código enviado %s; aguardando via --otp-cmd…\n", destino)
		return (&otp.Command{Cmd: o.cmd, Since: desde.Add(-time.Minute), Timeout: o.timeout, Log: verboseWriter(s)}).Fetch(cmd.Context())
	}
	for range 3 {
		fmt.Fprintf(s.err, "Código %s: ", destino)
		if c, ok := otp.Extract(lerLinha(s)); ok {
			return c, nil
		}
		fmt.Fprintln(s.err, "código inválido: digite os 6 dígitos")
	}
	return "", errors.New("código não informado; nada foi alterado")
}

func newContaAlterarCmd() *cobra.Command {
	var email, telefone string
	var o otpOpts
	cmd := &cobra.Command{
		Use:   "alterar",
		Short: "Troca o e-mail e o telefone de login (com código OTP)",
		Long: `Troca o e-mail e o telefone de login. A API recebe os dois juntos e o painel os mostra
mascarados, então informe os dois (repita o atual no que não muda). Telefone no formato
(DD)NNNNN-NNNN ou só os 11 dígitos.

Depois da confirmação, a Contabilizei envia um código (--via email|sms; padrão: o método do
2FA). O código é lido do --otp-cmd (ou CTBZ_OTP_CMD), como no login, ou digitado no terminal.

Risco alto: os próximos códigos de login vão para o novo e-mail (atualize CTBZ_USER) e um
erro pode tirar seu acesso. Dá para alterar de novo. Aceita --yes e --dry-run.`,
		Example: `  ctbz conta alterar --email novo@example.com --telefone "(11)98765-4321"
  CTBZ_OTP_CMD=./scripts/otp-gmail-gws.sh ctbz conta alterar --email novo@example.com --telefone 11987654321 --yes`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if email == "" || telefone == "" {
				return usageError{errors.New("informe --email e --telefone (a API recebe os dois juntos)")}
			}
			e, err := emailValido(email)
			if err != nil {
				return err
			}
			fone := ctbz.OnlyDigits(telefone)
			if len(fone) != 11 || strings.Trim(telefone, "0123456789()- ") != "" {
				return usageError{fmt.Errorf("telefone inválido %q: use (DD)NNNNN-NNNN", telefone)}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			atual, err := api.BuscarDadosLogin(cmd.Context(), g)
			if err != nil {
				return err
			}
			metodo, err := o.metodo(cmd, atual.Metodo)
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoAlto,
				Resumo:       fmt.Sprintf("Trocar o login de %s / %s para %s / (%s)%s-%s", semControle(atual.Email), semControle(atual.Telefone), e, fone[:2], fone[2:7], fone[7:]),
				Consequencia: "Os próximos códigos de login vão para o novo e-mail (atualize CTBZ_USER), e um erro pode tirar seu acesso. Dá para alterar de novo."}
			enviado, err := escreverComOTP(cmd, op, metodo, o, func(snd api.Sender, codigo string) error {
				return api.AlterarDadosConta(cmd.Context(), snd, e, fone, codigo, metodo)
			})
			if err != nil || !enviado {
				return err
			}
			d, err := api.BuscarDadosLogin(cmd.Context(), g)
			if err != nil {
				return err
			}
			rec := resultadoEscrita("alterar", "alterado", nil)
			for _, c := range dadosLoginRecord(d).Fields {
				rec.Add(c.Key, c.Label, c.Value)
			}
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "novo e-mail de login")
	cmd.Flags().StringVar(&telefone, "telefone", "", "novo telefone, (DD)NNNNN-NNNN")
	addOTPFlags(cmd, &o)
	addWriteFlags(cmd)
	return cmd
}

func newContaSenhaCmd() *cobra.Command {
	var o otpOpts
	cmd := &cobra.Command{
		Use:   "senha",
		Short: "Troca a senha de login (com código OTP)",
		Long: `Troca a senha de login. A nova senha é pedida duas vezes, sem eco, e nunca por argumento;
fora de um terminal, é lida da entrada padrão (duas linhas). Regras do painel: pelo menos 8
caracteres, com minúsculas, maiúsculas, números e um caractere especial.

A senha atual não é pedida: a prova é o código que a Contabilizei envia depois da confirmação
(--via email|sms; padrão: o método do 2FA), lido do --otp-cmd (ou CTBZ_OTP_CMD) ou digitado.

Risco alto: atualize CTBZ_PASSWORD; outras sessões podem ser encerradas. Aceita --yes e
--dry-run (que mostra *** no lugar da senha).`,
		Example: `  ctbz conta senha
  ctbz conta senha --via sms`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			senha, err := lerSegredo(s, "Nova senha: ")
			if err != nil {
				return err
			}
			if err := senhaValida(senha); err != nil {
				return err
			}
			repetida, err := lerSegredo(s, "Repita a nova senha: ")
			if err != nil {
				return err
			}
			if repetida != senha {
				return usageError{errors.New("as senhas não conferem")}
			}
			atual, err := api.BuscarDadosLogin(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			metodo, err := o.metodo(cmd, atual.Metodo)
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoAlto, Resumo: "Trocar a senha de login de " + semControle(atual.Email),
				Consequencia: "Atualize CTBZ_PASSWORD; outras sessões podem ser encerradas. Dá para trocar de novo."}
			enviado, err := escreverComOTP(cmd, op, metodo, o, func(snd api.Sender, codigo string) error {
				return api.AlterarSenhaConta(cmd.Context(), snd, senha, codigo, metodo)
			})
			if err != nil || !enviado {
				return err
			}
			return output.Write(s.out, f, resultadoEscrita("senha", "alterada", nil))
		},
	}
	addOTPFlags(cmd, &o)
	addWriteFlags(cmd)
	return cmd
}

// senhaValida aplica as regras do painel para a senha nova.
func senhaValida(senha string) error {
	var minuscula, maiuscula, numero, especial bool
	for _, r := range senha {
		switch {
		case unicode.IsLower(r):
			minuscula = true
		case unicode.IsUpper(r):
			maiuscula = true
		case unicode.IsDigit(r):
			numero = true
		case !unicode.IsSpace(r):
			especial = true
		}
	}
	if utf8.RuneCountInString(senha) < 8 || !minuscula || !maiuscula || !numero || !especial {
		return usageError{errors.New("a senha precisa de pelo menos 8 caracteres, com minúsculas, maiúsculas, números e um caractere especial")}
	}
	return nil
}
