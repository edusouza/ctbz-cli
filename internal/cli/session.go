package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edusouza/ctbz-cli/internal/ctbz"
	"github.com/edusouza/ctbz-cli/internal/output"
)

// authedAPI faz uma chamada com a sessão salva. Se a sessão tiver expirado e
// CTBZ_OTP_CMD estiver configurado, refaz o login sozinho e repete a chamada.
func authedAPI(ctx context.Context, s streams, method, path string, body []byte) (*ctbz.Response, error) {
	return chamarComSessao(ctx, s, method, path, body, true)
}

// chamarComSessao é authedAPI com o re-login opcional.
func chamarComSessao(ctx context.Context, s streams, method, path string, body []byte, relogar bool) (*ctbz.Response, error) {
	store, err := ctbz.DefaultStore()
	if err != nil {
		return nil, err
	}
	sess, err := store.LoadSession()
	if err != nil {
		return nil, err
	}
	c := sess.Client()
	resp, err := c.API(ctx, method, path, bodyReader(body))
	if relogar && errors.Is(err, ctbz.ErrUnauthorized) && os.Getenv("CTBZ_OTP_CMD") != "" {
		sess, err = relogin(ctx, store, s, sess)
		if err != nil {
			return nil, err
		}
		c = sess.Client()
		resp, err = c.API(ctx, method, path, bodyReader(body))
	}
	if err != nil {
		return nil, err
	}
	saveCookies(store, sess, c, s.err)
	return resp, nil
}

// reloginMu garante um único re-login quando várias chamadas em paralelo recebem 401.
var reloginMu sync.Mutex

// relogin refaz o login de uma sessão expirada. Quem chega depois de outra goroutine já ter
// refeito o login usa a sessão nova gravada por ela, sem pedir outro OTP.
func relogin(ctx context.Context, store *ctbz.Store, s streams, expirada *ctbz.Session) (*ctbz.Session, error) {
	reloginMu.Lock()
	defer reloginMu.Unlock()
	if atual, err := store.LoadSession(); err == nil && !atual.CreatedAt.Equal(expirada.CreatedAt) {
		return atual, nil
	}
	fmt.Fprintln(s.err, "Sessão expirada; refazendo login com CTBZ_OTP_CMD…")
	return login(ctx, store, s, loginOpts{
		otpCmd:     os.Getenv("CTBZ_OTP_CMD"),
		otpTimeout: envDuration("CTBZ_OTP_TIMEOUT", defaultOTPTimeout),
		cnpj:       expirada.CNPJ,
		restart:    true,
	})
}

// getJSON chama um endpoint com GET e decodifica a resposta em v.
func getJSON(ctx context.Context, s streams, path string, v any) error {
	return getJSONCom(ctx, s, path, v, true)
}

func getJSONCom(ctx context.Context, s streams, path string, v any, relogar bool) error {
	resp, err := chamarComSessao(ctx, s, "GET", path, nil, relogar)
	if err != nil {
		return err
	}
	if !sucesso(resp.Status) {
		return &ctbz.HTTPError{Step: path, Status: resp.Status, Body: resp.Body}
	}
	if err := json.Unmarshal(resp.Body, v); err != nil {
		return fmt.Errorf("resposta inesperada de %s: %w", path, err)
	}
	return nil
}

// sessionGetter implementa api.Getter com a sessão salva (e re-login automático).
type sessionGetter struct{ s streams }

func (g sessionGetter) GetJSON(ctx context.Context, path string, v any) error {
	return getJSON(ctx, g.s, path, v)
}

// releituraGetter implementa api.Getter sem re-login: depois de uma escrita que pode ter
// encerrado a sessão (troca do e-mail de login), um 401 na releitura não dispara outro OTP de
// login (#833).
type releituraGetter struct{ s streams }

func (g releituraGetter) GetJSON(ctx context.Context, path string, v any) error {
	return getJSONCom(ctx, g.s, path, v, false)
}

