# Mapa de comandos do ctbz

Gerado de `docs/referencia` por `scripts/gerar-comandos-skill.py`; não edite à mão.

Todos os comandos aceitam `-o table|json|csv`. Os de escrita aceitam também `--dry-run`
(mostra a requisição sem enviar) e `-y/--yes` (envia sem perguntar; obrigatório sem terminal).
Campos do JSON de cada comando: [campos-json.md](campos-json.md).

## acoes

`ctbz acoes [flags]`: Lista as ações de escrita enviadas à Contabilizei por esta CLI

```text
--desde string   só as ações a partir desta data (AAAA-MM-DD)
--limite int     quantas ações mais recentes mostrar (0 mostra todas) (default 50)
```

## api

`ctbz api CAMINHO [flags]`: Chama uma URL da plataforma com a sessão atual

```text
-d, --data string     corpo JSON da requisição (@arquivo lê de um arquivo, @- da entrada padrão)
-X, --method string   método HTTP (default "GET")
--raw             imprime o corpo da resposta como veio, sem formatar
```

## balancete

`ctbz balancete AAAA-MM`: Mostra o balancete de verificação de um mês

## balanco

`ctbz balanco AAAA[-MM]`: Mostra o balanço patrimonial (ativo, passivo e patrimônio líquido)

`ctbz balanco reabrir [flags]`: Reabre o balanço de um ano para regularizar a pendência documental do informe

```text
--ano int   ano do exercício a reabrir
```

## caixa

`ctbz caixa AAAA-MM`: Lista os lançamentos do caixa de um mês

`ctbz caixa adicionar [flags]`: Adiciona um recebimento ou pagamento no caixa de uma competência

```text
--competencia string   competência AAAA-MM (obrigatória)
--conta string         classificação: id ou descrição exata (ver ctbz caixa contas)
--data string          data do lançamento AAAA-MM-DD (não pode ser futura)
--descricao string     descrição do lançamento
--guia string          guia de imposto vinculada (contas "Impostos - …"; ver --vinculos)
--pagamento            saída de dinheiro do caixa
--recebimento          entrada de dinheiro no caixa
--sem-guia             classificação de imposto sem guia vinculada
--socio string         sócio vinculado (Distribuição de Lucros Antecipados; ver --vinculos)
--valor string         valor em reais, sempre positivo (ex.: 150,25)
```

`ctbz caixa contas [flags]`: Lista as classificações aceitas nos lançamentos do caixa de uma competência

```text
--competencia string   competência AAAA-MM (obrigatória)
--pagamento            só as classificações de pagamento
--recebimento          só as classificações de recebimento
--vinculos             lista as guias e os sócios aceitos como vínculo
```

`ctbz caixa editar ID [flags]`: Altera um lançamento manual do caixa

```text
--competencia string   competência AAAA-MM (obrigatória)
--conta string         classificação: id ou descrição exata (ver ctbz caixa contas)
--data string          data do lançamento AAAA-MM-DD (não pode ser futura)
--descricao string     descrição do lançamento
--guia string          guia de imposto vinculada (contas "Impostos - …"; ver --vinculos)
--pagamento            saída de dinheiro do caixa
--recebimento          entrada de dinheiro no caixa
--sem-guia             classificação de imposto sem guia vinculada
--socio string         sócio vinculado (Distribuição de Lucros Antecipados; ver --vinculos)
--valor string         valor em reais, sempre positivo (ex.: 150,25)
```

`ctbz caixa remover ID [flags]`: Exclui permanentemente um lançamento manual do caixa

```text
--competencia string   competência AAAA-MM (obrigatória)
```

## certificado

`ctbz certificado`: Mostra a situação do certificado digital da empresa e da renovação

## chamados

`ctbz chamados [flags]`: Lista os chamados de atendimento (em andamento ou finalizados)

```text
--finalizados   lista os chamados finalizados
```

## conta

