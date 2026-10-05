package api

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

// Importação de extrato pela tela "Importar extrato" (quatro passos). Ver
// docs/api/escrita/contabilidade-e-documentos.md, seções 1.7, 1.8 e 2.1.

// TamanhoMaximoExtrato é o limite de arquivo do painel (30 MB).
const TamanhoMaximoExtrato = 30 << 20

// PathPermiteImportacao diz se o período pode ser importado (passo 1, GET).
func PathPermiteImportacao(idConta int64, ano, mes int) string {
	return fmt.Sprintf("movimentacao-financeira/permiteImportacao/%d/%d/%d", idConta, ano, mes)
}

// PeriodoImportacao é um período da resposta de PathPermiteImportacao.
type PeriodoImportacao struct {
	Ano                   int    `json:"ano"`
	Mes                   int    `json:"mes"`
	Mensagem              string `json:"mensagem" contract:"optional"`
	TemLancamentoPendente bool   `json:"temLancamentoPendente"`
	TeveLancamento        bool   `json:"teveLancamento" contract:"optional"`
}

// PermiteImportacao é a resposta de PathPermiteImportacao (ADR-0021).
type PermiteImportacao struct {
	PeriodosDeImportacao []PeriodoImportacao `json:"periodosDeImportacao"`
}

// BuscarPermiteImportacao lê os períodos de importação da conta.
func BuscarPermiteImportacao(ctx context.Context, g Getter, idConta int64, ano, mes int) (*PermiteImportacao, error) {
	return get[PermiteImportacao](ctx, g, PathPermiteImportacao(idConta, ano, mes))
}

// PathUploadExtratoInit são os bancos, os formatos aceitos e as contas para o upload.
const PathUploadExtratoInit = "upload-documentos/extrato/v2/init"

// ContaUpload é uma conta em PathUploadExtratoInit.
type ContaUpload struct {
	ID               int64  `json:"id"`
	StatusIntegracao string `json:"statusIntegracao" contract:"optional"`
}

// BancoUpload é um banco em PathUploadExtratoInit, com os formatos que ele aceita.
type BancoUpload struct {
	ID                        int64         `json:"id"`
	Codigo                    string        `json:"codigo" contract:"optional"`
	FormatosDisponiveisUpload []string      `json:"formatosDisponiveisUpload"`
	ContasBancarias           []ContaUpload `json:"contasBancarias"`
}

// UploadExtratoInit é a resposta de PathUploadExtratoInit (ADR-0021).
type UploadExtratoInit struct {
	Bancos []BancoUpload `json:"bancos"`
}

// BuscarUploadExtratoInit lê bancos, formatos e contas do upload de extrato.
func BuscarUploadExtratoInit(ctx context.Context, g Getter) (*UploadExtratoInit, error) {
	return get[UploadExtratoInit](ctx, g, PathUploadExtratoInit)
}

// NomeArquivoExtrato monta o nome que o front dá ao arquivo no storage:
// "<cnpj>_<ano>_<mes>_<numeroConta>_<lastModified>.<ext>" (lastModified em ms).
func NomeArquivoExtrato(cnpj string, ano, mes int, numeroConta string, lastModified int64, ext string) string {
	return fmt.Sprintf("%s_%d_%d_%s_%d.%s", cnpj, ano, mes, numeroConta, lastModified, strings.TrimPrefix(strings.ToLower(ext), "."))
}

// PathEnviarExtrato recebe o arquivo (passo 2, multipart).
const PathEnviarExtrato = "upload-documentos/extrato/enviar/bucket"

// ArquivoExtrato é o arquivo e a identificação do upload.
type ArquivoExtrato struct {
	NomeArquivo string
	Ano, Mes    int
	IDConta     int64
	Arquivo     string // nome original
	Conteudo    []byte
}

// EnviarExtrato faz o upload do arquivo do extrato.
func EnviarExtrato(ctx context.Context, s Sender, a ArquivoExtrato) error {
	return s.SendMultipart(ctx, "POST", PathEnviarExtrato, []Campo{
		{Nome: "nomeArquivo", Valor: a.NomeArquivo},
		{Nome: "ano", Valor: strconv.Itoa(a.Ano)},
		{Nome: "mes", Valor: strconv.Itoa(a.Mes)},
		{Nome: "documento", Arquivo: a.Arquivo, Conteudo: a.Conteudo, Tipo: TipoArquivo(a.Arquivo)},
		{Nome: "idConta", Valor: strconv.FormatInt(a.IDConta, 10)},
	}, nil)
}

