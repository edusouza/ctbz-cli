package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// prefixoErroDocumento é o prefixo que o painel tira das mensagens de erro do envio.
const prefixoErroDocumento = "(ERRO-NA-VALIDACAO-DE-DOCUMENTO-API-CONTABIL)"

func semPrefixoDocumento(err error) error {
	var we *ctbz.WriteError
	if errors.As(err, &we) && strings.Contains(we.Message, prefixoErroDocumento) {
		return fmt.Errorf("a Contabilizei recusou o documento: %s", strings.Join(strings.Fields(strings.ReplaceAll(we.Message, prefixoErroDocumento, "")), " "))
	}
	return err
}

// idsPendenciasDocumento usa --pendencia ou, sem ele, as pendências das rotinas do painel.
func idsPendenciasDocumento(ctx context.Context, g api.Getter, ids []string) ([]string, error) {
	if len(ids) > 0 {
		return ids, nil
	}
	c, err := api.BuscarCentralRotinas(ctx, g)
	if err != nil {
		return nil, err
	}
	for _, r := range c.Rotinas {
		ids = append(ids, r.IDsPendentes()...)
	}
	return ids, nil
}

func competenciaDoc(d api.DocumentoPendente) any {
	if c := d.Propriedades.Competencia; c != nil && c.Mes > 0 {
		return fmt.Sprintf("%02d/%d", c.Mes, c.Ano)
	}
	return nil
}

func contaDoc(d api.DocumentoPendente) any {
	p := d.Propriedades
	if p.Banco == "" && p.ContaCorrente == "" {
		return nil
	}
	return strings.TrimSpace(fmt.Sprintf("%s ag. %s conta %s", p.Banco, p.Agencia, p.ContaCorrente))
}

func rawTexto(r []byte) any {
	s := strings.Trim(string(r), `"`)
	if s == "" || s == "null" {
		return nil
	}
	return s
}

func newDocumentosPendentesCmd() *cobra.Command {
	var ids []string
	cmd := &cobra.Command{
		Use:   "pendentes",
		Short: "Lista os documentos pedidos pelas pendências da Central de Rotinas",
		Long: `Lista os documentos que as pendências pedem (extrato de aplicação, informe de
investimentos, contratos, estoque, intermediações…): id da pendência, tipo, competência e
conta. Sem --pendencia, usa as pendências das rotinas do painel (coluna pendencias de
ctbz rotinas).

Envie com ctbz documentos enviar ARQUIVO --pendencia ID, ou declare que o documento não
existe com ctbz documentos sem-arquivo.`,
		Example: `  ctbz documentos pendentes
  ctbz documentos pendentes --pendencia 1000000000000601 -o json`,
		Args: exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			ids, err := idsPendenciasDocumento(ctx, g, ids)
			if err != nil {
				return err
			}
			l := &output.List{Columns: []output.Column{
				{Key: "id_pendencia", Header: "Pendência"},
				{Key: "tipo", Header: "Tipo"},
				{Key: "competencia", Header: "Competência"},
				{Key: "conta", Header: "Conta"},
				{Key: "tipo_investimento", Header: "Investimento"},
			}}
			if len(ids) == 0 {
				fmt.Fprintln(s.err, "Nenhuma pendência de documento nas rotinas.")
				return output.Write(s.out, f, l)
			}
			ini, err := api.BuscarEnvioDocumento(ctx, g, ids)
			if err != nil {
				return err
			}
			for _, d := range ini.Documentos {
				l.Append(d.Propriedades.IDTexto(), d.Tipo, competenciaDoc(d), contaDoc(d), rawTexto(d.Propriedades.TipoInvestimento))
			}
			return output.Write(s.out, f, l)
		},
	}
	cmd.Flags().StringSliceVar(&ids, "pendencia", nil, "id da pendência (repita ou separe por vírgula)")
	return cmd
}

