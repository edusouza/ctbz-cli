// Package api descreve as leituras e escritas da API da Contabilizei usadas pela CLI: o
// caminho de cada endpoint, o tipo Go da resposta (ou da requisição, nas escritas) e uma
// função que busca e decodifica (Getter) ou envia (Sender, em escrita.go).
//
// Os tipos são também o contrato verificado por internal/contract: um campo com tag
// json é obrigatório na resposta, a menos que tenha `contract:"optional"`.
// Declare só os campos que a CLI usa.
package api

import "context"

// Getter faz um GET autenticado e decodifica a resposta JSON em v. A implementação
// (sessão, re-login, cookies) fica na camada de cima.
type Getter interface {
	GetJSON(ctx context.Context, path string, v any) error
}

// TextGetter faz um GET autenticado e devolve o corpo como texto, para os poucos
// endpoints que não respondem JSON (ex.: situação da mensalidade, contrato em HTML).
type TextGetter interface {
	GetText(ctx context.Context, path string) (string, error)
}

// Endpoint liga um caminho ao tipo da resposta, para os testes de contrato.
type Endpoint struct {
	// Name identifica a fixture: testdata/<Name>.json.
	Name string
	// LivePath é o caminho chamado no modo ao vivo; vazio quando depende de dados
	// de outra resposta (ex.: um ID) e não pode ser chamado isoladamente.
	LivePath string
	// Type é o valor zero do tipo que decodifica a resposta.
	Type any
}