`ctbz conta`: Mostra os dados de login do usuário (e-mail, telefone e método do 2FA)

`ctbz conta alterar [flags]`: Troca o e-mail e o telefone de login (com código OTP)

```text
--email string           novo e-mail de login
--otp-cmd string         comando de shell que imprime o código (env CTBZ_OTP_CMD)
--otp-timeout duration   tempo máximo aguardando o --otp-cmd (env CTBZ_OTP_TIMEOUT) (default 3m0s)
--telefone string        novo telefone, (DD)NNNNN-NNNN
--via string             por onde receber o código: email ou sms (padrão: o método do 2FA)
```

`ctbz conta senha [flags]`: Troca a senha de login (com código OTP)

```text
--otp-cmd string         comando de shell que imprime o código (env CTBZ_OTP_CMD)
--otp-timeout duration   tempo máximo aguardando o --otp-cmd (env CTBZ_OTP_TIMEOUT) (default 3m0s)
--via string             por onde receber o código: email ou sms (padrão: o método do 2FA)
```

## contas-bancarias

`ctbz contas-bancarias`: Lista as contas bancárias cadastradas da empresa

`ctbz contas-bancarias adicionar [flags]`: Cadastra uma conta bancária PJ

```text
--abertura string        data de abertura da conta AAAA-MM-DD (define desde quando os extratos são pedidos)
--agencia string         agência, sem dígito
--banco string           banco: id ou código exato (ver ctbz contas-bancarias bancos)
--conta string           conta com dígito (ex.: 12345-6)
--declaro-conta-pj       declara que é uma conta PJ que movimenta só este CNPJ (obrigatório)
--saldo-inicial string   saldo na abertura, em reais (padrão: 0)
```

`ctbz contas-bancarias bancos`: Lista os bancos aceitos no cadastro de contas bancárias

`ctbz contas-bancarias editar ID [flags]`: Corrige os dados de uma conta bancária cadastrada

```text
--abertura string        data de abertura da conta AAAA-MM-DD (define desde quando os extratos são pedidos)
--agencia string         agência, sem dígito
--banco string           banco: id ou código exato (ver ctbz contas-bancarias bancos)
--conta string           conta com dígito (ex.: 12345-6)
--saldo-inicial string   saldo na abertura, em reais (padrão: 0)
```

`ctbz contas-bancarias remover ID [flags]`: Exclui permanentemente uma conta bancária

## contas

`ctbz contas [flags]`: Lista o plano de contas usado para classificar os lançamentos

```text
--busca string      trecho da descrição ou da conta contábil
--situacao string   ativo ou inativo (padrão: todas)
```

## correspondencias

`ctbz correspondencias`: Lista as correspondências recebidas no escritório virtual

`ctbz correspondencias autorizar on|off [flags]`: Autoriza ou não o recebimento de correspondências no endereço informado

`ctbz correspondencias baixar ID [flags]`: Baixa uma correspondência do escritório virtual

```text
--dir string   diretório onde salvar (default ".")
--force        sobrescreve o arquivo se existir
```

`ctbz correspondencias endereco [flags]`: Mostra ou troca o endereço de envio das correspondências

```text
--cep string           CEP do novo endereço
--complemento string   complemento do novo endereço
--numero string        número do novo endereço
```

## documentos

`ctbz documentos [flags]`: Lista os tipos de documento aceitos e os documentos enviados

```text
--tipo string   tipo de documento (veja a coluna tipo de "ctbz documentos")
```

`ctbz documentos enviar ARQUIVO [flags]`: Envia o documento pedido por uma ou mais pendências

```text
--pendencia strings   id da pendência (repita para enviar um arquivo único para várias)
```

`ctbz documentos pendentes [flags]`: Lista os documentos pedidos pelas pendências da Central de Rotinas

```text
--pendencia strings   id da pendência (repita ou separe por vírgula)
```

`ctbz documentos sem-aplicacao [flags]`: Declara que não houve aplicação financeira numa conta nas competências pendentes

