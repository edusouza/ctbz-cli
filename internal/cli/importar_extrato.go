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

// errosImportacao traduz os erros de negócio do upload de extrato (identificador → mensagem).
var errosImportacao = map[string]string{
	"exception/movimentacao-financeira-901": "por enquanto a importação deste banco aceita só OFX",
	"exception/movimentacao-financeira-902": "o extrato não corresponde ao banco da conta escolhida",
	"exception/movimentacao-financeira-903": "o extrato não corresponde à conta escolhida",
	"exception/movimentacao-financeira-904": "o extrato não corresponde à agência e à conta escolhidas",
	"exception/movimentacao-financeira-905": "o extrato não corresponde à agência escolhida",
}

// traduzErroImportacao torna os erros conhecidos do upload acionáveis.
func traduzErroImportacao(err error) error {
	var we *ctbz.WriteError
	if !errors.As(err, &we) {
		return err
	}
	if msg, ok := errosImportacao[we.Identificador]; ok {
		return fmt.Errorf("%s (%s)", msg, we.Message)
	}
	switch {
	case strings.Contains(we.Message, "CONTA-NAO-CADASTRADA"):
		return errors.New("a conta bancária do extrato não está cadastrada; cadastre-a com ctbz contas-bancarias adicionar e importe de novo")
	case strings.Contains(we.Message, "ERRO-NA-VALIDACAO-DE-DOCUMENTO-API-CONTABIL"):
		return errors.New("a Contabilizei não reconheceu o arquivo como extrato bancário; tente o OFX do banco ou o PDF " +
			"baixado direto do internet banking e, se continuar, abra um chamado no painel")
	}
	return err
}

// validarArquivoExtrato confere extensão e tamanho e devolve o formato (OFX ou PDF).
func validarArquivoExtrato(caminho string) (string, os.FileInfo, error) {
	ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(caminho), "."))
	if ext != "OFX" && ext != "PDF" {
		return "", nil, usageError{fmt.Errorf("o extrato deve ser .ofx ou .pdf: %s", caminho)}
	}
	fi, err := os.Stat(caminho)
	if err != nil {
		return "", nil, usageError{fmt.Errorf("lendo o arquivo: %w", err)}
	}
	if fi.Size() > api.TamanhoMaximoExtrato {
		return "", nil, usageError{fmt.Errorf("o arquivo tem %d MB; o limite é 30 MB", fi.Size()>>20)}
	}
	return ext, fi, nil
}

