# Escrita: impostos e pagamentos

> Análise **estática** dos chunks JS do painel (`/painel-de-controle/`, versão 2.0.88).
> Nenhuma requisição de rede foi feita. Itens não confirmados no código estão marcados **(incerto)**.
> Sem dados pessoais. Chaves/IDs públicos embutidos no bundle (Iugu account id, client keys) foram omitidos de propósito.

Convenções:

- Base `X["c"]` = `/api/plataforma/` (todas as chamadas abaixo usam essa base, salvo indicação).
  O módulo axios é o `d5c2` (exports `a`–`f`, ver `docs/frontend/README.md`).
- `X["f"]` = `/api/fintech/`.
- Ids de guia (`{idGuia}`) vêm de `GET impostos/v5/impostos-a-pagar/guias` (campo `id`, mais `origem` e `tipo`),
  de `GET impostos/v3/impostos-a-pagar/init` (`impostos[]`) ou de `GET impostos/v2/historico-impostos/guias`.
- Risco: **baixo** (só estado de UI / preferências), **médio** (estado fiscal informativo, reversível),
  **alto** (efeito financeiro, contratual ou fiscal difícil de desfazer).

---

## 1. Impostos

### 1.1 Confirmar / negar pagamento de guia ("Já paguei" / "Não paguei" / "Desmarcar")

Existem três versões do mesmo endpoint, uma por geração de tela. Todas são `PUT` com o mesmo body.

| Versão | Método + caminho | Chunk / tela |
|---|---|---|
| v5 (atual, rollout `v5`) | `PUT /api/plataforma/impostos/v5/impostos-a-pagar/guia/{idGuia}/confirmar-pagamento` | `app.*.js` (componente `DrawerDetalhesImposto`, usado em `impostosv5`, `pagamentosFintech`, histórico v5) |
| v3 (tela antiga "Impostos a pagar", ainda carregada se rollout ≠ v5) | `PUT /api/plataforma/impostos/v3/impostos-a-pagar/guia/{idGuia}/confirmar-pagamento` | `impostos.*.js` (abas "Este mês" / "Em atraso"), módulo `f344` compartilhado por `impostos-hub`, `historico-impostos-hub`, `impostos-parcelamentos` |
| v2 histórico (tela antiga de histórico) | `PUT /api/plataforma/impostos/v2/historico-impostos/guia/{idGuia}/confirmar-pagamento` | `historico-impostos.*.js` |

Body (idêntico nas três):

```json
{
  "tipo": "GUIA",              // string — campo `tipo` da guia na listagem (ex.: "GUIA", "PARCELA")
  "origem": "GUIAS",           // string — campo `origem` da guia (default do drawer v5: "GUIAS")
  "pagamentoConfirmado": true  // boolean — true = "Já paguei"/"Confirmar pagamento"; false = "Não paguei"/"Desmarcar"
}
```

Origem dos valores: v5 usa `detalhes.origem`/`detalhes.tipo` vindos de
`GET impostos/v5/impostos-a-pagar/guia/{id}` (fallback nas props `origem="GUIAS"`, `tipo="GUIA"`);
v3/v2 usam `guia.tipo`, `guia.origem` da listagem.

Resposta / tratamento:

- **v5**: o front lê só `movidaPara` (ex.: `"AGUARDANDO_PAGAMENTO"`, `"ESTE_MES"`), emite `pagamento-atualizado`
  e recarrega a lista. Toast de sucesso: **"{nome da guia} foi marcada como paga"**.
  Erro: "Não foi possível confirmar o pagamento. Tente novamente." / "Não foi possível atualizar o pagamento. Tente novamente."
- **v3**: a resposta é a **guia atualizada** (objeto completo com `aba`, `confirmacaoDePagamento{confirmado: "CONFIRMADO"|"NAO_CONFIRMADO"|"NENHUM", tipo: "MANUAL"…, desabilitado}`, `acaoBotao{id: "PAGAR"|"RECALCULAR"|"GERENCIAR"|"BAIXAR_GUIA"…}`, `vencimento{badge, venceMesAtual}`); o front faz `Object.assign` na linha e chama `previsao` de novo. Erro: toast "Erro ao registrar informação. Por favor, tente outra vez."
- **v2 histórico**: resposta é a guia atualizada (`confirmacaoPagamento: boolean`). Erro: "Não foi possível confirmar o pagamento da guia".

Textos da UI (semântica):

