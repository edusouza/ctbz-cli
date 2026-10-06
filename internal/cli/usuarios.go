package cli

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// avisoAcessoTotal é o aviso do painel ao convidar alguém.
const avisoAcessoTotal = "A pessoa convidada terá acesso total: poderá realizar todas as ações na plataforma e ver todas as informações da empresa, inclusive financeiras. Não há como revogar o convite antes do aceite; depois, só desativando o usuário."

func newUsuariosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usuarios",
		Short: "Lista quem tem acesso à empresa no painel (multiusuário)",
		Long: `Lista os usuários e os convites da empresa: ID, nome, e-mail, tipo (principal ou
secundário) e situação (ativo, inativo ou convite enviado).

Para convidar: ctbz usuarios convidar; para revogar ou devolver o acesso: ctbz usuarios
desativar|ativar.`,
		Example: `  ctbz usuarios
  ctbz usuarios -o json`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			us, err := api.BuscarUsuariosEmpresa(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{
				{Key: "id", Header: "ID"}, {Key: "nome", Header: "Nome"}, {Key: "email", Header: "E-mail"},
				{Key: "tipo", Header: "Tipo"}, {Key: "status", Header: "Situação"},
			}}
			for _, u := range us {
				l.Append(nilIfEmpty(idTexto(u.ID)), nilIfEmpty(u.Nome), nilIfEmpty(u.Email), nilIfEmpty(u.Tipo), nilIfEmpty(u.Status))
			}
			return output.Write(s.out, f, l)
		},
	}
	cmd.AddCommand(newUsuariosConvidarCmd(), newUsuariosAtivarCmd(true), newUsuariosAtivarCmd(false))
	return cmd
}

func newUsuariosConvidarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convidar EMAIL",
		Short: "Convida uma pessoa para acessar a empresa, com acesso total",
		Long: `Envia um convite por e-mail para uma pessoa acessar a empresa no painel. O acesso é
total: todas as ações e todas as informações, inclusive financeiras.

A CLI recusa e-mails da Contabilizei, e-mails que já têm acesso ou convite, e convites quando o
serviço está indisponível ou o limite de usuários secundários ativos foi atingido.

Risco alto: não há como revogar o convite antes do aceite; depois, só com ctbz usuarios
desativar. Aceita --yes e --dry-run.`,
		Example: `  ctbz usuarios convidar pessoa@example.com`,
		Args:    exactArgs(1, "o e-mail da pessoa"),
		RunE: func(cmd *cobra.Command, args []string) error {
			email, err := emailConvite(args[0])
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			st, err := api.BuscarStatusMultiusuario(cmd.Context(), g)
			if err != nil {
				return err
			}
			if !st.Ativo {
				return errors.New("o serviço de multiusuário está indisponível no momento")
			}
			us, err := api.BuscarUsuariosEmpresa(cmd.Context(), g)
			if err != nil {
				return err
			}
			ativos := 0
			for _, u := range us {
				if strings.EqualFold(u.Email, email) {
					return fmt.Errorf("%s já está na lista de usuários (%s)", email, u.Status)
				}
				if u.Tipo == api.UsuarioSecundario && u.Status == api.UsuarioAtivo {
					ativos++
				}
			}
			if ativos >= st.Limite {
				return fmt.Errorf("você já cadastrou o número máximo de %d usuários ativos; desative um antes de convidar", st.Limite)
			}
			op := operacao{Risco: riscoAlto, Resumo: "Convidar " + email + " para acessar a empresa", Consequencia: avisoAcessoTotal}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.ConvidarUsuario(cmd.Context(), snd, email) })
			if err != nil || !enviado {
				return err
			}
			u, err := usuarioPor(cmd, g, func(u api.UsuarioEmpresa) bool { return strings.EqualFold(u.Email, email) })
			situacao, id := "enviado", any(nil)
			if err == nil {
				situacao, id = firstNonEmpty(u.Status, situacao), nilIfEmpty(idTexto(u.ID))
			}
			return output.Write(s.out, f, resultadoEscrita("convidar", situacao, id).Add("email", "E-mail", email))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

