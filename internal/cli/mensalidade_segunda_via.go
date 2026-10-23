package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

func newMensalidadeSegundaViaCmd() *cobra.Command {
	var competencia string
	var soPix, soLinha bool
	cmd := &cobra.Command{
		Use:   "segunda-via",
		Short: "Gera a segunda via da mensalidade e mostra o Pix copia e cola e a linha digitável",
		Long: `Gera a segunda via da mensalidade em atraso, como o botão "Gerar segunda via s/ juros" do
card da fatura no painel, e mostra o que pagar: itens, total, vencimento, situação, Pix copia e
cola, linha digitável, código de barras e o link do checkout.

Risco médio: registra um boleto novo, com vencimento novo, mas não cobra nada. O caminho vem
da ação do card (ctbz mensalidade -o json, botao_adicional_acao); sem ela (ex.: fatura em
dia), o comando para com erro. Com --competencia, o card precisa oferecer a segunda via
daquele mês.

--pix e --linha imprimem só o código, sem decoração, para usar num pipe:

  inter-pj pix enviar --copia-e-cola "$(ctbz mensalidade segunda-via --pix --yes)"

O link do checkout tem o identificador da fatura, que dá acesso a ela sem login: ele sai só no
resultado, nunca no registro de ações. Depois de pagar, confirme com ctbz mensalidade situacao:
a fatura pública demora a dar baixa.

Pede confirmação (--yes em scripts) e aceita --dry-run, que mostra o GET sem enviar.`,
		Example: `  ctbz mensalidade segunda-via
  ctbz mensalidade segunda-via --competencia 2026-10 --yes -o json
  ctbz mensalidade segunda-via --pix --yes
  ctbz mensalidade segunda-via --linha --yes`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if soPix && soLinha {
				return usageError{errors.New("use --pix ou --linha, não os dois")}
			}
			var quero struct{ ano, mes int }
			if competencia != "" {
				m, err := parseMes(competencia)
				if err != nil {
					return err
				}
				quero.ano, quero.mes = m.Year(), int(m.Month())
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			fat, err := api.BuscarFatura(ctx, g)
			if err != nil {
				return err
			}
			caminho, err := acaoSegundaVia(fat)
			if err != nil {
				return err
			}
			ano, mes, temMes := api.CompetenciaSegundaVia(caminho)
			if competencia != "" && (!temMes || ano != quero.ano || mes != quero.mes) {
				oferece := "outra competência"
				if temMes {
					oferece = fmt.Sprintf("%02d/%d", mes, ano)
				}
				return fmt.Errorf("o card oferece a segunda via de %s, não de %02d/%d", oferece, quero.mes, quero.ano)
			}
			rotulo, id := "atual", ""
			if temMes {
				rotulo, id = fmt.Sprintf("de %02d/%d", mes, ano), fmt.Sprintf("%d%02d", ano, mes)
			}
			op := operacao{Risco: riscoMedio, ID: id,
				Resumo: fmt.Sprintf("Gerar a segunda via da mensalidade %s (registra um boleto novo, com vencimento novo; não cobra nada)", rotulo)}
			var nova *api.NovaFatura
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				var err error
				nova, err = api.GerarSegundaVia(ctx, snd, caminho)
				return err
			})
			if err != nil || !enviado {
				return err
			}
			idFatura, ok := api.IDFaturaDaURL(nova.URL)
			if !ok {
				return fmt.Errorf("a segunda via foi gerada, mas a resposta não trouxe o identificador da fatura; não gere outra: %s", linkOuPainel(nova.URL))
			}
			pub, err := api.BuscarFaturaPublica(ctx, g, idFatura)
			if err != nil {
				return falhaLeituraSegundaVia(err, idFatura, nova.URL)
			}
			resumo, err := api.BuscarFaturaPublicaResumo(ctx, g, idFatura)
			if err != nil {
				return falhaLeituraSegundaVia(err, idFatura, nova.URL)
			}
			pix, linha := textoPagamento(pub.MeioDePagamento.Pix), textoPagamento(pub.MeioDePagamento.Boleto)
			switch {
			case soPix && pix == "":
				return fmt.Errorf("a segunda via foi gerada, mas a fatura não oferece Pix; abra o link para pagar: %s", nova.URL)
			case soLinha && linha == "":
				return fmt.Errorf("a segunda via foi gerada, mas a fatura não oferece boleto; abra o link para pagar: %s", nova.URL)
			case soPix:
				fmt.Fprintln(s.out, pix)
			case soLinha:
				fmt.Fprintln(s.out, linha)
			default:
				if err := output.Write(s.out, f, segundaViaRecord(id, ano, mes, temMes, pub, resumo, pix, linha, nova.URL)); err != nil {
					return err
				}
			}
			fmt.Fprintln(s.err, "Depois de pagar, confirme com `ctbz mensalidade situacao`: a fatura pública demora a dar baixa.")
			return nil
		},
	}
	cmd.Flags().StringVar(&competencia, "competencia", "", "competência AAAA-MM esperada; o card precisa oferecer a segunda via dela")
	cmd.Flags().BoolVar(&soPix, "pix", false, "imprime só o Pix copia e cola (para pipe)")
	cmd.Flags().BoolVar(&soLinha, "linha", false, "imprime só a linha digitável do boleto (para pipe)")
	addWriteFlags(cmd)
	return cmd
}

