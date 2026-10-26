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

// conferirPendenciaReclassificacao recusa, antes do envio, uma pendência do init sem os ids
// que o corpo reenvia: um campo ausente vira erro de leitura, nunca uma escrita com id zero
// ou nulo (ADR-0021). O valor só aparece no resumo e pode faltar.
func conferirPendenciaReclassificacao(p api.PendenciaReclassificacao) error {
	nome := fmt.Sprintf("%q", p.Descricao)
	if id := p.IDTexto(); id != "" && id != "null" {
		nome = id
	}
	var falta string
	switch {
	case rawTexto(p.IDPendencia) == nil:
		falta = "sem idPendencia"
	case p.IDLancamento == 0:
		falta = "sem idLancamento"
	case len(p.Classificacoes) == 0:
		falta = "sem opções de classificação"
	}
	if falta != "" {
		return fmt.Errorf("lendo a pendência de reclassificação %s: a resposta veio %s; nada foi enviado, conclua a rotina pelo painel", nome, falta)
	}
	return nil
}

// opcoesReclassificacao devolve as opções da pendência que casam com --classificacao, na ordem
// de precedência: o id; senão o nome exato; senão o nome sem diferenciar maiúsculas. Um id
// sempre escolhe uma só opção, mesmo que outra se chame assim.
func opcoesReclassificacao(p api.PendenciaReclassificacao, classificacao string) []api.OpcaoReclassificacao {
	var porID, exatas, semCaixa []api.OpcaoReclassificacao
	for _, o := range p.Classificacoes {
		switch {
		case strconv.FormatInt(o.IDClassificacao, 10) == classificacao:
			porID = append(porID, o)
		case o.Nome == classificacao:
			exatas = append(exatas, o)
		case strings.EqualFold(o.Nome, classificacao):
			semCaixa = append(semCaixa, o)
		}
	}
	switch {
	case len(porID) > 0:
		return porID
	case len(exatas) > 0:
		return exatas
	}
	return semCaixa
}

// escolherReclassificacao acha a opção (id ou nome) na pendência e resolve o sócio: com um
// único sócio, ele é escolhido sozinho, como no painel. Um nome que casa com mais de uma opção
// é recusado, com os candidatos.
func escolherReclassificacao(p api.PendenciaReclassificacao, classificacao, socio string) (api.Reclassificacao, string, error) {
	opcoes := opcoesReclassificacao(p, classificacao)
	switch {
	case len(opcoes) == 0:
		return api.Reclassificacao{}, "", usageError{fmt.Errorf("a classificação %q não é opção para a pendência %s; rode sem --classificacao para ver as opções", classificacao, p.IDTexto())}
	case len(opcoes) > 1:
		candidatos := make([]string, len(opcoes))
		for i, o := range opcoes {
			candidatos[i] = fmt.Sprintf("%d %q", o.IDClassificacao, o.Nome)
		}
		return api.Reclassificacao{}, "", usageError{fmt.Errorf("a classificação %q é ambígua na pendência %s: casa com %s; use o id da opção em --classificacao",
			classificacao, p.IDTexto(), strings.Join(candidatos, ", "))}
	}
	o := opcoes[0]
	if o.IDClassificacao == 0 {
		return api.Reclassificacao{}, "", fmt.Errorf("lendo a pendência de reclassificação %s: a opção %q veio sem idClassificacao; nada foi enviado, conclua a rotina pelo painel", p.IDTexto(), o.Nome)
	}
	r := api.Reclassificacao{IDPendencia: p.IDPendencia, IDClassificacao: o.IDClassificacao, IDLancamento: p.IDLancamento}
	escolher := func(s api.SocioReclassificacao) (api.Reclassificacao, string, error) {
		if s.IDSocio == 0 {
			return api.Reclassificacao{}, "", fmt.Errorf("lendo a pendência de reclassificação %s: o sócio %q da opção %q veio sem idSocio; nada foi enviado, conclua a rotina pelo painel", p.IDTexto(), s.Nome, o.Nome)
		}
		id := s.IDSocio
		r.IDSocio = &id
		return r, o.Nome + " / " + s.Nome, nil
	}
	switch {
	case len(o.Socios) == 0 && socio != "":
		return r, "", usageError{fmt.Errorf("a classificação %q não usa sócio", o.Nome)}
	case len(o.Socios) == 0:
		return r, o.Nome, nil
	case socio == "" && len(o.Socios) == 1:
		return escolher(o.Socios[0])
	case socio == "":
		return r, "", usageError{fmt.Errorf("a classificação %q exige --socio (veja as opções sem --classificacao)", o.Nome)}
	}
	for _, s := range o.Socios {
		if strconv.FormatInt(s.IDSocio, 10) == socio {
			return escolher(s)
		}
	}
	return r, "", usageError{fmt.Errorf("sócio %q não é opção em %q", socio, o.Nome)}
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

--classificacao casa primeiro com o id da opção, depois com o nome escrito exatamente como na
lista e, por último, com o nome sem diferenciar maiúsculas. Um nome que casa com mais de uma
opção é recusado como uso incorreto, listando os candidatos: nesse caso, use o id.

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
				if err := conferirPendenciaReclassificacao(p); err != nil {
					return err
				}
				if itens[i], nomes[i], err = escolherReclassificacao(p, classificacao, socio); err != nil {
					return err
				}
				valor := "valor não informado"
				if p.Valor != nil {
					valor = formatarCentavos(centavos(*p.Valor))
				}
				resumo = append(resumo, fmt.Sprintf("%q %s → %s", p.Descricao, valor, nomes[i]))
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
