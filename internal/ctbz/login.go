package ctbz

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Stage indica em que ponto do fluxo de login a sessão está.
type Stage string

const (
	StageOTP           Stage = "otp"            // aguardando o código enviado por e-mail
	StageSelectCompany Stage = "select_company" // aguardando a escolha da empresa (CNPJ)
	StageDone          Stage = "done"
)

// Company é uma empresa oferecida na tela "Selecione a empresa".
type Company struct {
	CNPJ string `json:"cnpj"`
	Name string `json:"name"`
}

// LoginState é o estado intermediário do login. É serializável para que o
// fluxo possa ser retomado em outra execução (ex.: `ctbz login --otp 123456`).
type LoginState struct {
	Stage       Stage                   `json:"stage"`
	Token       string                  `json:"token"`
	Cookies     map[string]*http.Cookie `json:"cookies"`
	MaskedEmail string                  `json:"masked_email,omitempty"`
	OTPSentAt   time.Time               `json:"otp_sent_at,omitempty"`
	Companies   []Company               `json:"companies,omitempty"`
	// Preenchido quando Stage == StageDone.
	Storage map[string]json.RawMessage `json:"-"`
}

// ErrInvalidOTP é retornado quando a Contabilizei rejeita o código.
var ErrInvalidOTP = errors.New("código OTP incorreto ou expirado")

// Mensagens que a página de login exibe a partir do fragmento da URL.
var loginErrors = map[string]string{
	"incorreto":      "usuário ou senha incorretos",
	"bloqueado":      "acesso bloqueado temporariamente por segurança; tente de novo em alguns minutos",
	"limite-sessoes": "limite de acessos simultâneos atingido; saia da plataforma em outro aparelho",
	"sem-permissao":  "a empresa ainda não está pronta para usar a plataforma",
	"expirou":        "sessão de login expirou; recomece com `ctbz login`",
	"google":         "conta Google não permitida",
	"problema":       "a Contabilizei encontrou um problema no acesso; contate o suporte",
	"error-session":  "sessão expirada",
}

var (
	reFormToken   = regexp.MustCompile(`name="token"\s+value="([^"]*)"`)
	reMaskedEmail = regexp.MustCompile(`seu e-mail:\s*<strong>([^<]*)</strong>`)
	reCompany     = regexp.MustCompile(`(?s)name="cnpj"\s+value="(\d+)".*?<span>([^<]*)</span>`)
	reStorage     = regexp.MustCompile(`localStorage\.setItem\("([^"]+)",\s*"([^"]*)"\)`)
)

// StartLogin abre a página de login e envia usuário/senha. Em geral a
// Contabilizei responde com a tela de OTP e envia o código por e-mail.
func (c *Client) StartLogin(ctx context.Context, user, password string) (*LoginState, error) {
	if user == "" || password == "" {
		return nil, errors.New("usuário e senha são obrigatórios (CTBZ_USER / CTBZ_PASSWORD)")
	}
	c.Cookies = map[string]*http.Cookie{}
	page, err := c.Do(ctx, http.MethodGet, "/login", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("abrindo página de login: %w", err)
	}
	if page.Status != http.StatusOK {
		return nil, &HTTPError{Step: "abrindo página de login", Status: page.Status, Body: page.Body}
	}
	m := reFormToken.FindSubmatch(page.Body)
	if m == nil {
		return nil, &HTTPError{Step: "página de login sem token", Status: page.Status, Body: page.Body}
	}
	st := &LoginState{Token: string(m[1])}

	resp, err := c.postForm(ctx, "/login", url.Values{
		"user":     {user},
		"password": {password},
		"token":    {st.Token},
	})
	if err != nil {
		return nil, fmt.Errorf("enviando credenciais: %w", err)
	}
	resp, err = c.followRedirects(ctx, resp)
	if err != nil {
		return nil, err
	}
	if resp.Status != http.StatusOK {
		return nil, &HTTPError{Step: "login", Status: resp.Status, Body: resp.Body}
	}
	if err := st.absorbPage(resp.Body); err != nil {
		return nil, err
	}
	if st.Stage == StageOTP {
		st.OTPSentAt = time.Now()
	}
	st.Cookies = c.Cookies
	return st, nil
}

// Resume restaura os cookies de um LoginState salvo.
func (c *Client) Resume(st *LoginState) {
	c.Cookies = map[string]*http.Cookie{}
	for k, v := range st.Cookies {
		c.Cookies[k] = v
	}
}

// VerifyOTP envia o código recebido por e-mail.
func (c *Client) VerifyOTP(ctx context.Context, st *LoginState, otp string) error {
	if st.Stage != StageOTP {
		return fmt.Errorf("o login não está aguardando OTP (etapa atual: %s)", st.Stage)
	}
	otp = strings.TrimSpace(otp)
	// O navegador envia exatamente este corpo via fetch() (text/plain).
	resp, err := c.Do(ctx, http.MethodPost, "/login", strings.NewReader("v=2&otp="+url.QueryEscape(otp)),
		map[string]string{"Content-Type": "text/plain;charset=UTF-8"})
	if err != nil {
		return fmt.Errorf("enviando OTP: %w", err)
	}
	st.Cookies = c.Cookies
	switch resp.Status {
	case http.StatusNotFound:
		return ErrInvalidOTP
	case http.StatusPartialContent, http.StatusOK:
		// 206 = novo fragmento de tela (ex.: seleção de empresa);
		// 200 = script final que grava o localStorage e redireciona.
		return st.absorbPage(resp.Body)
	case http.StatusFound, http.StatusSeeOther, http.StatusMovedPermanently:
		resp, err := c.followRedirects(ctx, resp)
		if err != nil {
			return err
		}
		st.Cookies = c.Cookies
		if resp.Status != http.StatusOK {
			return &HTTPError{Step: "verificação do OTP", Status: resp.Status, Body: resp.Body}
		}
		return st.absorbPage(resp.Body)
	default:
		return &HTTPError{Step: "verificação do OTP", Status: resp.Status, Body: resp.Body}
	}
}

