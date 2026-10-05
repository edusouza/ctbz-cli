package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// termo descreve um termo ou carta com aceite na Central de Rotinas.
type termo struct {
	slug, chave, titulo string
	conteudo            func(api.PendenciaCentral) string
	pathAceite          string
	// consequencia é o que o painel diz sobre o aceite, mostrado na confirmação.
	consequencia string
}

var termos = []termo{
	{"carta-responsabilidade", api.PendenciaCartaResponsabilidade, "Carta de Responsabilidade da Administração",
		func(p api.PendenciaCentral) string { return p.ConteudoCartaResponsabilidade }, api.PathAceitarCartaResponsabilidade,
		"É a declaração anual exigida pelo CFC de que todas as informações e documentos foram entregues à contabilidade. Não dá para revogar."},
	{"termo-debitos", api.PendenciaTermoDebitos, "Termo de Ciência e Responsabilidade (retiradas de lucros)",
		func(p api.PendenciaCentral) string { return p.ConteudoAceiteTermoDebito }, api.PathAceitarTermoDebitos,
		"Declara ciência dos riscos das retiradas de lucros (fiscalização pela Receita Federal, multas, juros e autuações). Não dá para revogar."},
	{"termo-totalpass", api.PendenciaTermoTotalPass, "Termo de adesão ao TotalPass",
		func(p api.PendenciaCentral) string { return p.ConteudoTermoAdesaoTotalPass }, api.PathAceitarTermoTotalPass,
		"Adesão ao programa de benefícios (TotalPass e Starbem). Não dá para revogar pela CLI."},
}

func slugsTermos() string {
	var ss []string
	for _, t := range termos {
		ss = append(ss, t.slug)
	}
	return strings.Join(ss, ", ")
}

func acharTermo(slug string) (termo, error) {
	for _, t := range termos {
		if t.slug == slug {
			return t, nil
		}
	}
	return termo{}, usageError{fmt.Errorf("termo desconhecido %q: use %s", slug, slugsTermos())}
}

// prazoTacito devolve o prazo do aceite tácito (dias) do termo de débitos, quando houver.
func prazoTacito(t termo, p api.PendenciaCentral) any {
	if t.chave != api.PendenciaTermoDebitos || p.PrazoAceiteTacito == nil {
		return nil
	}
	return p.PrazoAceiteTacito
}

func newPendenciasTermosCmd() *cobra.Command {
	var todos bool
	cmd := &cobra.Command{
		Use:   "termos",
		Short: "Lista os termos e cartas com aceite pendente na Central de Rotinas",
		Long: `Lista os termos e cartas que a Central de Rotinas pede para aceitar: carta de
responsabilidade, termo de ciência sobre retiradas de lucros (com o prazo do aceite tácito, em
dias) e termo de adesão ao TotalPass. --todos inclui os que não estão pendentes.

Para ler um termo inteiro: ctbz pendencias termo CHAVE. Para aceitar: ctbz pendencias aceitar CHAVE.`,
		Example: `  ctbz pendencias termos
  ctbz pendencias termos --todos -o json`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			c, err := api.BuscarCentralRotinasInit(cmd.Context(), sessionGetter{s})
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{
				{Key: "chave", Header: "Chave"},
				{Key: "titulo", Header: "Termo"},
				{Key: "pendente", Header: "Pendente"},
				{Key: "prazo_aceite_tacito_dias", Header: "Aceite tácito em (dias)"},
			}}
			for _, t := range termos {
				p, ok := c.Pendencia(t.chave)
				if !ok || (!p.PossuiPendencia && !todos) {
					continue
				}
				l.Append(t.slug, t.titulo, p.PossuiPendencia, prazoTacito(t, p))
			}
			return output.Write(s.out, f, l)
		},
	}
	cmd.Flags().BoolVar(&todos, "todos", false, "inclui os termos sem aceite pendente")
	return cmd
}

// lerTermo busca a pendência e o texto (HTML convertido) de um termo.
func lerTermo(ctx context.Context, g api.Getter, t termo) (api.PendenciaCentral, string, error) {
	c, err := api.BuscarCentralRotinasInit(ctx, g)
	if err != nil {
		return api.PendenciaCentral{}, "", err
	}
	p, ok := c.Pendencia(t.chave)
	if !ok {
		return p, "", nil
	}
	texto := ""
	if html := t.conteudo(p); strings.TrimSpace(html) != "" {
		texto = htmlParaTexto(html)
	}
	return p, texto, nil
}

