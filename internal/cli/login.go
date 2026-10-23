package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/otp"
)

type loginOpts struct {
	otpCode    string
	otpCmd     string
	otpTimeout time.Duration
	cnpj       string
	restart    bool
	quiet      bool
}

func newLoginCmd() *cobra.Command {
	var o loginOpts
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Autentica na Contabilizei (usuário, senha e código por e-mail)",
		Long: `Autentica na Contabilizei usando CTBZ_USER e CTBZ_PASSWORD.

O código enviado por e-mail pode vir de:
  1. --otp-cmd / CTBZ_OTP_CMD: comando executado repetidamente até imprimir 6 dígitos;
  2. o terminal, se interativo;
  3. uma segunda chamada: "ctbz login --otp 123456" (o login fica salvo como pendente
     e o processo termina com código 3).

Com mais de uma empresa, escolha com --cnpj (ou CTBZ_CNPJ).`,
		Example: `  ctbz login --cnpj 00.000.000/0001-00
  CTBZ_OTP_CMD=./scripts/otp-gmail-gws.sh ctbz login
  ctbz login --otp 123456`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := ctbz.DefaultStore()
			if err != nil {
				return err
			}
			o.applyEnv(cmd)
			s := streamsOf(cmd)
			sess, err := login(cmd.Context(), store, s, o)
			if err != nil {
				return err
			}
			fmt.Fprintf(s.err, "Login concluído: %s\n", describeSession(sess))
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.otpCode, "otp", "", "código OTP recebido por e-mail (retoma um login pendente)")
	f.StringVar(&o.otpCmd, "otp-cmd", "", "comando de shell que imprime o OTP (env CTBZ_OTP_CMD)")
	f.DurationVar(&o.otpTimeout, "otp-timeout", defaultOTPTimeout, "tempo máximo aguardando o --otp-cmd (env CTBZ_OTP_TIMEOUT)")
	f.StringVar(&o.cnpj, "cnpj", "", "CNPJ da empresa a selecionar (env CTBZ_CNPJ)")
	f.BoolVar(&o.restart, "restart", false, "descarta um login pendente e recomeça")
	return cmd
}

const defaultOTPTimeout = 3 * time.Minute

// applyEnv preenche com variáveis de ambiente as flags não informadas. É feito na
// execução (e não como padrão da flag) para a ajuda e a documentação não exibirem
// valores do ambiente de quem as gerou.
func (o *loginOpts) applyEnv(cmd *cobra.Command) {
	if o.otpCmd == "" {
		o.otpCmd = os.Getenv("CTBZ_OTP_CMD")
	}
	if o.cnpj == "" {
		o.cnpj = os.Getenv("CTBZ_CNPJ")
	}
	if !cmd.Flags().Changed("otp-timeout") {
		o.otpTimeout = envDuration("CTBZ_OTP_TIMEOUT", defaultOTPTimeout)
	}
}