// Endpoints lista todas as leituras tipadas, na ordem do roadmap.
func Endpoints() []Endpoint {
	return []Endpoint{
		{Name: "dadosempresa", LivePath: PathDadosEmpresa, Type: DadosEmpresa{}},
		{Name: "appbar", LivePath: PathAppBar, Type: AppBar{}},
		{Name: "menu", LivePath: PathMenu, Type: []MenuItem{}},
		{Name: "sessao_empresa", Type: EmpresaSessao{}}, // localStorage "e" do login
		{Name: "certificado_status", LivePath: PathCertificadoStatus, Type: CertificadoStatus{}},
		{Name: "certificado_processo", LivePath: PathCertificadoProcesso, Type: CertificadoProcesso{}},
		{Name: "certificado_card", LivePath: PathCertificadoCard, Type: CertificadoCard{}},
		{Name: "socios", LivePath: PathSocios, Type: []Socio{}},
		{Name: "cnaes", LivePath: PathCNAEs, Type: []CNAEEmpresa{}},
		{Name: "guias_a_pagar", LivePath: PathGuiasAPagar, Type: GuiasAPagar{}},
		{Name: "rollout", LivePath: PathRollout, Type: Rollout{}}, // escrita a partir de endpoints-verificados.md
		{Name: "guia_detalhe", Type: GuiaDetalhe{}},               // o ID vem de guias_a_pagar
		{Name: "recalculo_init", Type: RecalculoInit{}},           // sintética (ADR-0021); o ID vem de guias_a_pagar
		{Name: "calculo_imposto", LivePath: PathCalculoImposto, Type: CalculoImposto{}},
		{Name: "tabela_irrf", LivePath: PathTabelaIRRF, Type: []FaixaIRRF{}},
		{Name: "link_guia", Type: LinkDownload{}}, // o ID vem de guias_a_pagar
		{Name: "historico_resumo", LivePath: PathHistoricoResumo, Type: HistoricoResumo{}},
		{Name: "historico_guias", LivePath: PathHistoricoGuias + "?pagina=1", Type: HistoricoGuias{}},
		{Name: "impostos_pagos_no_ano", Type: DadosGrafico{}}, // o ano vem do histórico de faturamento
		{Name: "parcelamentos", LivePath: PathImpostosV3, Type: ParcelamentosV3{}},
		{Name: "debitos_federais", LivePath: PathDebitosFederais, Type: DebitosFederais{}},
		{Name: "simulacao_parcelamento", Type: SimulacaoParcelamento{}}, // sintética (ADR-0021); sem dívidas a API responde 560
		{Name: "pendencias_empresa", LivePath: PathPendenciasEmpresa, Type: []PendenciaEmpresa{}},
		{Name: "central_rotinas", LivePath: PathCentralRotinas, Type: CentralRotinas{}},
		{Name: "checklist_onboarding", LivePath: PathChecklistInit, Type: ChecklistOnboarding{}},     // sintética (ADR-0021)
		{Name: "central_rotinas_init", LivePath: PathCentralRotinasInit, Type: CentralRotinasInit{}}, // sintética (ADR-0021)
		{Name: "reclassificar_init", Type: []PendenciaReclassificacao{}},                             // sintética (ADR-0021); ids de central_rotinas
		{Name: "conciliacao_resumo", LivePath: PathConciliacaoInit, Type: ConciliacaoResumo{}},
		{Name: "conciliacao_pendencias", LivePath: PathConciliacaoPendencias + "?status=PENDENTE&tipoPendencia=" + PendenciaNotaSemRecebimento, Type: PaginaConciliacao{}},
		{Name: "chamados_em_andamento", LivePath: PathChamadosEmAndamento, Type: []Chamado{}},
		{Name: "fatura", LivePath: PathFatura, Type: Fatura{}},
		{Name: "mensalidade", LivePath: PathMensalidade, Type: Mensalidade{}},
		{Name: "recorrencia", LivePath: PathRecorrenciaInit, Type: Recorrencia{}},
		{Name: "recorrencia_historico", LivePath: PathRecorrenciaHistorico, Type: HistoricoRecorrencia{}},
		{Name: "contrato_servico", LivePath: PathContratoServico, Type: ContratoServico{}},
		{Name: "notas_emitidas", LivePath: PathNotasEmitidas + "?pagina=1&limite=10&ano=2026&mes=10", Type: ListaNotas{}},
		{Name: "tomadores", LivePath: PathTomadores, Type: Tomadores{}},
		{Name: "cep", LivePath: PathCEP("01001000"), Type: EnderecoCEP{}},                           // sintética (ADR-0021); CEP público (Praça da Sé)
		{Name: "paises_emissao", LivePath: PathPaisesEmissao, Type: []PaisEmissao{}},                // sintética (ADR-0021)
		{Name: "cadastro_cliente", Type: CadastroCliente{}},                                         // sintética (ADR-0021); o id vem de tomadores
		{Name: "consulta_cnpj", LivePath: PathConsultaCNPJ("00000000000191"), Type: ConsultaCNPJ{}}, // CNPJ público (Banco do Brasil)
		{Name: "emissor_init", LivePath: PathEmissorInit, Type: EmissorInit{}},
		{Name: "versao_emissor", LivePath: PathVersaoEmissor, Type: VersaoEmissor{}},
		{Name: "aliquotas_emissor", LivePath: PathAliquotasEmissor, Type: AliquotasEmissor{}},
		{Name: "notas_entrada", LivePath: PathNotasEntradaManifestacao + "0?mes=9&ano=2026&empresa=&qtdPagina=10&cursor=", Type: ListaNotasEntrada{}},
		{Name: "notas_entrada_classificacao", LivePath: PathNotasEntradaClassificacao + "?mes=9&ano=2026&empresa=&tipo=0&limite=10&cursor=&offset=0", Type: ListaNotasEntrada{}},
		{Name: "prolabore_central", LivePath: PathProlaboreCentral, Type: ProlaboreCentral{}},
		{Name: "prolabore_dashboard", LivePath: PathProlaboreDashboard, Type: ProlaboreDashboard{}},
		{Name: "prolabore_historico", Type: []ProlaboreMes{}}, // o id do sócio vem de prolabore_central
		{Name: "prolabore_parametros", LivePath: PathProlaboreParametros, Type: ProlaboreParametros{}},
		{Name: "simulador_impostos", LivePath: PathSimuladorImpostos, Type: SimuladorImpostos{}},
		{Name: "balancete", LivePath: PathBalancete(2026, 9), Type: []ContaRelatorio{}},
		{Name: "balanco", LivePath: PathBalanco(2026, 9), Type: []ContaRelatorio{}},
		{Name: "razao", LivePath: PathRazao(2026, 9), Type: []ContaRelatorio{}},
		{Name: "caixa", LivePath: PathCaixa(2026, 9), Type: Caixa{}},
		{Name: "caixa_categorias", LivePath: PathCaixa(2026, 9), Type: CaixaCategorias{}}, // sintética (ADR-0021)
		{Name: "contas_usuario", LivePath: PathContasUsuario, Type: []ContaUsuario{}},
		{Name: "contas_usuario_competencia", LivePath: PathContasUsuarioCompetencia(2026, 9, OrigemExtrato), Type: []ContaUsuarioCompetencia{}}, // sintética (ADR-0021)
		{Name: "extratos", LivePath: PathExtratos, Type: []Extrato{}},
		{Name: "lancamentos_extrato", Type: PaginaLancamentosExtrato{}},                           // sintética (ADR-0021); a conta vem de extratos
		{Name: "permite_importacao", Type: PermiteImportacao{}},                                   // sintética (ADR-0021); a conta vem de extratos
		{Name: "upload_extrato_init", LivePath: PathUploadExtratoInit, Type: UploadExtratoInit{}}, // sintética (ADR-0021)
		{Name: "info_extrato", Type: InfoExtrato{}},                                               // sintética (ADR-0021); só existe depois de um upload
		{Name: "extrato_info", Type: ExtratoInfo{}},                                               // sintética (ADR-0021); a conta vem de extratos
		{Name: "contas_bancarias", LivePath: PathContasBancarias, Type: ContasBancarias{}},
		{Name: "detalhe_conta_bancaria", Type: DetalheContaBancaria{}}, // sintética (ADR-0021); o ID vem de contas_bancarias
		{Name: "documentos_tipos", LivePath: PathTiposDocumento(AreaDocumentosContabeis), Type: []TipoDocumento{}},
		{Name: "envio_documento_init", Type: EnvioDocumentoInit{}}, // sintética (ADR-0021); ids das pendências da Central de Rotinas
		{Name: "documentos_enviados", LivePath: PathDocumentosEnviados("EXTRATO_BANCARIO_MOVIMENTACOES", 12, 0), Type: PaginaDocumentos{}},
		{Name: "distribuicao_lucros", LivePath: PathDistribuicaoLucros, Type: DistribuicaoLucros{}},
		{Name: "restricoes_informe", LivePath: PathRestricoesInforme(2025), Type: RestricoesInforme{}},
		{Name: "socios_informe", LivePath: PathSociosInforme(2025), Type: []SocioInforme{}},
		// getValoresComprovanteRendimento fica sem contrato: nenhum exemplo real (sem sócios com informe).
	}
}

func get[T any](ctx context.Context, g Getter, path string) (*T, error) {
	var v T
	if err := g.GetJSON(ctx, path, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