- "Confirmação de pagamento — Se você já pagou esta guia e mesmo assim ela continua aparecendo como disponível, clique no botão de Confirmar pagamento. Ao confirmar, vamos marcá-la como paga para não haver risco de duplicidade."
- Guia vencida: "Esta guia venceu. Confirme o pagamento somente se você já tiver pago e ela ainda continuar como vencida. Se ainda não pagou, peça o recálculo para evitar o acúmulo de juros e multas."
- Após confirmar: "A guia fica com o status de Paga. Se o mês de vencimento ainda estiver aberto, ela continua em Contas a pagar. Se já tiver fechado, ela vai direto para o Histórico."
- Após negar: "Você marcou como não paga … Ela volta para contas a pagar com o status Disponível."
- "**Auditorias** — Essa confirmação não serve como comprovante oficial. Fazemos auditorias periódicas e podemos atualizar o status com dados oficiais do governo."
- v3: "As opções de pagamento, recálculo e/ou parcelamento, serão habilitadas após você confirmar que o pagamento não foi feito." / "Preencha a Confirmação de pagamento. Você só pode recalcular impostos marcados como não pago."

Pré-condições: guia existente; o botão "Não paguei" (`permitirNegar`) só aparece quando a guia está `A_CONFIRMAR`
no modo detalhado; pendências críticas (Central de Rotinas) bloqueiam confirmação/recálculo ("Resolva as pendências").

Semântica: **declaração do cliente** de que a guia foi (ou não) paga; muda status/aba na plataforma; não paga nada.

Reversível: **sim** — mesma chamada com `pagamentoConfirmado` invertido (o drawer v5 tem o botão "Desmarcar", que chama
exatamente `pagamentoConfirmado:false`).

Risco: **médio** — sem efeito financeiro direto, mas uma confirmação falsa pode esconder um imposto não pago
(a auditoria mensal da Contabilizei corrige depois: "Nossa auditoria não identificou o pagamento desta guia").

### 1.2 Marcar guia como "A confirmar" (efeito colateral de baixar boleto)

- `POST /api/plataforma/impostos/v5/impostos-a-pagar/guia/registrar-a-confirmar/{idGuia}` — **sem body**.
- Chunk: `pagamentosFintech.*.js` (componente `BaixarGuia` do checkout), método `marcarGuiaAConfirmar`.
- Quando: só com rollout `v5`, **depois** que o usuário baixa a guia (`GET impostos/v3/impostos-a-pagar/guia/{id}/baixar-guia`) ou copia o código de barras, se o status ≠ `A_CONFIRMAR`. Erros são engolidos (`.catch(()=>{})`).
- Texto: "Você pode baixar a guia e pagar em qualquer caixa eletrônico… O status do pagamento será atualizado na plataforma após a compensação do pagamento, o que pode levar até 3 dias."
- Semântica: muda o status da guia para `A_CONFIRMAR` ("pagamento provavelmente feito, aguardando compensação").
- Reversível: indiretamente, via `confirmar-pagamento` com `false` (incerto se volta exatamente ao status anterior).
- Risco: **baixo/médio**. Observação: `ctbz impostos baixar` (que usa só o GET `baixar-guia`) **não** dispara esse POST — comportamento seguro hoje.

### 1.3 Recálculo de guia vencida (serviço pago)

Pré-carga (GET sem efeito):
`GET /api/plataforma/impostos/v2/impostos-a-pagar/recalculo/init?idGuia={id}&origem={origem}&tipo={tipo}`
→ `descricao, mesCompetencia, anoCompetencia, vencimento, valorOriginal, valorRecalculo, dataRecomendada (AAAA-MM-DD), finalDoMes, datasIndisponiveis[], cobrarRecalculo (bool), motivoSemCobranca`.

Escrita:

- `PUT /api/plataforma/impostos/v2/impostos-a-pagar/guia/{idGuia}/v2/recalcular` (note o `/v2/` repetido no fim)
- Chunk: `meus-impostos-a-pagar.*.js`, componente `RecalculoDeImposto`, rota `#/meus-impostos-a-pagar/recalculo-de-imposto?idGuia=&origem=&tipo=`
  (as telas v3 e v5 navegam para essa rota ao clicar "Pedir recálculo"/"Recalcular").
- Body:

```json
{
  "tipo": "GUIA",                 // da guia
  "origem": "GUIAS",              // da guia
  "dataVencimento": "DD/MM/AAAA"  // novo vencimento escolhido no DatePicker (default = dataRecomendada); front converte AAAA-MM-DD → DD/MM/AAAA
}
```

- Resposta: `{ modalSucessoRecalculo, guia: { dataVencimento }, ofertaDebitoAutomatico }`; o front volta à lista e mostra
  "Recebemos seu pedido para recalcular a guia, com novo vencimento em …" / "A guia será recalculada em até 3 dias úteis. Avisaremos por e-mail quando o novo valor, já atualizado com juros e multas, estiver disponível."
