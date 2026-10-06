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
			total := centavosPtr(d.Saldo) + centavosPtr(d.TotalDistribuido)
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

// montarDistribuicao calcula, em centavos, o valor e o percentual de cada sócio da
// distribuição (os não citados ficam com zero) e confere que a soma é o total. A diferença
// de arredondamento dos percentuais vai para o último sócio citado em percentual.
func montarDistribuicao(lucros []api.LucroSocio, nomes map[string]string, partes []parteSocio, total int64) ([]api.DistribuicaoSocio, string, error) {
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
	var ultimoPct *parteSocio
	var soma int64
	for i := range partes {
		p := &partes[i]
		if p.percentual {
			p.centavos = (total*p.centesimos + 5000) / 10000
			ultimoPct = p
		}
		soma += p.centavos
	}
	if d := total - soma; ultimoPct != nil && d != 0 && abs64(d) <= int64(len(partes)) {
		ultimoPct.centavos += d
		soma += d
	}
	if soma != total {
		return nil, "", usageError{fmt.Errorf("a soma das partes (%s) deve ser igual ao lucro total do exercício (%s)", formatarCentavos(soma), formatarCentavos(total))}
	}
	itens := make([]api.DistribuicaoSocio, 0, len(lucros))
	var resumo []string
	for _, l := range lucros {
		id := idTexto(l.ID)
		var cent, pct int64
		if p, ok := porID[id]; ok {
			cent, pct = p.centavos, p.centesimos
			if !p.percentual {
				pct = (p.centavos*10000 + total/2) / total
			}
		}
		itens = append(itens, api.NovaDistribuicaoSocio(l.ID, cent, pct))
		resumo = append(resumo, fmt.Sprintf("%s %s%% (%s)", firstNonEmpty(nomes[id], id), strings.Replace(fmt.Sprintf("%d.%02d", pct/100, pct%100), ".", ",", 1), formatarCentavos(cent)))
	}
	return itens, strings.Join(resumo, "; "), nil
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