```text
--conta-bancaria int   id da conta bancária das pendências
--pendencia strings    id da pendência de extrato de aplicação (repita para várias)
```

`ctbz documentos sem-arquivo [flags]`: Declara que o documento pedido por pendências não existe

```text
--pendencia strings   id da pendência (repita para várias do mesmo tipo)
```

## empresa

`ctbz empresa [flags]`: Mostra os dados da empresa selecionada

```text
--json   atalho para -o json
```

`ctbz empresa atividades`: Lista os CNAEs da empresa e os anexos do Simples Nacional

`ctbz empresa certificado`: Mostra a situação do certificado digital da empresa e da renovação

`ctbz empresa credenciais [flags]`: Mostra os dados de acesso a órgãos públicos (código do Simples, prefeitura, Dataprev)

```text
--mostrar   mostra o código e as senhas (só num terminal)
```

`ctbz empresa credenciais atualizar [flags]`: Troca o código do Simples ou o usuário e a senha da prefeitura ou do Dataprev

```text
--codigo-simples   troca o código de acesso do Simples Nacional
--dataprev         troca o usuário e a senha do Dataprev
--prefeitura       troca o usuário e a senha da prefeitura
```

`ctbz empresa credenciais confirmar-prefeitura [flags]`: Responde "Sim, já atualizei a senha" ao alerta da senha da prefeitura

`ctbz empresa socios`: Lista os sócios da empresa

`ctbz empresa usar CNPJ`: Troca a empresa da sessão

## empresas

`ctbz empresas`: Lista as empresas do usuário, marcando a atual

## extratos

`ctbz extratos [flags]`: Lista a situação dos extratos bancários por mês e conta

```text
--ano int   só os meses deste ano
```

`ctbz extratos classificar ID [flags]`: Troca a classificação de um lançamento do extrato

```text
--competencia string   competência AAAA-MM (obrigatória)
--conta string         classificação: id ou descrição exata (ver ctbz extratos contas)
--conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
--socio string         sócio, nas classificações de sócio
```

`ctbz extratos contas [flags]`: Lista as classificações aceitas nos lançamentos do extrato de uma competência

```text
--competencia string   competência AAAA-MM (obrigatória)
--despesa              só as classificações de saída (DESPESA)
--receita              só as classificações de entrada (RECEITA)
```

`ctbz extratos desfazer-desmembramento ID [flags]`: Volta um lançamento desmembrado do extrato ao original

```text
--competencia string   competência AAAA-MM (obrigatória)
--conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
```

`ctbz extratos desmembrar ID [flags]`: Divide um lançamento do extrato em partes com classificações diferentes

```text
--competencia string   competência AAAA-MM (obrigatória)
--conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
--parte stringArray    parte "descrição:valor:conta[:sócio]" (repita para cada parte)
```

`ctbz extratos excluir [flags]`: Exclui o extrato importado de uma conta num mês

```text
--competencia string   competência AAAA-MM (obrigatória)
--conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
```

`ctbz extratos importar ARQUIVO [flags]`: Importa o extrato bancário (OFX ou PDF) de uma conta num mês

```text
--competencia string   competência AAAA-MM (obrigatória)
--conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
--saldo-final string   saldo do último dia esperado (OFX); diferente do lido, a importação para
--substituir           se o extrato do mês já foi importado, exclui-o antes e importa este arquivo (como excluir e importar no painel)
```

`ctbz extratos lancamentos [flags]`: Lista os lançamentos do extrato de uma conta bancária num mês

```text
--competencia string   competência AAAA-MM (obrigatória)
--conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
```

## impostos

`ctbz impostos [flags]`: Lista as guias de impostos a pagar (em atraso, do mês e do próximo mês)

```text
--atrasadas        mostra só as guias em atraso
--fail-on-atraso   termina com código 4 se houver guias em atraso
```

