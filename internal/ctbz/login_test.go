package ctbz

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// Fragmentos fiéis às telas reais da Contabilizei, com dados fictícios.
const (
	loginPage = `<form class="form" id="form-login" action="/login" method="post">
<input class="input-user" type="text" name="user" id="user" required autofocus>
<input type="hidden" name="token" value="TOKEN123"/></form>`

	otpPage = `<span class="subtitle">
  Enviamos um código de 6 dígitos para o seu e-mail: <strong>fu***@ex***com</strong>
</span><div class="form-div"><input class="input-token" type="text" id="otp" name="otp" maxlength="6"></div>`

	companyPage = `<div class="title">Selecione a empresa</div><form action="/selecionarempresa" method="post">
<label class="mdl-radio" for="11111111000111-1">
  <input type="radio" id="11111111000111-1" class="mdl-radio__button" name="cnpj" value="11111111000111">
  <div><b><span class="mdl-radio__label">11111111000111</span> </b></div>
  <div><span>FULANO MEI</span></div>
</label>
<label class="mdl-radio" for="22222222000122-1">
  <input type="radio" id="22222222000122-1" class="mdl-radio__button" name="cnpj" value="22222222000122">
  <div><b><span class="mdl-radio__label">22222222000122</span> </b></div>
  <div><span>FULANO TECNOLOGIA LTDA</span></div>
</label>
<input type="hidden" name="token" value="TOKEN123" /></form>`
)

func storageValue(v any) string {
	j, _ := json.Marshal(v)
	return base64.StdEncoding.EncodeToString([]byte(url.PathEscape(string(j))))
}

func finalPage(cnpj string) string {
	l := storageValue(map[string]any{"token": "abc", "email": "fulano@example.com",
		"empresa": map[string]any{"cnpj": cnpj, "razaoSocial": "FULANO TECNOLOGIA LTDA"}})
	e := storageValue(map[string]any{"cnpj": cnpj, "razaoSocial": "FULANO TECNOLOGIA LTDA"})
	return `<script>localStorage.clear();localStorage.setItem("l","` + l + `");localStorage.setItem("e","` + e +
		`");location.replace("/painel-de-controle/#/home");</script>`
}