func newExtratosImportarCmd() *cobra.Command {
	var saldoFinal string
	var substituir bool
	cmd := &cobra.Command{
		Use:   "importar ARQUIVO",
		Short: "Importa o extrato bancário (OFX ou PDF) de uma conta num mês",
		Long: `Importa o extrato do mês de uma conta bancária, como a tela "Importar extrato" do painel
(risco médio: entra na contabilidade; desfaça com ctbz extratos excluir).

Antes de enviar, a CLI confere a extensão (.ofx ou .pdf), o tamanho (até 30 MB), os formatos
aceitos pelo banco, se a conta é integrada (essas não precisam de importação), se há um
período anterior pendente e se a competência já foi importada ou está em processamento (o
painel também não deixa importar por cima). Depois:

  1. envia o arquivo;
  2. no OFX, lê o saldo do último dia que a Contabilizei encontrou e pede para confirmar
     (ou compara com --saldo-final; diferente, a importação para);
  3. conclui a importação.

Ao final, a CLI relê a competência: em processamento, avisa para conferir depois; importada
sem nenhum lançamento, avisa no stderr.

Para refazer a importação de um mês já importado (ex.: um extrato vazio criado pelo
processamento de outro mês, ou o arquivo errado), use --substituir: a CLI exclui o extrato
atual e importa o arquivo, com uma confirmação só, como excluir e importar no painel. Se o
extrato atual tem lançamentos, o risco é alto (as classificações se perdem; digite
"confirmo"); se está vazio, é médio. Se os contadores já concluíram a classificação, o painel
não deixa excluir, e a CLI também não. Se a exclusão der certo e a importação falhar, o mês
fica pendente: corrija e rode o comando de novo, sem --substituir. Se a conexão cair ou o
servidor responder com erro 5xx ao concluir, não se sabe se o novo arquivo entrou: confira
antes de qualquer nova tentativa, e não use --substituir, que excluiria o extrato novo.

Se a importação parar depois do envio (saldo recusado, extrato de outra competência), o
arquivo fica enviado mas o extrato não é efetivado; corrija e rode o comando de novo.

Se a conexão cair ou o servidor responder com erro 5xx (502, 504…) no envio do arquivo, a CLI
relê a situação do extrato: se a competência continua como estava (ex.: PENDENTE), nada foi
efetivado e é seguro rodar o mesmo comando de novo (o arquivo vai com o mesmo nome e substitui
o anterior). Se isso acontecer ao concluir a importação, ou se a competência já aparecer como
IMPORTADO ou PROCESSANDO, a importação pode ter sido aplicada: confira com ctbz extratos e ctbz
extratos lancamentos antes de repetir, para não importar duas vezes.

Pede confirmação (--yes em scripts) e aceita --dry-run, que mostra o envio e a conclusão sem
enviar nada.`,
		Example: `  ctbz extratos importar extrato-setembro.ofx --conta-bancaria 1000000000000001 --competencia 2026-09
  ctbz extratos importar extrato.ofx --conta-bancaria 1000000000000001 --competencia 2026-09 \
    --saldo-final 1.234,56 --yes
  ctbz extratos importar extrato-setembro.pdf --conta-bancaria 1000000000000001 --competencia 2026-09 \
    --substituir --dry-run`,
		Args: exactArgs(1, "o ARQUIVO do extrato"),
		RunE: func(cmd *cobra.Command, args []string) error {
			alvo, err := extratoFlags(cmd)
			if err != nil {
				return err
			}
			formatoArq, fi, err := validarArquivoExtrato(args[0])
			if err != nil {
				return err
			}
			var saldoEsperado *int64
			if saldoFinal != "" {
				if formatoArq != "OFX" {
					return usageError{errors.New("--saldo-final só vale para extratos OFX")}
				}
				c, err := parseValor(saldoFinal)
				if err != nil {
					return usageError{err}
				}
				saldoEsperado = &c
			}
			f, err := outputFormat(cmd, "")
			if err != nil {
				return err
			}
			s := streamsOf(cmd)
			ctx := cmd.Context()
			g := sessionGetter{s}
			ini, err := api.BuscarUploadExtratoInit(ctx, g)
			if err != nil {
				return err
			}
			if err := validarContaUpload(ini, alvo.conta, formatoArq); err != nil {
				return err
			}
			per, err := api.BuscarPermiteImportacao(ctx, g, alvo.conta, alvo.ano(), alvo.mesNum())
			if err != nil {
				return err
			}
			if err := periodoLiberado(per, alvo); err != nil {
				return err
			}
			existente, err := competenciaImportavel(ctx, s, g, alvo, substituir)
			if err != nil {
				return err
			}
			lista, err := api.BuscarContasBancarias(ctx, g)
			if err != nil {
				return err
			}
			numeroConta := ""
			for _, cb := range lista.ContasBancarias {
				if cb.ID == alvo.conta {
					numeroConta = normalizarConta(cb.ContaCorrente)
				}
			}
			if numeroConta == "" {
				return fmt.Errorf("conta bancária %d não encontrada; veja ctbz contas-bancarias", alvo.conta)
			}
			conteudo, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			store, err := ctbz.DefaultStore()
			if err != nil {
				return err
			}
			sess, err := store.LoadSession()
			if err != nil {
				return err
			}
			cnpj := currentCNPJ(sess)
			nome := api.NomeArquivoExtrato(cnpj, alvo.ano(), alvo.mesNum(), numeroConta, fi.ModTime().UnixMilli(), formatoArq)
			arquivo := api.ArquivoExtrato{NomeArquivo: nome, Ano: alvo.ano(), Mes: alvo.mesNum(), IDConta: alvo.conta,
				Arquivo: filepath.Base(args[0]), Conteudo: conteudo}
			yes, _ := cmd.Flags().GetBool("yes")
			mes := alvo.mes.Format("01/2006")
			op := operacao{Risco: riscoMedio, ID: fmt.Sprint(alvo.conta), Resumo: fmt.Sprintf("Importar o extrato %s (%s, %s) da conta %d em %s",
				arquivo.Arquivo, formatoArq, tamanhoArquivo(len(conteudo)), alvo.conta, mes)}
			if existente != nil {
				atual := "quantidade de lançamentos desconhecida"
				if existente.lancamentos != nil {
					atual = fmt.Sprintf("%d lançamentos", *existente.lancamentos)
				}
				op.Resumo = fmt.Sprintf("Substituir o extrato de %s da conta %d (atual: %s) pelo arquivo %s (%s, %s)",
					mes, alvo.conta, atual, arquivo.Arquivo, formatoArq, tamanhoArquivo(len(conteudo)))
				// Excluir um extrato vazio não perde nada; com lançamentos, perde as classificações.
				if !existente.vazio() {
					op.Risco = riscoAlto
					op.Consequencia = "O extrato atual é excluído antes do envio: as classificações e informações adicionais das movimentações se perdem e não dá para desfazer."
				}
			}
			var saldo *float64
			// importarArquivo envia o arquivo, confirma o saldo (OFX) e conclui a importação.
			importarArquivo := func(snd api.Sender) error {
				if err := api.EnviarExtrato(ctx, snd, arquivo); err != nil {
					if falha, causa, ok := envioIncerto(err); ok {
						return orientarEnvioIncerto(ctx, g, alvo, falha, causa)
					}
					return traduzErroImportacao(err)
				}
				evento := api.EventoUploadExtrato{Ano: alvo.ano(), Mes: alvo.mesNum(), CNPJ: cnpj, IDContaBancaria: alvo.conta, NomeArquivoStorage: nome}
				if formatoArq == "OFX" {
					valor, err := saldoConfirmado(ctx, snd, s, g, alvo, nome, saldoEsperado, yes)
					if err != nil {
						return err
					}
					saldo = &valor
					evento.Respostas = []api.RespostaSaldo{{Tipo: "SALDO_ULTIMO_MES", Data: now().UTC().Format("2006-01-02T15:04:05.000Z"), Valor: valor}}
				}
				if err := api.ConcluirImportacaoExtrato(ctx, snd, evento); err != nil {
					if falha, causa, ok := envioIncerto(err); ok {
						return &importacaoIncerta{fmt.Errorf("o arquivo foi enviado, mas %s ao concluir a importação, que pode ou não ter sido aplicada; confira com %s antes de repetir: %w",
							falha, comandoConferirExtrato(alvo), causa)}
					}
					return fmt.Errorf("o arquivo foi enviado, mas a importação não foi concluída: %w", err)
				}
				return nil
			}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error {
				if existente != nil {
					if err := api.ExcluirExtrato(ctx, snd, alvo.conta, alvo.ano(), alvo.mesNum()); err != nil {
						return fmt.Errorf("excluindo o extrato atual de %s (o novo arquivo não foi enviado): %w", mes, err)
					}
				}
				err := importarArquivo(snd)
				var incerta *importacaoIncerta
				switch {
				case err == nil || existente == nil:
					return err
				case errors.As(err, &incerta):
					return fmt.Errorf("o extrato antigo de %s foi excluído e não se sabe se o novo arquivo foi importado "+
						"(não use --substituir antes de conferir: ele excluiria o extrato novo): %w", mes, err)
				}
				return fmt.Errorf("o extrato antigo de %s foi excluído e a importação voltou a ficar pendente, mas o novo arquivo não foi importado "+
					"(rode ctbz extratos importar de novo, sem precisar de --substituir): %w", mes, err)
			})
			if err != nil || !enviado {
				return err
			}
			situacao := "importado"
			var situacaoExtrato, quantidade any
			if sit, err := situacaoExtratoAlvo(ctx, g, alvo); err != nil {
				fmt.Fprintln(s.err, "aviso: extrato importado, mas não foi possível reler os extratos:", err)
			} else {
				situacaoExtrato = nilIfEmpty(sit)
				quantidade = conferirImportado(ctx, s, g, alvo, sit)
			}
			rec := resultadoEscrita("importar", situacao, alvo.conta).
				Add("competencia", "Competência", mes).
				Add("arquivo", "Arquivo", nome).
				Add("saldo_ultimo_dia", "Saldo do último dia", moneyOrNil(saldo)).
				Add("situacao_extrato", "Situação do extrato", situacaoExtrato).
				Add("quantidade_lancamentos", "Lançamentos", quantidade).
				Add("substituiu_extrato", "Substituiu o extrato anterior", existente != nil)
			return output.Write(s.out, f, rec)
		},
	}
	addContaBancariaFlag(cmd)
	addCompetenciaFlag(cmd)
	cmd.Flags().StringVar(&saldoFinal, "saldo-final", "", "saldo do último dia esperado (OFX); diferente do lido, a importação para")
	cmd.Flags().BoolVar(&substituir, "substituir", false, "se o extrato do mês já foi importado, exclui-o antes e importa este arquivo (como excluir e importar no painel)")
	addWriteFlags(cmd)
	return cmd
}