- Erro: exibe `response.data.detalhes[0].detalhe`, ou "Ocorreu um erro inesperado. Por favor, tente novamente".
- Textos: "O recálculo de guia de imposto é um serviço adicional, que será **cobrado na sua próxima mensalidade**." /
  "Contratação de serviço adicional: Recálculo de guia" (quando `cobrarRecalculo=true`; senão mostra `motivoSemCobranca`).
  "Ao definir a data de vencimento para o próximo mês, você vai pagar um valor maior em juros e multas."
- Pré-condições: guia vencida e marcada como **não paga** (ver 1.1); sem pendências críticas; data fora de `datasIndisponiveis`.
- Reversível: **não** (não há endpoint de cancelamento).
- Risco: **alto** — gera cobrança adicional na mensalidade e nova guia com juros/multa.

Variante morta: `PUT impostos/v2/historico-impostos/guia/{id}/recalcular` com o mesmo body existe em
`historico-impostos.*.js` (`recalcularGuia`) mas **não é chamada** por nenhum componente (código morto).

### 1.4 "Memória de cálculo visualizada"

- `POST /api/plataforma/impostos/impostos-a-pagar/memoria-de-calculo-visualizada` — **sem body**.
- Chunks: `impostosv5`, `impostos` (v3), `impostos-hub`, `historico-impostos-hub`.
- Quando: ao fechar/confirmar o modal de memória de cálculo exibido quando `GET impostos/v5/impostos-a-pagar/init` traz `exibirMemoriaDeCalculo`/`exibeMemoriaDeCalculo=true` (`dadosMemoriaDeCalculo`).
- Semântica: registra que o usuário viu o modal (não exibe de novo). Resposta ignorada.
- Reversível: não; Risco: **baixo**.

### 1.5 Parcelamentos (contratar)

Todos em `impostos.*.js` (módulo `5367`), rotas `#/impostos/parcelamento/{tipo}/contratar`. Simulação = GET `…/init`.

| Tipo | Simulação (GET) | Contratação | Body |
|---|---|---|---|
| Simples Nacional (parcelamento digital, Receita) | `impostos/parcelamento/negociacao-automatica/simples-nacional/init` | `POST impostos/parcelamento/negociacao-automatica/simples-nacional/contratar` | **nenhum** |
| PGFN Simples Nacional (dívida ativa) | `…/pgfn-simples-nacional/init` | `POST …/negociacao-automatica/pgfn-simples-nacional/contratar` | `{"quantidadeParcelas": <int>}` |
| PGFN previdenciário | `…/pgfn-previdenciario/init` | `POST …/negociacao-automatica/pgfn-previdenciario/contratar` | `{"quantidadeParcelas": <int>}` |
| PGFN não previdenciário | `…/pgfn-nao-previdenciario/init` | `POST …/negociacao-automatica/pgfn-nao-previdenciario/contratar` | `{"quantidadeParcelas": <int>}` |
| Especializado (feito por especialista) | `impostos/parcelamento/negociacao-especializada/init?tipoNegociacao=` | `POST impostos/parcelamento/negociacao-especializada/contratar` | `{"tipoNegociacao": "VENCIDOS" \| "DIVIDA_ATIVA" \| "DIVIDA_ATIVA_E_VENCIDOS"}` |

Detalhes:

- Todos com base `/api/plataforma/`.
- `quantidadeParcelas` vem da opção escolhida em `init.parcelas[]` (cada item: `quantidade, valorEntrada, valorDemaisParcelas`;
  default = `parcelas[0]`). Tabela "Valor primeira parcela + Valor demais parcelas".
- `init` também traz `cobrarServicoAdicional, valorServicoAdicional, valorEmissaoGuia, taxaReparcelamento, resumoParcelamento{quantidadeParcelas, valorEntrada, valorDemaisParcelas}, negociacao.impostos[]`.
  Sem dívidas, o `init` responde 560 "Empresa não possui guias para simulação" (já verificado).
- Resposta: especializado retorna `{ idTicket }` (vira pedido de atendimento); automáticos: resposta ignorada, abre modal de sucesso
  ("Recebemos seu pedido de parcelamento!").
- Erro: toast "Não foi possível contratar o parcelamento. Por favor, tente outra vez."
- Detalhes pós-contratação: `GET …/{tipo}/detalhes/init/{idParcelamento}` (ou `…/negociacao-especializada/detalhes/init/{idTicket}`).
- Textos: "Após a contratação, o prazo para efetivar o parcelamento e liberar as parcelas é de …", "Se a primeira parcela não for paga, o parcelamento será cancelado automaticamente e você terá que fazer uma nova contratação", "O valor das parcelas é corrigido mensalmente pela taxa Selic + 1% de juros", "Contratação de Serviços adicionais: … Emissão de guia de parcelamento: …" ("Cobrado na sua próxima mensalidade apenas se você aceitar a proposta…"), taxa de reparcelamento ("1ª parcela deverá cobrir de 10% a 20% do valor total").
- Reversível: **não** pela API (cancelamento só por falta de pagamento ou atendimento).
- Risco: **alto** — ato perante Receita/PGFN, com custo e consequências legais (confissão de dívida).

