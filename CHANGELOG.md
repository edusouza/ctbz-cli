# Changelog

Todas as mudanças relevantes deste projeto são documentadas aqui.

O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e o projeto
usa [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Fixed

- `ctbz impostos parcelamento contratar` não contrata mais quando a simulação traz uma opção sem a
  quantidade de parcelas (antes enviava `{"quantidadeParcelas":0}`) ou a opção escolhida sem o
  valor das parcelas (antes mostrava R$ 0,00 na confirmação): pede para contratar pelo painel.
  Custos adicionais sem valor aparecem como "não informado" (#489).

### Security

- `ctbz api` com uma URL absoluta de outro host enviava para lá os cookies da sessão. Agora o
  cliente HTTP recusa toda requisição cujo esquema, host ou porta difira do painel (ou com
  usuário e senha antes do host), e o `ctbz api` recusa a URL antes de qualquer envio, com
  código de saída 2; URLs do próprio painel continuam aceitas (#941).

## [1.9.2] - 2026-10-26

O que travou o extrato do Banco Inter no uso real: a CLI recusava o OFX, que o painel importa,
e `extratos excluir` informava a exclusão de um mês preso em `PROCESSANDO`.

### Fixed

- `extratos importar` aceita OFX mesmo quando o banco não o lista entre os formatos de upload,
  como a tela "Importar extrato" do painel, com um aviso no stderr. Antes, a CLI recusava o OFX do
  Banco Inter, que lista só PDF, XLS e CSV, e o painel o importava (#1011).
- `extratos excluir` não informa mais "excluido" quando a competência estava e continua
  `PROCESSANDO` depois do envio: a Contabilizei aceita a exclusão sem mudar nada nesse estado. A
  confirmação avisa do risco antes, e o comando termina com erro (código 1), orientando importar
  pelo painel ou abrir um chamado. Um extrato `IMPORTADO` que passa a `PROCESSANDO` depois do
  envio continua informado como excluído, com aviso para esperar e conferir (#1012, #1022).

## [1.9.0] - 2026-10-23

O que faltou no primeiro uso real: gerar a segunda via da mensalidade atrasada com Pix copia
e cola e linha digitável, importar extratos em CSV e XLS pela Central de Documentos e
diagnosticar uma chamada que falha, com o modo verbose e o envio `multipart/form-data` no
`ctbz api`.

### Added

- Modo verbose: `--verbose` (em qualquer comando) ou `CTBZ_DEBUG=1` mostra no stderr cada
  requisição, com método, URL, status, tempo, tamanho e cabeçalhos enviados e recebidos.
  `--verbose --verbose` ou `CTBZ_DEBUG=2` mostra também os corpos, truncados. Saem sempre
  como `***`: valores de cookies, cabeçalhos e campos de credencial (senha, token, código,
  chave, sessão, `Authorization`, JWT), dados de pagamento das respostas (Pix, linha digitável,
  código de barras, chave do gateway), UUIDs, JWTs e CPFs/CNPJs (inclusive o alfanumérico) no
  caminho da URL. No nível 1, os valores da query também (#902, #909, #911, #935, #936).
- `ctbz api -F campo=valor -F campo=@arquivo[;type=tipo]`: envia `multipart/form-data`, como o
  `curl -F`. Assim dá para fazer pela ferramenta de baixo nível os uploads que nenhum comando
  cobre. Sem `-X`, envia com `POST`; não combina com `-d`. Cada arquivo tem até 30 MB, e o
  prazo do envio é o dos uploads (#900).
- `ctbz acoes` ganha a coluna `arquivos`: nos envios de arquivo (`ctbz api -F`,
  `extratos importar`, `documentos enviar`…), o nome e o tamanho de cada arquivo, nunca o
  conteúdo (#900).
- `ctbz mensalidade segunda-via`: gera a segunda via da mensalidade atrasada, como o botão
  "Gerar segunda via s/ juros" do card, e mostra itens, total, vencimento, situação, Pix copia
  e cola, linha digitável, código de barras e o link do checkout. `--pix` e `--linha` imprimem
  só o código, para pipe. O caminho vem da ação do card, lida a cada vez. Risco médio, com
  confirmação e `--dry-run`. O identificador da fatura não vai para o registro de ações (#851).
- `ctbz mensalidade` mostra os botões do card e as ações deles: `botao_titulo`, `botao_acao`,
  `botao_adicional_titulo` e `botao_adicional_acao` (#851).
- `ctbz extratos importar` aceita os formatos que o banco da conta declara além de OFX e PDF
  (CSV, XLS…). Como a tela "Importar extrato" do painel só aceita OFX e PDF, esses arquivos vão
  pela Central de Documentos (`POST upload-documentos/extrato/enviar`), num envio só, e a
  Contabilizei os importa manualmente, em até 45 dias. O resultado sai com
  `situacao: "enviado"` e a chave nova `importacao` (`automatica` ou `manual`). Até a
  importação, a competência continua `PENDENTE`: a CLI avisa para não reenviar e conferir na
  Central de Documentos. Esses formatos não aceitam `--substituir` e são recusados sobre um
  extrato já importado (#899, #920, #921).
- Documentado o que a plataforma exige numa chamada: o cookie `oauth-token` e um `User-Agent`
  de navegador. Com o `User-Agent` do `curl`, a resposta é 401 (#902).

### Changed

- `ctbz extratos importar` confere o formato do arquivo pela lista do banco, numa mensagem só:
  "o banco desta conta aceita extratos em PDF, XLS ou CSV; este arquivo é .ofx". Antes, a
  extensão era conferida contra uma lista fixa (`.ofx`/`.pdf`) que contradizia a do banco (#899).

### Fixed

- `extratos importar --substituir`: quando a exclusão dá certo e a importação falha, a mensagem
  vem da competência relida, e não de uma suposição. Antes dizia "voltou a ficar pendente" e
  mandava rodar de novo, mas a competência já foi vista `PROCESSANDO` por minutos depois de um
  arquivo recusado. Agora, com `PROCESSANDO`, a CLI manda esperar e conferir; com `IMPORTADO`,
  conferir os lançamentos; e só manda repetir quando a competência está pendente (#901).

## [1.8.1] - 2026-10-23

Correções do primeiro uso real de ponta a ponta: lançamentos do extrato, cadastro de conta
bancária, importação de extrato, histórico da mensalidade e `ctbz api` no Git Bash do Windows.
Traz também o plugin do Claude Code com as skills.

### Added

- Plugin do Claude Code (`.claude-plugin/`, `skills/`) com skills que fazem um agente operar a
  Contabilizei como um assessor contábil: `assessor` (base: login, protocolo de escrita e
  dossiê da empresa), `checkup`, `fechamento-mensal`, `impostos`, `planejamento-tributario`
  (com simulador de Fator R), `fechamento-anual`, `notas-e-clientes` e `cadastro-e-acessos`.
- `scripts/gerar-comandos-skill.py`: gera o mapa de comandos das skills a partir de
  `docs/referencia`.
- READMEs das skills (`skills/README.md` e um por skill): instalação, configuração do login,
  exemplos de uso, o que cada skill pode alterar e com que risco, limites e solução de problemas.

### Deprecated

- Em `ctbz mensalidade historico`, os campos `debito_automatico_ativo`,
  `debito_automatico_habilitado`, `status`, `competencia` e `proximo_pagamento` descrevem o débito
  automático de impostos (`ctbz impostos recorrente`), não a mensalidade. Continuam na saída por
  compatibilidade.

### Fixed

- `ctbz rotinas`, `documentos pendentes` e as demais leituras da Central de Rotinas não
  falham mais quando a Contabilizei responde HTTP 206: as leituras aceitam 2xx, como o painel,
  menos o 204, que continua querendo dizer "não encontrado".
- `ctbz impostos faturamento` funciona em empresas com menos de um ano: os meses anteriores à
  abertura saem sem impostos pagos, com um aviso no stderr, em vez de o comando falhar com
  "Ano inválido".
- `ctbz resumo` não lista mais como vencida uma guia já paga.
- `ctbz extratos lancamentos` (e `classificar`, `desmembrar` e `desfazer-desmembramento`)
  funciona com a resposta real da Contabilizei, um array de lançamentos sem total. A paginação
  tem limite e dá erro numa página repetida, em vez de truncar a lista. A lista ganha a
  situação e a classificação sugerida, com a probabilidade; mês sem lançamentos avisa no stderr
  e sai com 0 (#852).
- `ctbz contas-bancarias adicionar --banco 077` escolhe o Banco Inter S.A. (código exato), e não
  o Banco Inter - Internacional (077-1). `--banco` usa o id, depois o código exato e só então o
  código sem sufixo; código de mais de um banco dá erro com a lista de candidatos. O resumo e o
  `--dry-run` mostram o código do banco (#853).
- `ctbz extratos importar`: se a conexão cair ou o servidor responder com erro 5xx (502, 504…)
  no envio do arquivo, a CLI relê a competência e diz se é seguro repetir (continua `PENDENTE`:
  nada foi importado) ou se é preciso conferir antes. A queda ou o 5xx ao concluir a importação
  não diz mais "não foi concluída", e sim para conferir; "não foi concluída" fica para as
  recusas (4xx e regra de negócio). O envio do arquivo tem prazo de 3 minutos, em vez de 60
  segundos (#854, #594, #875).
- `ctbz extratos importar` recusa competência já `IMPORTADO` ou `PROCESSANDO`, como o painel, e
  explica quando o extrato importado está vazio: o processamento de um mês pode criar o mês
  seguinte sem lançamentos. Depois de importar, avisa se a competência terminou sem
  lançamentos (`quantidade_lancamentos` no JSON). `--substituir` refaz a importação de um mês
  já importado: exclui o extrato atual e importa o arquivo, numa confirmação só (risco alto se
  o extrato tem lançamentos). A recusa "não reconheceu o arquivo" sugere
  alternativas (#855).
- `ctbz mensalidade historico` lista as faturas pagas, da mesma fonte da tela "Consultar
  histórico" do painel, com vencimento, multa e juros e forma de pagamento. Antes lia o histórico
  do débito automático de impostos e saía vazio para quem paga por boleto ou Pix (#856).
- `ctbz api` aceita caminho absoluto começando com `//` (`//api/legado/…`), a forma que o Git
  Bash do Windows não converte, e recusa com explicação um caminho que chegou convertido pelo
  MSYS (`C:/Program Files/Git/api/…`), em vez de mandar a requisição e receber 404 (#857).

## [1.8.0] - 2026-10-04

Conta, usuários e acessos: quem acessa a empresa, dados de acesso a órgãos públicos, e-mail,
telefone e senha de login, e o escritório virtual.

### Added

- `ctbz usuarios`, `convidar`, `desativar` e `ativar`: quem acessa a empresa no painel
  (multiusuário), com convite de acesso total e revogação por desativação.
- `ctbz empresa credenciais`, `atualizar` e `confirmar-prefeitura`: dados de acesso a órgãos
  públicos (código do Simples, prefeitura e Dataprev), com segredos mascarados e lidos sem eco.
- `ctbz conta`, `alterar` e `senha`: e-mail, telefone e senha de login, confirmados com código
  OTP lido do `--otp-cmd` ou do terminal.
- `ctbz correspondencias`, `baixar`, `endereco` e `autorizar`: escritório virtual
  (correspondências, endereço de envio e autorização de recebimento).

### Changed

- O `--dry-run` também mascara o código de acesso do Simples (`chaveAcesso*`).

## [1.7.0] - 2026-10-04

Pró-labore, distribuição de lucros e informe de rendimentos: mudar a gestão do pró-labore,
zerá-lo nos meses sem faturamento, entrar e sair da gestão inteligente, distribuir lucros, dar
os aceites do informe e reabrir o balanço.

### Added

- `ctbz prolabore definir`: muda a gestão do pró-labore de um sócio (salário mínimo, teto do
  INSS, valor personalizado ou gestão inteligente com piso opcional).
- `ctbz prolabore zerar-sem-faturamento on|off`: preferência de não ter pró-labore nos meses
  sem faturamento.
- `ctbz prolabore gestao-inteligente ativar|sair`: põe a empresa no cálculo automático do
  pró-labore ou a tira dele.
- `ctbz prolabore` mostra o `id` de cada sócio, usado em `--socio`.
- `ctbz lucros distribuir`: registra a distribuição de lucros do exercício, em percentual ou
  em reais, com a soma conferida em centavos.
- `ctbz lucros` mostra o `id` e o percentual de cada sócio; o `id` é o que `--socio` pede.
- `ctbz lucros informe restricoes` e `carta`: o que bloqueia o informe de rendimentos de um ano
  e o texto da carta de responsabilidade.
- `ctbz lucros informe aceitar carta-responsabilidade|termo-debitos` e `decidir`: aceites e
  decisões que liberam o informe (regularizar ou não a pendência documental, não distribuir
  lucros).
- `ctbz balanco reabrir`: reabre o exercício contábil de um ano para regularizar a pendência
  documental do informe de rendimentos.

## [1.6.0] - 2026-10-04

Notas fiscais: cadastrar, editar e excluir tomadores; manifestar e classificar as notas de
entrada.

### Added

- `ctbz notas tomadores adicionar` e `editar`: cadastro de tomadores nacionais (CNPJ preenchido
  pela Receita, CPF validado) e do exterior. `ctbz notas tomadores` mostra o `id`.
- `ctbz notas tomadores remover`: exclui um tomador da lista de clientes.
- `ctbz notas entrada manifestar`: manifestação do destinatário (ciência, confirmação,
  desconhecimento ou operação não realizada) das notas de entrada.
- `ctbz notas entrada classificar --como`: classifica notas de entrada inteiras (estoque,
  insumo, uso e consumo, ativo imobilizado ou prestação de serviço).
- `ctbz notas entrada classificar --produto`: classificação item a item, com as regras do
  painel (sem negativos, soma igual à quantidade total).
- `ctbz notas entrada reclassificar`: muda a classificação de uma nota já classificada.
- `ctbz notas entrada produtos`: itens de uma nota de entrada e a distribuição entre as
  classificações.

### Changed

- `ctbz notas entrada` ganha a coluna `classificacao`, no fim.

## [1.5.0] - 2026-10-04

Extratos, contas bancárias e documentos: cadastrar contas, importar e excluir extratos,
enviar documentos e declarar a ausência deles.

### Added

- `ctbz contas-bancarias bancos`, `adicionar` e `editar`: bancos aceitos, cadastro de conta PJ
  (com a declaração do painel) e correção dos dados de uma conta.
- `ctbz contas-bancarias remover`: exclui uma conta bancária (risco alto), quando o painel permite.
- `ctbz extratos importar`: importa o extrato do mês (OFX ou PDF) com as validações do painel e
  a confirmação do saldo do último dia (ou `--saldo-final`).
- `ctbz extratos excluir`: exclui o extrato importado de um mês (risco alto), quando o painel permite.
- `ctbz documentos pendentes` e `ctbz documentos enviar`: documentos pedidos pelas pendências e
  o envio do arquivo (um por pendência ou único para várias), com tipo e metadados da pendência.
- `ctbz documentos sem-arquivo` e `ctbz documentos sem-aplicacao`: declaram que não há estoque,
  intermediações, contrato de AFAC ou aplicação financeira, fechando as pendências.

## [1.4.0] - 2026-10-04

Pendências: ler e aceitar termos, declarar a procuração do e-CAC, resolver a conciliação fiscal
e concluir os primeiros passos.

### Added

- `ctbz pendencias termos` e `ctbz pendencias termo CHAVE`: termos e cartas com aceite pendente
  na Central de Rotinas (com o prazo do aceite tácito) e o texto completo.
- `ctbz pendencias aceitar CHAVE`: aceita a carta de responsabilidade, o termo de débitos ou o
  termo do TotalPass, mostrando o texto completo antes da confirmação de risco alto.
- `ctbz pendencias procuracao [--ja-criei]`: situação da procuração do e-CAC (CNPJ a outorgar e
  link) e a declaração de que ela foi criada.
- `ctbz pendencias conciliacao resolver`: concilia pendências vinculando notas ou recebimentos,
  ou justificando com um motivo (com `motivos`, `candidatos` e `detalhes` para preparar).
- `ctbz primeiros-passos`: tarefas do checklist do painel, com `concluir ETAPA`, `dispensar` e
  `reativar`.

### Fixed

- Textos convertidos de HTML (contrato, termos) não levam mais caracteres de controle ao terminal.

## [1.3.0] - 2026-10-04

Impostos: confirmar ou negar o pagamento de guias, pedir recálculo e simular ou contratar
parcelamentos.

### Added

- `ctbz impostos confirmar ID [--nao-paguei]` e `ctbz impostos desmarcar ID`: informam que uma
  guia foi ou não paga (rota v5 ou v3 conforme a tela da empresa), com o aviso de auditoria do
  painel.
- `ctbz impostos recalcular ID [--vencimento AAAA-MM-DD]`: mostra os dados do recálculo de uma
  guia vencida ou pede o recálculo (risco alto: serviço cobrado na mensalidade).
- `ctbz impostos parcelamento simular TIPO`: opções de parcelas e custos adicionais de um
  parcelamento de débitos (Simples, PGFN ou especializado); sem débitos, avisa e sai com 0.
- `ctbz impostos parcelamento contratar TIPO`: contrata o parcelamento depois de simular, com os
  avisos do painel (confissão de dívida, cancelamento, Selic) e os custos na mensalidade.

## [1.2.0] - 2026-10-04

Primeiras escritas: caixa e classificação de lançamentos do extrato, com confirmação,
`--dry-run` e releitura do resultado.

### Added

- `ctbz caixa` mostra o id de cada lançamento (coluna `id`, no fim).
- `ctbz caixa contas` e `ctbz extratos contas`: classificações aceitas numa competência, com o
  filtro do painel e o vínculo exigido (guia de imposto ou sócio); `--vinculos` lista as guias
  e os sócios do caixa.
- `ctbz caixa adicionar`: registra um recebimento ou pagamento no caixa de uma competência,
  com as validações do painel, confirmação e `--dry-run`.
- `ctbz caixa editar`: altera data, valor, tipo, classificação, vínculo ou descrição de um
  lançamento manual do caixa, mostrando o antes e o depois.
- `ctbz caixa remover`: exclui um lançamento manual do caixa e mostra como recriá-lo.
- `ctbz extratos lancamentos`: lançamentos do extrato de uma conta num mês, com a classificação
  atual; `ctbz extratos classificar` troca a classificação de um deles.
- `ctbz extratos desmembrar` divide um lançamento do extrato em partes com classificações
  diferentes (com as regras do painel), e `ctbz extratos desfazer-desmembramento` desfaz.
- `ctbz rotinas` mostra os ids das pendências de cada rotina (coluna `pendencias`), e
  `ctbz rotinas reclassificar` confirma ou altera a classificação pedida e conclui a rotina.

## [1.1.0] - 2026-10-04

Fundação das ações que alteram dados na Contabilizei
([ADR-0018](https://github.com/edusouza/ctbz-cli/blob/main/docs/adr/0018-escrita-com-confirmacao.md)).
Os comandos de escrita de cada contexto chegam a partir da 1.2.

### Added

- `ctbz mensalidade historico`: pagamentos anteriores da mensalidade e situação do débito automático.
- Base dos comandos que alteram dados: resumo e confirmação por risco (`confirmo` no risco
  alto), `--yes` para scripts e `--dry-run` para ver a requisição sem enviar
  ([guia](https://github.com/edusouza/ctbz-cli/blob/main/docs/guia/escrita.md)).
- `ctbz acoes`: registro local das escritas enviadas (`$CTBZ_HOME/acoes.jsonl`, sem corpos),
  com `--desde` e `--limite`.

### Changed

- O monitoramento semanal passa a acompanhar as escritas: o catálogo de endpoints inclui
  chamadas `axios({method, url})`, sufixos concatenados, caminhos em variáveis e o front de
  notas de entrada, e o relatório aponta escritas usadas pela CLI que sumiram do front.
- `ctbz api -X` com método diferente de `GET` confere a sessão antes e envia uma única vez:
  não refaz o login nem repete a chamada depois do envio, e avisa quando uma falha de
  conexão deixa o resultado incerto.

## [1.0.0] - 2026-10-03

Primeira versão estável: comandos, flags, chaves de JSON/CSV e códigos de saída passam a
seguir o SemVer ([ADR-0017](https://github.com/edusouza/ctbz-cli/blob/main/docs/adr/0017-contrato-publico-da-1-0.md)).

### Added

- Binários para Linux, macOS e Windows (amd64 e arm64) na página de releases, gerados a
  cada versão.

## [0.8.0] - 2026-10-03

### Added

- `ctbz documentos`: tipos da central de documentos com quantidade enviada; `--tipo` lista os
  arquivos enviados (competência, envio, descrição, valor e link).
- `ctbz certificado`: o mesmo que `ctbz empresa certificado`, que agora mostra também o
  alerta de vencimento do painel e a etapa da compra ou renovação do certificado.

## [0.7.0] - 2026-10-03

### Added

- `ctbz balancete AAAA-MM`: contas com saldo anterior, débitos, créditos e saldo, recuadas
  por nível na tabela e planas (com coluna `nivel`) em CSV e JSON.
- `ctbz balanco AAAA[-MM]`: ativo, passivo e patrimônio líquido com saldo do exercício e do
  exercício anterior (só o ano: dezembro).
- `ctbz razao`: lançamentos por conta no período (`--de`/`--ate`, `--conta` por código ou
  prefixo), com contrapartida e saldo acumulado.
- `ctbz caixa AAAA-MM`: lançamentos do caixa com conta de classificação, tipo, valor e
  situação; total do mês no stderr.
- `ctbz extratos`: situação do extrato por conta bancária e mês (`--ano`);
  `ctbz contas-bancarias`: banco, agência, conta, saldo inicial e integração.
- `ctbz contas`: plano de contas usado nas classificações, com conta contábil, classificação
  e situação; `--busca` e `--situacao`.

## [0.6.0] - 2026-10-03

### Added

- `ctbz prolabore`: tipo de gerenciamento, total, competências do painel e, por sócio, valor,
  responsável na Receita, gestão, atualização e dependentes; `historico` lista pró-labore e
  descontos por competência (`--ano`, `--socio`).
- `ctbz prolabore parametros`: salário mínimo, alíquota e teto do INSS e o valor a partir
  do qual incide IRRF.
- `ctbz prolabore fator-r`: motor do Fator R, percentual atual, pró-labore e faturamento de 12
  meses e os anexos do Simples possíveis para cada atividade.
- `ctbz lucros`: saldo, total distribuído, adiantamentos, limite, prazo e valor por sócio do
  exercício aberto, com as restrições do informe (pendência documental, débitos federais,
  reabertura do balanço).
- `ctbz lucros informe`: valores do comprovante de rendimentos de cada sócio (rendimentos,
  INSS, IRRF, 13º, lucros isentos) para a declaração de IR.

## [0.5.0] - 2026-10-03

### Added

- `ctbz notas`: NFS-e emitidas no período (`--de`/`--ate`, até 24 meses) com número, emissão,
  tomador, documento, valor, status e situação; filtros `--tomador` (nome ou CPF/CNPJ) e
  `--numero`; total do período no stderr.
- `ctbz notas tomadores`: tomadores cadastrados no emissor; `consulta CNPJ` mostra o cadastro
  de um CNPJ na Receita (situação cadastral, atividade, Simples, endereço).
- `ctbz notas config`: emissor habilitado, versão, instabilidade reportada, certificado e
  município; `ctbz notas aliquotas`: alíquota, ISS, Fator R e item de serviço por atividade,
  para tomadores no Brasil e no exterior.
- `ctbz notas entrada`: NF-e recebidas pela empresa (a manifestar, manifestadas, a
  classificar e classificadas), com `--mes` e `--emitente`.

## [0.4.0] - 2026-10-03

### Added

- `ctbz mensalidade`: competência, valor, vencimento, situação e observações da mensalidade
  atual, e se há competência anterior em atraso (`--fail-on-atraso`, código 4).
- `ctbz impostos recorrente`: situação do pagamento recorrente (débito automático) de impostos,
  sem expor dados de cartão; `historico` lista as guias pagas e o custo de operação por mês.
- `ctbz mensalidade situacao`: indica se a empresa está em dia com a Contabilizei
  (`--fail-on-inadimplencia`, código 4).
- `ctbz plano`: plano contratado (descrição, categoria, valor de tabela e ramos);
  `contrato` e `proposta` exportam o contrato de serviço e a proposta do plano em HTML ou,
  com `--texto`, em texto simples.

## [0.3.0] - 2026-10-03

### Added

- `ctbz pendencias`: pendências da empresa com tipo, detalhe, criação, prazo e situação;
  coluna `alerta` (`vencida`/`próxima`), `--todas` e `--fail-on-vencidas` (código 4).
- `ctbz rotinas`: rotinas do mês da empresa e obrigações da Contabilizei, com prazo, status,
  valor e alerta; `--mes AAAA-MM` e `--fail-on-vencidas`.
- `ctbz pendencias conciliacao`: notas sem recebimento e recebimentos sem nota, com a
  competência de referência; `--listar notas|recebimentos` e `--fail-on-pendencias`.
- `ctbz chamados [--finalizados]`: chamados de atendimento com link para a central de ajuda.
  Quando o servidor não consegue listar os finalizados, usa os 100 chamados mais recentes
  de `dadosempresa/get` e avisa no stderr.
- `ctbz resumo`: numa lista só, impostos em atraso e do mês, pendências, rotinas do mês,
  mensalidade e pendências críticas do painel, com consultas em paralelo e
  `--fail-on-atencao` (código 4) para alertas em cron.

## [0.2.0] - 2026-10-03

### Added

- `ctbz impostos`: guias a pagar em atraso, do mês e do próximo mês, com totais por grupo;
  `--atrasadas` e `--fail-on-atraso` (código de saída 4 quando há atraso).
- `ctbz impostos guia ID`: detalhe de uma guia (valores, juros e multa, situação, ações).
- `ctbz impostos calculo` e `ctbz impostos tabela-irrf`: memória de cálculo do mês e tabela
  do IRRF.
- `ctbz impostos baixar`: PDFs de guias (por ID ou `--pendentes`) com nomes padronizados,
  sem sobrescrever arquivos existentes (`--force`).
- `ctbz impostos historico`: guias anteriores com valor pago e situação, filtros por ano,
  mês e situação.
- `ctbz impostos faturamento`: faturamento, pró-labore e impostos pagos nos últimos 12 meses,
  com RBT12 e Fator R.
- `ctbz impostos parcelamentos`, `ctbz impostos parcelamento ID` e `ctbz impostos debitos`
  (`--fail-on-debitos`).
- Guia de uso por contexto em `docs/guia/`.

## [0.1.0] - 2026-10-03

### Added

- `ctbz login`: autenticação com usuário/senha e código OTP enviado por e-mail, obtido por
  comando externo (`--otp-cmd`/`CTBZ_OTP_CMD`), pelo terminal ou em duas etapas (`--otp`).
- Seleção de empresa no login (`--cnpj`/`CTBZ_CNPJ`) e re-login automático quando a sessão expira.
- `ctbz status`, `ctbz empresa`, `ctbz api` e `ctbz logout`.
- Formatos de saída `table`, `json` e `csv` (`-o`/`--output`, `CTBZ_OUTPUT`).
- Script `scripts/otp-gmail-gws.sh` para ler o OTP do Gmail com o Google Workspace CLI,
  com mensagem clara (código 127) quando `gws` ou `jq` não estão instalados.
- Documentação da engenharia reversa, ADRs e roadmap.
- `ctbz version` (e `--version`), com versão, commit e data do build.
- Ajuda em português e referência de comandos gerada em `docs/referencia/`.
- Códigos de saída: 0 (sucesso), 1 (erro), 2 (uso incorreto) e 3 (login pendente).
- Completion de shell: `ctbz completion bash|zsh|fish|powershell`.
- `ctbz empresas`: lista as empresas do usuário (inclusive inativas), marcando a atual.
- `ctbz empresa usar CNPJ`: troca a empresa da sessão refazendo o login.
- `ctbz empresa` mostra também natureza jurídica, data de abertura, início na Contabilizei,
  inscrição estadual e endereço.
- `ctbz empresa certificado`: situação, vencimento e dias para vencer do certificado digital.
- `ctbz empresa socios`: sócios com CPF, papel (administrador, responsável na Receita),
  categoria, salário-base e data de entrada.
- `ctbz empresa atividades`: CNAEs da empresa, a principal e os anexos do Simples Nacional.

[Unreleased]: https://github.com/edusouza/ctbz-cli/compare/v1.9.2...HEAD
[1.9.2]: https://github.com/edusouza/ctbz-cli/compare/v1.9.0...v1.9.2
[1.9.0]: https://github.com/edusouza/ctbz-cli/compare/v1.8.1...v1.9.0
[1.8.1]: https://github.com/edusouza/ctbz-cli/compare/v1.8.0...v1.8.1
[1.8.0]: https://github.com/edusouza/ctbz-cli/compare/v1.7.0...v1.8.0
[1.7.0]: https://github.com/edusouza/ctbz-cli/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/edusouza/ctbz-cli/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/edusouza/ctbz-cli/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/edusouza/ctbz-cli/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/edusouza/ctbz-cli/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/edusouza/ctbz-cli/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/edusouza/ctbz-cli/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/edusouza/ctbz-cli/compare/v0.8.0...v1.0.0
[0.8.0]: https://github.com/edusouza/ctbz-cli/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/edusouza/ctbz-cli/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/edusouza/ctbz-cli/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/edusouza/ctbz-cli/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/edusouza/ctbz-cli/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/edusouza/ctbz-cli/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/edusouza/ctbz-cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/edusouza/ctbz-cli/releases/tag/v0.1.0
