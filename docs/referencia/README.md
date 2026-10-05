# Referência de comandos

Gerada a partir da própria CLI (`go run ./tools/gendocs`); não edite à mão.

| Comando | Descrição |
|---|---|
| [`ctbz`](ctbz.md) | CLI para a Contabilizei |
| [`ctbz acoes`](ctbz_acoes.md) | Lista as ações de escrita enviadas à Contabilizei por esta CLI |
| [`ctbz api`](ctbz_api.md) | Chama uma URL da plataforma com a sessão atual |
| [`ctbz balancete`](ctbz_balancete.md) | Mostra o balancete de verificação de um mês |
| [`ctbz balanco`](ctbz_balanco.md) | Mostra o balanço patrimonial (ativo, passivo e patrimônio líquido) |
| [`ctbz caixa`](ctbz_caixa.md) | Lista os lançamentos do caixa de um mês |
| [`ctbz caixa adicionar`](ctbz_caixa_adicionar.md) | Adiciona um recebimento ou pagamento no caixa de uma competência |
| [`ctbz caixa contas`](ctbz_caixa_contas.md) | Lista as classificações aceitas nos lançamentos do caixa de uma competência |
| [`ctbz caixa editar`](ctbz_caixa_editar.md) | Altera um lançamento manual do caixa |
| [`ctbz caixa remover`](ctbz_caixa_remover.md) | Exclui permanentemente um lançamento manual do caixa |
| [`ctbz certificado`](ctbz_certificado.md) | Mostra a situação do certificado digital da empresa e da renovação |
| [`ctbz chamados`](ctbz_chamados.md) | Lista os chamados de atendimento (em andamento ou finalizados) |
| [`ctbz contas`](ctbz_contas.md) | Lista o plano de contas usado para classificar os lançamentos |
| [`ctbz contas-bancarias`](ctbz_contas-bancarias.md) | Lista as contas bancárias cadastradas da empresa |
| [`ctbz documentos`](ctbz_documentos.md) | Lista os tipos de documento aceitos e os documentos enviados |
| [`ctbz empresa`](ctbz_empresa.md) | Mostra os dados da empresa selecionada |
| [`ctbz empresa atividades`](ctbz_empresa_atividades.md) | Lista os CNAEs da empresa e os anexos do Simples Nacional |
| [`ctbz empresa certificado`](ctbz_empresa_certificado.md) | Mostra a situação do certificado digital da empresa e da renovação |
| [`ctbz empresa socios`](ctbz_empresa_socios.md) | Lista os sócios da empresa |
| [`ctbz empresa usar`](ctbz_empresa_usar.md) | Troca a empresa da sessão |
| [`ctbz empresas`](ctbz_empresas.md) | Lista as empresas do usuário, marcando a atual |
| [`ctbz extratos`](ctbz_extratos.md) | Lista a situação dos extratos bancários por mês e conta |
| [`ctbz extratos classificar`](ctbz_extratos_classificar.md) | Troca a classificação de um lançamento do extrato |
| [`ctbz extratos contas`](ctbz_extratos_contas.md) | Lista as classificações aceitas nos lançamentos do extrato de uma competência |
| [`ctbz extratos desfazer-desmembramento`](ctbz_extratos_desfazer-desmembramento.md) | Volta um lançamento desmembrado do extrato ao original |
| [`ctbz extratos desmembrar`](ctbz_extratos_desmembrar.md) | Divide um lançamento do extrato em partes com classificações diferentes |
| [`ctbz extratos lancamentos`](ctbz_extratos_lancamentos.md) | Lista os lançamentos do extrato de uma conta bancária num mês |
| [`ctbz impostos`](ctbz_impostos.md) | Lista as guias de impostos a pagar (em atraso, do mês e do próximo mês) |
| [`ctbz impostos baixar`](ctbz_impostos_baixar.md) | Baixa o PDF de guias de imposto |
| [`ctbz impostos calculo`](ctbz_impostos_calculo.md) | Mostra como o imposto do mês foi calculado |
| [`ctbz impostos confirmar`](ctbz_impostos_confirmar.md) | Informa que uma guia foi paga (ou, com --nao-paguei, que não foi) |
| [`ctbz impostos debitos`](ctbz_impostos_debitos.md) | Indica se a empresa tem débitos federais em aberto |
| [`ctbz impostos desmarcar`](ctbz_impostos_desmarcar.md) | Desfaz a confirmação de pagamento de uma guia |
| [`ctbz impostos faturamento`](ctbz_impostos_faturamento.md) | Mostra o faturamento, o pró-labore e os impostos pagos nos últimos 12 meses |
| [`ctbz impostos guia`](ctbz_impostos_guia.md) | Mostra o detalhe de uma guia de imposto |
| [`ctbz impostos historico`](ctbz_impostos_historico.md) | Lista o histórico de guias de impostos |
| [`ctbz impostos parcelamento`](ctbz_impostos_parcelamento.md) | Mostra o detalhe de um parcelamento de impostos |
| [`ctbz impostos parcelamento contratar`](ctbz_impostos_parcelamento_contratar.md) | Contrata um parcelamento de débitos |
| [`ctbz impostos parcelamento simular`](ctbz_impostos_parcelamento_simular.md) | Simula um parcelamento de débitos (opções de parcelas e custos) |
| [`ctbz impostos parcelamentos`](ctbz_impostos_parcelamentos.md) | Lista os parcelamentos de impostos (em andamento, ativos e encerrados) |
| [`ctbz impostos recalcular`](ctbz_impostos_recalcular.md) | Mostra ou pede o recálculo de uma guia vencida |
| [`ctbz impostos recorrente`](ctbz_impostos_recorrente.md) | Mostra a situação do pagamento recorrente (débito automático) de impostos |
| [`ctbz impostos recorrente historico`](ctbz_impostos_recorrente_historico.md) | Lista os pagamentos recorrentes de impostos dos últimos meses |
| [`ctbz impostos tabela-irrf`](ctbz_impostos_tabela-irrf.md) | Mostra a tabela progressiva do IRRF usada no cálculo do pró-labore |
| [`ctbz login`](ctbz_login.md) | Autentica na Contabilizei (usuário, senha e código por e-mail) |
| [`ctbz logout`](ctbz_logout.md) | Apaga a sessão local |
| [`ctbz lucros`](ctbz_lucros.md) | Mostra a distribuição de lucros do exercício e o que a impede |
| [`ctbz lucros informe`](ctbz_lucros_informe.md) | Mostra os valores do informe de rendimentos dos sócios (para o IR) |
| [`ctbz mensalidade`](ctbz_mensalidade.md) | Mostra a mensalidade atual da Contabilizei (valor, vencimento e situação) |
| [`ctbz mensalidade historico`](ctbz_mensalidade_historico.md) | Lista os pagamentos anteriores da mensalidade e a situação do débito automático |
| [`ctbz mensalidade situacao`](ctbz_mensalidade_situacao.md) | Indica se a empresa está em dia com a Contabilizei |
| [`ctbz notas`](ctbz_notas.md) | Lista as notas fiscais de serviço (NFS-e) emitidas |
| [`ctbz notas aliquotas`](ctbz_notas_aliquotas.md) | Lista as alíquotas e os códigos de serviço por atividade |
| [`ctbz notas config`](ctbz_notas_config.md) | Mostra a configuração do emissor de notas |
| [`ctbz notas entrada`](ctbz_notas_entrada.md) | Lista as notas fiscais de entrada (NF-e recebidas pela empresa) |
| [`ctbz notas tomadores`](ctbz_notas_tomadores.md) | Lista os tomadores (clientes) cadastrados no emissor de notas |
| [`ctbz notas tomadores consulta`](ctbz_notas_tomadores_consulta.md) | Consulta os dados cadastrais de um CNPJ na Receita |
| [`ctbz pendencias`](ctbz_pendencias.md) | Lista as pendências da empresa (tipo, detalhe, prazo e situação) |
| [`ctbz pendencias aceitar`](ctbz_pendencias_aceitar.md) | Aceita um termo ou carta pendente da Central de Rotinas |
| [`ctbz pendencias conciliacao`](ctbz_pendencias_conciliacao.md) | Mostra as pendências de conciliação fiscal (notas e recebimentos) |
| [`ctbz pendencias procuracao`](ctbz_pendencias_procuracao.md) | Mostra a pendência de procuração do e-CAC ou declara que ela foi criada |
| [`ctbz pendencias termo`](ctbz_pendencias_termo.md) | Mostra o texto completo de um termo ou carta da Central de Rotinas |
| [`ctbz pendencias termos`](ctbz_pendencias_termos.md) | Lista os termos e cartas com aceite pendente na Central de Rotinas |
| [`ctbz plano`](ctbz_plano.md) | Mostra o plano contratado com a Contabilizei |
| [`ctbz plano contrato`](ctbz_plano_contrato.md) | Exporta o contrato de prestação de serviços (HTML ou texto) |
| [`ctbz plano proposta`](ctbz_plano_proposta.md) | Exporta a proposta do plano contratado, com a tabela de preços (HTML ou texto) |
| [`ctbz prolabore`](ctbz_prolabore.md) | Mostra o pró-labore vigente por sócio e o tipo de gerenciamento |
| [`ctbz prolabore fator-r`](ctbz_prolabore_fator-r.md) | Mostra a situação do Fator R e os anexos possíveis de cada atividade |
| [`ctbz prolabore historico`](ctbz_prolabore_historico.md) | Lista o histórico mensal de pró-labore dos sócios |
| [`ctbz prolabore parametros`](ctbz_prolabore_parametros.md) | Mostra os valores usados no cálculo do pró-labore (INSS e IRRF) |
| [`ctbz razao`](ctbz_razao.md) | Lista os lançamentos do razão contábil por conta |
| [`ctbz resumo`](ctbz_resumo.md) | Mostra numa lista só o que precisa de atenção |
| [`ctbz rotinas`](ctbz_rotinas.md) | Lista as rotinas e obrigações do mês (da empresa e da Contabilizei) |
| [`ctbz rotinas reclassificar`](ctbz_rotinas_reclassificar.md) | Confirma ou altera a classificação de lançamentos pedida pela Central de Rotinas |
| [`ctbz status`](ctbz_status.md) | Mostra a sessão atual e testa se ainda é válida |
| [`ctbz version`](ctbz_version.md) | Mostra a versão do ctbz |