func escreverTermo(w io.Writer, f output.Format, t termo, p api.PendenciaCentral, texto string) error {
	if f == output.FormatTable {
		_, err := fmt.Fprintf(w, "%s\n\n%s", t.titulo, texto)
		return err
	}
	rec := &output.Record{}
	rec.Add("chave", "Chave", t.slug).
		Add("titulo", "Termo", t.titulo).
		Add("pendente", "Pendente", p.PossuiPendencia).
		Add("prazo_aceite_tacito_dias", "Aceite tácito em (dias)", prazoTacito(t, p)).
		Add("texto", "Texto", nilIfEmpty(texto))
	return output.Write(w, f, rec)
}

func newPendenciasTermoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "termo CHAVE",
		Short: "Mostra o texto completo de um termo ou carta da Central de Rotinas",
		Long: `Mostra o texto completo de um termo ou carta, com o HTML convertido em texto. Leia antes
de aceitar com ctbz pendencias aceitar CHAVE.

CHAVE: ` + slugsTermos() + `.`,
		Example: `  ctbz pendencias termo carta-responsabilidade | less`,
		Args:    exactArgs(1, "a CHAVE do termo"),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := acharTermo(args[0])
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			p, texto, err := lerTermo(cmd.Context(), sessionGetter{s}, t)
			if err != nil {
				return err
			}
			if texto == "" {
				return fmt.Errorf("a Central de Rotinas não trouxe o texto de %q (o termo pode não estar disponível agora)", t.slug)
			}
			return escreverTermo(s.out, f, t, p, texto)
		},
	}
}

// erroAdmin traduz o 403 dos aceites e declarações: a sessão é de um administrador da
// Contabilizei personificando o cliente.
func erroAdmin(err error) error {
	var we *ctbz.WriteError
	if errors.As(err, &we) && we.Status == http.StatusForbidden {
		return fmt.Errorf("a Contabilizei recusou porque a sessão é de administrador (%s): aceites e declarações só podem ser feitos pelo próprio cliente", we.Message)
	}
	return err
}

func newPendenciasAceitarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aceitar CHAVE",
		Short: "Aceita um termo ou carta pendente da Central de Rotinas",
		Long: `Assina um aceite cobrado pela Central de Rotinas (risco alto: são declarações legais e
contábeis, sem como revogar). O texto completo do termo é sempre impresso no stderr antes da
confirmação, que no terminal exige digitar "confirmo" (--yes em scripts).

Se não houver aceite pendente, a CLI avisa "nada a aceitar" e termina com código 0.
Depois do envio, relê a Central de Rotinas para confirmar que a pendência sumiu.

CHAVE: ` + slugsTermos() + `.`,
		Example: `  ctbz pendencias aceitar carta-responsabilidade
  ctbz pendencias aceitar termo-debitos --dry-run`,
		Args: exactArgs(1, "a CHAVE do termo"),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := acharTermo(args[0])
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			p, texto, err := lerTermo(ctx, g, t)
			if err != nil {
				return err
			}
			if !p.PossuiPendencia {
				fmt.Fprintf(s.err, "Nada a aceitar: %s não está pendente.\n", t.titulo)
				return output.Write(s.out, f, resultadoEscrita("aceitar", "nada_a_aceitar", t.slug))
			}
			if texto == "" {
				return fmt.Errorf("a Central de Rotinas não trouxe o texto de %q; leia e aceite pelo painel", t.slug)
			}
			fmt.Fprintf(s.err, "%s\n\n%s\n", t.titulo, texto)
			consequencia := t.consequencia
			if dias := prazoTacito(t, p); dias != nil {
				consequencia += fmt.Sprintf(" Sem resposta em %v dias, a Contabilizei considera o termo aceito (aceite tácito).", dias)
			}
			op := operacao{Risco: riscoAlto, ID: t.slug, Resumo: "Aceitar: " + t.titulo, Consequencia: consequencia}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return erroAdmin(api.Aceitar(ctx, snd, t.pathAceite)) })
			if err != nil || !enviado {
				return err
			}
			situacao := "aceito"
			if novo, _, err := lerTermo(ctx, g, t); err != nil {
				fmt.Fprintln(s.err, "aviso: aceite enviado, mas não foi possível reler a Central de Rotinas:", err)
				situacao = "enviado"
			} else if novo.PossuiPendencia {
				fmt.Fprintln(s.err, "aviso: a pendência ainda aparece na Central de Rotinas; confira com ctbz pendencias termos")
				situacao = "enviado"
			}
			return output.Write(s.out, f, resultadoEscrita("aceitar", situacao, t.slug).Add("termo", "Termo", t.titulo))
		},
	}
	addWriteFlags(cmd)
	return cmd
}