// GetText implementa api.TextGetter: devolve o corpo de um GET como texto.
func (g sessionGetter) GetText(ctx context.Context, path string) (string, error) {
	resp, err := authedAPI(ctx, g.s, "GET", path, nil)
	if err != nil {
		return "", err
	}
	if !sucesso(resp.Status) {
		return "", &ctbz.HTTPError{Step: path, Status: resp.Status, Body: resp.Body}
	}
	return string(resp.Body), nil
}

// sucesso diz se um status de leitura traz dados. Como o painel, aceita 2xx (a Central de
// Rotinas às vezes responde 206 com a resposta completa), menos o 204: sem corpo, ele quer
// dizer "não encontrado" ou "sem dados" (ex.: o cadastro de um tomador inexistente).
func sucesso(status int) bool {
	return status >= 200 && status < 300 && status != http.StatusNoContent
}

// saveCookies persiste cookies renovados pelo servidor durante as chamadas.
func saveCookies(store *ctbz.Store, sess *ctbz.Session, c *ctbz.Client, errOut io.Writer) {
	sess.Cookies = c.Cookies
	if err := store.SaveSession(sess); err != nil {
		fmt.Fprintln(errOut, "aviso: não foi possível atualizar a sessão:", err)
	}
}

func bodyReader(b []byte) io.Reader {
	if b == nil {
		return nil
	}
	return strings.NewReader(string(b))
}

type sessInfo struct {
	Email       string
	RazaoSocial string
	CNPJ        string
}

// sessionInfo lê usuário e empresa dos dados do localStorage salvos no login.
func sessionInfo(sess *ctbz.Session) sessInfo {
	var login struct {
		Email   string `json:"email"`
		Empresa struct {
			CNPJ        string `json:"cnpj"`
			RazaoSocial string `json:"razaoSocial"`
		} `json:"empresa"`
	}
	if raw, ok := sess.Storage["l"]; ok {
		_ = json.Unmarshal(raw, &login)
	}
	info := sessInfo{Email: login.Email, RazaoSocial: login.Empresa.RazaoSocial, CNPJ: login.Empresa.CNPJ}
	if info.CNPJ == "" {
		info.CNPJ = sess.CNPJ
	}
	return info
}

// describeSession resume a sessão numa linha (mensagens em stderr).
func describeSession(sess *ctbz.Session) string {
	info := sessionInfo(sess)
	var parts []string
	if info.RazaoSocial != "" {
		parts = append(parts, fmt.Sprintf("%s (CNPJ %s)", info.RazaoSocial, output.FormatCNPJ(output.NewCNPJ(info.CNPJ))))
	} else if info.CNPJ != "" {
		parts = append(parts, "CNPJ "+output.FormatCNPJ(output.NewCNPJ(info.CNPJ)))
	}
	if info.Email != "" {
		parts = append(parts, "usuário "+info.Email)
	}
	if len(parts) == 0 {
		return "sessão ativa"
	}
	return strings.Join(parts, " — ")
}

// currentCNPJ extrai o CNPJ da empresa selecionada dos dados do login.
func currentCNPJ(sess *ctbz.Session) string {
	for _, key := range []string{"e", "l"} {
		raw, ok := sess.Storage[key]
		if !ok {
			continue
		}
		var v struct {
			CNPJ    string `json:"cnpj"`
			Empresa struct {
				CNPJ string `json:"cnpj"`
			} `json:"empresa"`
		}
		if json.Unmarshal(raw, &v) == nil {
			if v.CNPJ != "" {
				return ctbz.OnlyDigits(v.CNPJ)
			}
			if v.Empresa.CNPJ != "" {
				return ctbz.OnlyDigits(v.Empresa.CNPJ)
			}
		}
	}
	return ""
}

func baseURL() string {
	if v := os.Getenv("CTBZ_BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return ctbz.DefaultBaseURL
}

// verboseWriter é onde vai o log do --otp-cmd: o stderr com CTBZ_VERBOSE ou no modo verbose.
func verboseWriter(ctx context.Context, s streams) io.Writer {
	if os.Getenv("CTBZ_VERBOSE") != "" || ctbz.TracerFrom(ctx) != nil {
		return s.err
	}
	return nil
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
