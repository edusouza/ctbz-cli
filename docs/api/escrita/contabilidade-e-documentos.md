# Escrita: caixa, extratos, contas bancárias, contabilidade, documentos e certificado

Análise **estática** dos chunks minificados do painel (`/painel-de-controle/`, Vue 2), formatados
com `prettier` para leitura. Nenhuma requisição de rede foi feita.

Convenções:

- Bases: `X["c"]` = `/api/plataforma/` (BFF), `X["d"]`/`X["b"]` = `/api/legado/`. Salvo indicação,
  **todos os caminhos abaixo estão em `/api/plataforma/`**. Caminhos com barra inicial
  (`/movimentacao-financeira/...`) são relativos à mesma base: o axios junta `baseURL` e caminho.
- "Chunk" é o arquivo em `/painel-de-controle/js/` em que a chamada e a tela foram encontradas.
- Risco: **baixo** (marcação de UI ou evento, reversível), **médio** (altera dados contábeis,
  mas há como desfazer), **alto** (irreversível, envolve cobrança, segredo ou documento legal).
- Itens que não foi possível confirmar estão marcados como **(incerto)**.
- Também são registradas as chamadas que **faltam no `catalogo.md`**. Elas escapam ao gerador
  porque usam `axios({method, url})`, um caminho guardado em variável ou um helper que monta a query.

---

## 1. Caixa e movimentação financeira

### 1.1 `POST caixa/lancamentousuario/novo/`: criar **e editar** lançamento do caixa

- **Chunk/tela:** `lancamentos-caixa` (rota `/movimentacao-financeira/caixa`), modal "Novo"/"Editar"
  (`acao` = `CRIAR`/`EDITAR`).
- **Existe "editar lançamento"?** **Sim, pelo mesmo endpoint.** Se `lancamentoUsuario.id` vem
  preenchido, a operação é uma edição. Não existe `PUT` separado.
- **Corpo (JSON):**

```json
{
  "ano": "2026",                 // string: competência selecionada no filtro (String(ano))
  "mes": 9,                      // competência do filtro (número ou string, conforme o filtro; incerto)
  "lancamentoUsuario": {
    "data": "2026-09-15T03:00:00.000Z", // new Date(timestamp_ms) serializado; meia-noite local (BRT)
    "descricao": "texto livre",
    "id": 123,                   // AUSENTE na criação; id do lançamento na edição
    "idContaUsuario": 456,       // classificação (ver abaixo)
    "idVinculo": null,           // id de guia de imposto / sócio, quando a classificação exigir
    "valor": -150.25             // número; POSITIVO = Recebimento, NEGATIVO = Pagamento
  }
}
```

- **Origem dos valores:**
  - `tipo`: dois botões, **"Recebimento"** (`tipo=true`, o valor é forçado a positivo) e
    **"Pagamento"** (`tipo=false`, valor forçado a negativo, em `verificarSinalTrocado`).
    O backend não recebe nenhum campo "tipo": **o sinal de `valor` é o tipo**.
  - `data`: `<input type=date>` com `max` = hoje (uma data futura volta para hoje). O padrão é o dia 01
    da competência, ou hoje se a competência for o mês corrente.
  - `idContaUsuario`: vem de `serializedList.categoriasVinculoDTO` na resposta de
    `GET caixa/listpaginada/{ano}/{mes}/1000/null`. A lista é filtrada assim:
    - `tipo === "CONTA_USUARIO"` e `entrada === true` para Recebimento, ou `entrada === false` para Pagamento;
    - para Pagamento, é removido o id fixo `5981343255101440` (`categoriasIdNaoPermitidas`);
    - fica só o que tem `exibir: true` em `GET movimentacao-financeira/contas-usuario/{ano}/{mes}?origem=CAIXA`
      (lista `{id, exibir, ...}`).
  - `idVinculo`: obrigatório só quando a descrição da classificação é:
    - um imposto gerado pela Contabilizei (`"Impostos - COFINS"`, `"- CSLL"`, `"- FGTS"`, `"- INSS"`,
      `"- IRPJ"`, `"- IRRF"`, `"- ISS"`, `"- Simples Nacional"`). Nesse caso, a escolha é entre os itens
      `tipo === "GUIA_IMPOSTO"` de `categoriasVinculoDTO`, mais `{id:"0", descricao:"SEM GUIA"}`;
    - `"Sócios - Distribuição de Lucros Antecipados"`. Nesse caso, a escolha é entre os itens
      `tipo === "SOCIO"`.

    Nos demais casos, `idVinculo = null` (o validador o ignora).
  - Validação do front: todos os campos devem ser "truthy"; o valor 0 é recusado. Na edição, o botão
    só habilita se algo mudou.
- **Resposta/UI:** sucesso mostra o snackbar "Lançamento salvo com sucesso!". Erro mostra
  "Não foi possível salvar o lançamento! Tente novamente mais tarde. [<response.data>]".
  Nos dois casos a tela recarrega a listagem; o corpo da resposta não é usado.
- **Pré-condições:** `GET caixa/listpaginada/...` (categorias) e `GET movimentacao-financeira/contas-usuario/{ano}/{mes}?origem=CAIXA`.
  Lançamentos com `confirmadoViaSistema: true` mostram "SISTEMA" e **não têm botões de editar
  nem de excluir** (a CLI deve recusar o mesmo).
- **Semântica:** cria ou altera um lançamento manual do caixa (dinheiro em espécie) na competência.
- **Reversível:** sim (editar de novo ou excluir). **Risco: médio**, porque entra na contabilidade
  do mês.

### 1.2 `DELETE caixa/lancamentousuario/remover/{ano}/{mes}/{idLancamento}`

- **Chunk:** `lancamentos-caixa`. Não tem corpo. `ano`/`mes` vêm da competência do filtro e `idLancamento` é o `id` do item
  de `caixa/listpaginada`.
- **UI:** `DecisionModal` com o título "Excluir lançamento", o texto "Tem certeza de que deseja excluir o lançamento:
  <descrição> de <data> no valor de <valor>" e os botões "Não" e **"Excluir permanentemente"**.
  Sucesso mostra "Lançamento excluído permanentemente!". Erro mostra "Não foi possível excluir o lançamento! Tente
  novamente mais tarde. [<response.data>]".
- **Reversível:** não há undo (seria preciso recriar o lançamento à mão). **Risco: médio/alto.**

### 1.3 `PUT /movimentacao-financeira/classificar`: reclassificar lançamento de extrato

- **Chunk/tela:** `detalhes-do-extrato` (modal "ExtratoDetalhesModalEditarLancamento", botão
  "Salvar"). O serviço está em `detalhes-do-extrato~lancamentos-caixa` e em `importar-extrato~movimentacao-financeira`
  (módulo `9fe3`).