// documentosDasPendencias lê o init e devolve os documentos das pendências pedidas, na ordem,
// todos do mesmo tipo.
func documentosDasPendencias(ctx context.Context, g api.Getter, ids []string) ([]api.DocumentoPendente, error) {
	ini, err := api.BuscarEnvioDocumento(ctx, g, ids)
	if err != nil {
		return nil, err
	}
	porID := map[string]api.DocumentoPendente{}
	for _, d := range ini.Documentos {
		porID[d.Propriedades.IDTexto()] = d
	}
	var ds []api.DocumentoPendente
	for _, id := range ids {
		d, ok := porID[id]
		if !ok {
			return nil, fmt.Errorf("a pendência %s não pede documento (ou já foi resolvida); veja ctbz documentos pendentes", id)
		}
		if len(ds) > 0 && d.Tipo != ds[0].Tipo {
			return nil, usageError{fmt.Errorf("as pendências têm tipos diferentes (%s e %s); envie um arquivo por tipo", ds[0].Tipo, d.Tipo)}
		}
		ds = append(ds, d)
	}
	return ds, nil
}

// pendenciasAbertas relê o init e conta quantas das pendências ainda pedem documento.
func pendenciasAbertas(ctx context.Context, g api.Getter, ids []string) (int, error) {
	ini, err := api.BuscarEnvioDocumento(ctx, g, ids)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, d := range ini.Documentos {
		for _, id := range ids {
			if d.Propriedades.IDTexto() == id {
				n++
			}
		}
	}
	return n, nil
}

