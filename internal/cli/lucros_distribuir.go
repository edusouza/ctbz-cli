package cli

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// parteSocio é o que --socio pede para um sócio: um percentual (em centésimos de ponto) ou
// um valor (em centavos).
type parteSocio struct {
	id         string
	centesimos int64 // 6000 = 60,00%
	centavos   int64
	percentual bool
}

func newLucrosDistribuirCmd() *cobra.Command {
	var especs []string
	cmd := &cobra.Command{
		Use:   "distribuir",
		Short: "Registra quanto do lucro do exercício cabe a cada sócio",
		Long: `Registra a distribuição de lucros do exercício aberto, que vai para o informe de
rendimentos dos sócios. Cada --socio ID=parte diz a parte de um sócio, em percentual (60%) ou
em reais (30000,00); os sócios não citados ficam com zero. A soma precisa ser o lucro total
(saldo na empresa + o já distribuído), como no painel.

Os IDs dos sócios são o id de cada sócio em ctbz lucros. A distribuição só é aceita
enquanto o painel deixa alterar (até a data limite) e sem restrições no informe (pendência
documental ou débitos federais, também em ctbz lucros).

Risco alto: define os rendimentos isentos que os sócios declaram no IRPF. Dá para refazer
enquanto a distribuição puder ser alterada. Aceita --yes e --dry-run.`,
		Example: `  ctbz lucros distribuir --socio 1000000000000001=60% --socio 1000000000000002=40%
  ctbz lucros distribuir --socio 1000000000000001=30000,00 --socio 1000000000000002=20000,00`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			partes, err := lerPartesSocios(especs)
			if err != nil {
				return err
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			g := sessionGetter{s}
			d, err := api.BuscarDistribuicaoLucros(cmd.Context(), g)
			if err != nil {
				return err
			}
			if err := podeDistribuir(cmd, g, d); err != nil {
				return err
			}
			nomes, err := nomesSocios(cmd, g, d)
			if err != nil {
				return err
			}
			total, err := lucroTotal(d)
			if err != nil {
				return err
			}
			if total <= 0 {
				return errors.New("não há lucro para distribuir no exercício")
			}
			itens, resumo, err := montarDistribuicao(d.LucrosSocios, nomes, partes, total)
			if err != nil {
				return err
			}
			op := operacao{Risco: riscoAlto, ID: fmt.Sprint(d.Ano),
				Resumo:       fmt.Sprintf("Distribuir %s de lucro do exercício %d: %s", formatarCentavos(total), d.Ano, resumo),
				Consequencia: "Define os rendimentos isentos que os sócios declaram no IRPF; dá para refazer enquanto a distribuição puder ser alterada."}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				return api.SalvarDistribuicaoLucros(cmd.Context(), snd, itens)
			})
			if err != nil || !enviado {
				return err
			}
			relida, err := api.BuscarDistribuicaoLucros(cmd.Context(), g)
			if err != nil {
				return err
			}
			valores := map[string]int64{}
			for _, l := range relida.LucrosSocios {
				if l.Valor != nil {
					valores[idTexto(l.ID)] = centavos(*l.Valor)
				}
			}
			recs := make([]output.Record, len(itens))
			for i, it := range itens {
				id := idTexto(it.IDSocio)
				enviadoCentavos, _ := parseValor(it.Valor)
				situacao := "enviado"
				if v, ok := valores[id]; ok && v == enviadoCentavos {
					situacao = "registrado"
				}
				recs[i] = *resultadoEscrita("distribuir", situacao, id).Add("socio", "Sócio", nilIfEmpty(nomes[id])).
					Add("percentual", "Percentual", it.Percentual).Add("valor", "Valor", output.Money(reais(enviadoCentavos)))
			}
			return output.Write(s.out, f, output.RecordsToList(recs))
		},
	}
	cmd.Flags().StringArrayVar(&especs, "socio", nil, "parte de um sócio: ID=60% ou ID=30000,00 (repetível)")
	addWriteFlags(cmd)
	return cmd
}

// lerPartesSocios lê os --socio ID=parte.
func lerPartesSocios(especs []string) ([]parteSocio, error) {
	if len(especs) == 0 {
		return nil, usageError{errors.New("informe a parte de cada sócio com --socio ID=60% ou --socio ID=30000,00")}
	}
	vistos := map[string]bool{}
	partes := make([]parteSocio, 0, len(especs))
	for _, e := range especs {
		id, parte, ok := strings.Cut(e, "=")
		id, parte = strings.TrimSpace(id), strings.TrimSpace(parte)
		if !ok || id == "" || parte == "" {
			return nil, usageError{fmt.Errorf("--socio deve ser ID=60%% ou ID=30000,00: %q", e)}
		}
		if vistos[id] {
			return nil, usageError{fmt.Errorf("sócio %s repetido em --socio", id)}
		}
		vistos[id] = true
		p := parteSocio{id: id}
		var err error
		if num, ok := strings.CutSuffix(parte, "%"); ok {
			p.percentual = true
			p.centesimos, err = parseValor(num)
		} else {
			p.centavos, err = parseValor(parte)
		}
		if err != nil || p.centesimos < 0 || p.centavos < 0 || p.centesimos > 10000 {
			return nil, usageError{fmt.Errorf("--socio %s: parte inválida %q (use 60%% ou 30000,00)", id, parte)}
		}
		partes = append(partes, p)
	}
	return partes, nil
}

