# Escrita: pendências, termos, atendimento, conta e usuários

Análise **estática** dos chunks do painel (`/painel-de-controle/`, Vue 2). Nenhuma requisição foi feita.
Bases (módulo axios `d5c2`): `X["c"]` = `/api/plataforma/`, `X["e"]` = `/api/multiusuario/`.
O axios junta base e caminho removendo a barra duplicada, então `"/checklist-onboarding/x"` vira
`/api/plataforma/checklist-onboarding/x`.

Convenções:

- **Risco**: baixo = só estado de UI/telemetria; médio = altera cadastro/credencial/preferência ou gera
  efeito operacional; alto = efeito fiscal, legal (aceite de termo/declaração) ou financeiro (contratação, cobrança).
- "(incerto)" = inferido do código, sem confirmação no front.
- Os 403 com a mensagem *"Não é possível realizar a assinatura como admin!"* aparecem quando a sessão é de
  um administrador da Contabilizei personificando o cliente; para o usuário dono da sessão a chamada passa.

---

## 1. Pendências, rotinas e termos

### 1.1 Central de Rotinas: `central-rotinas/*`

Origem: `app.*.js` (composable `useCentralDeRotinas`, módulo de pendências) + chunk `CentralDeRotinas`
(componentes `SectionPendencias`/`SectionRotinas`). Tela `#/central-de-rotinas` (rota `CentralDeRotinas`).

**Pré-condição comum**: `GET /api/plataforma/central-rotinas/init` devolve
`{ pendencias: { pendenciasCriticas: {...}, outrasPendencias: {...} }, rotinas: [...], rotinasCtbz: [...], rotinasCalendario: [...] }`.
Cada pendência é um objeto com `possuiPendencia: boolean` e campos próprios. Chaves vistas no front:
`pendenciaCartaResponsabilidade`, `pendenciaProcuracaoEcac`, `pendenciaAceiteTermoDebitos`,
`pendenciaTermoAdesaoTotalPass`, `pendenciaCredencialPrefeitura`, `pendenciaCertificadoDigital`,
`pendenciaConciliacaoFiscal`, `pendenciaContratoPrestacaoServico`, `pendenciaImposto`, `pendenciaMensalidade`,
`pendenciaCadastroContaBancaria`, `pendenciaCadastroProlabore`, `pendenciaImportacaoExtrato`,
`pendenciaIntegracaoAExpirar`/`Expirada`, `pendenciaExigibilidadeDocumental`, `pendenciaContratoAFAC`,
`pendenciaContratoDeEmprestimo`, `pendenciaContratoDeFinanciamento`, `pendenciaContratoDeInvestimentoAnjo`,
`pendenciaAquisicaoAtivoImobilizado`, `pendenciaEstoque`, `pendenciaControleDeIntermediacoes`,
`pendenciaExtratoAplicacaoFinanceira`, `pendenciaInformeDeRendimentoInvestimentos`.
Só as quatro abaixo têm endpoint de resolução próprio. As demais apenas navegam para outra tela
(envio de documentos, impostos, conta bancária, dados de acesso etc.).

Os quatro POSTs **não têm corpo nem parâmetros**: o servidor deduz empresa e período pela sessão.
A resposta é lida (`.data`), mas o front não usa o conteúdo. Ao dar certo, o front marca
`possuiPendencia = false` e mostra o toast "Pendência resolvida".

#### `POST /api/plataforma/central-rotinas/aceitar-carta-responsabilidade`

- Corpo: nenhum.
- Fluxo: card "Aceite da carta de responsabilidade pendente", botão "Ver carta", que abre o
  `ModalCartaDeResponsabilidade` com o HTML de `pendenciaCartaResponsabilidade.conteudoCartaResponsabilidade`.
  Botões: "Regularizar mais tarde" (sem chamada) e **"Aceitar as declarações"** (dispara o POST).
- Semântica: aceite da *Carta de Responsabilidade da Administração*, a declaração anual exigida pelo CFC de que
  o cliente entregou todas as informações e documentos à contabilidade.
- Erros: 403 dá o toast "Não é possível realizar a assinatura como admin!". Outros erros são relançados.
- Reversível: **não** há endpoint de revogação.
- Risco: **alto** (declaração legal/contábil assinada).

#### `POST /api/plataforma/central-rotinas/aceitar-termo-debitos`

- Corpo: nenhum.
- Fluxo: card crítico "Aceite do Termo de Ciência e Responsabilidade" (template scroll com
  `bodyText = pendenciaAceiteTermoDebitos.conteudoAceiteTermoDebito`), botão **"Estou ciente do termo"**.
- Texto: "O documento explica os impactos das retiradas mensais de lucros a partir de 2026". Consequência:
  "Caso a gente não receba seu aceite no termo em até `{prazoAceiteTacito}` dias, vamos entender que não há
  impedimentos para as retiradas de lucros e a administração e/ou os sócios estão cientes dos riscos envolvidos,
  como fiscalização pela Receita Federal, multas, juros e autuações." Ou seja, existe **aceite tácito** por prazo.
- Erros: 403 dá "Não é possível realizar a assinatura como admin!". Os demais dão "Erro inesperado ao aceitar o
  termo. Por favor tente novamente.", e o card volta para pendente.
- Reversível: não.
- Risco: **alto** (ciência de risco fiscal sobre distribuição de lucros e débitos federais).
- Relacionado: `informerendimento/aceitar-termo-debitos/{ano}` (1.2), a versão por ano-calendário no informe
  de rendimentos.