func newDocumentosEnviarCmd() *cobra.Command {
	var ids []string
	cmd := &cobra.Command{
		Use:   "enviar ARQUIVO",
		Short: "Envia o documento pedido por uma ou mais pendências",
		Long: `Envia o arquivo que uma pendência pede (risco médio: envia um documento contábil e resolve
a pendência; não há exclusão pela API). O tipo do documento e os metadados (competência, conta,
investimento) vêm da pendência, nunca são digitados.

Com várias --pendencia do mesmo tipo, o arquivo é enviado uma vez para todas ("em um arquivo
único" no painel). Extratos de movimentação usam ctbz extratos importar.

Pede confirmação (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz documentos enviar aplicacao-agosto.pdf --pendencia 1000000000000601
  ctbz documentos enviar aplicacoes.pdf --pendencia 1000000000000601 --pendencia 1000000000000602`,
		Args: exactArgs(1, "o ARQUIVO"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(ids) == 0 {
				return usageError{errors.New("informe --pendencia ID (ver ctbz documentos pendentes)")}
			}
			fi, err := os.Stat(args[0])
			if err != nil {
				return usageError{fmt.Errorf("lendo o arquivo: %w", err)}
			}
			if fi.Size() > api.TamanhoMaximoExtrato {
				return usageError{fmt.Errorf("o arquivo tem %d MB; o limite é 30 MB", fi.Size()>>20)}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			ds, err := documentosDasPendencias(ctx, g, ids)
			if err != nil {
				return err
			}
			if ds[0].Tipo == api.TipoExtratoMovimentacoes {
				return usageError{errors.New("extrato de movimentação se envia com ctbz extratos importar")}
			}
			conteudo, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			nome := filepath.Base(args[0])
			var comps []string
			for _, d := range ds {
				if c, ok := competenciaDoc(d).(string); ok {
					comps = append(comps, c)
				}
			}
			op := operacao{Risco: riscoMedio, ID: strings.Join(ids, ","), Resumo: fmt.Sprintf("Enviar %s (%s) como %s para %d pendência(s) (%s)",
				nome, tamanhoArquivo(len(conteudo)), ds[0].Tipo, len(ds), strings.Join(comps, ", "))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				if len(ds) == 1 {
					return semPrefixoDocumento(api.EnviarDocumento(ctx, snd, ds[0], nome, conteudo))
				}
				return semPrefixoDocumento(api.EnviarDocumentoConsolidado(ctx, snd, ds, nome, conteudo))
			})
			if err != nil || !enviado {
				return err
			}
			situacao := "enviado"
			if n, err := pendenciasAbertas(ctx, g, ids); err != nil {
				fmt.Fprintln(s.err, "aviso: documento enviado, mas não foi possível reler as pendências:", err)
			} else if n == 0 {
				situacao = "resolvido"
			} else {
				fmt.Fprintf(s.err, "aviso: %d pendência(s) ainda pedem documento; confira com ctbz documentos pendentes\n", n)
			}
			rec := resultadoEscrita("enviar-documento", situacao, strings.Join(ids, ",")).
				Add("tipo", "Tipo", ds[0].Tipo).
				Add("arquivo", "Arquivo", nome)
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().StringSliceVar(&ids, "pendencia", nil, "id da pendência (repita para enviar um arquivo único para várias)")
	addWriteFlags(cmd)
	return cmd
}

// consequenciasSemArquivo são os textos do painel para cada declaração sem arquivo.
var consequenciasSemArquivo = map[string]string{
	"ESTOQUE":                    "Declara que a empresa não teve estoque no período, que será excluído. Para empresas de comércio, a falta desse documento pode gerar inconsistências na declaração anual.",
	"CONTROLE_DE_INTERMEDIACOES": "Declara que não houve intermediações no período.",
	"CONTRATO_DE_AFAC":           "Declara que o contrato de AFAC já foi enviado: o documento recebido antes será usado como comprovante.",
}

func newDocumentosSemArquivoCmd() *cobra.Command {
	var ids []string
	cmd := &cobra.Command{
		Use:   "sem-arquivo",
		Short: "Declara que o documento pedido por pendências não existe",
		Long: `Fecha pendências declarando que o documento não existe: "não tive estoque", "não tive
intermediações" ou "já enviei o contrato de AFAC" (risco alto: é uma declaração, sem
desfazer pela API). Só vale para os tipos ESTOQUE, CONTROLE_DE_INTERMEDIACOES e
CONTRATO_DE_AFAC; o tipo é lido da pendência.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz documentos sem-arquivo --pendencia 1000000000000603`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(ids) == 0 {
				return usageError{errors.New("informe --pendencia ID (ver ctbz documentos pendentes)")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			ds, err := documentosDasPendencias(ctx, g, ids)
			if err != nil {
				return err
			}
			tipo := ds[0].Tipo
			if !api.TiposSemArquivo[tipo] {
				return usageError{fmt.Errorf("documentos do tipo %s não aceitam a declaração sem arquivo (só ESTOQUE, CONTROLE_DE_INTERMEDIACOES e CONTRATO_DE_AFAC)", tipo)}
			}
			itens := make([]api.PendenciaSemArquivo, len(ds))
			var comps []string
			for i, d := range ds {
				itens[i] = api.PendenciaSemArquivo{IDPendencia: d.Propriedades.IDPendencia, Tipo: d.Tipo}
				if c, ok := competenciaDoc(d).(string); ok {
					comps = append(comps, c)
				}
			}
			op := operacao{Risco: riscoAlto, ID: strings.Join(ids, ","), Consequencia: consequenciasSemArquivo[tipo] + " Não dá para desfazer.",
				Resumo: fmt.Sprintf("Declarar que não há %s em %d pendência(s) (%s)", tipo, len(ds), strings.Join(comps, ", "))}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.EnviarSemArquivo(ctx, snd, itens) })
			if err != nil || !enviado {
				return err
			}
			situacao := "declarado"
			if n, err := pendenciasAbertas(ctx, g, ids); err != nil {
				fmt.Fprintln(s.err, "aviso: declaração enviada, mas não foi possível reler as pendências:", err)
				situacao = "enviado"
			} else if n > 0 {
				fmt.Fprintf(s.err, "aviso: %d pendência(s) ainda aparecem; confira com ctbz documentos pendentes\n", n)
				situacao = "enviado"
			}
			return output.Write(s.out, f, resultadoEscrita("sem-arquivo", situacao, strings.Join(ids, ",")).Add("tipo", "Tipo", tipo))
		},
	}
	cmd.Flags().StringSliceVar(&ids, "pendencia", nil, "id da pendência (repita para várias do mesmo tipo)")
	addWriteFlags(cmd)
	return cmd
}

func newDocumentosSemAplicacaoCmd() *cobra.Command {
	var ids []string
	var conta int64
	cmd := &cobra.Command{
		Use:   "sem-aplicacao",
		Short: "Declara que não houve aplicação financeira numa conta nas competências pendentes",
		Long: `Resolve pendências de extrato de aplicação financeira declarando que a conta não teve
aplicação no período ("Não tenho aplicação nesta conta" no painel; risco alto: é uma
declaração, sem desfazer pela API). As pendências precisam ser de extrato de aplicação da
conta informada.

Se a Contabilizei aceitar só parte das competências, a CLI lista as que falharam e termina
com código 1.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz documentos sem-aplicacao --conta-bancaria 7 --pendencia 1000000000000601 --pendencia 1000000000000602`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(ids) == 0 || conta <= 0 {
				return usageError{errors.New("informe --conta-bancaria ID e --pendencia ID")}
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			ds, err := documentosDasPendencias(ctx, g, ids)
			if err != nil {
				return err
			}
			if ds[0].Tipo != api.TipoExtratoAplicacao {
				return usageError{fmt.Errorf("as pendências são de %s; sem-aplicacao vale só para %s", ds[0].Tipo, api.TipoExtratoAplicacao)}
			}
			idConta := fmt.Sprint(conta)
			d := api.DeclaracaoSemAplicacao{TipoDocumento: api.TipoExtratoAplicacao, TipoPendencia: api.TipoExtratoAplicacao, IDContaBancaria: []byte(idConta)}
			var comps []string
			for _, doc := range ds {
				if rawTexto(doc.Propriedades.IDContaBancaria) != idConta {
					return usageError{fmt.Errorf("a pendência %s é de outra conta (%v)", doc.Propriedades.IDTexto(), rawTexto(doc.Propriedades.IDContaBancaria))}
				}
				c := api.CompetenciaPendente{ID: doc.Propriedades.IDPendencia}
				if doc.Propriedades.Competencia != nil {
					c.Mes, c.Ano = doc.Propriedades.Competencia.Mes, doc.Propriedades.Competencia.Ano
				}
				d.CompetenciasPendentes = append(d.CompetenciasPendentes, c)
				comps = append(comps, fmt.Sprintf("%02d/%d", c.Mes, c.Ano))
			}
			op := operacao{Risco: riscoAlto, ID: strings.Join(ids, ","), Consequencia: "Declara que não houve aplicação nem movimentação no período, o que resolve as pendências. Não dá para desfazer.",
				Resumo: fmt.Sprintf("Declarar que a conta %d não teve aplicação financeira em %s", conta, strings.Join(comps, ", "))}
			var res *api.ResultadoSemAplicacao
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				var err error
				res, err = api.DeclararSemAplicacao(ctx, snd, d)
				var we *ctbz.WriteError
				if errors.As(err, &we) {
					var parcial api.ResultadoSemAplicacao
					if api.DecodeResponse(we.Body, &parcial) == nil && len(parcial.CompetenciasComErro) > 0 {
						res = &parcial
						return nil
					}
				}
				return err
			})
			if err != nil || !enviado {
				return err
			}
			rec := resultadoEscrita("sem-aplicacao", "declarado", strings.Join(ids, ","))
			if res != nil && len(res.CompetenciasComErro) > 0 {
				falhas := res.CompetenciasComErroTexto()
				rec = resultadoEscrita("sem-aplicacao", "parcial", strings.Join(ids, ",")).Add("competencias_com_erro", "Competências com erro", falhas)
				if err := output.Write(s.out, f, rec); err != nil {
					return err
				}
				return fmt.Errorf("a Contabilizei não aceitou %d competência(s): %s; tente de novo", len(falhas), strings.Join(falhas, ", "))
			}
			return output.Write(s.out, f, rec)
		},
	}
	cmd.Flags().Int64Var(&conta, "conta-bancaria", 0, "id da conta bancária das pendências")
	cmd.Flags().StringSliceVar(&ids, "pendencia", nil, "id da pendência de extrato de aplicação (repita para várias)")
	addWriteFlags(cmd)
	return cmd
}