// emailConvite valida o e-mail como o front: endereço simples e fora da Contabilizei.
func emailConvite(v string) (string, error) {
	a, err := mail.ParseAddress(strings.TrimSpace(v))
	if err != nil || a.Name != "" || !strings.Contains(a.Address[strings.LastIndex(a.Address, "@")+1:], ".") {
		return "", usageError{fmt.Errorf("e-mail inválido: %q", v)}
	}
	email := strings.ToLower(a.Address)
	if strings.Contains(email[strings.LastIndex(email, "@"):], "@contabilizei.") {
		return "", usageError{errors.New("não é possível convidar e-mails da Contabilizei")}
	}
	return email, nil
}

func newUsuariosAtivarCmd(ativar bool) *cobra.Command {
	use, short, acao := "desativar ID", "Revoga o acesso de um usuário à empresa", "desativar"
	if ativar {
		use, short, acao = "ativar ID", "Devolve o acesso de um usuário desativado", "ativar"
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long: short + `. O ID está em ctbz usuarios. Desativar é a única forma de revogar o
acesso, e o administrador da empresa (usuário principal) não pode ser desativado.

Risco médio: reversível com ativar ou desativar. Aceita --yes e --dry-run.`,
		Example: "  ctbz usuarios " + acao + " 102",
		Args:    exactArgs(1, "o ID do usuário (ver ctbz usuarios)"),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			u, err := usuarioPor(cmd, g, func(u api.UsuarioEmpresa) bool { return idTexto(u.ID) == args[0] })
			if err != nil {
				return err
			}
			switch {
			case !ativar && u.Tipo == api.UsuarioPrincipal:
				return errors.New("não é possível desativar o administrador da empresa")
			case ativar && u.Status == api.UsuarioConviteEnviado:
				return errors.New("o convite ainda não foi aceito: não há acesso para ativar")
			case ativar == (u.Status == api.UsuarioAtivo) && u.Status != api.UsuarioConviteEnviado:
				fmt.Fprintf(s.err, "O usuário %s já está %s; nada foi enviado.\n", args[0], u.Status)
				return output.Write(s.out, f, resultadoEscrita(acao, "sem mudança", args[0]))
			}
			verbo := map[bool]string{true: "Ativar", false: "Desativar"}[ativar]
			op := operacao{Risco: riscoMedio, ID: args[0], Resumo: fmt.Sprintf("%s o acesso de %s (%s, %s)", verbo, firstNonEmpty(u.Nome, u.Email), u.Email, u.Status)}
			if u.Status == api.UsuarioConviteEnviado {
				op.Resumo += "; não se sabe se desativar cancela um convite ainda não aceito"
			}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.AtivarUsuarioEmpresa(cmd.Context(), snd, u.ID, ativar) })
			if err != nil || !enviado {
				return err
			}
			relido, err := usuarioPor(cmd, g, func(x api.UsuarioEmpresa) bool { return idTexto(x.ID) == args[0] })
			situacao := "enviado"
			if err == nil && relido.Status != u.Status {
				situacao = relido.Status
			}
			return output.Write(s.out, f, resultadoEscrita(acao, situacao, args[0]))
		},
	}
	addWriteFlags(cmd)
	return cmd
}

// usuarioPor acha um usuário da empresa.
func usuarioPor(cmd *cobra.Command, g api.Getter, achou func(api.UsuarioEmpresa) bool) (*api.UsuarioEmpresa, error) {
	us, err := api.BuscarUsuariosEmpresa(cmd.Context(), g)
	if err != nil {
		return nil, err
	}
	for i := range us {
		if achou(us[i]) {
			return &us[i], nil
		}
	}
	return nil, errors.New("usuário não encontrado (veja ctbz usuarios)")
}