// situacaoExtratoAlvo relê a situação do extrato da conta na competência ("" se não aparece).
func situacaoExtratoAlvo(ctx context.Context, g api.Getter, alvo extratoAlvo) (string, error) {
	es, err := api.BuscarExtratos(ctx, g)
	if err != nil {
		return "", err
	}
	for _, e := range es {
		if e.IDContaBancaria == alvo.conta && e.Ano == alvo.ano() && e.Mes == alvo.mesNum() {
			return e.Situacao, nil
		}
	}
	return "", nil
}

// extratoExistente é o extrato já importado que --substituir exclui antes de importar.
type extratoExistente struct {
	lancamentos *int // nil quando extrato/info não traz a quantidade
}

// vazio diz se o extrato existente sabidamente não tem lançamentos (nada se perde ao excluir).
func (e *extratoExistente) vazio() bool { return e.lancamentos != nil && *e.lancamentos == 0 }

// competenciaImportavel recusa importar sobre um extrato já importado ou em processamento,
// como o painel, que não mostra "Importar" nesses casos. Um extrato IMPORTADO pode estar
// vazio: o processamento de um extrato de outro mês pode criar o do mês seguinte sem
// lançamentos (#855). Com substituir, um extrato IMPORTADO que o painel deixa excluir é
// devolvido para ser excluído antes do envio, o mesmo caminho do painel (excluir e importar).
func competenciaImportavel(ctx context.Context, s streams, g api.Getter, alvo extratoAlvo, substituir bool) (*extratoExistente, error) {
	sit, err := situacaoExtratoAlvo(ctx, g, alvo)
	if err != nil {
		return nil, err
	}
	mes := alvo.mes.Format("01/2006")
	switch sit {
	case "PROCESSANDO":
		return nil, fmt.Errorf("o extrato de %s ainda está em processamento; aguarde e confira com `ctbz extratos`", mes)
	case "IMPORTADO":
		info, err := api.BuscarExtratoInfo(ctx, g, alvo.conta, alvo.ano(), alvo.mesNum())
		if substituir {
			if err != nil {
				return nil, err
			}
			if !info.PermiteExclusaoExtrato {
				return nil, fmt.Errorf("o extrato de %s não pode ser substituído: os contadores já concluíram a classificação dele (o painel também não deixa excluir)", mes)
			}
			return &extratoExistente{lancamentos: info.QuantidadeLancamentos}, nil
		}
		if err == nil && info.QuantidadeLancamentos != nil && *info.QuantidadeLancamentos == 0 {
			return nil, fmt.Errorf("o extrato de %s já aparece como IMPORTADO, mas sem nenhum lançamento (o processamento do extrato de outro mês "+
				"pode tê-lo criado vazio); para trocá-lo por este arquivo, rode o mesmo comando com --substituir", mes)
		}
		return nil, fmt.Errorf("o extrato de %s já foi importado (veja %s); para trocá-lo por este arquivo, rode o mesmo comando com --substituir "+
			"(risco alto: exclui o extrato atual e as classificações dele)", mes, comandoConferirExtrato(alvo))
	}
	if substituir {
		fmt.Fprintf(s.err, "Nada a substituir: o extrato de %s não foi importado; segue a importação normal.\n", mes)
	}
	return nil, nil
}