`ctbz impostos baixar [ID...] [flags]`: Baixa o PDF de guias de imposto

```text
-d, --dir string   diretório onde salvar os PDFs (default ".")
--force        sobrescreve arquivos existentes
--pendentes    baixa todas as guias a pagar
```

`ctbz impostos calculo`: Mostra como o imposto do mês foi calculado

`ctbz impostos confirmar ID [flags]`: Informa que uma guia foi paga (ou, com --nao-paguei, que não foi)

```text
--nao-paguei   informa que a guia não foi paga
```

`ctbz impostos debitos [flags]`: Indica se a empresa tem débitos federais em aberto

```text
--fail-on-debitos   termina com código 4 se houver débitos federais
```

`ctbz impostos desmarcar ID [flags]`: Desfaz a confirmação de pagamento de uma guia

`ctbz impostos faturamento`: Mostra o faturamento, o pró-labore e os impostos pagos nos últimos 12 meses

`ctbz impostos guia ID`: Mostra o detalhe de uma guia de imposto

`ctbz impostos historico [flags]`: Lista o histórico de guias de impostos

```text
--ano int         ano da competência
--mes int         mês da competência (1 a 12)
--status string   situação da guia (ex.: PAGO, PENDENTE)
```

`ctbz impostos parcelamento ID`: Mostra o detalhe de um parcelamento de impostos

`ctbz impostos parcelamento contratar TIPO [flags]`: Contrata um parcelamento de débitos

```text
--negociacao string   no tipo especializado: VENCIDOS, DIVIDA_ATIVA, DIVIDA_ATIVA_E_VENCIDOS
--parcelas int        quantidade de parcelas (tipos pgfn-*; padrão: a primeira opção da simulação)
```

`ctbz impostos parcelamento simular TIPO [flags]`: Simula um parcelamento de débitos (opções de parcelas e custos)

```text
--negociacao string   no tipo especializado: VENCIDOS, DIVIDA_ATIVA, DIVIDA_ATIVA_E_VENCIDOS
```

`ctbz impostos parcelamentos`: Lista os parcelamentos de impostos (em andamento, ativos e encerrados)

`ctbz impostos recalcular ID [flags]`: Mostra ou pede o recálculo de uma guia vencida

```text
--vencimento string   novo vencimento AAAA-MM-DD (sem ele, só mostra os dados do recálculo)
```

`ctbz impostos recorrente`: Mostra a situação do pagamento recorrente (débito automático) de impostos

`ctbz impostos recorrente historico`: Lista os pagamentos recorrentes de impostos dos últimos meses

`ctbz impostos tabela-irrf`: Mostra a tabela progressiva do IRRF usada no cálculo do pró-labore

## login

`ctbz login [flags]`: Autentica na Contabilizei (usuário, senha e código por e-mail)

```text
--cnpj string            CNPJ da empresa a selecionar (env CTBZ_CNPJ)
--otp string             código OTP recebido por e-mail (retoma um login pendente)
--otp-cmd string         comando de shell que imprime o OTP (env CTBZ_OTP_CMD)
--otp-timeout duration   tempo máximo aguardando o --otp-cmd (env CTBZ_OTP_TIMEOUT) (default 3m0s)
--restart                descarta um login pendente e recomeça
```

## logout

`ctbz logout`: Apaga a sessão local

## lucros

`ctbz lucros`: Mostra a distribuição de lucros do exercício e o que a impede

`ctbz lucros distribuir [flags]`: Registra quanto do lucro do exercício cabe a cada sócio

```text
--socio stringArray   parte de um sócio: ID=60% ou ID=30000,00 (repetível)
```

`ctbz lucros informe [flags]`: Mostra os valores do informe de rendimentos dos sócios (para o IR)

```text
--ano int   ano-calendário (padrão: o ano anterior)
```

`ctbz lucros informe aceitar carta-responsabilidade|termo-debitos [flags]`: Aceita a carta de responsabilidade ou o termo de débitos federais do informe

