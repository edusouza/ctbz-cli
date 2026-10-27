# Campos do `-o json`

Chaves em `snake_case`. Listas saem como array de objetos e registros como objeto. Valores
em reais são números, datas `AAAA-MM-DD`, horários RFC 3339, CNPJ/CPF só dígitos, ausentes
`null`. Campos marcados *como a API devolve* não têm formato garantido: leia o que vier.

## Sessão e empresa

| Comando | Forma | Campos |
|---|---|---|
| `status` | objeto | `empresa`, `cnpj`, `usuario`, `login_em`, `expira_em`, `situacao` (`válida`/`expirada`) |
| `empresa` | objeto | `razao_social`, `cnpj`, `situacao`, `regime_tributario`, `nome_fantasia`, `natureza_juridica`, `abertura`, `inicio_contabilidade`, `inscricao_municipal`, `inscricao_estadual`, `endereco`, `ramos_atividade`, `plano`, `certificado_situacao`, `certificado_validade`, `outras_empresas` |
| `empresas` | lista | `cnpj`, `razao_social`, `situacao`, `atual` |
| `empresa socios` | lista | `nome`, `cpf`, `administrador`, `responsavel_receita`, `categoria`, `salario_base`, `entrada`, `situacao` (o `id` de `--socio` vem de `prolabore -o json`) |
| `empresa atividades` | lista | `cnae`, `descricao`, `principal`, `ramo`, `anexos_simples` (ex. `["V","III"]`: depende do Fator R) |
| `certificado` | objeto | `situacao`, `valido`, `vencimento`, `dias_para_vencer`, `apto_renovacao`, `vencido_no_painel`, `prazo_finalizado`, `renovacao_etapa`, `renovacao_fluxo` |
| `plano` | objeto | `plano`, `categoria`, `valor`, `ramos_atividade` |

## Atenção do dia

| Comando | Forma | Campos |
|---|---|---|
| `resumo` | lista | `secao` (`impostos`, `pendencias`, `rotinas`, `mensalidade`, `painel`), `item`, `prazo`, `valor`, `situacao`, `alerta` (`vencida`, `próxima`, `crítica` ou `null`; indicadores não críticos do painel saem com `situacao: "aviso"` e `alerta: null`) |
| `pendencias` | lista | `id`, `tipo`, `detalhe` (texto completo), `criada`, `prazo`, `situacao`, `alerta` |
| `rotinas` | lista | `responsavel` (`empresa`/`contabilizei`), `rotina`, `prazo`, `status`, `valor`, `alerta`, `pendencias` (ids) |
| `pendencias termos` | lista | `chave`, `titulo`, `pendente`, `prazo_aceite_tacito_dias` |
| `pendencias procuracao` | objeto | `pendente`, `origem`, `situacao`, `cnpj_outorgado`, `link_ecac` |
| `pendencias conciliacao` | objeto | `competencia_referencia`, `notas_fiscais_pendentes`, `recebimentos_pendentes`, `conciliacoes_automaticas` |
| `pendencias conciliacao --listar …`, `candidatos`, `detalhes` | lista | como a API devolve |
| `primeiros-passos` | lista | `etapa`, `tarefa`, `situacao`, `concluida`, `conclusao_manual` |
| `chamados` | lista | `id`, `assunto`, `status`, `canal`, `criado`, `atualizado`, `previsao_retorno`, `link` |

## Impostos

| Comando | Forma | Campos |
|---|---|---|
| `impostos` | lista | `grupo` (`em_atraso`, `este_mes`, `proximo_mes`), `id`, `imposto`, `competencia`, `vencimento`, `valor`, `situacao` (guias `PAGA` também aparecem), `tipo` |
| `impostos guia ID` | objeto | `id`, `imposto`, `identificador`, `tipo`, `competencia`, `vencimento`, `valor_total`, `valor_original`, `juros_e_multa`, `valor_estimado`, `valor_em_atraso`, `situacao`, `acoes` (`PAGAR`, `INFORMAR_PAGAMENTO`, `BAIXAR_GUIA`…), `descricao`, `frequencia`, `impacto_do_atraso`. Aqui `situacao` é uma lista (`["PAGA"]`) e `vencimento` pode vir `null` |
| `impostos calculo` | objeto | `mes`, `faturamento`, `das_imposto_bruto`, `das_deducao_retencao`, `das_total`, `das_inconsistente`, `inss_prolabore`, `inss_aliquota`, `inss_total`, `irrf_base_calculo`, `irrf_aliquota`, `irrf_deducao`, `irrf_total`, `darf_total`, `faturamento_12_meses`, `prolabore_12_meses`, `fator_r` (%) |
| `impostos tabela-irrf` | lista | `base_calculo`, `aliquota`, `deducao` |
| `impostos faturamento` | lista | `competencia`, `faturamento`, `prolabore`, `impostos_pagos` (12 meses, do mais antigo ao mais recente; RBT12 e Fator R no stderr) |
| `impostos historico` | lista | `id`, `imposto`, `competencia`, `vencimento`, `valor`, `valor_pago`, `situacao`, `tipo` |
| `impostos baixar` | lista | `id`, `arquivo`, `situacao` (`baixado` ou `já existe`) |
| `impostos debitos` | objeto | `possui_debitos_federais` |
| `impostos parcelamentos` | lista | `grupo`, `id`, `titulo`, `tipo`, `situacao`, `parcela_atual` |
| `impostos parcelamento simular` | lista | `tipo`, `parcelas`, `entrada`, `demais_parcelas`, `servico_adicional`, `emissao_guia` |
| `impostos recalcular ID` | objeto | `id`, `descricao`, `vencimento`, `valor_original`, `valor_recalculo` (estimado), `data_recomendada`, `datas_indisponiveis`, `cobrar_recalculo`, `motivo_sem_cobranca` |
| `impostos recorrente` | objeto | `disponivel`, `ativo`, `situacao`, `competencia`, `proximo_pagamento`, `proxima_tentativa`, `cartoes_salvos`, `pagamentos_agendados`, `pagamentos_concluidos`, `pagamentos_recusados` |
| `impostos recorrente historico` | lista | `competencia`, `item`, `situacao`, `valor` |