#### `POST /api/plataforma/central-rotinas/resolver-pendencia-procuracao-ecac`

- Corpo: nenhum.
- Fluxo: card "Emissão de procuração eletrônica pendente" (`pendenciaProcuracaoEcac`, com
  `tipoPendencia` ∈ `SEM_PROCURACAO` | `EM_EXPIRACAO` | outro valor tratado como "vencida", e `cnpjOutorgado`).
  Botões: "Portal de serviços da receita" (só abre o e-CAC) e **"Já criei a procuração"** (dispara o POST).
  No caso `EM_EXPIRACAO`, o card mostra "Renovar procuração" e abre o link da Receita.
- Semântica: o cliente **declara** que já outorgou ou renovou a procuração eletrônica no e-CAC para a
  Contabilizei (CNPJ em `cnpjOutorgado`). O servidor provavelmente revalida depois (incerto).
- Erros: 403 dá "Não é possível realizar a assinatura como admin!".
- Reversível: não há endpoint. Se a procuração não existir, a pendência deve reaparecer no próximo `init` (incerto).
- Risco: **médio** (é só uma declaração; o efeito legal está na procuração feita no e-CAC, fora da plataforma).
- Equivalente no checklist de onboarding: `POST /checklist-onboarding/procuracoes/compartilhamentos` (1.5).

#### `POST /api/plataforma/central-rotinas/aceitar-termo-aceite-tp`

- Corpo: nenhum.
- Fluxo: pendência "outras" `pendenciaTermoAdesaoTotalPass` (card "Termo de adesão ao programa TotalPass",
  botão "Ler termo") ou rotina `TermoAdesaoTotalPass`. Abre o `ModalTermoAdesaoTotalPass` com
  `conteudoTermoAdesaoTotalPass`. Botões: "Ler depois" (sem chamada) e **"Estou ciente do termo"** (dispara o POST).
- Texto: "Sem a assinatura, você não terá acesso ao TotalPass e ao Starbem. Além disso, a falta de aceite pode
  levar ao cancelamento do seu plano de benefícios."
- Resposta: o front devolve `{status:true}` quando dá certo. Erros: 403 dá "Não é possível realizar a assinatura
  como admin!". Os demais dão "Erro inesperado ao aceitar o termo de adesão. Por favor tente novamente."
- Reversível: não.
- Risco: **médio/alto** (adesão contratual a benefício; efeito financeiro indireto no plano de benefícios).

> Pela URL: `#/central-de-rotinas?pendencia=<chave>` abre o modal da pendência "outras" indicada, se ela existir.

### 1.2 Informe de rendimentos: `informerendimento/*`

Chunk `informe-de-rendimentos`. Tela `#/socio/comprovante-rendimentos` (rota `informe-de-rendimentos`). O parâmetro de caminho `{ano}` é
`filtro.anoCompetencia`, o **ano-calendário** do informe (lista de 2013 ou do início da responsabilidade
até o ano anterior ao atual).

#### `POST /api/plataforma/informerendimento/aceitar-carta-responsabilidade/{ano}`

- Corpo: nenhum.
- Pré-condição: `GET informerendimento/carta-responsabilidade/{ano}` devolve
  `{ deveAssinarCartaResponsabilidade: bool, html: string }`. Também existe
  `GET informerendimento/visualizar-carta-responsabilidade/{ano}`.
- Fluxo: com `deveAssinarCartaResponsabilidade`, aparece a checkbox **"Aceito os termos da Carta de
  Responsabilidade"**, obrigatória para pré-visualizar ou baixar o informe ("Para visualizar o informe você precisa
  aceitar os termos da Carta de Responsabilidade."). Ao visualizar ou baixar, o front chama o POST e depois gera o PDF.
- Reversível: não. Risco: **alto** (mesma declaração da 1.1, mas por ano).

#### `POST /api/plataforma/informerendimento/aceitar-termo-debitos/{ano}`

- Corpo: `{ "aceite": "PRIMEIRA_VEZ" | "NEGOU_DA_PRIMEIRA_VEZ" }`. É uma **string**: identifica em qual
  apresentação do termo o aceite aconteceu, não é booleano.
  - `PRIMEIRA_VEZ`: o modal abriu automaticamente na primeira visita.
  - `NEGOU_DA_PRIMEIRA_VEZ`: o usuário tinha clicado "Ler depois" e reabriu o termo pelo link.
- Pré-condição: `GET informerendimento/v2/{ano}/restricoes` devolve
  `{ restricoes: { debitosFederais: { possuiPendencia, debitos[], divergenciaContabilFiscal },
  pendenciaDocumental: { possuiPendencia, fluxoRegularizacao, pendencias[], valorServicoAdicional } },
  processoReabertura: { status } }`. Com `debitosFederais.possuiPendencia`, abre o `ModalDebitosFederais` e, em
  seguida, o modal do termo.
- Botões: **"Estou ciente do termo"** (este POST, depois recarrega a distribuição) e "Ler depois". Na primeira vez,
  "Ler depois" chama `informerendimento/aceite/{ano}` com `IR_NAO_DISTRIBUIR_LUCROS`, ver abaixo.
- Erro: "Erro ao processar a ação. Aguarde um pouco e tente novamente."
- Reversível: não. Risco: **alto**.

#### `POST /api/plataforma/informerendimento/aceite/{ano}`

