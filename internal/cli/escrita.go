package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edusouza/ctbz-cli/internal/api"
	"github.com/edusouza/ctbz-cli/internal/ctbz"
)

// sessionSender implementa api.Sender com a sessão salva: confere a sessão antes e envia
// cada escrita uma única vez, sem retentativa nem re-login depois do envio (ADR-0018).
// Cada envio vira uma linha no registro de ações, com a origem informada.
type sessionSender struct {
	s      streams
	origem origemEscrita
	// semRelogin vale para o envio que vem depois de outro já feito no mesmo fluxo (o código
	// OTP da conta): a conferência da sessão não refaz o login, que mandaria outro código e
	// invalidaria o que já foi pedido.
	semRelogin bool
}

// origemEscrita identifica, no registro de ações, o comando e o objeto de uma escrita.
type origemEscrita struct {
	comando string
	id      string
	// arquivos descreve os arquivos de um envio multipart (nome e tamanho, sem o conteúdo).
	arquivos string
}

func (w sessionSender) Send(ctx context.Context, method, path string, body any, v any) error {
	data, err := api.EncodeBody(body)
	if err != nil {
		return err
	}
	contentType := ""
	if data != nil {
		contentType = "application/json"
	}
	return w.send(ctx, method, path, data, contentType, 0, v)
}

// timeoutUpload é o prazo dos envios de arquivo, maior que o geral do cliente (60 s): um
// extrato pode ter até 30 MB.
var timeoutUpload = 3 * time.Minute

func (w sessionSender) SendMultipart(ctx context.Context, method, path string, campos []api.Campo, v any) error {
	data, contentType, err := api.EncodeMultipart(campos)
	if err != nil {
		return err
	}
	w.origem.arquivos = descreverArquivos(campos)
	return w.send(ctx, method, path, data, contentType, timeoutUpload, v)
}

func (w sessionSender) send(ctx context.Context, method, path string, body []byte, contentType string, timeout time.Duration, v any) error {
	resp, err := sendOnceCom(ctx, w.s, w.origem, method, path, body, contentType, timeout, !w.semRelogin)
	if err != nil {
		return err
	}
	if resp.Status < 200 || resp.Status > 299 {
		return ctbz.NewWriteError(method, path, resp)
	}
	return api.DecodeResponse(resp.Body, v)
}

// erroSemResposta é a falha de conexão durante o envio de uma escrita: ela pode ou não ter sido
// aplicada. Quem chama pode reler o estado para orientar melhor (ex.: extratos importar).
type erroSemResposta struct{ causa error }

func (e *erroSemResposta) Error() string {
	return "a conexão falhou durante o envio e a escrita pode ou não ter sido aplicada; confira antes de repetir: " + e.causa.Error()
}

func (e *erroSemResposta) Unwrap() error { return e.causa }

// errExpiredOnSend indica que o servidor recusou a escrita por sessão expirada: nada foi aplicado.
var errExpiredOnSend = errors.New("a sessão expirou antes do envio e a escrita não foi aplicada; rode `ctbz login` e repita")

// errExpiredAfterCode indica que a sessão expirou depois de a conta já ter pedido o código
// OTP: não houve novo login (ele mandaria outro código), e nada foi alterado.
var errExpiredAfterCode = errors.New("o código foi pedido, mas a sessão expirou antes do envio e nada foi alterado; rode `ctbz login` e repita")

// sendOnce confere a sessão com um GET (que pode refazer o login, porque nada foi enviado
// ainda) e então envia a escrita uma única vez. Respostas HTTP de erro voltam sem erro, para
// quem chama decidir; erros de rede (*erroSemResposta) avisam que o resultado da escrita é
// incerto. timeout > 0 troca o prazo do cliente. Todo envio, com ou sem sucesso, é registrado
// em acoes.jsonl.
func sendOnce(ctx context.Context, s streams, origem origemEscrita, method, path string, body []byte, contentType string, timeout time.Duration) (*ctbz.Response, error) {
	return sendOnceCom(ctx, s, origem, method, path, body, contentType, timeout, true)
}

// sendOnceCom é sendOnce com o re-login da conferência opcional. Sem relogar (segundo envio
// de um fluxo que já pediu um código), a sessão expirada vira errExpiredAfterCode.
func sendOnceCom(ctx context.Context, s streams, origem origemEscrita, method, path string, body []byte, contentType string, timeout time.Duration, relogar bool) (*ctbz.Response, error) {
	if _, err := chamarComSessao(ctx, s, "GET", api.PathAppBar, nil, relogar); err != nil {
		if !relogar && errors.Is(err, ctbz.ErrUnauthorized) {
			return nil, errExpiredAfterCode
		}
		return nil, fmt.Errorf("conferindo a sessão antes de enviar: %w", err)
	}
	store, err := ctbz.DefaultStore()
	if err != nil {
		return nil, err
	}
	sess, err := store.LoadSession()
	if err != nil {
		return nil, err
	}
	c := sess.Client()
	if timeout > 0 {
		c.HTTP.Timeout = timeout
	}
	resp, err := c.Send(ctx, method, path, bytes.NewReader(body), contentType)
	registrarAcao(store, sess, s, origem, method, path, resp)
	switch {
	case errors.Is(err, ctbz.ErrUnauthorized):
		return nil, errExpiredOnSend
	case err != nil:
		return nil, &erroSemResposta{err}
	}
	saveCookies(store, sess, c, s.err)
	return resp, nil
}

// registrarAcao grava a escrita em acoes.jsonl. Uma falha ao gravar só gera aviso: não
// desfaz nem esconde o resultado da escrita.
func registrarAcao(store *ctbz.Store, sess *ctbz.Session, s streams, origem origemEscrita, method, path string, resp *ctbz.Response) {
	a := ctbz.Acao{
		Data:      now(),
		CNPJ:      currentCNPJ(sess),
		Comando:   origem.comando,
		Metodo:    method,
		Caminho:   semQuery(ctbz.ResolvePath(path)),
		Resultado: ctbz.ResultadoSemResposta,
		ID:        origem.id,
		Arquivos:  origem.arquivos,
	}
	if a.CNPJ == "" {
		a.CNPJ = sess.CNPJ
	}
	if resp != nil {
		a.Status = resp.Status
		switch {
		case resp.Status >= 200 && resp.Status <= 299:
			a.Resultado = ctbz.ResultadoEnviada
		case ctbz.StatusIncerto(resp.Status):
			a.Resultado = ctbz.ResultadoIncerta
		default:
			a.Resultado = ctbz.ResultadoRecusada
		}
	}
	if err := store.AppendAcao(a); err != nil {
		fmt.Fprintln(s.err, "aviso: não foi possível registrar a ação em acoes.jsonl:", err)
	}
}

// descreverArquivos lista os arquivos de um envio multipart para o registro de ações:
// "documento: extrato.csv (42 KB)", separados por "; ".
func descreverArquivos(campos []api.Campo) string {
	var partes []string
	for _, c := range campos {
		if c.Arquivo != "" {
			partes = append(partes, fmt.Sprintf("%s: %s (%s)", c.Nome, c.Arquivo, tamanhoArquivo(len(c.Conteudo))))
		}
	}
	return strings.Join(partes, "; ")
}

// semQuery tira a query do caminho registrado: ela pode carregar dados pessoais.
func semQuery(path string) string {
	p, _, _ := strings.Cut(path, "?")
	return p
}
