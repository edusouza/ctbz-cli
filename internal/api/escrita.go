package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"
)

// Sender envia uma escrita (POST, PUT, PATCH ou DELETE) autenticada e decodifica a resposta
// em v. A implementação confere a sessão antes e envia uma única vez, sem retentativa
// (ADR-0018). v pode ser nil (resposta ignorada) ou *string (resposta em texto puro).
type Sender interface {
	// Send envia body como JSON. body pode ser nil (sem corpo), []byte ou json.RawMessage
	// (enviados como estão), JSONString (string JSON) ou qualquer valor serializável.
	Send(ctx context.Context, method, path string, body any, v any) error
	// SendMultipart envia um formulário multipart/form-data (upload de arquivos).
	SendMultipart(ctx context.Context, method, path string, campos []Campo, v any) error
}

// Campo é um campo de um formulário multipart: texto (Valor) ou arquivo (Arquivo e Conteudo).
type Campo struct {
	Nome  string
	Valor string
	// Arquivo é o nome do arquivo enviado; vazio indica um campo de texto.
	Arquivo  string
	Conteudo []byte
	// Tipo é o Content-Type do arquivo; vazio deduz pela extensão do nome.
	Tipo string
}

// JSONString envia V serializado como uma string JSON, como o painel faz em algumas
// escritas (JSON.stringify do objeto e o resultado como corpo JSON).
type JSONString struct{ V any }

// MarshalJSON serializa V e devolve o resultado como string JSON.
func (s JSONString) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(s.V)
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(b))
}

// EncodeBody serializa o corpo de Sender.Send; nil devolve nil (requisição sem corpo).
func EncodeBody(body any) ([]byte, error) {
	switch b := body.(type) {
	case nil:
		return nil, nil
	case []byte:
		return b, nil
	case json.RawMessage:
		return b, nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("montando o corpo da requisição: %w", err)
	}
	return data, nil
}

