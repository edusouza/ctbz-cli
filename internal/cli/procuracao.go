package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// linkECAC é o portal da Receita onde a procuração eletrônica é criada.
const linkECAC = "https://cav.receita.fazenda.gov.br"

// Origens da pendência de procuração.
const (
	origemCentralRotinas   = "central-de-rotinas"
	origemPrimeirosPassos  = "primeiros-passos"
	origemSemProcuracaoPen = "nenhuma"
)

// situacaoProcuracao junta a pendência da Central de Rotinas e a etapa do checklist.
type situacaoProcuracao struct {
	origem, situacao, cnpj string
}

func (p situacaoProcuracao) pendente() bool { return p.origem != origemSemProcuracaoPen }

// path é o endpoint de "Já criei a procuração" da origem.
func (p situacaoProcuracao) path() string {
	if p.origem == origemPrimeirosPassos {
		return api.PathCompartilharProcuracao
	}
	return api.PathResolverProcuracaoEcac
}

func lerProcuracao(ctx context.Context, g api.Getter) (situacaoProcuracao, error) {
	c, err := api.BuscarCentralRotinasInit(ctx, g)
	if err != nil {
		return situacaoProcuracao{}, err
	}
	if p, ok := c.Pendencia(api.PendenciaProcuracaoEcac); ok && p.PossuiPendencia {
		return situacaoProcuracao{origemCentralRotinas, p.TipoPendencia, p.CnpjOutorgado}, nil
	}
	ck, err := api.BuscarChecklist(ctx, g)
	if err != nil {
		return situacaoProcuracao{}, err
	}
	if e, ok := ck.Etapa(api.EtapaGerarProcuracao); ok && !e.Concluida() {
		return situacaoProcuracao{origemPrimeirosPassos, e.Situacao, ""}, nil
	}
	return situacaoProcuracao{origem: origemSemProcuracaoPen}, nil
}

func procuracaoRecord(p situacaoProcuracao) *output.Record {
	rec := &output.Record{}
	return rec.Add("pendente", "Pendente", p.pendente()).
		Add("origem", "Origem", p.origem).
		Add("situacao", "Situação", nilIfEmpty(p.situacao)).
		Add("cnpj_outorgado", "CNPJ a outorgar", output.NewCNPJ(p.cnpj)).
		Add("link_ecac", "e-CAC", linkECAC)
}

func newPendenciasProcuracaoCmd() *cobra.Command {
	var jaCriei bool
	cmd := &cobra.Command{
		Use:   "procuracao",
		Short: "Mostra a pendência de procuração do e-CAC ou declara que ela foi criada",
		Long: `Sem --ja-criei, mostra se há pendência de procuração eletrônica (SEM_PROCURACAO,
EM_EXPIRACAO…), o CNPJ a quem outorgar e o link do e-CAC. Nada é enviado.

A procuração é criada (ou renovada) no e-CAC da Receita, fora da Contabilizei. Depois disso,
--ja-criei declara que ela existe ("Já criei a procuração" no painel), pelo caminho da origem
da pendência: Central de Rotinas ou checklist de primeiros passos (risco médio: é uma
declaração; se a procuração não existir, a pendência deve voltar).

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz pendencias procuracao
  ctbz pendencias procuracao --ja-criei`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			p, err := lerProcuracao(ctx, g)
			if err != nil {
				return err
			}
			if !jaCriei {
				if p.pendente() {
					alvo := "à Contabilizei"
					if p.cnpj != "" {
						alvo = "ao CNPJ " + output.FormatCNPJ(output.NewCNPJ(p.cnpj))
					}
					fmt.Fprintf(s.err, "Crie ou renove no e-CAC (%s) uma procuração eletrônica %s e depois rode: ctbz pendencias procuracao --ja-criei\n", linkECAC, alvo)
				}
				return output.Write(s.out, f, procuracaoRecord(p))
			}
			if !p.pendente() {
				fmt.Fprintln(s.err, "Nada a declarar: não há pendência de procuração.")
				return output.Write(s.out, f, resultadoEscrita("procuracao", "nada_a_declarar", nil))
			}
			op := operacao{Risco: riscoMedio, ID: p.origem, Resumo: fmt.Sprintf(
				"Declarar que a procuração eletrônica do e-CAC já foi criada (pendência %s, %s)", p.origem, p.situacao)}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return erroAdmin(api.DeclararProcuracao(ctx, snd, p.path()))
			})
			if err != nil || !enviado {
				return err
			}
			situacao := "declarado"
			if novo, err := lerProcuracao(ctx, g); err != nil {
				fmt.Fprintln(s.err, "aviso: declaração enviada, mas não foi possível reler a pendência:", err)
				situacao = "enviado"
			} else if novo.pendente() {
				fmt.Fprintln(s.err, "aviso: a pendência ainda aparece; confira com ctbz pendencias procuracao")
				situacao = "enviado"
			}
			return output.Write(s.out, f, resultadoEscrita("procuracao", situacao, p.origem))
		},
	}
	cmd.Flags().BoolVar(&jaCriei, "ja-criei", false, "declara que a procuração já foi criada no e-CAC")
	addWriteFlags(cmd)
	return cmd
}