- **Corpo:** uma string JSON já serializada (`JSON.stringify`), com `Content-Type: application/json`:

```json
{ "idLancamentoUsuario": 789, "idContaUsuario": 456, "idSocio": null }
```

  - `idLancamentoUsuario`: item de
    `GET /movimentacao-financeira/lancamento-usuario/paginado?idContaBancaria&ano&mes&registroPorPagina=100&pagina=N`.
    A resposta é um **array** de lançamentos, sem total (`[]` no mês vazio), verificado numa
    conta real (#852): `{idLancamentoUsuario, data (epoch ms), descricao, valor, idContaUsuario,
    idVinculo, idLancamentoPai, situacao ("CONFIRMADO"…), predicts: [{idContaUsuario, descricao,
    probabilidade, tipo, situacao}]}`. A CLI pagina até a primeira página incompleta.
  - `idContaUsuario`: escolhido no autocomplete "Classificação". As opções vêm de
    `GET movimentacao-financeira/contas-usuario/{ano}/{mes}?origem=EXTRATO`. Ficam só as que têm `exibir`,
    `situacao === "ATIVO"` e `classificacao` igual a `RECEITA` (se `valor > 0`) ou `DESPESA` (se `valor < 0`).
    Os `predicts` do lançamento aparecem primeiro.
  - `idSocio`: as contas de id fixo `6199733752168448`, `5491304964816896`,
    `4975757608615936` e `5605376636485632` (contas de sócio) são desdobradas em uma opção por sócio,
    com `idComposto = "<idConta>/<idSocio>"`. Os sócios vêm de `GET /movimentacao-financeira/extrato/info`
    (campo `socios`). Sem sócio, o front manda `parseInt(undefined)` = `NaN`, que o JSON serializa como **`null`**.
- **Pré-condição:** `extrato/info` com `permiteEditarLancamento: true`; sem isso, a coluna "Ações" some.
- **UI:** sucesso (HTTP 200) mostra "Lançamento editado com sucesso!". Erro mostra "Erro ao salvar edição. [<msg>]".
  Um erro 406 `exception/erro-negocial-406` é mapeado para "Nossos contadores já classificaram as
  movimentações deste período para fazer a contabilidade. Em casos de correções entre em contato com
  nosso suporte."
- **Reversível:** sim, basta reclassificar de novo, enquanto o período estiver aberto. **Risco: médio.**

### 1.4 `PUT /movimentacao-financeira/desmembrar`: dividir um lançamento de extrato

- **Chunk:** `detalhes-do-extrato` (modal "Desmembrar"). O corpo é uma string JSON:

```json
{
  "idLancamentoPai": 789,
  "lancamentosFilho": [
    { "descricao": "parte 1", "valor": "50.00", "idContaUsuario": 456, "idVinculo": null },
    { "descricao": "parte 2", "valor": 25.5,    "idContaUsuario": 457, "idVinculo": 12 }
  ]
}
```

  - O modal começa com 2 filhos com metade do valor cada. **Atenção:** esses valores iniciais são
    *strings* (`toFixed(2)`); os que o usuário digita são números. O backend aceita os dois formatos (incerto).
  - `idVinculo`: id do sócio quando a classificação é composta (`"<idConta>/<idSocio>"`). Fora disso, `NaN`, que vira `null`.
  - O sinal de cada filho é forçado a ser igual ao do pai.
  - Validações do front: descrição e classificação são obrigatórias, nenhum filho pode ter valor 0 e a
    soma dos filhos tem de ser igual ao valor do pai ("O Valor total dos Lançamentos Desmembrados não pode ser diferente do
    valor do Lançamento Original."). Não é possível remover filhos quando restam 2 ou menos.
  - Desmembrar só é permitido se `valor != 0`, se o item não tiver `idLancamentoPai` e se não tiver `idVinculo`.
- **UI:** sucesso mostra "Lançamento desmembrado com sucesso!". Erro mostra "Erro ao salvar desmembramento. [<msg>]".
- **Reversível:** **sim**, pelo endpoint da seção 1.5. **Risco: médio.**

### 1.5 `DELETE /movimentacao-financeira/desmembrar/desfazer/{idLancamentoPai}?idLancamentoPai={idLancamentoPai}` (ausente do catálogo)

- **Chunk:** `detalhes-do-extrato`, modal "Cancelar desmembramento": "Este lançamento voltará ao
  formato original. Você poderá desmembrar novamente se precisar.", com o botão "Confirmar".
- O helper `_()` gera a query a partir do objeto `{idLancamentoPai}`, por isso o id aparece duas vezes.
  Disponível em um lançamento filho (`idLancamentoPai !== null`).
- **UI:** sucesso mostra "Desmembramento cancelado com sucesso!". Erro mostra "Erro ao cancelar desmembramento. [..]".
- **Risco: médio** (desfaz o desmembramento).

### 1.6 `DELETE /movimentacao-financeira/extrato?idContaBancaria={id}&ano={aaaa}&mes={m}` (ausente do catálogo)

- **Chunk:** `detalhes-do-extrato`, botão "Excluir Extrato". Se `extrato/info.permiteExclusaoExtrato`
  for falso, o modal mostrado é "Ops! Este extrato não pode ser excluído" ("Nossos contadores já concluíram a
  classificação deste extrato…").
- **Confirmação:** "Ao excluir você ficará com a **importação pendente** para este mês e deverá importar
  um novo arquivo. Informações adicionais das movimentações também deverão ser preenchidas novamente.
  Tem certeza que deseja excluir?"
- **UI:** sucesso mostra "O extrato de <Mês>/<ano> foi excluído com sucesso. Não se esqueça de importá-lo
  novamente." Erro abre o modal "não pode ser excluído".
- **Reversível:** não; é preciso reimportar, e as classificações se perdem. **Risco: alto.**

### 1.7 `POST movimentacao-financeira/eventoUploadExtrato`: concluir a importação do extrato

- Segundo passo, chamado depois de `upload-documentos/extrato/enviar/bucket` (ver 2.1). O corpo é:

```json
{
  "ano": 2026, "mes": 9, "cnpj": "<cnpj da empresa>", "idContaBancaria": 111,
  "nomeArquivoStorage": "<cnpj>_<ano>_<mes>_<numeroConta>_<lastModified>.ofx",
  "respostas": [ { "tipo": "SALDO_ULTIMO_MES", "data": "<new Date()>", "valor": 1234.56 } ]
}
```

  - **PDF:** chamado logo após o upload, com `respostas: []`. Abre o modal de sucesso do PDF.
  - **OFX:** depois de `GET movimentacao-financeira/info-extrato/{idConta}/{ano}/{mes}/{nomeArquivo}`
    (que devolve `infoExtrato.saldoUltimoDia`, `extratoForaPeriodo` etc.), a tela "ExtratoUploadSucesso"
    pede a confirmação do saldo do último dia ("Preenchimento obrigatório") e manda
    `respostas: [{tipo:"SALDO_ULTIMO_MES", data, valor:saldoUltimoDia}]`.
  - O ramo `extratoSemMovimentacao` envia `respostas: []`, mas nenhum código encontrado liga essa flag
    **(incerto: possivelmente código morto)**.
- **UI:** os modais "ComMovimentoSucesso" e "SemMovimentoSucesso". Erro mostra "Ops! Um erro ocorreu ao tentar processar o arquivo."
- Em seguida o front chama `GET /integracao-bancaria/parametros-integracao/{idContaBancaria}` (oferta
  de Open Finance).
- **Risco: médio.** Efetiva a importação; o desfazer é o `DELETE /movimentacao-financeira/extrato` da seção 1.6.
- **Processamento assíncrono (observado numa conta real, #855).** Depois do evento, a competência
  fica `PROCESSANDO` e o servidor converte o PDF em OFX. Ao terminar (cerca de 30 s depois), o
  processamento de um PDF de **agosto** gravou dois OFX no mesmo instante (5 ms de diferença no
  `DTSERVER`): o de agosto, com os 2 lançamentos, e um de **setembro** vazio (nenhum `STMTTRN`).
  Setembro passou a `IMPORTADO`, com `extrato/info.quantidadeLancamentos = 0` e
  `lancamento-usuario/paginado` = `[]`, **antes** de qualquer envio de setembro. Os dois envios
  seguintes do PDF de setembro foram recusados com HTTP 400 `ERRO-NA-VALIDACAO-DE-DOCUMENTO-API-CONTABIL`
  ("não reconheceu o arquivo"). Não se sabe se a recusa veio do arquivo ou da competência já
  importada, nem por que o processamento criou o mês seguinte. O que se sabe:
  - um mês pode ficar `IMPORTADO` sem nenhum lançamento por efeito do processamento de outro mês;
  - `GET movimentacao-financeira/extrato/info?idContaBancaria&ano&mes` traz `quantidadeLancamentos`
    (e `urlDownloadExtrato`, uma URL assinada do OFX gerado, que a CLI não usa nem mostra);
  - `GET movimentacao-financeira/v2/extratos` traz `exibirBotaoImportar: false` nas competências
    `IMPORTADO`: o painel não oferece importar de novo sem excluir antes (seção 1.6).

  Por isso a CLI recusa importar sobre uma competência `IMPORTADO` ou `PROCESSANDO` (com a
  mensagem certa quando ela está vazia) e, depois de importar, avisa quando a competência termina
  sem lançamentos.

### 1.8 `GET movimentacao-financeira/permiteImportacao/{idContaBancaria}/{ano}/{mes}` (só leitura)

- Validação prévia ao upload. Resposta: `{ periodosDeImportacao: [{ano, mes, mensagem, temLancamentoPendente,
  teveLancamento, ...}] }`. Se existir um período anterior pendente diferente do escolhido, a tela mostra o modal
  "ErroImportacaoPeriodoComPendencia" e **não** deixa importar.
- Não tem efeito colateral visível. A CLI deve chamá-lo antes de importar.

### 1.9 `POST movimentacao-financeira/interagiu-modal-integracao`

- **Chunk:** `movimentacao-financeira`, chamado ao clicar em "integrar" ou ao fechar o modal de integração
  automática. Por um **bug do front**, o corpo enviado é o objeto de config:
  `{"headers":{"Content-Type":"application/json"}}`. O servidor provavelmente ignora o corpo.
- **Semântica:** marca "usuário viu ou interagiu com o modal de integração" (flag de UI). **Risco: baixo.**
  Não é útil na CLI.

### 1.10 Outras escritas de UI nesta área

- `POST /api/legado/evento-tour` com o corpo `{cnpj, flow, evento:"visualizado"}`. Antes é feito
  `GET /api/legado/evento-tour/{cnpj}/{flow}/{evento}`. É um registro de tour do Appcues. Risco baixo; ignorar.

**Recomendação para a CLI (caixa e extratos):**

```text
ctbz caixa adicionar   --competencia 2026-09 --data 2026-09-15 --valor 150.25 \
                       (--recebimento|--pagamento) --conta <idContaUsuario|nome> \
                       --descricao "…" [--guia <id>|--sem-guia] [--socio <id>]