// PathInfoExtrato é o que a Contabilizei leu do arquivo enviado (passo 3, GET).
func PathInfoExtrato(idConta int64, ano, mes int, nomeArquivo string) string {
	return fmt.Sprintf("movimentacao-financeira/info-extrato/%d/%d/%d/%s", idConta, ano, mes, url.PathEscape(nomeArquivo))
}

// InfoExtrato é a resposta de PathInfoExtrato. Os campos podem vir dentro de infoExtrato ou
// na raiz (incerto; ADR-0021).
type InfoExtrato struct {
	InfoExtrato *struct {
		SaldoUltimoDia     *float64 `json:"saldoUltimoDia"`
		ExtratoForaPeriodo bool     `json:"extratoForaPeriodo" contract:"optional"`
	} `json:"infoExtrato" contract:"optional"`
	ExtratoForaPeriodo bool `json:"extratoForaPeriodo" contract:"optional"`
}

// Saldo devolve o saldo do último dia lido do arquivo, se houver.
func (i *InfoExtrato) Saldo() *float64 {
	if i.InfoExtrato == nil {
		return nil
	}
	return i.InfoExtrato.SaldoUltimoDia
}

// ForaDoPeriodo diz se o arquivo é de outra competência.
func (i *InfoExtrato) ForaDoPeriodo() bool {
	return i.ExtratoForaPeriodo || (i.InfoExtrato != nil && i.InfoExtrato.ExtratoForaPeriodo)
}

// BuscarInfoExtrato lê o que a Contabilizei extraiu do arquivo enviado.
func BuscarInfoExtrato(ctx context.Context, g Getter, idConta int64, ano, mes int, nomeArquivo string) (*InfoExtrato, error) {
	return get[InfoExtrato](ctx, g, PathInfoExtrato(idConta, ano, mes, nomeArquivo))
}

// PathEventoUploadExtrato conclui a importação (passo 4).
const PathEventoUploadExtrato = "movimentacao-financeira/eventoUploadExtrato"

// RespostaSaldo confirma o saldo do último dia (só no OFX).
type RespostaSaldo struct {
	Tipo  string  `json:"tipo"` // SALDO_ULTIMO_MES
	Data  string  `json:"data"` // ISO, momento da confirmação
	Valor float64 `json:"valor"`
}

// EventoUploadExtrato é o corpo de PathEventoUploadExtrato.
type EventoUploadExtrato struct {
	Ano                int             `json:"ano"`
	Mes                int             `json:"mes"`
	CNPJ               string          `json:"cnpj"`
	IDContaBancaria    int64           `json:"idContaBancaria"`
	NomeArquivoStorage string          `json:"nomeArquivoStorage"`
	Respostas          []RespostaSaldo `json:"respostas"` // [] no PDF
}

// ConcluirImportacaoExtrato efetiva a importação do arquivo enviado.
func ConcluirImportacaoExtrato(ctx context.Context, s Sender, e EventoUploadExtrato) error {
	if e.Respostas == nil {
		e.Respostas = []RespostaSaldo{}
	}
	return s.Send(ctx, "POST", PathEventoUploadExtrato, e, nil)
}

// TipoArquivo fixa o Content-Type dos uploads (PDF ou binário), sem depender da tabela de
// tipos do sistema operacional.
func TipoArquivo(nome string) string {
	if strings.EqualFold(filepath.Ext(nome), ".pdf") {
		return "application/pdf"
	}
	return "application/octet-stream"
}

// PathExcluirExtrato exclui o extrato importado de uma conta num mês (DELETE; fora do
// catálogo). Ver docs/api/escrita/contabilidade-e-documentos.md, seção 1.6.
func PathExcluirExtrato(idConta int64, ano, mes int) string {
	return fmt.Sprintf("movimentacao-financeira/extrato?idContaBancaria=%d&ano=%d&mes=%d", idConta, ano, mes)
}

// ExcluirExtrato apaga o extrato: a importação volta a ficar pendente e as classificações se perdem.
func ExcluirExtrato(ctx context.Context, s Sender, idConta int64, ano, mes int) error {
	return s.Send(ctx, "DELETE", PathExcluirExtrato(idConta, ano, mes), nil, nil)
}
