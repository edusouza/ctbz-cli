package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func opcoesReclassificacaoList(ps []api.PendenciaReclassificacao) *output.List {
	l := &output.List{Columns: []output.Column{
		{Key: "id_pendencia", Header: "Pendência"},
		{Key: "descricao", Header: "Lançamento"},
		{Key: "valor", Header: "Valor"},
		{Key: "classificacao_atual", Header: "Classificação atual"},
		{Key: "id_classificacao", Header: "ID"},
		{Key: "classificacao", Header: "Opção"},
		{Key: "socios", Header: "Sócios"},
	}}
	for _, p := range ps {
		for _, o := range p.Classificacoes {
			var socios []string
			for _, s := range o.Socios {
				socios = append(socios, fmt.Sprintf("%d %s", s.IDSocio, s.Nome))
			}
			l.Append(p.IDTexto(), output.Text(p.Descricao), moneyOrNil(p.Valor), p.ClassificacaoAtual, o.IDClassificacao, o.Nome, nonNil(socios))
		}
	}
	return l
}

// escolherReclassificacao acha a opção (id ou nome) na pendência e resolve o sócio: com um
// único sócio, ele é escolhido sozinho, como no painel.
func escolherReclassificacao(p api.PendenciaReclassificacao, classificacao, socio string) (api.Reclassificacao, string, error) {
	for _, o := range p.Classificacoes {
		if strconv.FormatInt(o.IDClassificacao, 10) != classificacao && !strings.EqualFold(o.Nome, classificacao) {
			continue
		}
		r := api.Reclassificacao{IDPendencia: p.IDPendencia, IDClassificacao: o.IDClassificacao, IDLancamento: p.IDLancamento}
		nome := o.Nome
		switch {
		case len(o.Socios) == 0 && socio != "":
			return r, "", usageError{fmt.Errorf("a classificação %q não usa sócio", o.Nome)}
		case len(o.Socios) == 0:
			return r, nome, nil
		case socio == "" && len(o.Socios) == 1:
			id := o.Socios[0].IDSocio
			r.IDSocio = &id
			return r, nome + " / " + o.Socios[0].Nome, nil
		case socio == "":
			return r, "", usageError{fmt.Errorf("a classificação %q exige --socio (veja as opções sem --classificacao)", o.Nome)}
		}
		for _, s := range o.Socios {
			if strconv.FormatInt(s.IDSocio, 10) == socio {
				id := s.IDSocio
				r.IDSocio = &id
				return r, nome + " / " + s.Nome, nil
			}
		}
		return r, "", usageError{fmt.Errorf("sócio %q não é opção em %q", socio, o.Nome)}
	}
	return api.Reclassificacao{}, "", usageError{fmt.Errorf("a classificação %q não é opção para a pendência %s; rode sem --classificacao para ver as opções", classificacao, p.IDTexto())}
}

func newRotinasReclassificarCmd() *cobra.Command {
	var classificacao, socio string
	cmd := &cobra.Command{
		Use:   "reclassificar ID_PENDENCIA...",
		Short: "Confirma ou altera a classificação de lançamentos pedida pela Central de Rotinas",
		Long: `Resolve as rotinas que pedem para confirmar ou alterar a classificação de um lançamento
bancário (financiamento, investimento anjo e outros). Os ids das pendências aparecem na
coluna "pendencias" de ctbz rotinas.

Sem --classificacao, só lista as opções de cada pendência (nada é enviado). Com
--classificacao (id ou nome da opção), grava a classificação e conclui a rotina (risco médio:
a rotina não reabre, mas a classificação ainda pode ser trocada com ctbz extratos classificar
enquanto o período estiver aberto). Opções com sócio exigem --socio, exceto quando há um só
sócio, que é escolhido sozinho como no painel.

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz rotinas reclassificar 1000000000000501
  ctbz rotinas reclassificar 1000000000000501 --classificacao "Aporte de Capital"`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return usageError{errors.New("informe pelo menos um ID de pendência (ver ctbz rotinas)")}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if socio != "" && classificacao == "" {
				return usageError{errors.New("--socio só vale com --classificacao")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			ps, err := api.BuscarReclassificacoes(ctx, g, args)
			if err != nil {
				return err
			}
			if len(ps) == 0 {
				return fmt.Errorf("nenhuma pendência de reclassificação encontrada para %s", strings.Join(args, ", "))
			}
			if classificacao == "" {
				return output.Write(s.out, f, opcoesReclassificacaoList(ps))
			}
			itens := make([]api.Reclassificacao, len(ps))
			var resumo []string
			nomes := make([]string, len(ps))
			for i, p := range ps {
				if itens[i], nomes[i], err = escolherReclassificacao(p, classificacao, socio); err != nil {
					return err
				}
				valor := int64(0)
				if p.Valor != nil {
					valor = centavos(*p.Valor)
				}
				resumo = append(resumo, fmt.Sprintf("%q %s → %s", p.Descricao, formatarCentavos(valor), nomes[i]))
			}
			op := operacao{Risco: riscoMedio, ID: strings.Join(args, ","), Resumo: fmt.Sprintf(
				"Reclassificar e concluir %d pendência(s) da Central de Rotinas: %s", len(ps), strings.Join(resumo, "; "))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.Reclassificar(ctx, snd, itens) })
			if err != nil || !enviado {
				return err
			}
			pendentes := map[string]bool{}
			if c, err := api.BuscarCentralRotinas(ctx, g); err != nil {
				fmt.Fprintln(s.err, "aviso: reclassificação enviada, mas não foi possível reler a Central de Rotinas:", err)
				for _, p := range ps {
					pendentes[p.IDTexto()] = true
				}
			} else {
				for _, r := range c.Rotinas {
					for _, id := range r.IDsPendentes() {
						pendentes[id] = true
					}
				}
			}
			recs := make([]output.Record, len(ps))
			for i, p := range ps {
				situacao := "concluido"
				if pendentes[p.IDTexto()] {
					situacao = "enviado"
				}
				recs[i] = *resultadoEscrita("reclassificar", situacao, p.IDTexto()).
					Add("id_lancamento", "Lançamento", p.IDLancamento).
					Add("classificacao", "Classificação", nomes[i])
			}
			return output.Write(s.out, f, output.RecordsToList(recs))
		},
	}
	cmd.Flags().StringVar(&classificacao, "classificacao", "", "opção de classificação: id ou nome (sem ela, lista as opções)")
	cmd.Flags().StringVar(&socio, "socio", "", "sócio, nas opções que exigem")
	addWriteFlags(cmd)
	return cmd
}
