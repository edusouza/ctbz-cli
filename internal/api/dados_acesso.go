package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Dados de acesso a órgãos públicos (código do Simples, prefeitura, Dataprev). Ver
// docs/api/escrita/pendencias-e-conta.md, seção 3.3. O GET devolve as senhas em texto
// claro: nunca registre nem imprima o formulário inteiro.
const (
	PathDadosAcesso                   = "empresa/dadosacesso/init"
	PathAtualizarDadosAcesso          = "empresa/dadosacesso/atualizarDadosAcesso"
	PathConfirmarCredencialPrefeitura = "empresa/dadosacesso/confirmar-credencial-prefeitura-valida"
)

// Campos do formulário de dados de acesso.
const (
	CampoChaveAcessoSimples = "chaveAcessoSimples"
	CampoUsuarioPrefeitura  = "usuarioPrefeitura"
	CampoSenhaPrefeitura    = "senhaPrefeitura"
	CampoUsuarioDataprev    = "usuarioDataprev"
	CampoSenhaDataprev      = "senhaDataprev"
)

// FormularioDadosAcesso são as credenciais que a Contabilizei usa nos órgãos públicos.
// Campos citados no front; sem captura real (403 para administrador, ADR-0021).
type FormularioDadosAcesso struct {
	ID                                  any     `json:"id" contract:"optional"`
	ChaveAcessoSimples                  *string `json:"chaveAcessoSimples" contract:"optional"`
	UsuarioPrefeitura                   *string `json:"usuarioPrefeitura" contract:"optional"`
	SenhaPrefeitura                     *string `json:"senhaPrefeitura" contract:"optional"`
	UsuarioDataprev                     *string `json:"usuarioDataprev" contract:"optional"`
	SenhaDataprev                       *string `json:"senhaDataprev" contract:"optional"`
	DataValidadeSenhaPrefeituraCuritiba any     `json:"dataValidadeSenhaPrefeituraCuritiba" contract:"optional"`
}

// DadosAcesso é a resposta de PathDadosAcesso.
type DadosAcesso struct {
	Formulario FormularioDadosAcesso `json:"formulario"`
	// AlertaPendenciaCredencial pede para redefinir a senha da prefeitura.
	AlertaPendenciaCredencial bool `json:"alertaPendenciaCredencial" contract:"optional"`
}

// DadosAcessoLidos são os dados de acesso com o formulário como veio, para reenviá-lo.
type DadosAcessoLidos struct {
	DadosAcesso
	FormularioBruto json.RawMessage
}

// BuscarDadosAcesso lê os dados de acesso. 403 para usuário administrador.
func BuscarDadosAcesso(ctx context.Context, g Getter) (*DadosAcessoLidos, error) {
	var corpo json.RawMessage
	if err := g.GetJSON(ctx, PathDadosAcesso, &corpo); err != nil {
		return nil, err
	}
	var d DadosAcessoLidos
	var bruto struct {
		Formulario json.RawMessage `json:"formulario"`
	}
	if err := json.Unmarshal(corpo, &d.DadosAcesso); err != nil {
		return nil, fmt.Errorf("dados de acesso: %w", err)
	}
	if err := json.Unmarshal(corpo, &bruto); err != nil {
		return nil, fmt.Errorf("dados de acesso: %w", err)
	}
	d.FormularioBruto = bruto.Formulario
	return &d, nil
}

// NovoFormularioDadosAcesso monta o corpo de AtualizarDadosAcesso como o front: o formulário
// inteiro como veio, com os campos de mudancas trocados e o id da empresa.
func NovoFormularioDadosAcesso(bruto json.RawMessage, idEmpresa any, mudancas map[string]string) (json.RawMessage, error) {
	campos := map[string]json.RawMessage{}
	if len(bruto) > 0 && string(bruto) != "null" {
		if err := json.Unmarshal(bruto, &campos); err != nil {
			return nil, fmt.Errorf("formulário de dados de acesso: %w", err)
		}
	}
	for k, v := range mudancas {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		campos[k] = b
	}
	if id, ok := campos["id"]; !ok || string(id) == "null" {
		if idEmpresa == nil {
			return nil, fmt.Errorf("o id da empresa não veio no formulário nem na sessão; atualize pelo painel")
		}
		b, err := json.Marshal(idEmpresa)
		if err != nil {
			return nil, err
		}
		campos["id"] = b
	}
	return json.Marshal(campos)
}

// AtualizarDadosAcesso grava o formulário de dados de acesso (204 em caso de sucesso); o
// servidor então verifica o acesso aos órgãos.
func AtualizarDadosAcesso(ctx context.Context, s Sender, formulario json.RawMessage) error {
	return s.Send(ctx, "POST", PathAtualizarDadosAcesso, formulario, nil)
}

// ConfirmarCredencialPrefeitura declara que a senha atual da prefeitura é válida e limpa o
// alerta de redefinição.
func ConfirmarCredencialPrefeitura(ctx context.Context, s Sender) error {
	return s.Send(ctx, "PUT", PathConfirmarCredencialPrefeitura, nil, nil)
}