// conferirImportado relê o extrato depois da importação e devolve quantos lançamentos ele tem
// (nil se não deu para saber). Em processamento, orienta a conferir depois; importado sem
// lançamentos, avisa: a competência pareceria resolvida sem nada dentro.
func conferirImportado(ctx context.Context, s streams, g api.Getter, alvo extratoAlvo, sit string) any {
	if sit == "PROCESSANDO" {
		fmt.Fprintf(s.err, "A Contabilizei está processando o extrato; em alguns minutos confira com `ctbz extratos` e %s.\n", comandoConferirExtrato(alvo))
	}
	if sit != "IMPORTADO" {
		return nil
	}
	info, err := api.BuscarExtratoInfo(ctx, g, alvo.conta, alvo.ano(), alvo.mesNum())
	if err != nil {
		fmt.Fprintln(s.err, "aviso: extrato importado, mas não foi possível contar os lançamentos:", err)
		return nil
	}
	if info.QuantidadeLancamentos == nil {
		return nil
	}
	if *info.QuantidadeLancamentos == 0 {
		fmt.Fprintf(s.err, "aviso: o extrato de %s ficou IMPORTADO sem nenhum lançamento; confira o arquivo e, se ele tem movimentações, "+
			"abra um chamado no painel\n", alvo.mes.Format("01/2006"))
	}
	return *info.QuantidadeLancamentos
}