- Corpo: `{ "tipoAceite": "<enum>" }`. Valores encontrados:
  - `IR_NAO_DISTRIBUIR_LUCROS`: enviado por "Ler depois" no modal de termo de débitos (primeira vez). O efeito é
    **não distribuir lucros** no informe daquele ano.
  - `NAO_REGULARIZAR_PENDENCIA_DOCUMENTAL`: botão "Não regularizar pendência" nos modais
    `ModalExigibilidadeDocumental`/`ModalPendenciasDocumentais`. Texto: "Caso opte por 'Não regularizar pendência',
    não haverá distribuição de lucros."
  - `REGULARIZAR_PENDENCIA_DOCUMENTAL`: botão "Regularizar pendência". Depois do POST, o front abre o formulário
    Zendesk `https://suporte.contabilizei.com.br/hc/pt-br/requests/new?ticket_form_id=360000562139`.
- Erros: 400 com `detalhes[0].identificador == "exception/erro-negocial-001"` mostra `detalhes[0].detalhe`. Fora isso,
  mostra "Erro ao tentar salvar aceite."
- Reversível: não há endpoint explícito (incerto se um novo `aceite` sobrescreve o anterior).
- Risco: **alto** (decide se haverá distribuição de lucros isenta no informe de rendimentos).

Adjacentes no mesmo serviço, fora deste escopo mas com escrita:

- `POST informerendimento/salvarconfiguracaocliente`: corpo `[{idSocio, percentual, valor}]`, distribuição de lucro
  por sócio.
- `POST informerendimento/historico`: corpo `{tipoEvento: "VISUALIZOU_IR"|"BAIXOU_IR", anoExercicio}`, telemetria.
- `POST informerendimento/reabrir-balanco/{ano}`: sem corpo. **Risco alto**: o texto diz que o serviço "gera uma
  cobrança de R$ …". Um 403 mostra "Usuário admin não pode fazer a reabertura do exercício contábil para um cliente."

### 1.3 Dashboard: `POST /api/plataforma/dashboard/v2/termo-exclusao`

- Chunk `chunk-b88c6b8a` (dashboard v2), componente `ModalTermoDeExclusao2023`.
- Pré-condição: o init do dashboard devolve `modalInitDashboard == "TERMO_EXCLUSAO_2023"`, e então o modal abre
  sozinho.
- Corpo: `{ "contratarServico": boolean }`
  - `false`: botão **"Estou ciente, não quero regularizar"**. Em seguida, mostra o modal "Regularize suas pendências".
  - `true`: botão **"Solicitar verificação de pendências"**. Em seguida, mostra o modal "Serviço contratado".
- Texto: "Sua empresa será excluída do regime do Simples Nacional… A Receita Federal notificou sua empresa com o
  TERMO DE EXCLUSÃO DO SIMPLES NACIONAL… a não regularização irá desenquadrar sua empresa do Simples Nacional,
  podendo aumentar significativamente o valor dos seus impostos e também acrescenta o valor de R$95 à sua
  mensalidade." Os prazos são de 2023/2024: campanha antiga, provavelmente inativa.
- Resposta: ignorada. Reversível: não.
- Risco: **alto** (`true` contrata um serviço; `false` registra ciência de exclusão do Simples).

### 1.4 Conciliação fiscal (receitas × recebimentos)

Serviço `04aa` em `app.*.js` (v1) e cópias no chunk `conciliacao-v2`. Telas `ConciliacaoFiscal` (v1) e
conciliação v2 (Vuex `conciliacaoFiscalV2`).

#### `POST /api/plataforma/conciliacao-fiscal/conciliar/resolver-pendencia`

- Pré-condição: `GET conciliacao-fiscal/pendencias?…` (v1) ou `GET conciliacao-fiscal/v2/pendencias?…` (v2) para os
  ids das pendências. Contrapartes vêm de `GET conciliacao-fiscal/conciliar/receitas` (notas) ou
  `…/conciliar/recebimentos` (movimentações). Sócios para `idVinculo`: getter `sociosDisponiveisParaVinculo` (incerto
  sobre o GET de origem).
- Corpo (três formas usadas):

```json
// a) conciliar: vincular pendência(s) a contraparte(s)
{ "idPendencia": [123, 124],
  "contraparte": [ { "origem": "NOTAFISCAL" | "MOVIMENTACAO", "id": 987 } ],
  "tipoResolucaoPendencia": null }

// b) reclassificar / justificar sem contraparte
{ "idPendencia": [123],
  "contraparte": [],
  "tipoResolucaoPendencia": "REEMBOLSO",
  "idVinculo": 55,                       // opcional: id do sócio quando o motivo exige sócio
  "numeroNotaAtivacaoContabil": "1234" } // opcional: fluxo "nota anterior/ativação contábil"

// c) conciliar com diferença justificada: contraparte preenchida + tipoResolucaoPendencia não nulo
```

- `origem`: quando a pendência é de **recebimento**, as contrapartes são notas (`NOTAFISCAL`). Quando é de **nota**,
  são movimentações (`MOVIMENTACAO`).