## Pró-labore e lucros

| Comando | Forma | Campos |
|---|---|---|
| `prolabore` | objeto | `gerenciamento` (`INTELIGENTE`…), `total`, `valor_card`, `competencia_atual`, `competencia_anterior`, `calculando`, `indisponivel`, `socios[]` com `id` (o de `--socio`), `nome`, `cpf`, `valor`, `possui_prolabore`, `responsavel_receita`, `gestao`, `atualizado`, `dependentes` |
| `prolabore parametros` | objeto | `salario_minimo`, `inss_aliquota`, `inss_contribuicao_maxima`, `prolabore_teto_inss`, `irrf_a_partir_de` |
| `prolabore fator-r` | objeto | `motor_fator_r`, `fator_r` (%), `prolabore_12_meses`, `faturamento_12_meses`, `simulador`, `atividades[]` com `cnae`, `atividade`, `anexos`, `aliquotas`, `anexo_fixo` (não use para saber se depende do Fator R: veja `empresa atividades`) |
| `prolabore historico` | lista | `competencia`, `socio`, `prolabore`, `descontos` |
| `prolabore zerar-sem-faturamento on\|off` | objeto | `acao`, `situacao`, `zerar` (estado depois do envio). Não há leitura do estado atual: o comando exige `on` ou `off` e, se o estado já for o pedido, não envia nada (`sem mudança`) |
| `lucros` | objeto | `exercicio`, `saldo`, `total_distribuido`, `total_adiantamentos`, `limite_distribuicao`, `exercicio_fechado`, `pode_alterar`, `data_limite`, `motivo`, `pendencia_documental`, `debitos_federais`, `reabertura_balanco`, `socios[]` com `id`, `socio`, `valor`, `porcentagem` |
| `lucros informe restricoes` | objeto | `ano`, `carta_pendente`, `debitos_federais`, `debitos`, `divergencia_contabil_fiscal`, `pendencia_documental`, `pendencias`, `fluxo_regularizacao` (`NAO_NECESSARIO`, `REABERTURA_BALANCO`…), `valor_servico_adicional`, `reabertura_balanco` |
| `lucros informe` | lista | `ano`, `socio`, `cpf`, `rendimentos`, `previdencia`, `irrf_retido`, `decimo_terceiro`, `irrf_decimo_terceiro`, `lucros_isentos` |

## Notas fiscais

| Comando | Forma | Campos |
|---|---|---|
| `notas` | lista | `numero`, `emissao`, `tomador`, `documento`, `valor`, `status`, `situacao` (total no stderr) |
| `notas tomadores` | lista | `id`, `nome`, `documento`, `email`, `telefone`, `inscricao_municipal`, `municipio`, `uf`, `exterior` |
| `notas tomadores consulta CNPJ` | objeto | `cnpj`, `razao_social`, `nome_fantasia`, `abertura`, `atividade_principal`, `natureza_juridica`, `situacao_cadastral`, `optante_simples`, `endereco`, `cep`, `municipio`, `uf`, `email`, `telefone` |
| `notas config` | objeto | `emissor_habilitado`, `versao`, `instabilidade`, `permite_exterior`, `certificado_validade`, `certificado_vencido`, `certificado_em_renovacao`, `municipio`, `uf` |
| `notas aliquotas` | lista | `mercado` (`interno`/`externo`), `tipo`, `cnae`, `atividade`, `item_servico`, `descricao_item`, `aliquota` (%), `aliquota_iss` (%), `fator_r` (%), `anexo_fixo` |
| `notas entrada` | lista | `id`, `emissao`, `emitente`, `cnpj_emitente`, `valor`, `situacao`, `chave`, `classificacao` |
| `notas entrada produtos ID` | lista | `id` (item), `descricao`, `ncm`, `quantidade`, `valor`, `estoque`, `insumo`, `uso_consumo`, `ativo_imobilizado`, `prestacao_servico` |