// comandoConferirExtrato é o comando que mostra o que entrou no extrato da competência.
func comandoConferirExtrato(alvo extratoAlvo) string {
	return fmt.Sprintf("`ctbz extratos lancamentos --conta-bancaria %d --competencia %s`", alvo.conta, alvo.mes.Format("2006-01"))
}

// envioIncerto diz se a falha de um envio deixa o resultado incerto e devolve como descrevê-la
// e a causa: sem resposta (*erroSemResposta) ou erro 5xx de servidor ou gateway (502, 504…),
// que pode chegar depois de o backend aplicar a escrita. 560 (regra de negócio) e os erros de
// negócio conhecidos são recusas definitivas.
func envioIncerto(err error) (falha string, causa error, ok bool) {
	var semResposta *erroSemResposta
	if errors.As(err, &semResposta) {
		return "a conexão falhou", semResposta.causa, true
	}
	var we *ctbz.WriteError
	if errors.As(err, &we) && we.Status >= 500 && we.Status != 560 && traduzErroImportacao(err) == err {
		return fmt.Sprintf("o servidor respondeu com erro (HTTP %d)", we.Status), err, true
	}
	return "", nil, false
}

// importacaoIncerta é a falha da importação cujo resultado só se sabe conferindo: quem a recebe
// não pode afirmar que nada foi importado nem mandar repetir.
type importacaoIncerta struct{ err error }

func (e *importacaoIncerta) Error() string { return e.err.Error() }

func (e *importacaoIncerta) Unwrap() error { return e.err }