- `tipoResolucaoPendencia` (enum encontrado): `ADIANTAMENTO_EMISSAO_NOTA_FISCAL_FUTURA`, `AFAC`, `ATIVACAO_CONTABIL`,
  `CAIXA_OU_PESSOA_FISICA`, `CASHBACK`, `DESCONTO_CONCEDIDO`, `DEVOLUCAO_DE_SAQUE_DE_LUCROS_ANTECIPADOS`,
  `DIFERENCA_REPASSAR`, `EMPRESTIMO_BANCARIO`, `EMPRESTIMO_DO_SOCIO_A_EMPRESA`, `ESTORNO_TRANSACAO_BANCARIA`,
  `INTEGRALIZACAO_CAPITAL`, `INVESTIMENTO_ANJO_MUTUO`, `JUROS_RECEBIDOS`, `NOTA_FISCAL_FUTURA`, `PERDA`,
  `RECEBIMENTO_DE_EMPRESTIMOS_CONCEDIDOS`, `RECEBIMENTO_DE_INTERMEDIACAO_DE_SERVICOS_A_SEREM_REPASSADOS`,
  `RECEBIMENTO_DE_VALORES_A_SEREM_REPASSADOS_AO_CLIENTE`, `RECEBIMENTO_PARCELADO`, `REEMBOLSO`,
  `REEMBOLSO_DE_CLIENTES_ADMINISTRACAO_DE_IMOVEIS`, `REEMBOLSO_DE_PAGAMENTO_ANTECIPADO_DE_DESPESAS_PARA_CLIENTES`,
  `REEMBOLSO_DE_TAXAS_E_CUSTAS_DE_PROCESSOS_JUDICIAIS_DE_CLIENTES`, `TAXAS_INTERMEDIADORAS`,
  `TRANSFERENCIA_ENTRE_CONTAS`, `VARIACAO_CAMBIAL_NEGATIVA`, `VARIACAO_CAMBIAL_POSITIVA`.
  Exemplos de rótulos na UI: "Recebi em dinheiro" / "Recebi em uma conta PF" → `CAIXA_OU_PESSOA_FISICA`;
  "Não vou mais receber por essa nota" → `PERDA`; "Valor adiantado para virar capital social" → `AFAC`.
  O motivo "Outro" não chama a API: manda para o atendimento.
- Restrição de UI: algumas opções só valem se cada item for menor que R$ 1.000,00. Também aparece
  "Você já usou essa opção para esta pendência."
- Resposta: ignorada. Depois do POST, o front recarrega as pendências e mostra "Pendência conciliada com sucesso." /
  "O recebimento foi c…". Erros: "Falha ao conciliar pendência. Tente novamente." /
  "Falha ao reclassificar recebimento. Tente novamente."
- Reversível: a v2 tem o fluxo "alteração" (`iniciarAlteracao`/`abrirModalConfirmarAlteracao`), que **reenvia o mesmo
  endpoint** sobre uma conciliação já feita (histórico em `selecionarConciliacaoHistorico`). Não há DELETE.
- Risco: **alto**. A classificação define a natureza contábil e fiscal do recebimento (receita tributável ou não,
  empréstimo, capital etc.).

#### `PATCH /api/plataforma/conciliacao-fiscal/conciliar/alerta-receitas/fechar`

- Corpo: nenhum (chamada sem argumento). Fecha o banner/alerta de receitas no drawer de conciliação.
- Risco: **baixo** (estado de UI). Sem endpoint de desfazer.

#### `POST /api/plataforma/conciliacao-fiscal/pendencia/detalhes` (leitura via POST)

- Corpo: `{ "idsRecebimento": [..], "idsReceita": [..] }` (a v1 manda `idsReceita = idsNotaFiscal` do item).
- Devolve os detalhes da pendência ou conciliação (justificativa/reclassificação). Campos da resposta: (incerto).
- **Sem efeito colateral**: pode ser tratado como GET na CLI. Risco: nenhum.

### 1.5 Checklist de onboarding: `checklist-onboarding/*`

Módulo `bd38` no chunk `chunk-b88c6b8a` (dashboard v2: `ChecklistAside`, card "Primeiros passos").

| Método e caminho (`/api/plataforma/…`) | Corpo | Semântica | Risco |
|---|---|---|---|
| `GET checklist-onboarding/aside/init` | | `{etapas:[{nome, situacao}], progresso}` | – |
| `GET checklist-onboarding/aside/trilha/{trilha}` | | `trilha` ∈ `configure-conta`, `aprenda-basico`; devolve `{etapas, progresso}` | – |
| `POST checklist-onboarding/aside/concluir-etapa` | `{"etapa": "<nome>"}` | marca uma tarefa como concluída | baixo |
| `POST checklist-onboarding/procuracoes/compartilhamentos` | nenhum | botão "Já criei a procuração" na etapa `GERAR_PROCURACAO_VIRTUAL`; declara a procuração e avança a etapa | médio |
| `POST checklist-onboarding/pular-tarefas` | nenhum | fechar o card do checklist (v1) = pular as tarefas restantes | baixo |
| `PATCH checklist-onboarding/exibir-primeiros-passos/dispensar` | nenhum | dispensa ou oculta o card "Primeiros passos" | baixo |
| `PATCH checklist-onboarding/reativar-tarefas` | nenhum | botão **"Desfazer"** depois de dispensar: desfaz o `dispensar` | baixo |
| `POST checklist-onboarding/tooltip/guia-inicio/concluir` | nenhum | conclui o tour guiado | baixo |

- Valores de `etapa` (`nome`): `CADASTRO_CONTA_PJ`, `CADASTRE_PRO_LABORE`, `GERAR_PROCURACAO_VIRTUAL`,
  `ACEITE_TERMOS_MULTIBENEFICIOS`, `REUNIAO_BOAS_VINDAS`, `LIVE_EMISSAO_NOTAS_FISCAIS`, `LIVE_VIDA_COM_CNPJ`.
  Nem toda etapa tem "concluir" manual: os botões são `marcarConcluido`/`concluirTarefa`/`agendarEConcluir`.