### 1.6 Pagar impostos com cartão de crédito (Adyen, via fintech)

Fluxo `#/pagar-impostos/checkout?ano=AAAA&mes=M[&idGuia=]` (`pagamentosFintech.*.js`):

1. `GET /api/plataforma/api/pagamentos-fintech/initial-data/{ano}/{mes}[?idGuia=]` → `ticket[] (guias: id, descricao, tipoImposto, valorCentavos, status, barcode), subtotal, operationCost, total, payAllButtonEnabled, maisDiasParaPagar, cartao (bool), paymentDetails{parcelamentos[{valorTotalComTaxa, valorParcela}], ultimoDiaPagamentoDisponivel}, paymentMethodsResponse{storedPaymentMethods[{id, name, lastFour, brand}]}, impostosAdicionaisPagamentoCartao, pagamentoRecorrenteHabilitado/Ativado`.
2. `GET /api/plataforma/api/card-payment/adyen/originKeys/domain` → `originKeys{<origin>: key}` (chave do Adyen Checkout).
3. **`POST /api/fintech/pagamento-cartao/adyen/payments`** (instância `X["f"]`):

```json
{
  "idGuias": [123, 456],          // ids de ticket[] (todas as guias listadas no checkout)
  "installments": 1,              // nº de parcelas do cartão (paymentDetails.parcelamentos)
  "storePaymentMethod": false,    // salvar cartão (só cartão novo)
  "paymentMethod": {              // gerado pelo Adyen Secured Fields (criptografado no navegador)
    "type": "scheme",
    "encryptedCardNumber": "adyenjs_…",
    "encryptedExpiryMonth": "adyenjs_…",
    "encryptedExpiryYear": "adyenjs_…",
    "encryptedSecurityCode": "adyenjs_…",
    "holderName": "…",               // só cartão novo
    "storedPaymentMethodId": "…"     // só cartão salvo
  },
  "browserInfo": { … }              // objeto do Adyen (3DS2)
}
```

- Existe uma versão alternativa não usada pela tela atual: `POST /api/plataforma/api/card-payment/adyen/payments` (mesmo body).
- Sucesso → rota `PagamentoSucesso`; erro → `PagamentoError`. Há "custo de operação" (taxa de conveniência) somado ao total.
- **Dados de cartão nunca vão em claro à Contabilizei**: o Adyen Web SDK (chunk `cadastroCartaoCredito~impostos-pagamento-recorrente~pagamentosFintech`) cifra no navegador (CSE). Uma CLI não consegue gerar `encrypted*` sem reimplementar a cifra do Adyen.
- Reversível: não (estorno só por atendimento). Risco: **alto** (débito no cartão, pagamento de tributo).

### 1.7 Débito automático de impostos ("pagamento recorrente")

Tela `#/impostos-a-pagar/debito-automatico` (`impostos-pagamento-recorrente.*.js`). O botão "Gerenciar débito automático"
aparece quando `GET impostos/v5/impostos-a-pagar/init` traz `podeGerenciarDebitoAutomatico=true`.
Estado: `GET /api/plataforma/payments/recorrencia/init` → `habilitado, habilitadoConta, ativado (cartão), ativadoConta (saldo conta PJ), aceiteConta, aceiteContaVigente, cartoesSalvos[{id, bandeira, principal…}], adyenClientKey{<origin>}, temContaPj, prazoParaDesativar, temLotesEmProcessamento, competencia, dataProximaTentativa, pagamentosAgendados/Concluido/Recusado, tentativaFalhou`.
Histórico: `GET payments/recorrencia/historico` → `pagamentos`.