```text
--ano int   ano-calendário (padrão: o ano anterior)
```

`ctbz lucros informe carta [flags]`: Mostra a carta de responsabilidade do informe de rendimentos de um ano

```text
--ano int   ano-calendário (padrão: o ano anterior)
```

`ctbz lucros informe decidir [flags]`: Decide o que fazer com a pendência documental ou os débitos do informe

```text
--ano int                     ano-calendário (padrão: o ano anterior)
--nao-distribuir-lucros       não distribuir lucros
--nao-regularizar-pendencia   não regularizar a pendência documental
--regularizar-pendencia       regularizar a pendência documental
```

`ctbz lucros informe restricoes [flags]`: Mostra o que bloqueia o informe de rendimentos de um ano e o que falta aceitar

```text
--ano int   ano-calendário (padrão: o ano anterior)
```

## mensalidade

`ctbz mensalidade [flags]`: Mostra a mensalidade atual da Contabilizei (valor, vencimento e situação)

```text
--fail-on-atraso   termina com código 4 se houver competência anterior em atraso
```

`ctbz mensalidade historico`: Lista as faturas pagas da mensalidade

`ctbz mensalidade situacao [flags]`: Indica se a empresa está em dia com a Contabilizei

```text
--fail-on-inadimplencia   termina com código 4 se a empresa não estiver em dia
```

## notas

`ctbz notas [flags]`: Lista as notas fiscais de serviço (NFS-e) emitidas

```text
--ate string       último mês, AAAA-MM (padrão: --de ou o mês atual)
--de string        primeiro mês, AAAA-MM (padrão: o mês atual)
--numero string    número da nota
--tomador string   nome do tomador ou CPF/CNPJ
```

`ctbz notas aliquotas`: Lista as alíquotas e os códigos de serviço por atividade

`ctbz notas config`: Mostra a configuração do emissor de notas

`ctbz notas entrada [flags]`: Lista as notas fiscais de entrada (NF-e recebidas pela empresa)

```text
--emitente string   filtra pela razão social do emitente
--lista string      a-manifestar, manifestadas, a-classificar ou classificadas (default "a-manifestar")
--mes string        mês, AAAA-MM (padrão: o mês atual)
```

`ctbz notas entrada classificar ID... [flags]`: Classifica notas de entrada (estoque, insumo, uso e consumo…), inteiras ou por produto

```text
--como string           classifica as notas inteiras: estoque, insumo, uso-consumo, ativo-imobilizado, prestacao-servico
--mes string            mês das notas, AAAA-MM (padrão: o mês atual)
--produto stringArray   distribui um item, ITEM:opcao=qtd,opcao=qtd (repetível)
```

`ctbz notas entrada manifestar ID... [flags]`: Envia a manifestação do destinatário das notas de entrada à SEFAZ

```text
--ciencia                ciência da operação (preliminar)
--confirmar              confirmação da operação (recebi)
--desconhecer            desconhecimento da operação
--justificativa string   motivo da operação não realizada (15 a 255 caracteres)
--mes string             mês das notas, AAAA-MM (padrão: o mês atual)
--nao-realizada          operação não realizada (não recebi); exige --justificativa
```

`ctbz notas entrada produtos ID`: Lista os itens de uma nota de entrada e como estão classificados

`ctbz notas entrada reclassificar ID [flags]`: Muda a classificação de uma nota de entrada já classificada

```text
--como string           toda a quantidade numa opção: estoque, insumo, uso-consumo, ativo-imobilizado, prestacao-servico
--mes string            mês da nota, AAAA-MM (padrão: o mês atual)
--produto stringArray   distribui um item, ITEM:opcao=qtd,opcao=qtd (repetível)
```

`ctbz notas tomadores`: Lista os tomadores (clientes) cadastrados no emissor de notas

`ctbz notas tomadores adicionar [flags]`: Cadastra um tomador (cliente) nacional ou do exterior

