package api

import "context"

// Candidatos para conciliar (GET; parâmetros não verificados): notas fiscais para um
// recebimento e movimentações para uma nota. Ver docs/api/escrita/pendencias-e-conta.md, 1.4.
const (
	PathCandidatosNotas        = "conciliacao-fiscal/conciliar/receitas"
	PathCandidatosRecebimentos = "conciliacao-fiscal/conciliar/recebimentos"
)

// PathDetalhesConciliacao é uma leitura feita com POST (sem efeito), com DetalhesConciliacao no corpo.
const PathDetalhesConciliacao = "conciliacao-fiscal/pendencia/detalhes"

// DetalhesConciliacao é o corpo de PathDetalhesConciliacao.
type DetalhesConciliacao struct {
	IDsRecebimento []int64 `json:"idsRecebimento"`
	IDsReceita     []int64 `json:"idsReceita"`
}

// Origens das contrapartes: notas para pendências de recebimento, movimentações para
// pendências de nota.
const (
	OrigemNotaFiscal   = "NOTAFISCAL"
	OrigemMovimentacao = "MOVIMENTACAO"
)

// Contraparte é o item vinculado a uma pendência na conciliação.
type Contraparte struct {
	Origem string `json:"origem"`
	ID     int64  `json:"id"`
}

// ResolucaoConciliacao é o corpo de PathResolverConciliacao, em três formas: só contrapartes
// (conciliar), só motivo (justificar sem contraparte) ou os dois (diferença justificada).
type ResolucaoConciliacao struct {
	IDPendencia                []int64       `json:"idPendencia"`
	Contraparte                []Contraparte `json:"contraparte"` // [] quando não há
	TipoResolucaoPendencia     *string       `json:"tipoResolucaoPendencia"`
	IDVinculo                  *int64        `json:"idVinculo,omitempty"` // sócio, nos motivos que exigem
	NumeroNotaAtivacaoContabil string        `json:"numeroNotaAtivacaoContabil,omitempty"`
}

// PathResolverConciliacao resolve (ou refaz) pendências de conciliação fiscal.
const PathResolverConciliacao = "conciliacao-fiscal/conciliar/resolver-pendencia"

// ResolverConciliacao concilia ou justifica pendências. Define a natureza fiscal do
// recebimento; pode ser refeito pelo mesmo endpoint, mas não apagado.
func ResolverConciliacao(ctx context.Context, s Sender, req ResolucaoConciliacao) error {
	if req.Contraparte == nil {
		req.Contraparte = []Contraparte{}
	}
	return s.Send(ctx, "POST", PathResolverConciliacao, req, nil)
}

// MotivoConciliacao é um valor de tipoResolucaoPendencia, com o rótulo para o usuário.
type MotivoConciliacao struct {
	Codigo, Rotulo string
	// ExigeSocio indica motivos que pedem o sócio em idVinculo (inferido; ver a documentação).
	ExigeSocio bool
}

// MotivosConciliacao lista os motivos encontrados no front, em ordem alfabética do código.
var MotivosConciliacao = []MotivoConciliacao{
	{"ADIANTAMENTO_EMISSAO_NOTA_FISCAL_FUTURA", "Adiantamento: a nota fiscal será emitida depois", false},
	{"AFAC", "Valor adiantado para virar capital social (AFAC)", true},
	{"ATIVACAO_CONTABIL", "Ativação contábil (nota anterior)", false},
	{"CAIXA_OU_PESSOA_FISICA", "Recebi em dinheiro ou em uma conta PF", false},
	{"CASHBACK", "Cashback", false},
	{"DESCONTO_CONCEDIDO", "Desconto concedido", false},
	{"DEVOLUCAO_DE_SAQUE_DE_LUCROS_ANTECIPADOS", "Devolução de saque de lucros antecipados", true},
	{"DIFERENCA_REPASSAR", "Diferença a repassar", false},
	{"EMPRESTIMO_BANCARIO", "Empréstimo bancário", false},
	{"EMPRESTIMO_DO_SOCIO_A_EMPRESA", "Empréstimo do sócio à empresa", true},
	{"ESTORNO_TRANSACAO_BANCARIA", "Estorno de transação bancária", false},
	{"INTEGRALIZACAO_CAPITAL", "Integralização de capital", true},
	{"INVESTIMENTO_ANJO_MUTUO", "Investimento anjo (mútuo)", false},
	{"JUROS_RECEBIDOS", "Juros recebidos", false},
	{"NOTA_FISCAL_FUTURA", "Nota fiscal futura", false},
	{"PERDA", "Não vou mais receber por essa nota", false},
	{"RECEBIMENTO_DE_EMPRESTIMOS_CONCEDIDOS", "Recebimento de empréstimos concedidos", false},
	{"RECEBIMENTO_DE_INTERMEDIACAO_DE_SERVICOS_A_SEREM_REPASSADOS", "Intermediação de serviços a repassar", false},
	{"RECEBIMENTO_DE_VALORES_A_SEREM_REPASSADOS_AO_CLIENTE", "Valores a repassar ao cliente", false},
	{"RECEBIMENTO_PARCELADO", "Recebimento parcelado", false},
	{"REEMBOLSO", "Reembolso", false},
	{"REEMBOLSO_DE_CLIENTES_ADMINISTRACAO_DE_IMOVEIS", "Reembolso de clientes (administração de imóveis)", false},
	{"REEMBOLSO_DE_PAGAMENTO_ANTECIPADO_DE_DESPESAS_PARA_CLIENTES", "Reembolso de despesas pagas pelos clientes", false},
	{"REEMBOLSO_DE_TAXAS_E_CUSTAS_DE_PROCESSOS_JUDICIAIS_DE_CLIENTES", "Reembolso de custas judiciais de clientes", false},
	{"TAXAS_INTERMEDIADORAS", "Taxas de intermediadoras", false},
	{"TRANSFERENCIA_ENTRE_CONTAS", "Transferência entre contas", false},
	{"VARIACAO_CAMBIAL_NEGATIVA", "Variação cambial negativa", false},
	{"VARIACAO_CAMBIAL_POSITIVA", "Variação cambial positiva", false},
}

// MotivoPorCodigo procura um motivo pelo código.
func MotivoPorCodigo(codigo string) (MotivoConciliacao, bool) {
	for _, m := range MotivosConciliacao {
		if m.Codigo == codigo {
			return m, true
		}
	}
	return MotivoConciliacao{}, false
}