| Ação | Método + caminho (base `/api/plataforma/`) | Body | Notas |
|---|---|---|---|
| Aceitar termo (cartão) | `POST payments/aceite-termo-recorrencia` | `{"tipo":"DEBITO_CARTAO"}` | Erro: "Falha ao aceitar os termos do serviço…" |
| Aceitar termo + ativar débito em conta | `POST payments/recorrencia/ativar-conta-com-termo` | nenhum | Sucesso: "Termos do débito automático aceitos com sucesso." Ativa `DEBITO_CONTA` |
| Ativar | `POST payments/recorrencia/ativar` | `{"tipo":"DEBITO_CARTAO"}` ou `{"tipo":"DEBITO_CONTA"}` | 400 com mensagem de termo ("termo de aceite", "precisa aceitar"…) → front abre modal de termo e repete. Botão "Ativar débito automático" |
| Desativar | `POST payments/recorrencia/desativar` | `{"motivo":"<texto livre>","tipo":"DEBITO_CARTAO"\|"DEBITO_CONTA"}` | Modal: "Ao desativar, você não terá mais agendamento de impostos e vai precisar pagá-los manualmente todo mês." / "A desativação dos pagamentos precisa de N dias para ser efetivada." Sucesso: "Débito automático de impostos desativado com sucesso." |
| Adicionar cartão | `POST payments/cartao/adicionar` | `{"browserInfo":{…},"paymentMethod":{…campos encrypted* do Adyen…,"holderName":"…"},"principal":true}` | Resposta = cartão salvo (`id, bandeira, principal`). 400 → `message` |
| Excluir cartão | `DELETE payments/cartao/{idCartao}/excluir` | — | id de `cartoesSalvos[].id`. 400 → `message` ("Houve um erro ao excluir o seu cartão.") |
| Aceite da conta digital | `POST fintech/conta-digital/aceite?origem=pagamento-recorrente` (base plataforma) | nenhum | Abertura de conta PJ (fora do escopo; listado por completude) |

- Reversível: ativar ↔ desativar (desativação pode levar `prazoParaDesativar` dias); adicionar ↔ excluir cartão.
- Risco: ativar/desativar **alto** (passa a debitar tributos automaticamente / deixa de pagar);
  aceite de termo **médio** (contratual); cartão **médio** (exige Adyen CSE, inviável na CLI).

### 1.8 Outras chamadas relacionadas (sem escrita)

- `GET fintech/metodos-pagamento-impostos/init` (base plataforma; chunk `impostos-metodos-de-pagamento`) é **só leitura**
  (tela informativa "Métodos de pagamento": boleto, débito automático no cartão, cartão de crédito). Nenhum POST.
- `GET api/pagamentos-fintech/card-impostos` (dashboard) — leitura.

### Recomendação para a CLI — Impostos

```text
ctbz impostos confirmar <idGuia> [--nao-paguei] [--yes]
    → PUT impostos/v5/impostos-a-pagar/guia/{id}/confirmar-pagamento {tipo, origem, pagamentoConfirmado}
      (buscar tipo/origem antes via GET impostos/v5/impostos-a-pagar/guia/{id}; fallback v3 se rollout != v5)
ctbz impostos desmarcar <idGuia>     (alias de --nao-paguei)
ctbz impostos recalcular <idGuia> [--vencimento DD/MM/AAAA] --yes
    → GET recalculo/init (mostrar valorOriginal, valorRecalculo, cobrarRecalculo, dataRecomendada)
      → PUT impostos/v2/impostos-a-pagar/guia/{id}/v2/recalcular; exigir confirmação explícita (cobrança na mensalidade)
ctbz impostos parcelamento simular <tipo>        (só GET init; seguro)
ctbz impostos parcelamento contratar <tipo> --parcelas N --yes   (alto risco; talvez só com --i-understand)
ctbz impostos debito-automatico status|historico (GETs)
```

Não implementar: pagamento com cartão (Adyen CSE), cadastro de cartão. Ativar/desativar débito automático só com
confirmação dupla. A data de pagamento **não** faz parte do body de "confirmar pagamento" — não há `--data`.

---

## 2. Mensalidade / pagamentos da Contabilizei

### 2.1 Cartões de crédito da mensalidade — gestão (Adyen)

Tela `#/formas-pagamento` (só se `GET /api/plataforma/cartoes-de-credito/init` → `cobrancaViaAdyen=true`; senão redireciona ao front legado `/pagto-formas`). Chunk `cadastroCartaoCredito.*.js`.

| Ação | Método + caminho (base `/api/plataforma/`) | Body |
|---|---|---|
| Listar | `GET cartoes-de-credito/gestao-pagamentos` | → lista `{id, digitosCartao, nomeCartaoCliente, dataExpiracaoCartao, cartaoPrincipal, bandeira}` |
| Chave Adyen | `GET billing/gestao-pagamentos/get-client-key` | → `{clientKey}` |
| Cadastrar (`#/cadastrar-cartao`) | `POST cartoes-de-credito/gestao-pagamentos` | `{"browserInfo":{…},"dadosDoCartao":{"encryptedCardNumber","encryptedExpiryMonth","encryptedExpiryYear","encryptedSecurityCode","holderName"},"cartaoPrincipal":true}` |
| Definir como principal / editar | `PUT cartoes-de-credito/gestao-pagamentos/{idCartao}` | `{"cartaoPrincipal": true\|false}` (único campo editável) |
| Excluir | `DELETE cartoes-de-credito/gestao-pagamentos/{idCartao}` | — |