ctbz caixa editar <id> --competencia 2026-09 [--data …] [--valor …] [--conta …] [--descricao …]
ctbz caixa remover <id> --competencia 2026-09            # pedir confirmação; --yes
ctbz caixa contas      --competencia 2026-09 [--entrada|--saida]   # categoriasVinculoDTO ∩ exibir
ctbz extrato classificar <idLancamentoUsuario> --conta <id> [--socio <id>]
ctbz extrato desmembrar  <idLancamentoUsuario> --parte "desc:valor:conta[:socio]" …
ctbz extrato desfazer-desmembramento <idLancamentoPai>
ctbz extrato excluir --conta-bancaria <id> --competencia 2026-09   # risco alto, confirmação dupla
```

Na CLI, o sinal deve ser derivado da flag (`--pagamento` grava o valor negativo). Antes de qualquer escrita, a CLI deve
validar as regras do front: nada de lançamento `confirmadoViaSistema`, nada de data futura e soma exata no
desmembramento.

---

## 2. Extratos e contas bancárias

### 2.1 Como o arquivo de extrato é enviado: **multipart direto ao BFF** (sem URL assinada)

Existem **dois** caminhos de upload. Nenhum usa URL assinada, GCS direto ou `storage.googleapis`.
O nome "bucket" indica apenas que o backend grava o arquivo no storage.

**(a) `POST upload-documentos/extrato/enviar/bucket`** (tela "Importar extrato" em Movimentação financeira)

- `multipart/form-data` com os campos `nomeArquivo`, `ano`, `mes`, `documento` (o arquivo) e `idConta`.
- `nomeArquivo` é gerado pelo front: `"<cnpjEmpresa>_<ano>_<mes>_<numeroConta>_<file.lastModified><.ext>"`.
  Esse nome é reaproveitado depois em `info-extrato/...` e em `eventoUploadExtrato.nomeArquivoStorage`.
- **Formatos:** `.ofx`, `.pdf` (`extencoesValidas`, `acceptedFiles: "application/pdf,.ofx"`), até **30 MB**.
  Se o banco tem `importacaoManualParaPdf` e o arquivo é PDF, a tela mostra o erro "ErroImportacaoManual".
- **Erros conhecidos:**
  - `exception/movimentacao-financeira-901`: "Por enquanto a importação do seu banco é compatível
    apenas com OFX…";
  - `-902`/`-903`/`-904`/`-905`: o extrato não corresponde ao banco, à conta, à agência e conta ou à agência
    (detalhe em JSON no campo `detalhes[0].detalhe`);
  - resposta `{account_validation:"account_dont_match", error:"…(CONTA-NAO-CADASTRADA)…", agency_number,
    account_number}`: abre o modal de cadastro de conta já preenchido. O texto é "A conta bancária identificada no extrato
    não foi encontrada no seu cadastro…". A tela salva com `contabancaria/salvar` e recarrega;
  - `(ERRO-NA-VALIDACAO-DE-DOCUMENTO-API-CONTABIL)`: o modal "Extrato não identificado".
- **Fluxo completo:** `permiteImportacao` → `extrato/enviar/bucket` → `GET info-extrato/…`
  (o erro `extratoForaPeriodo` abre o modal "mês incorreto") → `POST eventoUploadExtrato` (seção 1.7).

**(b) `POST upload-documentos/extrato/enviar`** (Central de Documentos, ausente do catálogo, usa `axios({method,url})`)

- `multipart/form-data` com os campos `documento` (arquivo), `competencia` (`"MM/AAAA"`) e `idContaBancaria`.
- Os bancos e as contas vêm de `GET upload-documentos/extrato/v2/init`, que devolve
  `{bancos:[{id, codigo, formatosDisponiveisUpload:["OFX","PDF",…], contasBancarias:[{id, statusIntegracao,…}]}],
  primeiraCompetencia:{mes, ano}}`. O formato aceito depende do banco. Contas `INTEGRADA` ficam bloqueadas
  ("Não é preciso fazer importação para esta conta, ela é feita automaticamente"). O tamanho máximo é 30 MB.
  Para o banco de código 323, a tela mostra um alerta do Mercado Pago.
- **UI:** sucesso mostra "Arquivo enviado com sucesso!". Erro mostra `response.data` ou "Arquivo não enviado. Verifique seu
  arquivo e tente enviar novamente."
- Neste caminho não há confirmação de saldo nem `eventoUploadExtrato` **(incerto se o backend
  processa como importação ou só arquiva)**.

**Risco do upload: médio.** O extrato importado entra na contabilidade; o desfazer é o
`DELETE /movimentacao-financeira/extrato` da seção 1.6.

### 2.2 `POST upload-documentos/extrato-aplicacao-financeira/enviar/sem-aplicacao-financeira` ("Não tenho aplicação nesta conta")

- **Chunk:** `central-de-documentos`, formulário de Extrato de aplicação financeira, botão
  **"Não tenho aplicação nesta conta"**. O corpo é:

```json
{
  "tipoDocumento": "EXTRATO_APLICACAO_FINANCEIRA",
  "competenciasPendentes": [ { "id": "<idPendencia>", "mes": 1, "ano": 2026 } ],
  "idContaBancaria": 111,
  "tipoPendencia": "EXTRATO_APLICACAO_FINANCEIRA"
}
```

  - `tipoDocumento` e `tipoPendencia` recebem o mesmo valor, o tipo da tela. O formulário também serve
    `INFORME_DE_RENDIMENTO_INVESTIMENTOS` (incerto se o botão aparece nesse tipo).
  - As competências são as pendências da conta e do `tipoInvestimento` que ainda não têm arquivo `SUCCESS`.
    Vêm de `GET documentos/envio-documento/init?id=<idPendencia>&id=…` (campo `documentos`, com
    `propriedades.{idPendencia, idContaBancaria, tipoInvestimento, competencia{mes,ano,periodo}, banco,
    agencia, contaCorrente}`).
- **Resposta:** em caso de erro, `response.data.competenciasComErro: [{mes, ano}]`. O front marca as demais como
  sucesso e mostra "Ocorreu um erro ao processar sua solicitação. Tente novamente."
- **Semântica:** declara que não houve aplicação ou movimentação no período, o que **resolve as pendências**.
- **Reversível:** não há endpoint de desfazer (incerto). **Risco: médio/alto** (é uma declaração).

### 2.3 `GET upload-documentos/extrato/v2/init`: só leitura (ver 2.1b)

### 2.4 `POST contabancaria/salvar`: cadastrar ou editar conta bancária

- **Chunks:** `conta-bancaria`, `detalhes-da-conta` e `movimentacao-financeira` (cadastro pelo
  upload). O serviço está no módulo `ffc2` e o modal é `ModalContaBancaria`.
- **Corpo:**

```json
{
  "id": 222,                       // só na edição
  "bancoId": 33,                   // id do banco (lista "bancos" de contabancaria/list)
  "agencia": "1234",
  "contaCorrente": "123456",       // conta + dígito concatenados, sem hífen
  "dataSaldoInicial": "2024-01-10",// "Abertura da conta*" (YYYY-MM-DD)
  "vlrSaldoInicial": 0
}
```

  Na edição, o objeto é a cópia do item de `contabancaria/list` **sem** os campos `permissoes`, `codigoBanco`,
  `dataExpiracao`, `dataIntegracao`, `fluxoIntegracao`, `idConsentimento`, `statusIntegracao`,
  `permiteVerDetalhes` e `identificadorInstituicao` (os demais campos do item vão junto: `nomeBanco` etc.).
  É **obrigatório** marcar o checkbox "Declaro que a conta acima é uma Conta PJ e suas movimentações
  refletem exclusivamente este CNPJ." (só no front).
- **Resposta:** `{ contaBancariaInfo: { contasBancarias: [...] }, exibirModalIntegracaoOpenFinance: bool }`.
  Na criação, se `exibirModalIntegracaoOpenFinance` for verdadeiro, a tela oferece integrar via Open Finance.
- **UI:** sucesso mostra "Conta Bancária salva com sucesso!". Erro mostra "Não foi possível identificar uma conta bancária.
  Tente novamente mais tarde. [<erro>]".
- Editar e excluir dependem de `permiteEditar`/`permiteExcluir`/`motivoPermissoesExclusaoEEdicao`, vindos de
  `GET contabancaria/detalhes-da-conta/init/{id}`.
- **Risco: médio.** A data de abertura define o período cobrado de extratos.

### 2.5 `DELETE contabancaria/excluir/{idContaBancaria}`

- **Chunk:** `detalhes-da-conta`, botão "Excluir conta" (visível só com `permiteExcluir`).
  `DecisionModal` "Excluir Conta Bancária": "Tem certeza de que deseja excluir a conta bancária:
  <banco> agência <ag> conta <cc>?", com os botões "Não" e **"Excluir permanentemente"**.
- **UI:** sucesso mostra "Conta Bancária excluída com sucesso!". Erro mostra "Não foi possível excluir a conta bancária…".
- **Reversível:** não (seria preciso recadastrar, e os vínculos se perdem). **Risco: alto.**

### 2.6 `POST /contabancaria/divergencia/atualizar-conta` e `POST /contabancaria/divergencia/cadastrar-conta`

- **Chunk:** `divergencia-de-contas`, tela "Divergência de dados bancários". Aparece quando a conta
  integrada (Open Finance/Belvo) não bate com a conta cadastrada.
- Pré-condição: `GET contabancaria/divergencia/init/{bancoId}` devolve
  `{contaIntegradaBelvo:{agencia, conta, codigoBanco}, contasNaoIntegradas:[...]}`. Os parâmetros da rota vêm
  do item da conta divergente (`id`, `bancoId`, `nomeBanco`).
- **atualizar-conta** ("vincular conta", quando o usuário escolhe uma conta cadastrada correspondente):
  `{ idContaBancaria: <id da conta cadastrada escolhida>, agenciaContaIntegrada, contaContaIntegrada,
  idContaBancariaDivergente: <id da conta divergente> }`.
- **cadastrar-conta** ("cadastrar como nova"):
  `{ codigoBanco, agenciaContaIntegrada, contaContaIntegrada, idContaBancariaDivergente }`.
- **UI:** os modais "Conta atualizada" e "Conta cadastrada". Erro mostra "Houve um erro ao inesperado. Por favor, tente novamente."
- **Risco: médio** (mexe no vínculo da integração). Pouco útil na CLI.

### 2.7 `POST /oferta-conta-integracao-bs2`

- **Chunk:** `movimentacao-financeira`, modal de oferta e integração do banco BS2 (Contabilizei Bank).
  O corpo é `{ visualizouModalIntegracao: bool, visualizouModalConta: bool }`
  (`theme === "integracao"` ou `"ativacao"`). É enviado tanto ao aceitar (o front então navega para
  `/conta-bancaria/bs2/integracao|pendencias`) quanto ao recusar (`204` atualiza a view).
- **Semântica:** registra que a oferta foi vista (flag de UI). **Risco: baixo.** Ignorar na CLI.

### 2.8 Integração bancária (Open Finance), ausente do catálogo (caminho em variável)

- `POST /integracao-bancaria/token` com `{ token, parceiro }`: conclui a conexão feita no widget do
  parceiro (Pluggy/Belvo). Depende de um token gerado no browser. **Fora do alcance da CLI.**
- `POST /integracao-bancaria/{idContaBancaria}/inativacao` com `{ idContaBancaria }`: "remover integração".
  UI: "Integração removida com sucesso!" / "Não foi possível remover sua integração…". **Risco: alto**
  (o texto do modal diz que "não poderá ser desfeito").
- `GET /integracao-bancaria/widget-info?parceiro&gerenciar-consentimento&idContaBancariaRenovacao&url-success&url-exit&url-event`
  e `GET /integracao-bancaria/parametros-integracao/{id}` (só leitura).
- BS2 (fora do escopo; anotado apenas): `PUT conta-bancaria/{chave}/{conta}?integrar={status}`
  (chunk `contaBancariaIntegracao`), `POST /api/bs2/sync/aceitar`,
  `POST api/bs2/envia-resolucao-pendencias…`.

**Recomendação para a CLI (contas e extratos):**

```text
ctbz contas-bancarias adicionar --banco <id|código> --agencia 1234 --conta 12345-6 \
                                --abertura 2024-01-10 [--saldo-inicial 0] --declaro-conta-pj