// acaoSegundaVia acha, entre as ações do card, a que gera a segunda via ("request:<caminho>"),
// e valida o caminho. Ações "redirect:" são navegação do painel.
func acaoSegundaVia(fat *api.Fatura) (string, error) {
	for _, a := range []string{fat.BotaoAdicionalAcao, fat.BotaoAcao} {
		acao, ok := api.ParseAcaoCard(a)
		if ok && acao.Tipo == "request" {
			return api.CaminhoSegundaVia(acao.Valor)
		}
	}
	return "", errors.New("o card da mensalidade não oferece segunda via agora (a fatura pode estar em dia; veja ctbz mensalidade)")
}

// textoPagamento devolve o código de um meio de pagamento exibido ("" se não houver).
func textoPagamento(m *api.MeioPagamento) string {
	if m == nil || !m.Exibir {
		return ""
	}
	return strings.TrimSpace(m.TextoPagamento)
}

// falhaLeituraSegundaVia explica uma leitura que falhou depois de gerar a segunda via: o
// identificador da fatura sai do erro (ele dá acesso à fatura) e só aparece no link.
func falhaLeituraSegundaVia(err error, idFatura, link string) error {
	causa := strings.ReplaceAll(err.Error(), idFatura, "***")
	return fmt.Errorf("a segunda via foi gerada, mas não foi possível ler a fatura (%s); não gere outra; abra o link para pagar: %s", causa, link)
}

func linkOuPainel(link string) string {
	if link == "" {
		return "veja o card da mensalidade no painel"
	}
	return "abra o link para pagar: " + link
}

func segundaViaRecord(id string, ano, mes int, temMes bool, pub *api.FaturaPublica, resumo *api.FaturaPublicaResumo, pix, linha, link string) *output.Record {
	var competencia, status, codigoBarras any
	if temMes {
		competencia = fmt.Sprintf("%02d/%d", mes, ano)
	}
	if st := resumo.Status; st != nil {
		status = nilIfEmpty(st.ID)
		if st.Descricao != "" {
			status = st.Descricao
		}
	}
	if b := pub.MeioDePagamento.Boleto; b != nil && b.Exibir {
		codigoBarras = nilIfEmpty(strings.TrimSpace(b.URLPagamento))
	}
	itens := make([]output.Record, 0, len(resumo.Resumo.Itens))
	for _, it := range resumo.Resumo.Itens {
		r := output.Record{}
		r.Add("descricao", "Descrição", semControle(it.Descricao)).Add("valor", "Valor", moneyOrNil(it.Valor))
		itens = append(itens, r)
	}
	return resultadoEscrita("segunda-via", "gerada", nilIfEmpty(id)).
		Add("competencia", "Competência", competencia).
		Add("vencimento", "Vencimento", dataDeValor(pub.DataVencimento)).
		Add("total", "Total", moneyOrNil(pub.Total)).
		Add("status", "Situação da fatura", status).
		Add("itens", "Itens", itens).
		Add("pix_copia_e_cola", "Pix copia e cola", nilIfEmpty(pix)).
		Add("linha_digitavel", "Linha digitável", nilIfEmpty(linha)).
		Add("codigo_barras", "Código de barras", codigoBarras).
		Add("link", "Link", nilIfEmpty(link))
}