```text
--bairro string        bairro (nacional; padrão: o do CEP)
--cep string           CEP (nacional)
--cidade string        cidade (exterior)
--complemento string   complemento
--documento string     CPF ou CNPJ do tomador nacional
--email string         e-mail
--exterior             tomador do exterior (sem documento brasileiro)
--im string            inscrição municipal (só pessoa jurídica)
--logradouro string    logradouro (padrão: o do CEP)
--nome string          razão social ou nome (CNPJ: padrão é o da Receita)
--numero string        número
--pais string          país: sigla (ex.: US), nome ou código (exterior)
--telefone string      telefone (só nacional)
```

`ctbz notas tomadores consulta CNPJ`: Consulta os dados cadastrais de um CNPJ na Receita

`ctbz notas tomadores editar DOCUMENTO|ID [flags]`: Corrige os dados de um tomador cadastrado

```text
--bairro string        bairro (nacional; padrão: o do CEP)
--cep string           CEP (nacional)
--cidade string        cidade (exterior)
--complemento string   complemento
--email string         e-mail
--im string            inscrição municipal (só pessoa jurídica)
--logradouro string    logradouro (padrão: o do CEP)
--nome string          razão social ou nome (CNPJ: padrão é o da Receita)
--numero string        número
--pais string          país: sigla (ex.: US), nome ou código (exterior)
--telefone string      telefone (só nacional)
```

`ctbz notas tomadores remover DOCUMENTO|ID [flags]`: Exclui um tomador da lista de clientes

## pendencias

`ctbz pendencias [flags]`: Lista as pendências da empresa (tipo, detalhe, prazo e situação)

```text
--fail-on-vencidas   termina com código 4 se houver pendência vencida
--todas              inclui as pendências finalizadas
```

`ctbz pendencias aceitar CHAVE [flags]`: Aceita um termo ou carta pendente da Central de Rotinas

`ctbz pendencias conciliacao [flags]`: Mostra as pendências de conciliação fiscal (notas e recebimentos)

```text
--fail-on-pendencias   termina com código 4 se houver algo a conciliar
--listar string        lista os itens pendentes: notas ou recebimentos
```

`ctbz pendencias conciliacao candidatos [flags]`: Lista as notas ou os recebimentos que podem ser vinculados numa conciliação

```text
--notas          notas fiscais para vincular a recebimentos
--recebimentos   recebimentos para vincular a notas
```

`ctbz pendencias conciliacao detalhes [flags]`: Mostra os detalhes de uma pendência ou conciliação (justificativa, vínculos)

```text
--nota strings          id de nota fiscal (repita ou separe por vírgula)
--recebimento strings   id de recebimento (repita ou separe por vírgula)
```

`ctbz pendencias conciliacao motivos`: Lista os motivos aceitos para justificar uma pendência de conciliação

`ctbz pendencias conciliacao resolver ID_PENDENCIA... [flags]`: Concilia ou justifica pendências de conciliação fiscal

```text
--motivo string          motivo da resolução (ver ctbz pendencias conciliacao motivos)
--nota-ativacao string   número da nota no fluxo de ativação contábil
--socio string           sócio, nos motivos que exigem
--vincular stringArray   contraparte ORIGEM:ID (NOTAFISCAL ou MOVIMENTACAO; repita para várias)
```

`ctbz pendencias procuracao [flags]`: Mostra a pendência de procuração do e-CAC ou declara que ela foi criada

```text
--ja-criei   declara que a procuração já foi criada no e-CAC
```

`ctbz pendencias termo CHAVE`: Mostra o texto completo de um termo ou carta da Central de Rotinas

`ctbz pendencias termos [flags]`: Lista os termos e cartas com aceite pendente na Central de Rotinas

```text
--todos   inclui os termos sem aceite pendente
```

## plano

`ctbz plano`: Mostra o plano contratado com a Contabilizei