ctbz contas-bancarias editar <id> [--agencia …] [--conta …] [--abertura …]
ctbz contas-bancarias remover <id>                 # risco alto: confirmação + --yes
ctbz extrato importar arquivo.ofx --conta-bancaria <id> --competencia 2026-09 \
                      [--saldo-final 1234.56]      # permiteImportacao → bucket → info-extrato → eventoUploadExtrato
ctbz extrato enviar arquivo.pdf --conta-bancaria <id> --competencia 2026-09   # rota Central de Documentos
ctbz aplicacoes sem-movimento --conta-bancaria <id> --pendencia <id>…        # sem-aplicacao-financeira
```

Para `importar`, recomenda-se o fluxo (a). Ele valida período, banco e conta e devolve o
saldo do último dia, que a CLI pode mostrar e confirmar ou receber via `--saldo-final`. O `nomeArquivo` deve ser
gerado no mesmo formato do front.

---

## 3. Contabilidade

### 3.1 `POST informerendimento/reabrir-balanco/{anoCompetencia}`

- **Chunk:** `informe-de-rendimentos`. Não tem corpo.
- **Quando aparece:** `GET informerendimento/v2/{ano}/restricoes` devolve
  `restricoes.pendenciaDocumental.{possuiPendencia, pendencias[], fluxoRegularizacao, valorServicoAdicional}`
  e `processoReabertura.status` (`NENHUM`/`EM_ANDAMENTO`/`ANALISANDO`). Com
  `possuiPendencia` e `fluxoRegularizacao === "REABERTURA_BALANCO"`, a tela abre o modal "ModalExigibilidadeDocumental":
  "Verificamos que você possui pendências documentais referentes ao exercício de <ano> … Caso sua
  empresa possua lucros no exercício e você opte por **não regularizar pendência**, não será possível
  realizar a distribuição de lucros…", com os botões "Não regularizar pendência" e **"Regularizar pendências"**.
- Clicar em "Regularizar pendências" dispara `reabrir-balanco`. Em caso de sucesso, a tela vai para a Central de Rotinas.
  Com `fluxoRegularizacao === "CONTRATAR_SERVICO_ADICIONAL"`, a tela **não** chama o endpoint: abre um modal de
  serviço pago ("Seu período contábil já foi fechado… Esse serviço gera uma cobrança de R$ <valor>") que
  leva à rota `ServicosAvulsosReaberturaDoBalanco`.
- **Erros:** `403` mostra "Usuário admin não pode fazer a reabertura do exercício contábil para um cliente."
- **Semântica:** reabre o exercício contábil (balanço) do ano para que as pendências documentais sejam
  regularizadas. O status passa a `EM_ANDAMENTO`/`ANALISANDO`, e o informe de rendimentos fica indisponível.
- **Reversível:** não há endpoint. **Risco: alto** (afeta balanço, informe e distribuição de lucros).
- Os demais `POST informerendimento/*` (`aceite/{ano}`, `aceitar-termo-debitos/{ano}` com `{aceite}`,
  `aceitar-carta-responsabilidade/{ano}`, `historico`, `salvarconfiguracaocliente`) estão no mesmo
  módulo e pertencem ao contexto de lucros e informe (outra página).

### 3.2 `POST documentos/reclassificar`: alterar a classificação de lançamento bancário (Central de Rotinas)

- **Chunk:** `app` (módulo da Central de Rotinas), modal "ModalCentralDocumentosReclassificacao".
  O título é "Alterar classificação de extrato bancário" para os tipos `CONTRATO_DE_FINANCIAMENTO` e
  `AQUISICAO_INVESTIMENTO_ANJO`; nos demais casos é "Alterar lançamento bancário".
- **Pré-condição:** `GET documentos/reclassificar/init?id=<idPendencia>&id=…` (ids de
  `rotina.propriedades.documentosPendentes`). A resposta é uma lista de
  `{idPendencia, idLancamento, classificacaoAtual, data, valor, descricao, banco,
  classificacoes:[{idClassificacao, nome, socios:[{idSocio, nome}]}]}`.
- **Corpo:** um array, um item por pendência:

```json
[ { "idPendencia": "…", "idClassificacao": 456, "idSocio": null, "idLancamento": 789 } ]
```

  `idSocio` é obrigatório quando a classificação escolhida tem `socios` (com um único sócio, o front o seleciona sozinho).
- **UI:** sucesso mostra "A nova classificação do extrato foi salva e a rotina concluída." (vindo da Central de
  Rotinas, a página recarrega). Erro mostra "Falha ao salvar nova classificação. Tente novamente."
- **Semântica:** reclassifica o lançamento e **conclui a pendência/rotina**. **Risco: médio**
  (é possível reclassificar de novo pela seção 1.3 enquanto o período estiver aberto; a rotina não reabre).

**Recomendação para a CLI (contabilidade):**

```text
ctbz balanco reabrir --ano 2025         # só se restricoes.fluxoRegularizacao == REABERTURA_BALANCO;
                                        # mostrar restrições e exigir --yes; nunca para CONTRATAR_SERVICO_ADICIONAL
ctbz rotinas reclassificar <idPendencia>… --classificacao <id> [--socio <id>]
```

---

## 4. Documentos (Central de Documentos)

### 4.1 Upload de documento **com arquivo**: multipart para o BFF (ausente do catálogo)

**`POST documentos/envio-documento/enviar`** (um arquivo para uma pendência)

- `multipart/form-data`:
  - `arquivo`: o arquivo;
  - `tipoDocumento`: o tipo da pendência (lista abaixo);
  - `competencia`: JSON `{"mes":1,"ano":2026}`;
  - `metadados`: JSON `{"valor", "idPendencia", "periodo", "idContaBancaria", "tipoInvestimento"}` (campos
    nulos quando não se aplicam);
  - `ordemArquivoCompetencia`: número da ordem do arquivo, ou a string `"null"` (FormData de `null`).

**`POST documentos/envio-documento/enviar/consolidado`** (um arquivo para várias competências,
"Em um arquivo único")

- `multipart/form-data`:
  - `arquivo`;
  - `tipoDocumento`;
  - `documentos`: JSON
    `[{competencia:{mes,ano}, metadados:{idPendencia, idContaBancaria, periodo, tipoInvestimento}}, …]`.

**Comportamento comum aos dois:**

- **Origem dos dados:** `GET documentos/envio-documento/init?id=<idPendencia>&id=…` devolve
  `{ tiposPermitidos: [...], documentos: [ {tipo, propriedades:{…}} ] }`. Os ids das pendências chegam pela
  Central de Rotinas (`query.id`).
- **Tipos** (mapa de formulários da tela "EnvioDeDocumentos"):
  - `EXTRATO_BANCARIO_MOVIMENTACOES` (usa 2.1b);
  - `EXTRATO_APLICACAO_FINANCEIRA`;
  - `INFORME_DE_RENDIMENTO_INVESTIMENTOS`;
  - `CONTRATO_DE_EMPRESTIMO`;
  - `CONTRATO_DE_FINANCIAMENTO`;
  - `CONTROLE_DE_INTERMEDIACOES`;
  - `ESTOQUE`;
  - `AQUISICAO_ATIVO_IMOBILIZADO`;
  - `AQUISICAO_INVESTIMENTO_ANJO`;
  - `CONTRATO_DE_AFAC`.
- O limite de tamanho no formulário de aplicações é `MAX_FILE_SIZE_MB` (30 MB, incerto para os outros tipos).
- **Resposta/erros:** cada envio é uma promessa independente (`Promise.allSettled`). Para cada erro, o front lê
  `response.data` e remove o prefixo `(ERRO-NA-VALIDACAO-DE-DOCUMENTO-API-CONTABIL)`. Também extrai
  `idPendencia` do FormData para marcar o item. Em caso de sucesso, a tela mostra o modal "ModalSuccessUploadFiles".
  Exceção: em `AQUISICAO_ATIVO_IMOBILIZADO` os itens não são marcados como enviados.
- **Risco: médio.** Envia um documento contábil e resolve a pendência. Não há endpoint de exclusão no
  front (incerto).

### 4.2 `POST documentos/envio-documento/enviar/sem-arquivo`: declarar "não tenho o documento"

- **Chunk:** `app` (Central de Rotinas e pendências críticas). O corpo é um array:

```json
[ { "idPendencia": "<id>", "tipo": "ESTOQUE" } ]
```

  `tipo` ∈ { `CONTROLE_DE_INTERMEDIACOES`, `ESTOQUE`, `CONTRATO_DE_AFAC` (ou o `tipo` da rotina) }.
  Os ids vêm de `rotina.propriedades.documentosPendentes`.
- **Modais e textos:**
  - estoque: "Confirmar ausência de estoque no período". "Ao confirmar que sua empresa não teve estoque
    em <ano>, esse período será excluído. Mas atenção: para empresas de comércio, a falta desse documento
    pode gerar inconsistências na declaração anual. Deseja confirmar?";
  - AFAC: "Confirmar envio do contrato". "Ao confirmar que já enviou o contrato, usaremos o documento
    recebido anteriormente como comprovante…";
  - intermediações: modal equivalente ("ModalConfirmarRelatorioIntermediacoes").
- **UI:** sucesso recarrega a página ou emite `excluiuPendencia`. Erro mostra "Não foi possível excluir o período. Por favor, tente
  outra vez."
- **Semântica:** fecha a pendência sem arquivo. **Reversível:** não. **Risco: médio/alto.**

### 4.3 `GET /documentos/tipos/init?area=DOCUMENTOS_CONTABEIS` (só leitura)

O único valor de `area` encontrado nos bundles é `DOCUMENTOS_CONTABEIS`. A CLI já tem essa leitura em `documentos.go`.

**Recomendação para a CLI (documentos):**

```text
ctbz documentos pendentes                       # envio-documento/init (sem ids → tiposPermitidos)
ctbz documentos enviar <arquivo> --pendencia <id> [--tipo …]          # envio-documento/enviar
ctbz documentos enviar <arquivo> --pendencia <id> --pendencia <id>…   # …/consolidado
ctbz documentos sem-arquivo --pendencia <id> --tipo ESTOQUE           # confirmação obrigatória
```

---

## 5. Certificado digital

Existem **três gerações** de fluxo convivendo no bundle. `GET certificado-digital/v2/verificar-rollout`
devolve `{certificadora}`, que decide a tela de upload. Os valores são:

- `SOLUTI`: fluxo antigo, rota `EnviarCertificado`;
- `CERTISIGN`: fluxo v2;
- `SAFEWEB`: fluxo v3.

### 5.1 Upload de certificado A1 (.pfx/.p12): três endpoints, todos multipart

| Certificadora | Método e caminho | Campos |
|---|---|---|
| SOLUTI (legado) | `POST certificado/envio-manual` (ausente do catálogo) | `file`, `password` |
| CERTISIGN (v2) | `POST certificado-digital/v2/upload-certificado` (ausente do catálogo) | `file`, `password` |
| SAFEWEB (v3) | `POST /certificado-digital/v3/upload` (ausente do catálogo, base em variável `"/certificado-digital/v3"`) | `file`, `password` |

- **Validação do front:**
  - o arquivo é obrigatório e precisa ter a extensão `/(\.pfx|\.p12)$/i` ("O arquivo deve ter a extensão pfx ou p12.");
  - tamanho máximo de **1 MB** ("O arquivo não pode ter mais que 1mb."; `h = 1e6` no v2);
  - a senha é obrigatória.
- **Resposta e erros:**
  - legado: devolve `{dataVencimento}`, gravado em `COMPANY_DATA.configuracao.dataValidadeCertificado`.
    O erro `certificado/InvalidTypeECPFException` mostra "Certificado inválido. Envie um certificado digital no formato e-CNPJ A1.";
  - v2: o erro `SENHA_INVALIDA_UPLOAD_CERTIFICADO` aparece no campo senha (`detalhes[0].detalhe`); outros
    `detalhes[0].detalhe` aparecem no campo arquivo; o erro genérico é "Ocorreu um erro durante o upload, tente novamente mais tarde.";
  - em caso de sucesso, a tela volta para "CertificadoDigital".
- v3: também tem `GET /certificado-digital/v3/init`, `/download` e `/senha`.
- **Semântica:** passa a usar esse certificado na empresa (emissão de NF, e-CAC etc.).
  **Risco: alto** (manipula a chave privada e a senha). É reversível só por novo upload.
- Para a CLI: a senha nunca deve ir para os argumentos (usar prompt ou stdin), e o arquivo deve ser validado localmente.

### 5.2 Remoção: `DELETE empresa/certificado/removercert`

- Existe no serviço (export `j`) e na action Vuex `certificadoDigital/removeCertificadoDigital`, que
  depois faz `setCertificadoDigital({situacao:"NAO_ENVIADO"})`. **Nenhum componente despacha essa
  action neste build (incerto: código morto ou chamado via outro app).** Não tem corpo.
- **Risco: alto** (remove o certificado da empresa; emissões automáticas param).

### 5.3 Fluxo v2 de aquisição (CERTISIGN), chunk `chunk-173ea0c1` (tela `/certificado-digital`)

Estado: `GET certificado-digital/v2/init` devolve `status` (`CRIANDO_SOLICITACAO`, `SOLICITACAO_CRIADA`,
`ATENDIMENTO_AGENDADO`, `ATENDIMENTO_REALIZADO`, `EMISSAO_INICIADA`, `CERTIFICADO_PRONTO_PARA_EMISSAO`,
`EMITIDO`, `ERRO_*`, `PENDENCIA_CNPJ_IRREGULAR`, `PENDENCIA_SOCIO_DIVERGENTE`, …) e os blocos
`dadosAquisicao`, `dadosAgendamento`, `dadosCertificado`, `documentos` e `possuiTicketPendenciaEmAberto`.

| Endpoint | Corpo | Semântica e UI | Risco |
|---|---|---|---|
| `POST certificado-digital/v2/salvar-aceite` | `{senhaCertificado: <senha criada pelo usuário>, confirmouSenhaSalvaLocalSeguro: true, estaRecriandoSenha: bool}` | **Inicia a aquisição.** O modal de termos exige "Declaro que li e concordo com os termos acima" (Termo de Titularidade Certisign) e, quando aplicável, "**Autorizo a cobrança do valor do certificado na minha próxima mensalidade**". Devolve `{status}`. Sucesso mostra "A senha foi cadastrada com sucesso." ou, ao recriar, "Seu processo foi reiniciado. Agende e refaça a entrevista para prosseguir.". O erro 400 `chaveErro` contendo `ERRO_INICIAR_SOLICITACAO` mostra "Emissão pausada". | **alto** (cobrança e segredo) |
| `PUT certificado-digital/v2/tipo-atendimento` | `{tipoAtendimento: "ONLINE"\|"PRESENCIAL", informouQueTemCNH?: bool}` | Muda a modalidade da entrevista. | baixo/médio |
| `POST certificado-digital/v2/agendar-atendimento` | `{dataHoraAgendamento: "<data> <hora>", posto: <objeto do ponto de atendimento> \| undefined}` | Agenda a videoconferência ou o atendimento presencial. Os horários vêm de `GET …/buscar-horarios-agendamento?data&postoId` e os postos de `GET …/pontos-atendimento/buscar?estado&cidade&bairro` (com `buscar-cidades` e `buscar-bairros`). A resposta é `{status, dadosAgendamento}`; `status:"HORARIO_INDISPONIVEL"` remove o horário da lista. O formato da data é incerto (provavelmente `YYYY-MM-DD HH:mm`). | médio |
| `GET certificado-digital/v2/reiniciar-atendimento` (**GET com efeito**) | — | "Deseja realizar uma nova entrevista?" (reagendar). Devolve `{status, dadosAtendimento}`. O erro 400 `FILA_EXPRESSA_BLOQUEADA` abre o modal "entrevista realizada". | médio |
| `GET certificado-digital/v2/iniciar-atendimento-expresso` (**GET com efeito**) | — | Entra na fila virtual. Devolve `{status, link, tipoAtendimento, tempoEspera}`. | médio |
| `PUT certificado-digital/v2/emitir-certificado-com-senha` | `{senhaCertificado}` | Emite após a entrevista. O status passa a `EMISSAO_INICIADA`. O erro 400 `SENHA_INVALIDA_EMISSAO_CERTIFICADO` mostra a mensagem de senha inválida. | alto |
| `POST certificado-digital/v2/solicitar-atendimento-pendencia` | sem corpo | Abre um ticket de suporte para a pendência (CNPJ irregular ou sócio divergente). Mostra o modal "Atendimento solicitado". | baixo |
| `PUT certificado-digital/v2/cancelar-solicitacao` | sem corpo | Usado no modal "Antes de continuar": "Ao confirmar que a pendência foi resolvida, você irá **recomeçar o processo** de emissão…", botão "Confirmar e recomeçar". Cancela a solicitação atual e recarrega o fluxo. | alto |
| `POST certificado-digital/v2/emitir-certificado` | `{usuario, senha, aceite}` | Tela "EmitirCertificado" (one-click SOLUTI no v2): o usuário digita as credenciais da certificadora (com aviso se tiver menos de 31 caracteres) e aceita as políticas de privacidade da Contabilizei e da Soluti. | alto |
| `POST certificado-digital/v2/informar-cnh` | `{temCNH: bool}` | Informa se o titular tem CNH (decide entre videoconferência e presencial). | baixo |

### 5.4 Fluxo legado (`certificado/*`, módulo `689a` em `app`)

| Endpoint | Corpo | Uso |
|---|---|---|
| `POST certificado/emitir-certificado` | `{usuario, senha, aceite}` | `fluxoEmitirCertificado`: emissão one-click (SOLUTI). |
| `POST certificado/fluxo-one-click/bird-id/valida-senha` | `{senha}` → `{isSenhaValida}` | Valida a senha nova (6–42 caracteres, com maiúscula, especial etc.). Recusada: "Não foi possível cadastrar sua senha. Digite novamente." |
| `POST certificado/fluxo-one-click/bird-id/emitir-certificado` | `{codigoOTP, senha}` | Emissão BirdID. O erro `certificado/codigoOTPInvalido` mostra "Token inválido". |
| `GET certificado/fluxo-one-click/alterarParametroEmpresa` (**GET com efeito**) | — | "AlertRemoveFromOneClick": tira a empresa do fluxo one-click. Erro mostra "Serviço não disponível!". |
| `PUT certificado/processo-aquisicao/etapa` | `{etapa}` ou `{etapa, sairFluxoBird: true}` | Máquina de estados do assistente. `etapa` ∈ `PRE_CHECKOUT`, `VERIFICACAO_COMPRA`, `VERIFICACAO_CNH`, `INFORMAR_DADOS`, `CADASTRO_SENHA`, `AGENDAMENTO_VIDEOCONFERENCIA`, `AGENDAMENTO_PRESENCIAL`, `AGENDAMENTO_REALIZADO`, `SELECIONAR_HORARIO`, `BIRDID_DADOS`, `EMISSAO`. Com `etapa` nula, o front só lê `etapaAtual`. |
| `POST certificado/agendamento-contabilizei` | `{horario: "<selectedDate> <selectedTime>"}` | Agenda o atendimento da Contabilizei (videoconferência). |
| `POST certificado/agendamento/dados-cliente` | `{telefone, email}` | Atualiza o contato do agendamento ("Atualizar"). Sem `await` no front. |
| `PUT certificado/agendamento-contabilizei/documento` (ausente do catálogo) | multipart `arquivo`, `tipo` (`"FRENTE"`/`"VERSO"`) | Envia a foto do documento de identidade do titular. **Dado pessoal sensível.** |
| `POST certificado/agendamento-contabilizei/senha` | `{senha}` | Cria a senha do certificado ("criar-senha"); em seguida `etapa=SELECIONAR_HORARIO`. |

**Riscos:** emissão, senha e documento: **alto**. Etapa e agendamento: **médio**. Sem undo, exceto reagendar ou reiniciar.

**Recomendação para a CLI (certificado):**

```text
ctbz certificado status                        # certificado/status + certificado-digital/v2/init (leitura)
ctbz certificado enviar empresa.pfx            # senha por prompt; escolhe o endpoint por verificar-rollout
                                               # (SOLUTI→envio-manual, CERTISIGN→v2/upload-certificado, SAFEWEB→v3/upload)
```

Não convém automatizar a aquisição e a emissão (aceite com cobrança, entrevista, documento com foto).
No máximo, `ctbz certificado agendamento` em modo leitura (horários e postos) e um link para a tela.

---

## 6. Escritas ausentes do `catalogo.md` (resumo)

| Método | Caminho | Motivo da ausência |
|---|---|---|
| DELETE | `/movimentacao-financeira/extrato?idContaBancaria&ano&mes` | helper `_()` de query |
| DELETE | `/movimentacao-financeira/desmembrar/desfazer/{id}?idLancamentoPai=` | helper `_()` |
| POST | `upload-documentos/extrato/enviar` | `axios({method,url})` |
| POST | `documentos/envio-documento/enviar` | `axios({method,url})` |
| POST | `documentos/envio-documento/enviar/consolidado` | `axios({method,url})` |
| POST | `certificado/envio-manual` | `axios({method,url})` |
| POST | `certificado-digital/v2/upload-certificado` | `axios({method,url})` |
| PUT | `certificado/agendamento-contabilizei/documento` | `axios({method,url})` |
| POST | `/certificado-digital/v3/upload` (+ GET `init`, `download`, `senha`) | base em variável |
| POST | `/integracao-bancaria/token` | caminho em variável |
| POST | `/integracao-bancaria/{id}/inativacao` | caminho em variável |
| GET | `/integracao-bancaria/widget-info`, `/integracao-bancaria/parametros-integracao/{id}` | caminho em variável (leitura) |

O `evento-tour` (base legado) já está no catálogo.

O catálogo não marca como escrita os GETs com efeito colateral `certificado-digital/v2/reiniciar-atendimento`,
`certificado-digital/v2/iniciar-atendimento-expresso` e `certificado/fluxo-one-click/alterarParametroEmpresa`.

## 7. Prioridade sugerida para o roadmap da CLI

1. **Caixa:** `adicionar`, `editar`, `remover` e `contas`. Payload simples, valor alto para o usuário e risco controlável.
2. **Extrato:** `importar` (OFX/PDF) com o fluxo completo e `classificar`, `desmembrar`, `desfazer-desmembramento`.
3. **Documentos:** `enviar` e `sem-arquivo` a partir das pendências da Central de Rotinas.
4. **Contas bancárias:** `adicionar` e `editar`. `remover` fica atrás de `--yes`.
5. **Certificado:** só `enviar` (upload A1) e `status`.
6. Não implementar (ou só com flag explícita): `reabrir-balanco`, `DELETE extrato`, aquisição e emissão de
   certificado, integração Open Finance, flags de UI (`interagiu-modal-integracao`, oferta BS2, `evento-tour`).