func login(ctx context.Context, store *ctbz.Store, s streams, o loginOpts) (*ctbz.Session, error) {
	base := baseURL()
	o.cnpj = ctbz.OnlyDigits(o.cnpj)
	client := ctbz.NewClient(base)

	var p *ctbz.Pending
	if o.restart {
		_ = store.ClearPending()
	} else if pend, err := store.LoadPending(); err == nil && pend.BaseURL == base && pend.State != nil &&
		(o.otpCode != "" || (pend.State.Stage == ctbz.StageSelectCompany && o.cnpj != "")) {
		p = pend
		client.Resume(p.State)
		if o.cnpj != "" {
			p.WantedCNPJ = o.cnpj
		}
		logf(s, o, "Retomando login pendente (etapa: %s)\n", p.State.Stage)
	}
	if p == nil {
		if o.otpCode != "" {
			return nil, errors.New("não há login pendente para usar --otp: o código vale só para o login que o gerou; rode `ctbz login` primeiro")
		}
		logf(s, o, "Enviando credenciais para %s…\n", base)
		st, err := client.StartLogin(ctx, os.Getenv("CTBZ_USER"), os.Getenv("CTBZ_PASSWORD"))
		if err != nil {
			return nil, err
		}
		p = &ctbz.Pending{BaseURL: base, WantedCNPJ: o.cnpj, State: st}
	}
	st := p.State

	// pause salva o login para ser retomado em outra execução.
	pause := func(format string, a ...any) error {
		if err := store.SavePending(p); err != nil {
			return err
		}
		fmt.Fprintf(s.err, format, a...)
		return errPending
	}

	for {
		switch st.Stage {
		case ctbz.StageOTP:
			code := o.otpCode
			o.otpCode = "" // um código informado na linha de comando é usado uma única vez
			if code == "" && o.otpCmd != "" {
				logf(s, o, "Código enviado para %s; aguardando via --otp-cmd…\n", st.MaskedEmail)
				fetched, err := (&otp.Command{
					Cmd: o.otpCmd,
					// Margem para atraso de relógio entre esta máquina e o servidor de e-mail.
					Since:   st.OTPSentAt.Add(-time.Minute),
					Timeout: o.otpTimeout,
					Log:     verboseWriter(ctx, s),
				}).Fetch(ctx)
				if err != nil {
					return nil, pause("%v\nO login ficou pendente; informe o código com: ctbz login --otp NNNNNN\n", err)
				}
				code = fetched
			}
			if code == "" && s.interactive {
				var err error
				code, err = otp.Prompt(s.in, s.err, fmt.Sprintf("Código enviado para %s: ", st.MaskedEmail))
				if err != nil {
					return nil, pause("\n%v\nO login ficou pendente; conclua com: ctbz login --otp NNNNNN\n", err)
				}
			}
			if code == "" {
				return nil, pause("Código de verificação enviado para %s.\nConclua com: ctbz login --otp NNNNNN\n", st.MaskedEmail)
			}
			if err := client.VerifyOTP(ctx, st, code); err != nil {
				if errors.Is(err, ctbz.ErrInvalidOTP) {
					return nil, pause("%v. Tente outro código com: ctbz login --otp NNNNNN (ou recomece com ctbz login --restart)\n", err)
				}
				return nil, err
			}

		case ctbz.StageSelectCompany:
			cnpj := p.WantedCNPJ
			if cnpj == "" && len(st.Companies) == 1 {
				cnpj = st.Companies[0].CNPJ
			}
			if cnpj == "" && s.interactive {
				var err error
				if cnpj, err = promptCompany(s, st.Companies); err != nil {
					return nil, err
				}
			}
			if cnpj == "" {
				var b strings.Builder
				for _, co := range st.Companies {
					fmt.Fprintf(&b, "  %s  %s\n", co.CNPJ, co.Name)
				}
				return nil, pause("Escolha a empresa:\n%sConclua com: ctbz login --cnpj <CNPJ>\n", b.String())
			}
			if !st.HasCompany(cnpj) {
				p.WantedCNPJ = ""
			}
			if err := client.SelectCompany(ctx, st, cnpj); err != nil {
				if !st.HasCompany(cnpj) {
					return nil, pause("%v\nConclua com: ctbz login --cnpj <CNPJ>\n", err)
				}
				return nil, err
			}

		case ctbz.StageDone:
			sess := &ctbz.Session{
				BaseURL:   base,
				CreatedAt: time.Now(),
				Cookies:   client.Cookies,
				Storage:   st.Storage,
			}
			sess.CNPJ = currentCNPJ(sess)
			if err := store.SaveSession(sess); err != nil {
				return nil, err
			}
			_ = store.ClearPending()
			return sess, nil

		default:
			return nil, fmt.Errorf("etapa de login desconhecida: %q", st.Stage)
		}
	}
}

func promptCompany(s streams, companies []ctbz.Company) (string, error) {
	fmt.Fprintln(s.err, "Selecione a empresa:")
	for i, co := range companies {
		fmt.Fprintf(s.err, "  [%d] %s  %s\n", i+1, co.CNPJ, co.Name)
	}
	for {
		fmt.Fprint(s.err, "Número ou CNPJ: ")
		var answer string
		if _, err := fmt.Fscanln(s.in, &answer); err != nil {
			return "", fmt.Errorf("lendo escolha: %w", err)
		}
		if n, err := strconv.Atoi(answer); err == nil && n >= 1 && n <= len(companies) {
			return companies[n-1].CNPJ, nil
		}
		d := ctbz.OnlyDigits(answer)
		for _, co := range companies {
			if co.CNPJ == d {
				return d, nil
			}
		}
		fmt.Fprintln(s.err, "opção inválida")
	}
}

func logf(s streams, o loginOpts, format string, a ...any) {
	if !o.quiet {
		fmt.Fprintf(s.err, format, a...)
	}
}