// EncodeMultipart monta o corpo de Sender.SendMultipart e devolve o Content-Type com o boundary.
func EncodeMultipart(campos []Campo) ([]byte, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for _, c := range campos {
		if err := writeCampo(w, c); err != nil {
			return nil, "", fmt.Errorf("montando o campo %q: %w", c.Nome, err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}

func writeCampo(w *multipart.Writer, c Campo) error {
	if c.Arquivo == "" {
		return w.WriteField(c.Nome, c.Valor)
	}
	tipo := c.Tipo
	if tipo == "" {
		tipo = mime.TypeByExtension(strings.ToLower(filepath.Ext(c.Arquivo)))
	}
	if tipo == "" {
		tipo = "application/octet-stream"
	}
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, c.Nome, c.Arquivo))
	h.Set("Content-Type", tipo)
	part, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = part.Write(c.Conteudo)
	return err
}

// DecodeResponse decodifica a resposta de uma escrita em v: nil ignora, *string recebe o
// texto puro e corpo vazio deixa v como está.
func DecodeResponse(data []byte, v any) error {
	switch p := v.(type) {
	case nil:
		return nil
	case *string:
		*p = string(data)
		return nil
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("resposta inesperada: %w", err)
	}
	return nil
}

// Escrita liga uma escrita a uma chamada de exemplo, com valores fictícios, para os testes
// de requisição: a requisição gerada é comparada com testdata/requisicoes/<Name>.json.
type Escrita struct {
	Name    string
	Exemplo func(ctx context.Context, s Sender) error
}

// Escritas lista todas as escritas tipadas, na ordem do roadmap. Toda escrita nova entra
// aqui; o teste de cobertura falha se faltar o golden.
func Escritas() []Escrita {
	return []Escrita{
		{Name: "caixa_salvar_lancamento", Exemplo: func(ctx context.Context, s Sender) error {
			return SalvarLancamento(ctx, s, SalvarLancamentoCaixa{Ano: "2026", Mes: 9, LancamentoUsuario: LancamentoCaixaUsuario{
				Data: DataISOBrasilia(time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)), Descricao: "Guia do Simples",
				IDContaUsuario: 1000000000000013, IDVinculo: json.RawMessage(`"1000000000000021"`), Valor: -150.25,
			}})
		}},
		{Name: "caixa_editar_lancamento", Exemplo: func(ctx context.Context, s Sender) error {
			id := int64(1000000000000002)
			return SalvarLancamento(ctx, s, SalvarLancamentoCaixa{Ano: "2026", Mes: 9, LancamentoUsuario: LancamentoCaixaUsuario{
				Data: DataISOBrasilia(time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC)), Descricao: "Venda à vista", ID: &id,
				IDContaUsuario: 1000000000000011, Valor: 900,
			}})
		}},
		{Name: "extrato_classificar", Exemplo: func(ctx context.Context, s Sender) error {
			socio := int64(1000000000000031)
			return Classificar(ctx, s, ClassificarLancamento{IDLancamentoUsuario: 1000000000000102, IDContaUsuario: 6199733752168448, IDSocio: &socio})
		}},
		{Name: "extrato_desmembrar", Exemplo: func(ctx context.Context, s Sender) error {
			socio := int64(1000000000000031)
			return Desmembrar(ctx, s, Desmembramento{IDLancamentoPai: 1000000000000102, LancamentosFilho: []LancamentoFilho{
				{Descricao: "Fornecedor", Valor: -1000, IDContaUsuario: 1000000000000012},
				{Descricao: "Antecipação de lucros", Valor: -234.56, IDContaUsuario: 6199733752168448, IDVinculo: &socio},
			}})
		}},
		{Name: "extrato_desfazer_desmembramento", Exemplo: func(ctx context.Context, s Sender) error {
			return DesfazerDesmembramento(ctx, s, 1000000000000100)
		}},
		{Name: "rotinas_reclassificar", Exemplo: func(ctx context.Context, s Sender) error {
			socio := int64(1000000000000031)
			return Reclassificar(ctx, s, []Reclassificacao{
				{IDPendencia: json.RawMessage(`"1000000000000501"`), IDClassificacao: 1000000000000042, IDSocio: &socio, IDLancamento: 1000000000000102},
			})
		}},
		{Name: "impostos_confirmar_pagamento_v5", Exemplo: func(ctx context.Context, s Sender) error {
			_, err := ConfirmarPagamento(ctx, s, "v5", 1000000000000001, ConfirmacaoPagamento{Tipo: "GUIA", Origem: OrigemGuiaPadrao, PagamentoConfirmado: true})
			return err
		}},
		{Name: "impostos_negar_pagamento_v3", Exemplo: func(ctx context.Context, s Sender) error {
			_, err := ConfirmarPagamento(ctx, s, "v3", 1000000000000001, ConfirmacaoPagamento{Tipo: "PARCELA", Origem: OrigemGuiaPadrao, PagamentoConfirmado: false})
			return err
		}},
		{Name: "impostos_recalcular", Exemplo: func(ctx context.Context, s Sender) error {
			_, err := Recalcular(ctx, s, 1000000000000001, PedidoRecalculo{Tipo: "GUIA", Origem: OrigemGuiaPadrao, DataVencimento: "20/10/2026"})
			return err
		}},
		{Name: "parcelamento_contratar_simples_nacional", Exemplo: func(ctx context.Context, s Sender) error {
			_, err := ContratarParcelamento(ctx, s, "simples-nacional", 0, "")
			return err
		}},
		{Name: "parcelamento_contratar_pgfn", Exemplo: func(ctx context.Context, s Sender) error {
			_, err := ContratarParcelamento(ctx, s, "pgfn-previdenciario", 24, "")
			return err
		}},
		{Name: "parcelamento_contratar_especializado", Exemplo: func(ctx context.Context, s Sender) error {
			_, err := ContratarParcelamento(ctx, s, TipoEspecializado, 0, "DIVIDA_ATIVA_E_VENCIDOS")
			return err
		}},
		{Name: "aceitar_carta_responsabilidade", Exemplo: func(ctx context.Context, s Sender) error {
			return Aceitar(ctx, s, PathAceitarCartaResponsabilidade)
		}},
		{Name: "aceitar_termo_debitos", Exemplo: func(ctx context.Context, s Sender) error {
			return Aceitar(ctx, s, PathAceitarTermoDebitos)
		}},
		{Name: "aceitar_termo_totalpass", Exemplo: func(ctx context.Context, s Sender) error {
			return Aceitar(ctx, s, PathAceitarTermoTotalPass)
		}},
		{Name: "procuracao_central_rotinas", Exemplo: func(ctx context.Context, s Sender) error {
			return DeclararProcuracao(ctx, s, PathResolverProcuracaoEcac)
		}},
		{Name: "procuracao_checklist", Exemplo: func(ctx context.Context, s Sender) error {
			return DeclararProcuracao(ctx, s, PathCompartilharProcuracao)
		}},
		{Name: "conciliacao_vincular", Exemplo: func(ctx context.Context, s Sender) error {
			return ResolverConciliacao(ctx, s, ResolucaoConciliacao{IDPendencia: []int64{123, 124},
				Contraparte: []Contraparte{{Origem: OrigemNotaFiscal, ID: 987}}})
		}},
		{Name: "conciliacao_motivo", Exemplo: func(ctx context.Context, s Sender) error {
			motivo, socio := "EMPRESTIMO_DO_SOCIO_A_EMPRESA", int64(55)
			return ResolverConciliacao(ctx, s, ResolucaoConciliacao{IDPendencia: []int64{123}, TipoResolucaoPendencia: &motivo, IDVinculo: &socio})
		}},
		{Name: "conciliacao_vincular_com_motivo", Exemplo: func(ctx context.Context, s Sender) error {
			motivo := "DESCONTO_CONCEDIDO"
			return ResolverConciliacao(ctx, s, ResolucaoConciliacao{IDPendencia: []int64{123},
				Contraparte: []Contraparte{{Origem: OrigemMovimentacao, ID: 456}}, TipoResolucaoPendencia: &motivo, NumeroNotaAtivacaoContabil: "1234"})
		}},
		{Name: "checklist_concluir_etapa", Exemplo: func(ctx context.Context, s Sender) error {
			return ConcluirEtapa(ctx, s, "REUNIAO_BOAS_VINDAS")
		}},
		{Name: "checklist_dispensar", Exemplo: func(ctx context.Context, s Sender) error { return DispensarChecklist(ctx, s) }},
		{Name: "checklist_reativar", Exemplo: func(ctx context.Context, s Sender) error { return ReativarChecklist(ctx, s) }},
		{Name: "conta_bancaria_cadastrar", Exemplo: func(ctx context.Context, s Sender) error {
			return SalvarContaBancaria(ctx, s, ContaBancariaSalvar{BancoID: 33, Agencia: "1234", ContaCorrente: "123456", DataSaldoInicial: "2024-01-10"})
		}},
		{Name: "conta_bancaria_editar", Exemplo: func(ctx context.Context, s Sender) error {
			id := int64(1000000000000001)
			return SalvarContaBancaria(ctx, s, ContaBancariaSalvar{ID: &id, BancoID: 33, Agencia: "4321", ContaCorrente: "654321", DataSaldoInicial: "2024-01-10", VlrSaldoInicial: 1234.56})
		}},
		{Name: "conta_bancaria_excluir", Exemplo: func(ctx context.Context, s Sender) error {
			return ExcluirContaBancaria(ctx, s, 1000000000000001)
		}},
		{Name: "caixa_remover_lancamento", Exemplo: func(ctx context.Context, s Sender) error {
			return RemoverLancamento(ctx, s, 2026, 9, 1000000000000002)
		}},
	}
}