// SelectCompany escolhe a empresa (CNPJ) e conclui o login.
func (c *Client) SelectCompany(ctx context.Context, st *LoginState, cnpj string) error {
	if st.Stage != StageSelectCompany {
		return fmt.Errorf("o login não está aguardando seleção de empresa (etapa atual: %s)", st.Stage)
	}
	cnpj = OnlyDigits(cnpj)
	if !st.HasCompany(cnpj) {
		return fmt.Errorf("CNPJ %s não está entre as empresas disponíveis: %s", cnpj, st.companyList())
	}
	resp, err := c.postForm(ctx, "/selecionarempresa", url.Values{"cnpj": {cnpj}, "token": {st.Token}})
	if err != nil {
		return fmt.Errorf("selecionando empresa: %w", err)
	}
	resp, err = c.followRedirects(ctx, resp)
	if err != nil {
		return err
	}
	st.Cookies = c.Cookies
	if resp.Status != http.StatusOK {
		return &HTTPError{Step: "seleção de empresa", Status: resp.Status, Body: resp.Body}
	}
	if err := st.absorbPage(resp.Body); err != nil {
		return err
	}
	if st.Stage != StageDone {
		return fmt.Errorf("seleção de empresa não concluiu o login (etapa: %s)", st.Stage)
	}
	return nil
}

func (st *LoginState) HasCompany(cnpj string) bool {
	for _, co := range st.Companies {
		if co.CNPJ == cnpj {
			return true
		}
	}
	return false
}

func (st *LoginState) companyList() string {
	var parts []string
	for _, co := range st.Companies {
		parts = append(parts, co.CNPJ+" ("+co.Name+")")
	}
	return strings.Join(parts, ", ")
}

// followRedirects segue redirecionamentos dentro do host, convertendo
// fragmentos de erro (/login#incorreto) em erros legíveis.
func (c *Client) followRedirects(ctx context.Context, resp *Response) (*Response, error) {
	for i := 0; i < 5 && resp.Status >= 300 && resp.Status < 400 && resp.Location != ""; i++ {
		loc, err := url.Parse(resp.Location)
		if err != nil {
			return nil, fmt.Errorf("redirecionamento inválido %q: %w", resp.Location, err)
		}
		if msg, ok := loginErrors[loc.Fragment]; ok {
			return nil, errors.New("login recusado: " + msg)
		}
		if loc.Path == "/login" && loc.Fragment == "" && loc.RawQuery == "" {
			return nil, errors.New("login recusado: a Contabilizei voltou para a tela de login")
		}
		target := loc.String()
		if !loc.IsAbs() {
			target = loc.RequestURI()
		}
		resp, err = c.Do(ctx, http.MethodGet, target, nil, nil)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// absorbPage identifica qual tela do fluxo foi retornada e atualiza o estado.
func (st *LoginState) absorbPage(body []byte) error {
	s := string(body)
	switch {
	case reStorage.MatchString(s):
		storage, err := ParseStorage(s)
		if err != nil {
			return err
		}
		st.Stage, st.Storage = StageDone, storage
	case reCompany.MatchString(s):
		st.Stage = StageSelectCompany
		st.Companies = ParseCompanies(s)
		if m := reFormToken.FindStringSubmatch(s); m != nil {
			st.Token = m[1]
		}
	case strings.Contains(s, `id="otp"`):
		st.Stage = StageOTP
		if m := reMaskedEmail.FindStringSubmatch(s); m != nil {
			st.MaskedEmail = html.UnescapeString(m[1])
		}
	case strings.Contains(s, `id="form-login"`):
		return errors.New("login recusado: a Contabilizei voltou para a tela de login (usuário/senha incorretos?)")
	default:
		return &HTTPError{Step: "tela de login não reconhecida", Status: 200, Body: body}
	}
	return nil
}

// ParseCompanies extrai as opções da tela "Selecione a empresa".
func ParseCompanies(page string) []Company {
	var out []Company
	seen := map[string]bool{}
	for _, m := range reCompany.FindAllStringSubmatch(page, -1) {
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		out = append(out, Company{CNPJ: m[1], Name: strings.TrimSpace(html.UnescapeString(m[2]))})
	}
	return out
}

// ParseStorage decodifica as chaves que a página final grava no localStorage
// (base64 de JSON passado por encodeURIComponent). Chaves conhecidas:
// "l" (dados do login), "r" (responsável/usuário) e "e" (empresa selecionada).
func ParseStorage(page string) (map[string]json.RawMessage, error) {
	out := map[string]json.RawMessage{}
	for _, m := range reStorage.FindAllStringSubmatch(page, -1) {
		raw, err := base64.StdEncoding.DecodeString(m[2])
		if err != nil {
			if raw, err = base64.RawStdEncoding.DecodeString(m[2]); err != nil {
				return nil, fmt.Errorf("localStorage %q: base64 inválido: %w", m[1], err)
			}
		}
		decoded, err := url.PathUnescape(string(raw))
		if err != nil {
			return nil, fmt.Errorf("localStorage %q: %w", m[1], err)
		}
		if !json.Valid([]byte(decoded)) {
			return nil, fmt.Errorf("localStorage %q: JSON inválido", m[1])
		}
		out[m[1]] = json.RawMessage(decoded)
	}
	if len(out) == 0 {
		return nil, errors.New("página final do login sem dados de sessão")
	}
	return out, nil
}

// OnlyDigits remove pontuação de CPF/CNPJ.
func OnlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
