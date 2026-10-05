package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// termo descreve um termo ou carta com aceite na Central de Rotinas.
type termo struct {
	slug, chave, titulo string
	conteudo            func(api.PendenciaCentral) string
}

var termos = []termo{
	{"carta-responsabilidade", api.PendenciaCartaResponsabilidade, "Carta de Responsabilidade da Administração",
		func(p api.PendenciaCentral) string { return p.ConteudoCartaResponsabilidade }},
	{"termo-debitos", api.PendenciaTermoDebitos, "Termo de Ciência e Responsabilidade (retiradas de lucros)",
		func(p api.PendenciaCentral) string { return p.ConteudoAceiteTermoDebito }},
	{"termo-totalpass", api.PendenciaTermoTotalPass, "Termo de adesão ao TotalPass",
		func(p api.PendenciaCentral) string { return p.ConteudoTermoAdesaoTotalPass }},
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