// podeDistribuir confere o que o painel confere: exercício aberto, alteração permitida e
// informe sem restrições.
func podeDistribuir(cmd *cobra.Command, g api.Getter, d *api.DistribuicaoLucros) error {
	if d.Ano == 0 {
		return errors.New("não há exercício aberto para distribuir lucros")
	}
	if !d.PodeAlterar {
		msg := "a distribuição de lucros não pode mais ser alterada"
		if m, ok := textoDeValor(d.MotivoNaoPodeAlterar).(string); ok {
			msg += ": " + m
		}
		return errors.New(msg)
	}
	if limite, ok := dataDeValor(d.DataLimite).(output.Date); ok && !limite.IsZero() && diasEntre(now(), limite.Time) < 0 {
		return fmt.Errorf("o prazo para distribuir os lucros de %d acabou em %s", d.Ano, limite.Format("02/01/2006"))
	}
	r, err := api.BuscarRestricoesInforme(cmd.Context(), g, d.Ano)
	if err != nil {
		return err
	}
	var restricoes []string
	if r.Restricoes.PendenciaDocumental.PossuiPendencia {
		restricoes = append(restricoes, "pendência documental")
	}
	if r.Restricoes.DebitosFederais.PossuiPendencia {
		restricoes = append(restricoes, "débitos federais")
	}
	if len(restricoes) > 0 {
		return fmt.Errorf("o informe de %d tem restrições (%s): resolva-as antes de distribuir (veja ctbz lucros)", d.Ano, strings.Join(restricoes, " e "))
	}
	return nil
}

// nomesSocios dá nome aos sócios da distribuição: o da própria distribuição ou o do cadastro.
func nomesSocios(cmd *cobra.Command, g api.Getter, d *api.DistribuicaoLucros) (map[string]string, error) {
	if len(d.LucrosSocios) == 0 {
		return nil, errors.New("a Contabilizei não listou os sócios da distribuição; distribua pelo painel")
	}
	socios, err := api.BuscarSocios(cmd.Context(), g)
	if err != nil {
		return nil, err
	}
	nomes := map[string]string{}
	for _, so := range socios {
		nomes[strconv.FormatInt(so.ID, 10)] = so.Nome
	}
	for _, l := range d.LucrosSocios {
		if l.Socio != "" {
			nomes[idTexto(l.ID)] = l.Socio
		}
	}
	return nomes, nil
}

// lucroTotal é o lucro que as partes precisam somar: o saldo na empresa mais o já
// distribuído, como no painel. Um dos dois ausente é erro de leitura, não zero (ADR-0021):
// sem ele, os percentuais sairiam de um total menor que o real.
func lucroTotal(d *api.DistribuicaoLucros) (int64, error) {
	if d.Saldo == nil {
		return 0, fmt.Errorf("a Contabilizei não informou o saldo de lucros do exercício %d; distribua pelo painel", d.Ano)
	}
	if d.TotalDistribuido == nil {
		return 0, fmt.Errorf("a Contabilizei não informou o total já distribuído do exercício %d; distribua pelo painel", d.Ano)
	}
	return centavos(*d.Saldo) + centavos(*d.TotalDistribuido), nil
}

// conferirIDsSocios recusa a distribuição cujos sócios não dá para identificar: o id de cada
// item vai como idSocio na escrita, e um id ausente ou repetido iria nulo ou ambíguo.
func conferirIDsSocios(lucros []api.LucroSocio) error {
	vistos := map[string]bool{}
	for _, l := range lucros {
		id := idTexto(l.ID)
		switch l.ID.(type) {
		case float64, string:
		default:
			id = ""
		}
		if strings.TrimSpace(id) == "" {
			return errors.New("a Contabilizei listou um sócio da distribuição sem id; distribua pelo painel")
		}
		if vistos[id] {
			return fmt.Errorf("a Contabilizei listou o sócio %s mais de uma vez na distribuição; distribua pelo painel", id)
		}
		vistos[id] = true
	}
	return nil
}

