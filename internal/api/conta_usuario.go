package api

import "context"

// Dados de login do usuário (e-mail, telefone e senha). Todas as escritas exigem um código
// OTP. Ver docs/api/escrita/pendencias-e-conta.md, seção 3.1.
const (
	PathContaUsuario      = "conta-usuario/init"
	PathEnviarOTPConta    = "conta-usuario/enviar-token-otp"
	PathAlterarDadosConta = "conta-usuario/alterar-dados"
	PathAlterarSenhaConta = "conta-usuario/alterar-senha"
)

// Métodos de envio do código (e de 2FA).
const (
	MetodoOTPEmail = "EMAIL"
	MetodoOTPSMS   = "SMS"
	MetodoOTPApp   = "APP"
)

// DadosLogin são os dados de login, mascarados com "*" pela API, e o método atual de 2FA.
// Campos citados no front; sem captura real (ADR-0021).
type DadosLogin struct {
	Email    string `json:"email" contract:"optional"`
	Telefone string `json:"telefone" contract:"optional"`
	Metodo   string `json:"metodo" contract:"optional"`
}

// BuscarDadosLogin lê os dados de login do usuário.
func BuscarDadosLogin(ctx context.Context, g Getter) (*DadosLogin, error) {
	return get[DadosLogin](ctx, g, PathContaUsuario)
}

// EnviarOTPConta pede um código por EMAIL ou SMS e devolve em quantos segundos dá para pedir
// outro. 422 com o tempo restante quando o pedido anterior é recente.
func EnviarOTPConta(ctx context.Context, s Sender, metodo string) (int, error) {
	var r struct {
		Tempo int `json:"tempo"`
	}
	err := s.Send(ctx, "POST", PathEnviarOTPConta, struct {
		MetodoEnvio string `json:"metodoEnvio"`
	}{metodo}, &r)
	return r.Tempo, err
}

// AlterarDadosConta troca o e-mail e o telefone (só dígitos) de login. 404 = código inválido.
func AlterarDadosConta(ctx context.Context, s Sender, email, telefone, codigo, metodo string) error {
	return s.Send(ctx, "POST", PathAlterarDadosConta, struct {
		Email    string `json:"email"`
		Telefone string `json:"telefone"`
		Codigo   string `json:"codigo"`
		Metodo   string `json:"metodo"`
	}{email, telefone, codigo, metodo}, nil)
}

// AlterarSenhaConta troca a senha; a prova é o código, não a senha atual. 404 = código
// inválido.
func AlterarSenhaConta(ctx context.Context, s Sender, senha, codigo, metodo string) error {
	return s.Send(ctx, "POST", PathAlterarSenhaConta, struct {
		Senha  string `json:"senha"`
		Codigo string `json:"codigo"`
		Metodo string `json:"metodo"`
	}{senha, codigo, metodo}, nil)
}