## Contabilidade

| Comando | Forma | Campos |
|---|---|---|
| `balancete AAAA-MM` | lista | `conta`, `descricao`, `nivel`, `saldo_anterior`, `debitos`, `creditos`, `saldo` |
| `balanco AAAA[-MM]` | lista | `conta`, `descricao`, `nivel`, `grupo`, `saldo`, `saldo_exercicio_anterior` |
| `razao` | lista | `data`, `conta`, `conta_descricao`, `historico`, `contrapartida`, `debito`, `credito`, `saldo` |
| `caixa AAAA-MM` | lista | `id`, `data`, `descricao`, `conta`, `classificacao`, `valor` (saída negativa), `situacao`, `automatico` |
| `caixa contas`, `extratos contas` | lista | `id`, `descricao`, `classificacao`/`tipo`, `vinculo`/`exige_socio` |
| `extratos` | lista | `competencia`, `banco`, `agencia`, `conta`, `situacao` (`ABERTO`/`FECHADO`), `integracao`, `id_conta` |
| `contas-bancarias` | lista | `id`, `banco`, `codigo_banco`, `agencia`, `conta`, `saldo_inicial`, `data_saldo_inicial`, `integracao`, `fluxo_integracao` |
| `extratos lancamentos` | lista | `id`, `data`, `descricao`, `valor`, `classificacao`, `id_classificacao`, `id_lancamento_pai`, `situacao`, `sugestao`, `probabilidade_sugestao` (0–100) |
| `contas` | lista | `id`, `descricao`, `conta_contabil`, `classificacao`, `situacao` |
| `rotinas reclassificar ID` | lista | `id_pendencia`, `descricao`, `valor`, `classificacao_atual`, `id_classificacao`, `classificacao`, `socios` |

## Documentos, mensalidade e acessos

| Comando | Forma | Campos |
|---|---|---|
| `documentos` | lista | `tipo`, `nome`, `quantidade`, `modificado` (com `--tipo`: `competencia`, `arquivo`, `enviado`, `descricao`, `valor`, `link`) |
| `documentos pendentes` | lista | `id_pendencia`, `tipo`, `competencia`, `conta`, `tipo_investimento` |
| `mensalidade` | objeto | `competencia`, `valor`, `vencimento` (texto, ex. `"Vencimento 15/10"`), `situacao` (ex. `Atrasada`), `observacao`, `competencia_anterior_atrasada`, `botao_titulo`, `botao_acao`, `botao_adicional_titulo`, `botao_adicional_acao` (`request:…` = dá para gerar a segunda via) |
| `mensalidade segunda-via` | objeto | `acao`, `situacao` (`gerada`), `id` (`AAAAMM`), `competencia`, `vencimento`, `total`, `status`, `itens[]` (`descricao`, `valor`), `pix_copia_e_cola`, `linha_digitavel`, `codigo_barras`, `link` |
| `mensalidade situacao` | objeto | `em_dia`, `situacao` (`OK`, `EM_ATRASO`…) |
| `mensalidade historico` | objeto | `pagamentos[]`: faturas pagas, da mais recente (`data` do pagamento, `competencia` `MM/AAAA`, `vencimento`, `valor`, `multa_juros`, `forma_pagamento`, `status`). Obsoletos, do débito automático de **impostos**: `debito_automatico_ativo`, `debito_automatico_habilitado`, `status`, `competencia`, `proximo_pagamento` |
| `usuarios` | lista | `id`, `nome`, `email`, `tipo`, `status` |
| `correspondencias` | lista | `id`, `data`, `remetente`, `descricao`, `situacao`, `valor_envio` |
| `acoes` | lista | `data`, `cnpj`, `comando`, `metodo`, `caminho`, `status`, `resultado` (`enviada`, `recusada`, `sem resposta`), `id`, `arquivos` (nome e tamanho, nos envios de arquivo) |

## Resultado de escritas

Toda escrita com `-o json` devolve pelo menos `acao`, `situacao` e `id`. `situacao: "enviado"`
quer dizer aceita pela Contabilizei, mas ainda não visível na releitura: não repita.
`--dry-run -o json` devolve uma lista com `metodo`, `caminho`, `content_type`, `corpo` e `campos`.

`extratos importar` acrescenta `competencia`, `importacao` (`automatica` para OFX/PDF; `manual`
para CSV/XLS, que a Contabilizei importa em até 45 dias, com `situacao: "enviado"`), `arquivo`,
`saldo_ultimo_dia` (o saldo enviado na conclusão do OFX), `saldo_lido` (o que a Contabilizei
leu do arquivo; nulo se não leu, no PDF e no envio manual), `saldo_divergente` (verdadeiro quando
foi o de `--saldo-final` no lugar de um lido diferente ou ausente), `situacao_extrato`,
`quantidade_lancamentos` e `substituiu_extrato`.