// montarDistribuicao calcula, em centavos, o valor e o percentual de cada sócio da
// distribuição (os não citados ficam com zero) e confere que a soma é o total. Só a diferença
// que o arredondamento dos percentuais explica é ajustada (ver ajustarArredondamento); os
// percentuais enviados saem dos valores finais e somam 100,00%.
func montarDistribuicao(lucros []api.LucroSocio, nomes map[string]string, partes []parteSocio, total int64) ([]api.DistribuicaoSocio, string, error) {
	if err := conferirIDsSocios(lucros); err != nil {
		return nil, "", err
	}
	naDistribuicao := map[string]bool{}
	for _, l := range lucros {
		naDistribuicao[idTexto(l.ID)] = true
	}
	porID := map[string]*parteSocio{}
	for i := range partes {
		if !naDistribuicao[partes[i].id] {
			return nil, "", usageError{fmt.Errorf("sócio %s não está na distribuição; os sócios dela são %s (veja ctbz lucros)", partes[i].id, sociosDaDistribuicao(lucros, nomes))}
		}
		porID[partes[i].id] = &partes[i]
	}
	var soma int64
	for i := range partes {
		p := &partes[i]
		if p.percentual {
			p.centavos = (total*p.centesimos + 5000) / 10000
		}
		soma += p.centavos
	}
	soma += ajustarArredondamento(partes, total, total-soma)
	if soma != total {
		return nil, "", usageError{fmt.Errorf("a soma das partes (%s) deve ser igual ao lucro total do exercício (%s)", formatarCentavos(soma), formatarCentavos(total))}
	}
	cents := make([]int64, len(lucros))
	for i, l := range lucros {
		if p, ok := porID[idTexto(l.ID)]; ok {
			cents[i] = p.centavos
		}
	}
	pcts := percentuaisDe(cents, total)
	itens := make([]api.DistribuicaoSocio, 0, len(lucros))
	var resumo []string
	for i, l := range lucros {
		id := idTexto(l.ID)
		cent, pct := cents[i], pcts[i]
		itens = append(itens, api.NovaDistribuicaoSocio(l.ID, cent, pct))
		resumo = append(resumo, fmt.Sprintf("%s %s%% (%s)", firstNonEmpty(nomes[id], id), strings.Replace(fmt.Sprintf("%d.%02d", pct/100, pct%100), ".", ",", 1), formatarCentavos(cent)))
	}
	return itens, strings.Join(resumo, "; "), nil
}

// ajustarArredondamento acerta a diferença d (total - soma) que vem do arredondamento dos
// percentuais e devolve quanto ajustou: 0 quando d não se explica por arredondamento, e a soma
// fica errada. Cada percentual não nulo arredonda no máximo meio centavo, então |d| não passa
// de meio centavo por sócio com percentual não nulo. Cada centavo vai (ou sai) de um sócio
// diferente, do último citado para o primeiro, entre os que arredondaram no sentido oposto:
// quem pediu 0% nunca ganha centavo e ninguém fica abaixo de zero (#779).
func ajustarArredondamento(partes []parteSocio, total, d int64) int64 {
	var naoNulos int64
	for _, p := range partes {
		if p.percentual && p.centesimos > 0 {
			naoNulos++
		}
	}
	if d == 0 || 2*abs64(d) > naoNulos {
		return 0
	}
	passo := int64(1)
	if d < 0 {
		passo = -1
	}
	var ajustar []*parteSocio
	for i := len(partes) - 1; i >= 0 && int64(len(ajustar)) < abs64(d); i-- {
		p := &partes[i]
		if !p.percentual || p.centesimos == 0 {
			continue
		}
		// erro > 0: arredondou para cima (pode perder um centavo); < 0: para baixo (pode ganhar).
		erro := p.centavos*10000 - total*p.centesimos
		if erro*passo < 0 {
			ajustar = append(ajustar, p)
		}
	}
	if int64(len(ajustar)) < abs64(d) {
		return 0
	}
	for _, p := range ajustar {
		p.centavos += passo
	}
	return d
}

// percentuaisDe calcula, em centésimos de ponto, o percentual de cada valor sobre o total
// (a soma dos valores é o total) pelo método do maior resto: os percentuais somam
// exatamente 100,00% e cada um fica coerente com o valor enviado (#789).
func percentuaisDe(cents []int64, total int64) []int64 {
	pcts := make([]int64, len(cents))
	restos := make([]int64, len(cents))
	falta := int64(10000)
	for i, c := range cents {
		pcts[i], restos[i] = c*10000/total, c*10000%total
		falta -= pcts[i]
	}
	ordem := make([]int, len(cents))
	for i := range ordem {
		ordem[i] = i
	}
	sort.SliceStable(ordem, func(a, b int) bool { return restos[ordem[a]] > restos[ordem[b]] })
	for _, i := range ordem[:falta] {
		pcts[i]++
	}
	return pcts
}

// sociosDaDistribuicao lista os ids aceitos em --socio, com o nome quando há.
func sociosDaDistribuicao(lucros []api.LucroSocio, nomes map[string]string) string {
	ids := make([]string, 0, len(lucros))
	for _, l := range lucros {
		id := idTexto(l.ID)
		if n := nomes[id]; n != "" {
			id += " (" + n + ")"
		}
		ids = append(ids, id)
	}
	return strings.Join(ids, ", ")
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