- Sucesso no cadastro → rota "SucessoCadastroCartaoCredito"; erro: "Ocorreu um erro ao cadastrar o cartão de crédito."; PUT/DELETE erro: "Houve um erro ao acessar esta funcionalidade."
- Tokenização: Adyen Secured Fields (CSE no navegador) — o número do cartão **não** chega em claro.
- Reversível: cadastrar ↔ excluir; principal ↔ outro principal. Risco: PUT/DELETE **médio** (muda a forma de cobrança da mensalidade); POST inviável na CLI.

### 2.2 Pagar fatura (mensalidade) com cartão — fluxo legado Iugu

Tela `#/pagar-com-cartao` (`pagamentoCartaoCredito.*.js`):

1. `GET /api/plataforma/pagamento-cartao-credito/faturas` → `{statusLabel ("Em aberto"|"Pendente"|"Atrasada"|"Paga"), vencimento, total, …}`; só permite pagar se `statusLabel ∈ {Pendente, Atrasada}`.
2. O número do cartão vai **para a Iugu** (`window.Iugu.createPaymentToken`, account id embutido no bundle em `VUE_APP_BILLING_GATEWAY_ACCOUNT_ID`, base64), que devolve um token.
3. `POST /api/plataforma/pagamento-cartao-credito/pagar`

```json
{ "token": "<token Iugu>", "descricao": "1234" /* últimos 4 dígitos */, "principal": true /* switch "usar como cobrança recorrente" */ }
```

- Sucesso → `#/pagar-com-cartao/sucesso`. Erro: `response.data` (texto) ou "Ocorreu algum problema com o pagamento, tente novamente ou entre em contato com a operadora do seu cartão."
- Reversível: não. Risco: **alto** (cobrança no cartão) e exige tokenização Iugu — não recomendado para a CLI.

### 2.3 Botão do card "Fatura" do dashboard (`dashboard/fatura` → `botaoAcao`)

- `GET /api/plataforma/dashboard/fatura` traz `botaoTitulo/botaoAcao` e `botaoAdicionalTitulo/botaoAdicionalAcao`.
- Formato de `botaoAcao`: `"<tipo>:<valor>"`.
  - `redirect:<rota>` → navegação interna (sem rede).
  - qualquer outro tipo → **`GET /api/plataforma/<valor>`** (o caminho vem do servidor) → `{url}` aberto em nova aba.
    Toast: "Seu boleto para pagamento será aberto em uma nova aba. Os boletos gerados são registrados, por isso caso tenha dificuldades no pagamento, recomendamos tentar novamente após 2 horas."