`ctbz plano contrato [flags]`: Exporta o contrato de prestação de serviços (HTML ou texto)

```text
--texto   converte o HTML em texto simples
```

`ctbz plano proposta [flags]`: Exporta a proposta do plano contratado, com a tabela de preços (HTML ou texto)

```text
--texto   converte o HTML em texto simples
```

## primeiros-passos

`ctbz primeiros-passos`: Lista as tarefas do checklist de primeiros passos do painel

`ctbz primeiros-passos concluir ETAPA [flags]`: Marca uma tarefa dos primeiros passos como concluída

`ctbz primeiros-passos dispensar [flags]`: Oculta o card de primeiros passos do painel

`ctbz primeiros-passos reativar [flags]`: Volta a mostrar as tarefas dos primeiros passos

## prolabore

`ctbz prolabore`: Mostra o pró-labore vigente por sócio e o tipo de gerenciamento

`ctbz prolabore definir [flags]`: Muda como o pró-labore de um sócio é definido (salário mínimo, teto, valor ou automático)

```text
--minimo string   piso da gestão inteligente, em reais (opcional)
--socio int       ID do sócio (de ctbz prolabore -o json)
--tipo string     salario-minimo, teto-inss, personalizado ou inteligente
--valor string    pró-labore personalizado, em reais (3000,00)
```

`ctbz prolabore fator-r`: Mostra a situação do Fator R e os anexos possíveis de cada atividade

`ctbz prolabore gestao-inteligente`: Põe a empresa no cálculo automático do pró-labore ou a tira dele

`ctbz prolabore gestao-inteligente ativar [flags]`: Põe a empresa no cálculo automático do pró-labore

`ctbz prolabore gestao-inteligente sair [flags]`: Tira a empresa do cálculo automático do pró-labore

`ctbz prolabore historico [flags]`: Lista o histórico mensal de pró-labore dos sócios

```text
--ano int     só as competências deste ano
--socio int   ID do sócio (de "ctbz prolabore -o json"); padrão: todos
```

`ctbz prolabore parametros`: Mostra os valores usados no cálculo do pró-labore (INSS e IRRF)

`ctbz prolabore zerar-sem-faturamento on|off [flags]`: Liga ou desliga o pró-labore zerado nos meses sem faturamento

## razao

`ctbz razao [flags]`: Lista os lançamentos do razão contábil por conta

```text
--ate string     último mês, AAAA-MM (padrão: --de ou o mês atual)
--conta string   código da conta ou prefixo (ex.: 1.01)
--de string      primeiro mês, AAAA-MM (padrão: o mês atual)
```

## resumo

`ctbz resumo [flags]`: Mostra numa lista só o que precisa de atenção

```text
--fail-on-atencao   termina com código 4 se houver algo vencido ou crítico
```

## rotinas

`ctbz rotinas [flags]`: Lista as rotinas e obrigações do mês (da empresa e da Contabilizei)

```text
--fail-on-vencidas   termina com código 4 se houver rotina da empresa vencida
--mes string         mês do prazo, AAAA-MM (padrão: o mês atual)
```

`ctbz rotinas reclassificar ID_PENDENCIA... [flags]`: Confirma ou altera a classificação de lançamentos pedida pela Central de Rotinas

```text
--classificacao string   opção de classificação: id ou nome (sem ela, lista as opções)
--socio string           sócio, nas opções que exigem
```

## status

`ctbz status`: Mostra a sessão atual e testa se ainda é válida

## usuarios

`ctbz usuarios`: Lista quem tem acesso à empresa no painel (multiusuário)

`ctbz usuarios ativar ID [flags]`: Devolve o acesso de um usuário desativado

`ctbz usuarios convidar EMAIL [flags]`: Convida uma pessoa para acessar a empresa, com acesso total

`ctbz usuarios desativar ID [flags]`: Revoga o acesso de um usuário à empresa

## version

`ctbz version`: Mostra a versão do ctbz
