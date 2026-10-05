# Changelog

Todas as mudanças relevantes deste projeto são documentadas aqui.

O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e o projeto
usa [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Added

- `ctbz caixa contas` e `ctbz extratos contas`: classificações aceitas numa competência, com o
  filtro do painel e o vínculo exigido (guia de imposto ou sócio); `--vinculos` lista as guias
  e os sócios do caixa.

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

[Unreleased]: https://github.com/edusouza/ctbz-cli/compare/v1.1.0...HEAD
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