- Semântica: **GET com efeito colateral**: gera e registra a segunda via da mensalidade.
- **Verificado numa conta real (23/10/2026, #851):** com a mensalidade atrasada, o card traz
  `"botaoTitulo": "Pagar com cartão"`, `"botaoAcao": "redirect:/painel-de-controle/#/pagar-com-cartao"`,
  `"botaoAdicionalTitulo": "Gerar segunda via s/ juros"` e
  `"botaoAdicionalAcao": "request:dashboard/nova-fatura/202610"`. O tipo é `request:` e o caminho é
  `dashboard/nova-fatura/{AAAAMM}`. Com a fatura em dia, os campos `botaoAdicional*` vêm `null`
  (e `botaoAcao` é, por exemplo, `redirect:/painel-de-controle/#/cadastrar-cartao`).
- `GET /api/plataforma/dashboard/nova-fatura/{AAAAMM}` registra a segunda via com vencimento novo
  (no caso, D+2) e devolve `{"url": "https://app.contabilizei.com.br/checkout/#/fatura-contabilizei?idFatura=<UUID>"}`.
- O checkout é uma SPA (`VUE_APP_CONTABILIZEI_FINTECH_URL = "/api/fintech/"`). Os endpoints são
  **públicos**: o UUID é a credencial, e eles funcionam sem cookie.

| Uso | Método e caminho | Testado |
|---|---|---|
| versão do checkout | `GET /api/fintech/public/v2/fatura-contabilizei/{uuid}/init` → `{"versao":"V4"}` | ✅ |
| fatura com meios de pagamento (v2) | `GET /api/fintech/public/v2/fatura-contabilizei/{uuid}` | ✅ |
| fatura com itens e status (v1) | `GET /api/fintech/public/fatura-contabilizei/{uuid}` | ✅ |
| Pix | `GET /api/fintech/public/v2/fatura-contabilizei/{uuid}/pix` (mesmo corpo da v2) | ✅ |
| cobrança | `GET /api/fintech/public/fatura-contabilizei/{uuid}/cobranca` (mesmo corpo da v2) | ✅ |
| boleto | `GET /api/fintech/public/v2/fatura-contabilizei/{uuid}/boleto` | ⚠️ 400 depois de paga |
| adiantar pagamento | `GET /api/fintech/public/fatura-contabilizei/{uuid}/link-adiantar-pagamento` | não testado |
| pagar com cartão | `POST /api/fintech/public/v2/fatura-contabilizei/{uuid}/pagar` | não testado, risco alto |

- Campos usados:
  - v2: `dataVencimento`, `total` e `meioDePagamento`. Nele, `pix.textoPagamento` é o copia e
    cola (`pix.urlPagamento` é o QR em data-URI), `boleto.textoPagamento` é a **linha digitável**
    (47 dígitos) e `boleto.urlPagamento` é, apesar do nome, o **código de barras** (44 dígitos).
    Cada meio tem `exibir`. A v2 também traz `token`, que a CLI não lê nem mostra.
  - v1: `resumo.itens[{descricao, valor}]` (mensalidade, multa e juros da fatura anterior,
    serviços adicionais) e `status{id, descricao}`.
- Depois do Pix pago, a v1 continuou `PENDENTE`, mas `inadimplencia/consultasituacaomensalidadeempresa`
  já dizia `OK`: a baixa da fatura pública atrasa.
- Não há endpoint só de leitura com o UUID da fatura atual; o caminho é gerar a segunda via.
- Os endpoints públicos mostram razão social, endereço, CNPJ mascarado e o código de pagamento a quem
  tiver o UUID. A CLI não o grava no registro de ações e o mascara no `--verbose`.
- Risco: **médio** (registra boleto; não cobra). Na CLI: `ctbz mensalidade segunda-via`, que lê a
  ação do card a cada vez e só segue uma ação `request:` com caminho relativo do BFF.

### 2.4 Troca de plano (upsell Multibenefícios)

- `PUT /api/plataforma/home/pagtoplano/atualizar-plano` (com `/` inicial; axios trata igual).
- Chunk `landing-beneficios.*.js`, tela `#/beneficios/landing-upsell`, modal `ModalUpsell` ("Confirmar alteração de plano?").
- Body: `{"idPagtoPlano": <id>, "gerarContrato": true}` — `idPagtoPlano` = `comparacaoPlanos.planoUpgrade.id`, vindo de
  `GET cross-sell/multibeneficios/cms/v1/landing-upsell` (provável origem de `landingData`, **incerto**).
- Pré-condição: checkbox "Li e aceito o contrato…" (contrato via `GET /contrato/buscarContratoServico` → `html`).
- Sucesso: "Parabéns, seu novo plano está ativo!" / "A troca foi feito com sucesso! O novo valor entrará em vigor a partir da próxima fatura." Erro: "Não foi possível trocar seu plano".
- Efeitos paralelos: envia lead/eventos ao HubSpot (`X["a"]`, base `/api/leads/hubspot/`: `POST send`, `POST track-event`).
- Reversível: não pela API. Risco: **alto** (altera contrato e valor da mensalidade).

### 2.5 Oferta "Plano Manutenção" (dashboard)

- `POST /api/plataforma/planos/aceite-plano-manutencao`, body `{"aceite": true|false}`.
- Chunks: `chunk-b88c6b8a` (dashboard v2) e `chunk-ecc0fd76` (dashboard v1), módulo `8ae59`.
- Quando: `dashboard/v2/init` → `mensalidade.ofertaPlanoManutencao.exibir=true` (v1: `mensalidadeAlertas.ofertaPlanoManutencao`).
  Botões "aceitar novo plano" (`true`) / "manter plano atual" (`false`).
- Textos: "Plano Manutenção — O jeito mais econômico para sua empresa permanecer regularizada enquanto não estiver faturando";
  checkboxes "Estou ciente de que o plano Manutenção só é válido enquanto minha empresa não estiver faturando."; após aceitar:
  "O plano Manutenção somente será ativado após a regularização das mensalidades pendentes." → botão "Regularizar" (`#/pagto-pendente`).
- Erro: "Não foi possível salvar sua opção, tente novamente mais tarde."
- Reversível: não (incerto). Risco: **alto** com `true` (troca de plano); `false` = só registra recusa (baixo).

### 2.6 Retenção — "Solicitar mais dias de cobertura"

- `POST /api/plataforma/retencao/paywall/salvar-interacao`, body `{}`.
- Chunks: `chunk-vendors` (app-shell), `chunk-b88c6b8a`, `chunk-2371324a`.
- Quando: modal "Sua mensalidade está atrasada." → botão "Solicitar mais dias de cobertura", e só se `competenciaAnteriorAtrasada=true`.
  Em seguida abre o modal "prolongamos o prazo de regularização da mensalidade até o final do mês".
- Semântica: registra a interação / concede prorrogação do paywall (incerto se o servidor efetivamente estende o prazo ou só registra).
- Reversível: não; Risco: **baixo**.

### 2.7 Fora do escopo, vistos de passagem

- `POST loans/aceite-oferta-dash/OFERTA_LOANS_FINTECH` com `{aceite: …}` (oferta de crédito no dashboard) — risco alto, relatar em outro contexto.
- `POST dashboard/v2/termo-exclusao`, `PUT fintech/conta-digital/desistir`, `POST fintech/conta-digital/aceite?origem=…`.

### Recomendação para a CLI — Mensalidade

```text
ctbz mensalidade status                 → GET dashboard/fatura, pagamento-cartao-credito/faturas (leitura)
ctbz mensalidade boleto                 → segue botaoAcao (GET dinâmico) e imprime a URL (avisar que registra boleto)
ctbz mensalidade cartoes                → GET cartoes-de-credito/gestao-pagamentos
ctbz mensalidade cartoes principal <id> → PUT cartoes-de-credito/gestao-pagamentos/{id} {"cartaoPrincipal":true}
ctbz mensalidade cartoes remover <id> --yes → DELETE cartoes-de-credito/gestao-pagamentos/{id}
```

Não implementar: cadastro/pagamento com cartão (Adyen CSE / Iugu), troca de plano, aceite de Plano Manutenção
(efeito contratual), salvar-interacao (sem utilidade fora da UI).

---

## 3. Resumo rápido (tabela)

| Método | Caminho (base `/api/plataforma/` salvo indicação) | Body | Risco |
|---|---|---|---|
| PUT | `impostos/v5/impostos-a-pagar/guia/{id}/confirmar-pagamento` | `{tipo, origem, pagamentoConfirmado}` | médio |
| PUT | `impostos/v3/impostos-a-pagar/guia/{id}/confirmar-pagamento` | idem | médio |
| PUT | `impostos/v2/historico-impostos/guia/{id}/confirmar-pagamento` | idem | médio |
| POST | `impostos/v5/impostos-a-pagar/guia/registrar-a-confirmar/{id}` | — | baixo |
| PUT | `impostos/v2/impostos-a-pagar/guia/{id}/v2/recalcular` | `{tipo, origem, dataVencimento:"DD/MM/AAAA"}` | alto (cobrado) |
| PUT | `impostos/v2/historico-impostos/guia/{id}/recalcular` (código morto) | idem | alto |
| POST | `impostos/impostos-a-pagar/memoria-de-calculo-visualizada` | — | baixo |
| POST | `impostos/parcelamento/negociacao-automatica/simples-nacional/contratar` | — | alto |
| POST | `impostos/parcelamento/negociacao-automatica/pgfn-{simples-nacional,previdenciario,nao-previdenciario}/contratar` | `{quantidadeParcelas}` | alto |
| POST | `impostos/parcelamento/negociacao-especializada/contratar` | `{tipoNegociacao}` | alto |
| POST | `/api/fintech/pagamento-cartao/adyen/payments` | `{idGuias[], installments, storePaymentMethod, paymentMethod(Adyen), browserInfo}` | alto |
| POST | `api/card-payment/adyen/payments` (não usado) | idem | alto |
| POST | `payments/aceite-termo-recorrencia` | `{tipo:"DEBITO_CARTAO"}` | médio |
| POST | `payments/recorrencia/ativar` | `{tipo}` | alto |
| POST | `payments/recorrencia/ativar-conta-com-termo` | — | alto |
| POST | `payments/recorrencia/desativar` | `{motivo, tipo}` | alto |
| POST | `payments/cartao/adicionar` | `{browserInfo, paymentMethod(Adyen), principal}` | médio |
| DELETE | `payments/cartao/{id}/excluir` | — | médio |
| POST | `cartoes-de-credito/gestao-pagamentos` | `{browserInfo, dadosDoCartao(Adyen), cartaoPrincipal}` | médio |
| PUT | `cartoes-de-credito/gestao-pagamentos/{id}` | `{cartaoPrincipal}` | médio |
| DELETE | `cartoes-de-credito/gestao-pagamentos/{id}` | — | médio |
| POST | `pagamento-cartao-credito/pagar` | `{token(Iugu), descricao, principal}` | alto |
| GET* | `<caminho de dashboard/fatura.botaoAcao>` | — (gera boleto) | baixo/médio |
| PUT | `home/pagtoplano/atualizar-plano` | `{idPagtoPlano, gerarContrato:true}` | alto |
| POST | `planos/aceite-plano-manutencao` | `{aceite}` | alto/baixo |
| POST | `retencao/paywall/salvar-interacao` | `{}` | baixo |