// orientarEnvioIncerto explica uma falha de resultado incerto no envio do arquivo (passo 1).
// Sem a conclusão (passo 3), nada entra na contabilidade: se a competência não está importada
// nem em processamento, repetir é seguro e só substitui o arquivo, que vai com o mesmo nome.
func orientarEnvioIncerto(ctx context.Context, g api.Getter, alvo extratoAlvo, falha string, causa error) error {
	mes := alvo.mes.Format("01/2006")
	sit, rerr := situacaoExtratoAlvo(ctx, g, alvo)
	switch {
	case rerr != nil:
		return &importacaoIncerta{fmt.Errorf("%s no envio do arquivo e a escrita pode ou não ter sido aplicada; confira antes de repetir: %w (e não foi possível reler a situação do extrato: %v)",
			falha, causa, rerr)}
	case sit == "IMPORTADO" || sit == "PROCESSANDO":
		return &importacaoIncerta{fmt.Errorf("%s no envio do arquivo e o extrato de %s aparece como %s; confira com %s antes de repetir: %w",
			falha, mes, sit, comandoConferirExtrato(alvo), causa)}
	}
	situacao := "não aparece entre os extratos"
	if sit != "" {
		situacao = "continua " + sit
	}
	return fmt.Errorf("%s no envio do arquivo, mas o extrato de %s %s: nada foi importado e é seguro rodar o mesmo comando de novo (%v)",
		falha, mes, situacao, causa)
}

// validarContaUpload confere se a conta aceita upload no formato do arquivo.
func validarContaUpload(ini *api.UploadExtratoInit, idConta int64, formato string) error {
	for _, b := range ini.Bancos {
		for _, c := range b.ContasBancarias {
			if c.ID != idConta {
				continue
			}
			if c.StatusIntegracao == "INTEGRADA" {
				return errors.New("a conta é integrada: a importação é feita automaticamente, não precisa enviar extrato")
			}
			for _, fmtOK := range b.FormatosDisponiveisUpload {
				if strings.EqualFold(fmtOK, formato) {
					return nil
				}
			}
			return usageError{fmt.Errorf("o banco desta conta aceita extratos em %s, não %s", strings.Join(b.FormatosDisponiveisUpload, " ou "), formato)}
		}
	}
	return fmt.Errorf("a conta %d não aparece entre as contas para importação; veja ctbz contas-bancarias", idConta)
}

// periodoLiberado recusa a importação quando há um período anterior com importação pendente.
func periodoLiberado(p *api.PermiteImportacao, alvo extratoAlvo) error {
	for _, per := range p.PeriodosDeImportacao {
		anterior := per.Ano < alvo.ano() || (per.Ano == alvo.ano() && per.Mes < alvo.mesNum())
		if per.TemLancamentoPendente && anterior {
			msg := per.Mensagem
			if msg == "" {
				msg = "importe os períodos anteriores primeiro"
			}
			return fmt.Errorf("há importação pendente em %02d/%d: %s", per.Mes, per.Ano, msg)
		}
	}
	return nil
}