// fakeServer simula o fluxo de login da Contabilizei.
func fakeServer(t *testing.T, validOTP string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.SetCookie(w, &http.Cookie{Name: "__C", Value: "pre-session", Path: "/"})
			io.WriteString(w, loginPage)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if strings.HasPrefix(string(body), "v=") {
			if ck, err := r.Cookie("__C"); err != nil || ck.Value != "pre-session" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			q, _ := url.ParseQuery(string(body))
			if q.Get("otp") != validOTP {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusPartialContent)
			io.WriteString(w, companyPage)
			return
		}
		form, _ := url.ParseQuery(string(body))
		if form.Get("token") != "TOKEN123" || form.Get("password") != "s3nha" {
			http.Redirect(w, r, "/login#incorreto", http.StatusFound)
			return
		}
		io.WriteString(w, otpPage)
	})
	mux.HandleFunc("/selecionarempresa", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		http.SetCookie(w, &http.Cookie{Name: "oauth-token", Value: "authed", Path: "/"})
		io.WriteString(w, finalPage(r.PostForm.Get("cnpj")))
	})
	mux.HandleFunc("/api/plataforma/dadosempresa/get", func(w http.ResponseWriter, r *http.Request) {
		if ck, err := r.Cookie("oauth-token"); err != nil || ck.Value != "authed" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		io.WriteString(w, `{"empresaAtual":{"cnpj":"22.222.222/0001-22"}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestLoginFlow(t *testing.T) {
	srv := fakeServer(t, "123456")
	ctx := context.Background()
	c := NewClient(srv.URL)

	st, err := c.StartLogin(ctx, "00000000000", "s3nha")
	if err != nil {
		t.Fatal(err)
	}
	if st.Stage != StageOTP || st.MaskedEmail != "fu***@ex***com" {
		t.Fatalf("estado inesperado: %+v", st)
	}

	// Simula a retomada em outra execução: serializa e restaura o estado.
	raw, _ := json.Marshal(st)
	var resumed LoginState
	if err := json.Unmarshal(raw, &resumed); err != nil {
		t.Fatal(err)
	}
	c2 := NewClient(srv.URL)
	c2.Resume(&resumed)

	if err := c2.VerifyOTP(ctx, &resumed, "000000"); !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("esperava ErrInvalidOTP, veio %v", err)
	}
	if err := c2.VerifyOTP(ctx, &resumed, "123456"); err != nil {
		t.Fatal(err)
	}
	if resumed.Stage != StageSelectCompany || len(resumed.Companies) != 2 ||
		resumed.Companies[1] != (Company{CNPJ: "22222222000122", Name: "FULANO TECNOLOGIA LTDA"}) {
		t.Fatalf("seleção de empresa inesperada: %+v", resumed)
	}
	if err := c2.SelectCompany(ctx, &resumed, "99.999.999/0001-99"); err == nil {
		t.Fatal("esperava erro para CNPJ fora da lista")
	}
	if err := c2.SelectCompany(ctx, &resumed, "22.222.222/0001-22"); err != nil {
		t.Fatal(err)
	}
	if resumed.Stage != StageDone {
		t.Fatalf("login não concluído: %s", resumed.Stage)
	}
	var e struct{ CNPJ string }
	json.Unmarshal(resumed.Storage["e"], &e)
	if e.CNPJ != "22222222000122" {
		t.Fatalf("localStorage decodificado incorretamente: %s", resumed.Storage["e"])
	}

	resp, err := c2.API(ctx, http.MethodGet, "dadosempresa/get", nil)
	if err != nil || resp.Status != 200 {
		t.Fatalf("API: %v %v", resp, err)
	}
	if _, err := NewClient(srv.URL).API(ctx, http.MethodGet, "dadosempresa/get", nil); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("esperava ErrUnauthorized sem cookies, veio %v", err)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	srv := fakeServer(t, "123456")
	_, err := NewClient(srv.URL).StartLogin(context.Background(), "00000000000", "errada")
	if err == nil || !strings.Contains(err.Error(), "incorretos") {
		t.Fatalf("esperava erro de senha incorreta, veio %v", err)
	}
}

func TestParseStorageSpecialChars(t *testing.T) {
	page := `localStorage.setItem("e","` + storageValue(map[string]any{"razaoSocial": "AÇÃO & CIA + 100%"}) + `")`
	st, err := ParseStorage(page)
	if err != nil {
		t.Fatal(err)
	}
	var v struct{ RazaoSocial string }
	json.Unmarshal(st["e"], &v)
	if v.RazaoSocial != "AÇÃO & CIA + 100%" {
		t.Fatalf("decodificado: %q", v.RazaoSocial)
	}
}

// O WAF da Contabilizei (Azion) devolve 403 para um User-Agent de Chrome sem a versão completa
// ("Chrome/128.0"); um navegador real sempre manda quatro partes ("Chrome/128.0.0.0").
func TestUserAgentPadraoTemVersaoCompletaDeChrome(t *testing.T) {
	if !regexp.MustCompile(`Chrome/\d+\.\d+\.\d+\.\d+ `).MatchString(defaultUserAgent) {
		t.Fatalf("User-Agent %q sem versão de Chrome com quatro partes: o WAF recusa o login com 403", defaultUserAgent)
	}
}

func TestStartLoginBloqueadoPeloWAF(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, `<!DOCTYPE html><html><head><title>Azion - Default error page</title></head><body></body></html>`)
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL).StartLogin(context.Background(), "fulano", "segredo")
	if err == nil {
		t.Fatal("esperava erro com o 403 do WAF")
	}
	for _, quer := range []string{"403", "bloqueou", "User-Agent"} {
		if !strings.Contains(err.Error(), quer) {
			t.Errorf("erro %q deveria mencionar %q", err, quer)
		}
	}
	if strings.Contains(err.Error(), "sem token") {
		t.Errorf("erro %q ainda trata o bloqueio como falta de token", err)
	}
}