- Resposta **304** em `concluir-etapa`/`tooltip/.../concluir` é tratada como "já estava concluído" (não é erro).
  Outros erros: "Ocorreu um erro ao concluir a tarefa. Tente novamente mais tarde."
- Reversibilidade: `dispensar` ↔ `reativar-tarefas`. Não há endpoint para "desconcluir" uma etapa.

### 1.6 Dashboard: notificações "primeiros passos"

#### `POST /api/plataforma/dashboard/interacao-notificacao`

- Chunk `chunk-2371324a` (dashboard v1). Pré-condição: `GET dashboard/primeiros-passos` devolve
  `[{tipo, visto}]`.
- Corpo: `{ "tipo": "DUPLO_VINCULO" }` (único tipo mapeado no front: pop-up "Sabia que você pode estar pagando um
  valor de INSS maior que o necessário?", botão "Saber mais").
- Semântica: marca a notificação como vista (`visto = true`) e abre um artigo de suporte.
- Risco: **baixo**. Não há como desfazer.

Outras escritas de interação ou telemetria no dashboard (todas `/api/plataforma/`, risco baixo):
`home/nps` `{nota: 0..10, comentario}` (NPS); `upsell/banner/clicar`; `retencao/paywall/salvar-interacao` `{}`;
`dashboard/v2/notas-fiscais/salvar-clique-emitir`; `evento-tour` (base legado `X["d"]`, corpo `{flow, evento}`)
(incerto); `hubspot/enviar` (lead/CRM).

### 1.7 Ativação: `ativacao/timeline/*`

Chunk `ativacao` (timeline de abertura de empresa, várias versões A–G do mesmo componente).
Pré-condição: `GET /api/plataforma/ativacao/timeline` (steps, `mfaAtivo`, `hasCertificadoDigital`,
`contaDigital`, `regimeTributario`, `inscricaoMunicipal`…).

| Método e caminho | Corpo | Semântica | Risco |
|---|---|---|---|
| `POST ativacao/timeline/salvar-interacao` | `{"step": "CONTA_DIGITAL"\|"CERTIFICADO_DIGITAL"\|"PLANO_SAUDE", "label": "<texto livre>"}`; labels vistos: `"ver step"`, `"aceite"`, `"emitir e-cnpj"` e rótulos de botão | registra a interação num passo (colapsa ou expande o step) | baixo |
| `POST ativacao/timeline/finalizar` | nenhum ou `{}` | botão "Acessar plataforma": encerra a timeline (`auth/updateTimelineFinalizada`). Resposta `{rolloutOnboarding: bool}` | baixo/médio (sem volta) |
| `POST ativacao/timeline/finish-step-plano-saude` | `{"nomeBotao": "QUERO_SABER_MAIS"\|"AGORA_NAO"}` | conclui o step de plano de saúde | baixo |
| `POST ativacao/timeline/aceite/mfa-gov` | `{}` | declaração "Declaro que desativei a verificação em duas etapas [gov.br] e estou ciente que: 1. A verificação impede a Contabilizei de concluir a abertura e entregar as obrigações, o que pode gerar multas…" (botão "Continuar") | médio (declaração) |

Erro genérico: "Não foi possível prosseguir! Tente novamente." Nenhum desses endpoints tem desfazer.
A timeline também chama `fintech/conta-digital/aceite?origem=timeline` e `fintech/conta-digital/notificacao-conclusao`
(escopo fintech).

**Recomendação para a CLI (contexto 1)**

- `ctbz pendencias listar` (GET `central-rotinas/init`): mostra as chaves com `possuiPendencia`, o tipo e o texto.
- `ctbz pendencias ver <chave>`: imprime o conteúdo do termo ou da carta (`conteudo*`) para o usuário ler **antes**.
- `ctbz pendencias aceitar carta|termo-debitos|totalpass` e `ctbz pendencias procuracao-ecac --ja-criei`: exigir
  confirmação interativa e `--yes` explícito, exibir o texto do termo e avisar que é irreversível.
- `ctbz ir aceite --ano 2025 --tipo NAO_REGULARIZAR_PENDENCIA_DOCUMENTAL|...` e
  `ctbz ir termo-debitos --ano 2025`: só com confirmação dupla, porque decidem a distribuição de lucros.
- `ctbz conciliacao resolver <idPendencia...> --contraparte NOTAFISCAL:987 | --motivo REEMBOLSO [--socio id]`:
  útil, mas é a escrita de maior impacto fiscal desta página. Sugere-se `--dry-run` padrão que mostra o JSON, e
  `ctbz conciliacao detalhes` (POST de leitura).
- `ctbz onboarding concluir <ETAPA>`, `ctbz onboarding dispensar|reativar`: baixo valor; opcionais.
- Não expor: `termo-exclusao` (campanha 2023), `interacao-notificacao`, `salvar-interacao`, `finalizar` da
  timeline e o restante da telemetria.

---

## 2. Chamados de atendimento (abrir ou responder)

**Conclusão: o painel não tem endpoint de API para abrir chamado nem para responder a um.**

- `GET /api/plataforma/atendimento/chamados?em-andamento=true` e `…?finalizados=true` (tela `#/chamados`,
  rota `MeusChamados`, chunk `chunk-49028f4c`) são **somente leitura**. Itens:
  `{id, assunto, canal, status, atualizado, previsaoRetorno}`. Cada card é um **link para o Zendesk Help Center**:
  `https://suporte.contabilizei.com.br/hc/pt-br/requests/{id}`, e é lá que o cliente "adiciona mais informações".
  A resposta acontece no Zendesk, com outra sessão e autenticação, e não pela API da Contabilizei.
- "Fale conosco" / "Ajuda" abre um iframe do app separado `/widget-window/#/helpdesk` (CRM widget, que conversa por
  `postMessage` com eventos `helpdesk/open`, `chat/open`, `whatsapp/open`, `zendesk/open`). As chamadas desse widget
  não estão nos bundles analisados. O canal WhatsApp é `https://wa.me/<numero>?text=Olá, eu quero falar sobre o
  tema … e assunto …`.
- Formulário Zendesk direto: `https://suporte.contabilizei.com.br/hc/pt-br/requests/new?ticket_form_id=360000562139`
  (aberto pelo informe de rendimentos).
- `GET zendesk/artigos?filtro=` só busca artigos da central de ajuda.

Escritas que se parecem com "pedir atendimento" (efeito: o time entra em contato):

| Endpoint | Corpo | Contexto | Risco |
|---|---|---|---|
| `POST /api/plataforma/empresa-inativa/registrar-solicitacao` | `{"tipoSolicitacao": "AJUDA"}` | tela `#/central-inativos` (empresa sem acesso: mensalidade atrasada ou empresa encerrada), botão "Solicitar atendimento" → "Recebemos o seu pedido!". Pré-condição: `GET /empresa-inativa/permite-registrar-solicitacao` → `{permiteRegistrarSolicitacao, tipoSolicitacao}` (se `false`, já existe pedido). 403 → "Usuário administrador não tem permissão para executar essa ação." | baixo |
| `POST certificado-digital/v2/solicitar-atendimento-pendencia` | nenhum | fluxo de certificado digital (outra página) | baixo |
| `POST fintech/conta-digital/lead-atendimento` | nenhum | conta digital (fintech) | baixo |

**Recomendação para a CLI (contexto 2)**: `ctbz chamados listar [--finalizados]` (leitura) imprimindo o link
`…/hc/pt-br/requests/{id}`, e `ctbz chamados abrir` apenas abrindo o navegador no formulário Zendesk ou no
`/widget-window/`. Uma integração real exigiria a API do Zendesk (`/api/v2/requests`) com autenticação própria,
fora do escopo da sessão Contabilizei. `empresa-inativa/registrar-solicitacao` só faz sentido para empresa inativa.

---

## 3. Conta, usuários e empresa

### 3.1 Dados da conta do usuário: `conta-usuario/*`

Chunk `DadosDaContaDoUsuario`, tela `#/conta-do-usuario` "Dados da minha conta" ("Alterações no e-mail, senha e telefone impactam seu
acesso à plataforma Contabilizei."). Todas as escritas exigem um **código OTP** de 6 dígitos.

Pré-condição: `GET /api/plataforma/conta-usuario/init` → `{ email, telefone, metodo }`. Os valores vêm
mascarados com `*`, e `metodo` é o método atual de 2FA: `EMAIL` | `SMS` | `APP`.

#### `POST /api/plataforma/conta-usuario/enviar-token-otp`

- Corpo: `{ "metodoEnvio": "EMAIL" | "SMS" }`. Com `APP`, o front não chama: o código vem do app autenticador.
- Resposta: `{ "tempo": <segundos> }`, usado no contador "Reenviar por …".
- **422** com `data.error` numérico = rate-limit. O número é o tempo restante e o modal abre mesmo assim.
  Outro erro: "Não foi possível enviar o código. Tente novamente mais tarde."
- Risco: **baixo** (envia e-mail ou SMS ao titular).

#### `POST /api/plataforma/conta-usuario/alterar-dados`

- Corpo: `{ "email": string, "telefone": "11987654321", "codigo": "123456", "metodo": "EMAIL"|"SMS"|"APP" }`.
  `telefone` sai do formulário `(DD)NNNNN-NNNN` sem `()-`. `metodo` é o mesmo usado no OTP.
- Validação no front: e-mail por regex; telefone `^\(\d{2}\)\d{4,5}-\d{4,5}$` com 14 caracteres.
- Resposta: 2xx → recarrega o `init` e mostra o modal "As alterações foram salvas!". **404 = código incorreto**
  ("O código está incorreto ou inválido"). Outros erros → "As alterações não foram salvas!".
- Reversível: só alterando de novo. Risco: **médio/alto** (troca o e-mail de login e o telefone do 2FA; pode
  causar perda de acesso).

#### `POST /api/plataforma/conta-usuario/alterar-senha`

- Corpo: `{ "senha": "<nova>", "codigo": "123456", "metodo": "EMAIL"|"SMS"|"APP" }`. Não envia a senha atual:
  a prova é o OTP.
- Regras no front: no mínimo 8 caracteres, minúsculas e maiúsculas, números e caractere especial; a confirmação
  precisa ser igual.
- Respostas: iguais às de `alterar-dados` (404 = código inválido).
- Risco: **alto** (credencial). Pode invalidar outras sessões (incerto).

#### `POST /api/plataforma/conta-usuario/validar-codigo-app`

- Pré-condição: `GET conta-usuario/qr-code-aplicativo-autenticacao` → QR code (imagem/otpauth; formato incerto).
- Corpo: `{ "codigo": "123456", "metodo": "APP" }`.
- Semântica: sincroniza o app autenticador (TOTP) e passa a usá-lo como método de 2FA.
  Sucesso: "O aplicativo de autenticação foi sincronizado!". 404 = código inválido.
- Risco: **médio/alto** (muda o segundo fator). Para voltar a EMAIL/SMS não há endpoint visível (incerto).

### 3.2 Multiusuário: `/api/multiusuario/`

Módulo `9649` (`app.*.js`), telas `#/multiusuario` (lista) e `#/convite-usuario`.

- `GET /api/multiusuario/usuarios-e-invites/consultar` → `[{id, nome, email, tipo: "USUARIO_PRINCIPAL"|"USUARIO_SECUNDARIO", status: "ATIVO"|"INATIVO"|"CONVITE_ENVIADO"}]`.
- `GET /api/multiusuario/status/servico` → `{ativo: bool, limite: number}`. O front bloqueia o convite se
  `!ativo` ("O serviço está indisponivel no momento.") ou se os secundários ativos ≥ `limite` ("Você já cadastrou o
  número máximo de N usuários."). A CLI conta da mesma forma e, se algum item vier com `tipo` ou `status` vazio ou
  fora dos valores acima, recusa o convite sem escrita, porque não dá para conferir o limite.

#### `POST /api/multiusuario/invites/enviar`

- Corpo: `{ "email": "<convidado>" }`. O front recusa e-mails `@contabilizei.`.
- Texto: "Ao convidar uma pessoa com acesso total, ela terá permissão para realizar todas as ações na plataforma e
  visualizar todas as informações da empresa, inclusive financeiras." Depois: "O convite foi enviado para o novo
  usuário…". Erro: "Não foi possível enviar o convite! Tente novamente mais tarde."
- Resposta: sem conteúdo usado.
- Reversível: **não há endpoint para revogar ou reenviar convite** no front. Só é possível desativar o usuário
  depois que ele aceitar (`ativar` com `false`), e isso vale para convites pendentes também (incerto).
- Risco: **alto** (concede acesso total à empresa a terceiros).

#### `PUT /api/multiusuario/usuarios/ativar`

- Corpo: `{ "idUsuarioEmpresa": <id da lista>, "ativo": true|false }`.
- Bloqueio no front: `tipo == "USUARIO_PRINCIPAL"` → "Não é possível desativar o administrador da empresa."
- Mensagens: "O usuário foi ativado com sucesso." / "O usuário foi desativado com sucesso." / "Ação não realizada."
- Reversível: **sim** (alternar `ativo`). Esta é a única forma de "revogar" acesso; não existe DELETE.
- Risco: **médio**.

### 3.3 Dados de acesso a órgãos públicos: `empresa/dadosacesso/*`

Chunk `dados-de-acesso`, tela `#/dados-de-acesso` (aba "Dados de acesso").

- Pré-condição: `GET /api/plataforma/empresa/dadosacesso/init` →
  `{ formulario: { chaveAcessoSimples, usuarioPrefeitura, senhaPrefeitura, usuarioDataprev, senhaDataprev, dataValidadeSenhaPrefeituraCuritiba }, alertaPendenciaCredencial: bool, … }`.
  **Atenção: o GET devolve as credenciais em texto claro.** A CLI nunca deve logá-las. Com 403 (admin), o front
  mostra "Edição bloqueada" (`ADMIN_SEM_AUTORIZACAO_AOS_DADOS`).

#### `POST /api/plataforma/empresa/dadosacesso/atualizarDadosAcesso`

- Corpo:

```json
{ "id": <empresa.id da sessão (getLoginData.empresa.id)>,
  "chaveAcessoSimples": "123456789012",   // 12 dígitos; obrigatório se empresa.optanteSimples
  "usuarioPrefeitura": "...", "senhaPrefeitura": "...",   // mínimo de 3 caracteres cada, se preenchidos
  "usuarioDataprev": "...", "senhaDataprev": "...",       // só para quem retirou pró-labore antes de 10/2021
  "dataValidadeSenhaPrefeituraCuritiba": null }
```

- O formulário inteiro é reenviado: os campos não alterados vão com o valor do `init`.
- Resposta: **204** = sucesso ("Dados de acesso salvos com sucesso"). Qualquer outro status → "Ocorreu um erro
  inesperado. Verifique as informações no formulário e tente novamente". Erro de validação local: "Código de acesso
  deve possuir 12 caracteres apenas numéricos".
- Depois de salvar, o servidor verifica o acesso: "Verificação de acesso iniciada… Caso seja negado, enviaremos um
  e-mail". Status mostrados no chip: Pendente, Verificando Acesso, Vinculada, Acesso Negado, Atualização Necessária.
- Reversível: só regravando. Risco: **alto** (credenciais usadas para transmitir declarações; um erro gera
  "multas e juros").

#### `PUT /api/plataforma/empresa/dadosacesso/confirmar-credencial-prefeitura-valida`

- Corpo: nenhum.
- Fluxo: alerta "Redefina a senha da prefeitura e atualize os dados de acesso" (`alertaPendenciaCredencial`), link
  "clique aqui", modal "Você atualizou a senha de acesso à prefeitura e garante que o acesso está correto?", botão
  **"Sim, já atualizei a senha"**.
- Semântica: declara que a credencial da prefeitura atual é válida e limpa o alerta.
  Também resolve a pendência crítica `CREDENCIAL_NAO_PREENCHIDA` da Central de Rotinas (incerto).
- Risco: **médio** (declaração; se for falsa, as obrigações municipais falham).

### 3.4 Escritório virtual (correspondências): `escritorio-virtual/*`

Chunk `escritorio-virtual`, tela `#/escritorio-virtual` ("Correspondências"; abas "Disponíveis" e "Configurações
de envio").

Leituras: `GET escritorio-virtual/endereco-entrega`, `GET escritorio-virtual/recuperar-mensagens?{cursor,…}`
(paginação `meta.cursor`/`meta.count`), `GET escritorio-virtual/verifica-autorizacao-recebimento` → `{autoriza}`,
`GET escritorio-virtual/pesquisarEndereco/{cep}` (CEP com 8 dígitos), e
`GET escritorio-virtual/gerar-link-download-correspondencia/{id}` ou `…/gerar-link-download-comprovante/{id}` →
`{url}` (link assinado, sem efeito colateral conhecido).

#### `PUT /api/plataforma/escritorio-virtual/autorizar-recebimento/{true|false}`

- Valor no caminho: o booleano do switch "Autorizo o recebimento de correspondências e documentos entregues no
  endereço informado."
- Corpo: nenhum. Mensagens: "Autorização de recebimento alterada!" / "Ocorreu um erro ao alterar a autorização de
  recebimento."
- Reversível: **sim** (enviar o valor oposto). Risco: **médio**. A lista mostra "Valor do envio" (`taxa`), então o
  envio de correspondência pode ser cobrado (incerto).

#### `POST /api/plataforma/escritorio-virtual/salvar-endereco-correspondencia`

- Corpo: `{ "cep": "01310100", "numero": "100", "complemento": "sala 1" }`. Logradouro, bairro, município e UF vêm
  do `pesquisarEndereco/{cep}` e não são enviados.
- Resposta: o novo `infoEndereco` (mesmo formato de `endereco-entrega`). Mensagens: "Endereço alterado!" /
  "Ocorreu um erro ao alterar o endereço." Aviso: "As alterações de endereço são apenas para o envio das
  correspondências."
- Reversível: sim, regravando o endereço anterior. Risco: **médio** (para onde vão documentos oficiais).

### 3.5 Outros achados neste contexto

- O catálogo não deixou de fora nenhuma escrita de conta, usuários ou empresa: todas as chamadas desses módulos
  usam caminho literal. A busca por caminhos em variáveis só encontrou chamadas do emissor de notas, da integração
  bancária e de `notafiscaltomada`, fora deste escopo.
- Não existe endpoint de logout no servidor, nem de exclusão de conta ou usuário (`DELETE`).

**Recomendação para a CLI (contexto 3)**

- `ctbz conta ver` (GET `init`).
- `ctbz conta otp enviar --via email|sms`.
- `ctbz conta alterar --email X --telefone Y --codigo N` e `ctbz conta senha --codigo N` (lendo a senha por prompt
  oculto, nunca por argumento). As duas devem ser interativas: primeiro enviam o OTP, depois pedem o código.
- `ctbz conta 2fa app` (mostra o QR no terminal e chama `validar-codigo-app`): opcional e de risco.
- `ctbz usuarios listar`, `ctbz usuarios convidar <email> --yes`, `ctbz usuarios ativar|desativar <id>` (o par
  ativar/desativar é reversível e seguro).
- `ctbz empresa credenciais ver` (mascarar senhas por padrão; `--mostrar` explícito),
  `ctbz empresa credenciais atualizar --simples 123456789012 [--prefeitura-usuario …]` (com merge do `init`: o POST
  exige o formulário completo) e `ctbz empresa credenciais confirmar-prefeitura`.
- `ctbz correspondencias listar|baixar <id>`, `ctbz correspondencias endereco --cep … --numero …` e
  `ctbz correspondencias autorizar on|off`.
- Ordem sugerida de implementação: primeiro as leituras; depois as escritas reversíveis (usuários ativar/desativar,
  autorização de correspondência, endereço); por último as declarações e termos irreversíveis, sempre com confirmação.

---

## Resumo de risco e reversibilidade

| Endpoint | Risco | Desfazer |
|---|---|---|
| central-rotinas/aceitar-carta-responsabilidade, informerendimento/aceitar-carta-responsabilidade/{ano} | alto | não |
| central-rotinas/aceitar-termo-debitos, informerendimento/aceitar-termo-debitos/{ano} | alto | não |
| informerendimento/aceite/{ano} | alto | não (incerto) |
| conciliacao-fiscal/conciliar/resolver-pendencia | alto | refazer pelo fluxo "alteração" (mesmo endpoint) |
| dashboard/v2/termo-exclusao | alto | não |
| empresa/dadosacesso/atualizarDadosAcesso | alto | regravar |
| conta-usuario/alterar-senha, alterar-dados, validar-codigo-app | alto/médio | alterar de novo |
| invites/enviar | alto | só desativando o usuário |
| central-rotinas/aceitar-termo-aceite-tp | médio/alto | não |
| central-rotinas/resolver-pendencia-procuracao-ecac, checklist procuracoes/compartilhamentos | médio | não |
| empresa/dadosacesso/confirmar-credencial-prefeitura-valida, ativacao/timeline/aceite/mfa-gov | médio | não |
| usuarios/ativar, escritorio-virtual/* | médio | sim |
| checklist-onboarding/* (exceto procuração), interacao-notificacao, timeline/salvar-interacao, alerta-receitas/fechar | baixo | dispensar ↔ reativar; o resto não |
| conciliacao-fiscal/pendencia/detalhes | nenhum (leitura) | – |