// saldoConfirmado lê o saldo do último dia que a Contabilizei extraiu do OFX e o confirma:
// com --saldo-final compara, no terminal pergunta, com --yes aceita. No --dry-run não há
// arquivo enviado: usa --saldo-final ou zero.
func saldoConfirmado(ctx context.Context, snd api.Sender, s streams, g api.Getter, alvo extratoAlvo, nome string, esperado *int64, yes bool) (float64, error) {
	if _, simulado := snd.(dryRunSender); simulado {
		if esperado != nil {
			return reais(*esperado), nil
		}
		fmt.Fprintln(s.err, "Simulação: o saldo do último dia só é conhecido depois do envio; o corpo abaixo usa 0.")
		return 0, nil
	}
	info, err := api.BuscarInfoExtrato(ctx, g, alvo.conta, alvo.ano(), alvo.mesNum(), nome)
	if err != nil {
		return 0, fmt.Errorf("o arquivo foi enviado, mas não foi possível ler o saldo (importação não concluída): %w", err)
	}
	if info.ForaDoPeriodo() {
		return 0, fmt.Errorf("o extrato não é de %s (importação não concluída)", alvo.mes.Format("01/2006"))
	}
	lido := info.Saldo()
	if lido == nil {
		return 0, errors.New("a Contabilizei não leu o saldo do último dia do arquivo (importação não concluída)")
	}
	msg := fmt.Sprintf("Saldo do último dia lido pela Contabilizei: %s", formatarCentavos(centavos(*lido)))
	switch {
	case esperado != nil && *esperado != centavos(*lido):
		return 0, fmt.Errorf("%s, diferente de --saldo-final %s (importação não concluída)", msg, formatarCentavos(*esperado))
	case esperado != nil, yes || !s.interactive:
		fmt.Fprintln(s.err, msg)
	case !perguntar(s, msg+". Confere com o extrato?"):
		return 0, errors.New("saldo não confirmado (o arquivo foi enviado, mas a importação não foi concluída)")
	}
	return *lido, nil
}

// tamanhoArquivo mostra o tamanho em bytes, KB ou MB.
func tamanhoArquivo(n int) string {
	switch {
	case n < 1<<10:
		return fmt.Sprintf("%d bytes", n)
	case n < 1<<20:
		return fmt.Sprintf("%d KB", (n+1<<10-1)>>10)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
}

func newExtratosExcluirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "excluir",
		Short: "Exclui o extrato importado de uma conta num mês",
		Long: `Exclui o extrato importado de uma conta na competência, por exemplo depois de enviar o
arquivo errado (risco alto). A importação volta a ficar pendente e as classificações e
informações adicionais das movimentações se perdem; não dá para desfazer, só reimportar e
reclassificar. Se os contadores já concluíram a classificação, o painel não deixa excluir, e a
CLI também não.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.`,
		Example: `  ctbz extratos excluir --conta-bancaria 1000000000000001 --competencia 2026-09`,
		Args:    exactArgs(0, "nenhum argumento"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			alvo, err := extratoFlags(cmd)
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
			info, err := api.BuscarExtratoInfo(ctx, g, alvo.conta, alvo.ano(), alvo.mesNum())
			if err != nil {
				return err
			}
			if !info.PermiteExclusaoExtrato {
				return errors.New("este extrato não pode ser excluído: os contadores já concluíram a classificação dele")
			}
			op := operacao{Risco: riscoAlto, ID: fmt.Sprint(alvo.conta), Resumo: fmt.Sprintf("Excluir o extrato de %s da conta %d", alvo.mes.Format("01/2006"), alvo.conta),
				Consequencia: "A importação volta a ficar pendente e você terá de importar um novo arquivo. As classificações e informações adicionais das movimentações se perdem. Não dá para desfazer."}
			enviado, err := escrever(cmd, op, func(snd api.Sender) error { return api.ExcluirExtrato(ctx, snd, alvo.conta, alvo.ano(), alvo.mesNum()) })
			if err != nil || !enviado {
				return err
			}
			var situacaoExtrato any
			if sit, err := situacaoExtratoAlvo(ctx, g, alvo); err != nil {
				fmt.Fprintln(s.err, "aviso: extrato excluído, mas não foi possível reler os extratos:", err)
			} else {
				situacaoExtrato = nilIfEmpty(sit)
			}
			fmt.Fprintf(s.err, "Não se esqueça de importar o extrato de %s de novo (ctbz extratos importar).\n", alvo.mes.Format("01/2006"))
			rec := resultadoEscrita("excluir-extrato", "excluido", alvo.conta).
				Add("competencia", "Competência", alvo.mes.Format("01/2006")).
				Add("situacao_extrato", "Situação do extrato", situacaoExtrato)
			return output.Write(s.out, f, rec)
		},
	}
	addContaBancariaFlag(cmd)
	addCompetenciaFlag(cmd)
	addWriteFlags(cmd)
	return cmd
}
