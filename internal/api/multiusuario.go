package api

import "context"

// Multiusuário: quem acessa a empresa no painel (base /api/multiusuario/). Ver
// docs/api/escrita/pendencias-e-conta.md, seção 3.2.
const (
	PathUsuariosEmpresa      = "/api/multiusuario/usuarios-e-invites/consultar"
	PathStatusMultiusuario   = "/api/multiusuario/status/servico"
	PathConvidarUsuario      = "/api/multiusuario/invites/enviar"
	PathAtivarUsuarioEmpresa = "/api/multiusuario/usuarios/ativar"
)

// Tipos e situações dos usuários da empresa.
const (
	UsuarioPrincipal      = "USUARIO_PRINCIPAL"
	UsuarioSecundario     = "USUARIO_SECUNDARIO"
	UsuarioAtivo          = "ATIVO"
	UsuarioInativo        = "INATIVO"
	UsuarioConviteEnviado = "CONVITE_ENVIADO"
)

// UsuarioEmpresa é um usuário ou convite da empresa. Campos citados no front; sem captura
// real (ADR-0021).
type UsuarioEmpresa struct {
	ID     any    `json:"id" contract:"optional"` // o idUsuarioEmpresa de PathAtivarUsuarioEmpresa
	Nome   string `json:"nome" contract:"optional"`
	Email  string `json:"email" contract:"optional"`
	Tipo   string `json:"tipo" contract:"optional"`
	Status string `json:"status" contract:"optional"`
}

// BuscarUsuariosEmpresa lista os usuários e os convites da empresa.
func BuscarUsuariosEmpresa(ctx context.Context, g Getter) ([]UsuarioEmpresa, error) {
	u, err := get[[]UsuarioEmpresa](ctx, g, PathUsuariosEmpresa)
	if err != nil {
		return nil, err
	}
	return *u, nil
}

// StatusMultiusuario diz se o serviço está disponível e quantos usuários secundários ativos
// ele permite.
type StatusMultiusuario struct {
	Ativo  bool `json:"ativo" contract:"optional"`
	Limite int  `json:"limite" contract:"optional"`
}

// BuscarStatusMultiusuario lê o status do serviço de multiusuário.
func BuscarStatusMultiusuario(ctx context.Context, g Getter) (*StatusMultiusuario, error) {
	return get[StatusMultiusuario](ctx, g, PathStatusMultiusuario)
}

// ConvidarUsuario envia um convite com acesso total à empresa, inclusive financeiro. Não há
// como revogar o convite antes do aceite.
func ConvidarUsuario(ctx context.Context, s Sender, email string) error {
	return s.Send(ctx, "POST", PathConvidarUsuario, struct {
		Email string `json:"email"`
	}{email}, nil)
}

// AtivarUsuarioEmpresa ativa ou desativa o acesso de um usuário (a única forma de revogar).
func AtivarUsuarioEmpresa(ctx context.Context, s Sender, idUsuarioEmpresa any, ativo bool) error {
	return s.Send(ctx, "PUT", PathAtivarUsuarioEmpresa, struct {
		IDUsuarioEmpresa any  `json:"idUsuarioEmpresa"`
		Ativo            bool `json:"ativo"`
	}{idUsuarioEmpresa, ativo}, nil)
}
