# Escrita: notas fiscais, tomadores, notas de entrada, pró-labore e lucros

Análise **estática** dos bundles do painel (`/painel-de-controle/js/*.js`, Vue 2 + axios) e do
front de notas de entrada (`/nota-entrada/static/js/app.*.js`, Vue 1 + vue-resource). Nenhuma
requisição foi feita. Os bundles foram formatados com `prettier` para leitura; as referências
"chunk:linha" abaixo valem para a versão formatada (formatada com `prettier`, não versionada).

## 0. Convenções

| Base | Instância axios | Uso |
|---|---|---|
| `/api/plataforma/` | `X["c"]` (BFF) | quase tudo desta página |
| `/api/legado/` | `X["d"]` / `X["b"]` (monólito) | cancelamento de nota importada, anexos, notas tomadas |
| `/api/emissor/` | vue-resource absoluto | front de notas de entrada |

- Autenticação só por cookie de sessão (`withCredentials`); nenhum cabeçalho extra (exceto
  `strinfs-token` em `parametrosEmpresa/create` do front de notas de entrada).
- Corpo JSON salvo quando indicado `multipart/form-data`.
- Caminhos com `/` inicial no código (`"/novo-emissor/..."`) somam-se à base:
  `POST /api/plataforma/novo-emissor/...`.
- **Risco**: baixo = preferência/telemetria; médio = dado cadastral ou contábil reversível;
  alto = documento fiscal, obrigação legal ou valor financeiro/tributário.
- "(incerto)" marca o que não pôde ser confirmado só pelo código.

---

## 1. NFS-e: novo emissor

### 1.1 Mapa do fluxo de emissão (v2)

```text
GET  novo-emissor/listagem/init                → emissorEnabled, certificadoDigital{vencido…}, hasInstability
GET  novo-emissor/v2/versao-emissor            → qual emissor (v2 x legado)
GET  novo-emissor/v2/feature-flag              → listagem v2
GET  novo-emissor/tomadores/init               → tomadores[], emissaoSemTomador, permiteEmissaoExterior
GET  novo-emissor/tomador/{cpfCnpj|id}         → tomador escolhido (rota recebe documentoTomador)
GET  novo-emissor/v2/emissao/init              → regimeTributario (SIMPLES|LUCRO_PRESUMIDO), motorFatorR,
                                                 exibirNovaExperienciaNbsCclass, exibirNovaExperienciaMoedaEstrangeira,
                                                 exibirAssistenteNBS, aliquotas, codIbgePrestador,
                                                 enderecoTomadorPFObrigatorio, exibirOfertaCtbzBank
GET  novo-emissor/emissao/atividades           → cnaes[] {isMainService, listaCaracteristica,
                                                 cnaeValues[{dadosParaEmissao{idCnaeEmpresa, codigoCnae}}]}
GET  novo-emissor/v2/trilhas-empresa[?exportacaoServico=true]
                                               → "trilhas": [{cnae, principal, preferencia,
                                                 listaCaracteristica[{idCnaeEmpresa}],
                                                 listaCodigoNacionalOuItemServico[{codigo, tipoExcecao (OBRA|EVENTO|null),
                                                   listaCTribMun[{codigo}], listaNbs[{codigo, listaCclass[{codigo}]}]}]}]
GET  novo-emissor/v2/emissao/atividades/cindop?codigoNacionalItemServico=&nbs=[&exportacaoServico=true]
                                               → [{codigo}] (indicador de operação, IBS/CBS)
GET  novo-emissor/v2/emissao/atividades/codigo-municipal?codigoIbge=&codigoNacional=
GET  novo-emissor/v2/emissor-async             → boolean: permite emissão assíncrona
GET  banco-central/moeda ; banco-central/moeda/{moeda}/data-referencia/{data} → {saleRate} (tomador exterior)
POST novo-emissor/v2/calculo/resumo-impostos   → cálculo (só leitura, ver 1.3)
POST novo-emissor/v2/emissao/emitir | emitir-async   ← EMISSÃO
```

Arquivo: `NovoEmissorNF.b75ee2a3.js` (serviço no módulo `0f71`, montagem do corpo nas
funções `rs`/`cs`, ~linhas 20480–20900). O catálogo não tem esses caminhos porque ficam em
variáveis.

### 1.2 Emitir NFS-e

| | |
|---|---|
| **Método/caminho** | `POST /api/plataforma/novo-emissor/v2/emissao/emitir?origem=EMISSOR_SIMPLIFICADO_EMISSAO` |
| **Variante assíncrona** | `POST /api/plataforma/novo-emissor/v2/emissao/emitir-async?origem=EMISSOR_SIMPLIFICADO_EMISSAO`. Usada quando `tipoEmissor == "EMISSAO"`, não é replicação e `GET v2/emissor-async` devolveu `true`. Mesmo corpo. |
| **Tela** | `#/emissor/...` rota `emitir-nota` (chunk `NovoEmissorNF`), botão **"Confirmar e emitir"** |
| **Risco** | **alto**: gera documento fiscal na prefeitura. Desfazer = cancelar (1.6), com prazo municipal e custo após o dia 5 do mês seguinte |

**Corpo** (função `rs` + complementos de `b()`):

| Campo | Tipo | Origem / regra |
|---|---|---|
| `cpfCnpj` | string \| null | Documento do tomador só com alfanuméricos, em maiúsculas (pronto para CNPJ alfanumérico). Tomador estrangeiro: `tomador.id`. Emissão sem tomador: `null`. |
| `discriminacao` | string | Descrição digitada (obrigatória, máx. **1400** caracteres). Se houver complementos (`complementoDescricaoNota`), anexa `"\n---------------\n"` + linhas. |
| `enviarEmailTomador` | bool | Checkbox de envio ao cliente **e** tomador com e-mail. |
| `idCnaeEmpresa` | string | `selectedCaracteristica.idCnaeEmpresa` (das atividades/trilhas). Padrão: atividade principal com uma única característica. |
| `valorServico` | number | Valor (> 0, senão "Informe um valor para o serviço."). |
| `issRetido` | bool | Toggle de ISS retido. |
| `aliquotaIssRetido` | number \| null | Com ISS retido: `parseFloat(aliquotaISS.replace(",", "."))`. Validação: "A alíquota deve ser de 2% a 5%." |
| `inscricaoMunicipal` | string \| null | Inscrição municipal do tomador (enviada quando há alíquota de ISS). |
| `valorIss` | number \| null | Com ISS retido: valor calculado. |
| `incidenciaExterior` / `servicoExportacao` | bool | `tomador.estrangeiro && isForeignService == "1"` (os dois iguais). |
| `codigoNbs`, `codigoCclass`, `codigoNacionalServico` | string | Só na "nova experiência NBS/cClass" (reforma tributária). Vêm da trilha escolhida: `listaCodigoNacionalOuItemServico[].codigo` → `listaNbs[].codigo` → `listaCclass[].codigo` (cClass só se `exibirCclass`). |
| `cTribMun` | string | Código tributário municipal, da trilha (`listaCTribMun`). Obrigatório quando o código nacional tem exceção municipal. |
| `cIndOp` | string | Indicador de operação, do `GET .../cindop`. Obrigatório na nova experiência ("Selecione uma indicação da operação."). |
| `definirNovaTrilhaPreferencial` | true | Só presente quando o usuário marca a trilha como preferida. |
| `descontoIncondicionado` | number | Padrão 0; não pode passar do valor ("O desconto não pode ser maior que o valor do serviço."). |
| `naturezaOperacao` | string | Só se a tela calcular `naturezaOperacaoParaEnvio` (incerto: valores por município). |
| `moeda`, `cotacao`, `valorMoedaEstrangeira`, `dataInvoice` | | Só tomador estrangeiro. Recebimento em BRL: `moeda:"BRL"`, `cotacao:"1.0000"`, `valorMoedaEstrangeira = valorServico`. Em moeda estrangeira: `cotacao = saleRate` da API do BCB, `valorMoedaEstrangeira` digitado, `dataInvoice:"AAAA-MM-DDT12:00:00"`. |
| `localPrestacao` | `{uf, codIbgeMunicipio, nomeMunicipio}` | Quando a tela mostra "local de prestação". |
| `obra` | `{tipoIdentificacao: CIB\|CODIGO_OBRA\|ENDERECO, cib, codigoObra, endereco}` | Quando o código de serviço tem `tipoExcecao` de obra. `endereco` nacional `{pais:"BR", cep, logradouro, numero, semNumero, complemento, bairro, uf, codIbgeMunicipio, nomeMunicipio}` ou exterior `{pais, cidade, nomeMunicipio, regiao, cep, …}`. |
| `evento` | `{nome, dataInicio, dataFim ("AAAA-MM-DDT12:00:00"), endereco}` | Quando `tipoExcecao == "EVENTO"`. |
| `irRetido`, `aliquotaIr`, `valorIr`, `inssRetido`, `valorInss`, `pisRetido`, `valorPis`, `cofinsRetido`, `valorCofins`, `csllRetido`, `valorCsll`, `valorLiquidoNfse` | bool/number | **Só no Lucro Presumido**. Os valores vêm do cálculo `resumo-impostos` (1.3); valor nulo quando a retenção está desligada. |

- **Não há campo de data de competência nem de data de emissão** na emissão: a competência é a
  da data de emissão. Os textos da tela confirmam: *"Você só paga os impostos dessa nota no
  próximo mês"* e *"Se você emitir uma nota errada, basta cancelar e emitir uma nova dentro do
  mesmo mês do serviço prestado."*
- Outras validações do front: "Selecione um CNAE.", "Selecione um código de serviço.",
  "Selecione um NBS.", "Selecione um cClass." (se obrigatório), "Selecione uma característica."
  (quando a atividade tem duas ou mais), "Informe a descrição do serviço.".
- **Resposta**: o front ignora o corpo de sucesso. Em sucesso mostra *"Estamos processando a
  nota fiscal"* / *"Assim que a nota for autorizada pela prefeitura você receberá o arquivo
  por e-mail."*. Em erro lê `response.data.{identificador, mensagemAmigavel | message}` e mostra
  **"Falha na emissão"** (padrão: *"Não foi possível emitir a nota fiscal. Tente novamente mais
  tarde."*), com botões "Editar cadastro do cliente" / "Voltar para dados da nota". Em ambos os
  casos chama `notas/notasInit` (relê a listagem).

### 1.3 Cálculo de impostos (POST sem efeito)

`POST /api/plataforma/novo-emissor/v2/calculo/resumo-impostos`. Chamado com debounce e
cancelamento a cada mudança no formulário. Corpo:
`{valorServico, idCnaeEmpresa, issRetido, aliquotaIssRetido|null, tomadorPessoaJuridica,
incidenciaExterior, pisRetido, cofinsRetido, csllRetido, inssRetido, irRetido, aliquotaIr|null,
codigoNacionalServico|null, tomadorExterior, [cTribMun, codIbgeLocalPrestacao]}`.
Responde valores de ISS, retenções, base de cálculo e valor líquido. **Só leitura**, sem risco:
serve para um `ctbz notas emitir --dry-run`.

### 1.4 Replicar ("emitir de novo" a partir de uma nota)

| | |
|---|---|
| **Caminho** | `POST /api/plataforma/novo-emissor/v2/emissao/replicar?origem=EMISSOR_SIMPLIFICADO_REPLICACAO` |
| **Corpo** | igual ao 1.2 **+ `idNotaReplicada`** (id da nota de origem) |
| **Pré** | `GET novo-emissor/v2/emissao/init?idNotaReplicada={id}` → `dadosReplicacao{tomador, idCnaeEmpresa, valorServico, discriminacao, issRetido, aliquotaIssRetido, servicoPrestadoExterior, nbs, cclass, codigoNacionalServico, moeda, cotacao, valorMoedaEstrangeira, dataInvoice, localPrestacao…}` para preencher o formulário |
| **Tela** | listagem → botão "Replicar" (só se `permiteReplicacao`), rota `/emissor/notas/:hash` com `hash = btoa(idNota)` e `?idNotaReplicada=` |
| **Risco** | **alto**: é uma emissão nova (não há "duplicar como rascunho") |

### 1.5 Registro de nota emitida fora (tipoEmissor = REGISTRO)

Para empresas que emitem no site da prefeitura e só lançam a nota no Contabilizei.

- `POST /api/plataforma/novo-emissor/v2/registro-de-notas/registrar`: corpo do 1.2 **+**
  `dataEmissao`, `numeroNota`, `competencia` (as duas datas em `"AAAA-MM-DDT12:00:00"`, não
  futuras: "A data de emissão não pode ser futura.", "A competência não pode ser futura.").
- `POST /api/plataforma/novo-emissor/v2/registro-de-notas/replicar`: o mesmo **+**
  `idNotaReplicada`. Pré-carga: `GET v2/emissao/init?idNotaReplicada={id}&isRegistro=true`.
- Sucesso: *"Nota registrada com sucesso!"*. Risco **médio** (lançamento contábil e fiscal;
  as notas "Registrada" aparecem na listagem e podem ficar "Cancelada").

### 1.6 Emissor legado (v1)

Usado quando `versao-emissor` não é v2 (chunk `notas`, serviço `0de2` no `app`):
`POST /api/plataforma/novo-emissor/emissao/nova-nota?origem=EMISSOR_SIMPLIFICADO_EMISSAO` e
`POST /api/plataforma/novo-emissor/emissao/replicar/{idNota}?origem=EMISSOR_SIMPLIFICADO_REPLICACAO`.
Corpo: `{cpfCnpj, discriminacao, enviarEmailTomador, idCnaeEmpresa, valorServico, issRetido,
aliquotaIssRetido, inscricaoMunicipal, incidenciaExterior, [codigoNbs, codigoCclass,
codigoNacionalServico], [moeda, cotacao, valorMoedaEstrangeira, dataInvoice], [naturezaOperacao,
descontoCondicionado, descontoIncondicionado, valorCofins, valorCsll, valorIr, valorIss,
valorLiquidoNfse, valorPis, servicoExportacao]}`. Pré: `GET novo-emissor/formulario-emissao/init[?idNotaReplicada=]`.
A CLI deve atender só o v2 e recusar o legado.

### 1.7 Como o front sabe que a prefeitura autorizou

**Não há polling.** Depois do POST o front só mostra "processando" e relê a listagem. O estado
vem da listagem `GET novo-emissor/v2/listagem/notas/filtro?pagina=&limite=&ano=&mes=`, item a
item:
`situacaoNota` (ex.: `PROCESSADO_SUCESSO`, `CANCELADA`), `statusNotaFiscal{text, badgeType}`
(rótulo pronto do servidor), `errosNotaFiscal[{descricao}]` (popover "detalhes" com o erro da
prefeitura), `numero` (só aparece com `PROCESSADO_SUCESSO`) e `permiteReplicacao`. O comprovante
chega por e-mail. Para a CLI: depois de emitir, consultar a listagem do mês em intervalos até
`numero` aparecer ou `errosNotaFiscal` vir preenchido. O POST não devolve id confiável (incerto),
então use tomador, valor e horário para achar a nota.

### 1.8 Cancelar NFS-e emitida pelo emissor

| | |
|---|---|
| **Caminho** | `POST /api/plataforma/novo-emissor/emissao/cancelar/{idNota}` |
| **Corpo** | **nenhum**. Não há motivo, código de cancelamento nem justificativa. |
| **Pré** | `GET /api/legado/nota{perfil.codigo}/obtercodverificacao/{codigoNota}` → objeto com `idNota` (e dados mostrados: código de verificação, tomador). `perfil.codigo` vem do `COMPANY_DATA` em localStorage (base64). O caminho resultante é provavelmente `notafiscal/obtercodverificacao/…` (incerto). O `idNota` deve ser o `id` da listagem (incerto). |
| **Tela** | rotas `#/cancelamento-nota-emitida/inicio/:codigoNota` → `detalhes` → `finalizacao` (no `app`, ~29780–30250). A entrada vem do front antigo (`/sistema`, "consultar notas"); a listagem v2 não tem botão de cancelar. |
| **Textos** | Início: *"Notas canceladas depois do dia 5 do próximo mês possuem um custo de operação contábil."* e *"O tempo de cancelamento da nota fiscal pode demorar um pouco em certas prefeituras."*. Confirmação: *"Tem certeza que deseja cancelar a seguinte nota fiscal?"*, botão "Confirmar cancelamento". Sucesso: toast *"Nota cancelada com sucesso"* + *"Sua solicitação de cancelamento foi registrada com sucesso."* (assíncrono na prefeitura). |
| **Erro** | Se `response.data.error.errorClassName` contém `CancelamentoPrefeituraException`, mostra `response.data.error.errorMessage` (mensagem da prefeitura); senão "Falha ao cancelar nota". |
| **Risco** | **alto** e irreversível (nota cancelada não volta; corrigir = emitir outra). Pode gerar custo. |

**Cancelamento de nota importada da prefeitura (legado, GET com efeito colateral)**

- `GET /api/legado/notafiscal/cancelarAPI?notas={numero};{numeroNotaSubstituta|0};servico`
  (montado por `Hn()`: `numero;substituta;tipoNota`; vários itens por vírgula é incerto).
- Pré: `GET /api/legado/notafiscal/buscarNotaEmpresa?numero={numero}` → `{nota:{…}}`.
- Fluxo `#/cancelamento-nota-importada/…`: *"Primeiro faça o cancelamento no site da sua
  prefeitura"* + checkbox **"Confirmo que cancelei a nota na prefeitura"**. Só baixa a nota na
  contabilidade, sem falar com a prefeitura. Prazo mostrado: até o dia 5 (Simples) ou dia 1
  (demais regimes) do mês seguinte.
- Erro: `response.data.errorMessage` (caminho diferente do fluxo anterior). Risco **médio/alto**
  (contábil) e é **GET**: nunca chamar em sondagem.

### 1.9 Agendamento, recorrência e rascunho: não existem

- `novo-emissor/rollout/agendamento` → `{permiteAgendamento}` controla o card *"Precisa de ajuda
  na emissão da sua 1ª nota fiscal?"* (**agendar reunião** com especialista), não agendamento
  de nota.
- Os chunks `agendamentoEmissao`/`agendamentoRealizado` são do **certificado digital** (Soluti).
- "Rascunho" é só estado local: `loadSnapshot()` na store Vuex e `localStorage["selected-tomador"]`.
  Não há endpoint de rascunho, duplicação sem emitir, nota recorrente nem emissão programada.

### 1.10 Outros endpoints de escrita ligados a notas

| Método/caminho | Corpo | Semântica | Risco |
|---|---|---|---|
| `POST /api/plataforma/novo-emissor/assistente-nbs/session` | `{stateVariables:{trilha:[…], trilha_selecionado:{cnae, codigoNacionalServico, cTribMun, nbs, …}}}` (objetos internos enviados como string JSON) | Abre sessão do **assistente de IA de NBS** (ChatKit); responde `{clientSecret}`. Há também um `GET` da mesma rota (simulação). | baixo (sem efeito fiscal); inútil para a CLI |
| `POST /api/plataforma/dashboard/v2/notas-fiscais/salvar-clique-emitir` | nenhum | Registra que o usuário clicou "Emitir" no card do dashboard (flag `clicouEmitir`), quando não há certificado ou o registro é manual | baixo (telemetria) |
| `PUT /api/legado/notafiscal/salvaranexosprincipais` | `[ {cnae:{id, descricao, codTabelaSimples…}, anexos:[{codTabelaSimples, anexoPrincipal: bool, …}]} ]` (um CNAE por chamada) | Tela "Configurar atividades da nota fiscal" (`CnaesParaSelecao`): escolhe o **anexo do Simples principal** por CNAE para notas importadas. Pré: `GET /api/legado/notafiscal/cnaeanexosmultiplos/list`. "Reverter" manda o mesmo item com todos `anexoPrincipal:false`. | **médio/alto** (muda a tributação) |
| `PUT /api/plataforma/experts/cashflow/notas-fiscais` | `{cnpj, nfsId, expectedPaymentDate: ISO\|null, paymentDate: ISO\|null}` | Fluxo de caixa (tela "experts"): informa data prevista e data de recebimento de uma NF. Pré: `GET experts/cashflow/notas-fiscais?mes=&ano=` | baixo/médio (gerencial) |

### 1.11 Notas tomadas (legado, ainda no painel)

O front `/nota-tomada/` morreu, mas os chunks `registrar-nota` e `listagem-notas` do painel
usam `/api/legado/notafiscaltomada/*`:
`POST notafiscaltomada/verificarNotaTomadaExiste` → `{exists, mes, ano}` (mensagem *"Nota tomada
já registrada na competência de MM/AAAA!"*), `POST notafiscaltomada/salvarnota` com
`{dataEmissao (timestamp), descricaoServico, numeroNota, numeroSerie|null, prestador, servico,
categoria, valorTotalServico, valorCofins:0, valorCsll:0, valorIr:0, valorPagar:0, valorPis:0}`,
`DELETE notafiscaltomada/removernota/{id}`. Pré: `listarservicos`, `listarcategorias`,
`registroempresa/localizarempresa/{cnpj}`, `validarmesinclusao/{…}`. Fora do catálogo. Risco médio.

### Recomendação para a CLI (notas)

```text
ctbz notas simular  --tomador <doc|id> --valor 1500 --atividade <idCnaeEmpresa> [--iss-retido 2%]   # resumo-impostos
ctbz notas emitir   --tomador <doc|id>|--sem-tomador --valor 1500.00 --descricao "..." \
                    --atividade <idCnaeEmpresa> [--codigo-servico <codNacional> --nbs <nbs> --cclass <c> \
                    --ctribmun <c> --cindop <c>] [--iss-retido --aliquota-iss 2] [--exterior --moeda USD --valor-moeda 300 --data-invoice 2026-10-01] \
                    [--enviar-email] [--dry-run] [--yes] [--aguardar]
ctbz notas replicar <idNota> [--valor ...] [--yes]
ctbz notas cancelar <idNota> [--yes]          # confirmar sempre; avisar custo após o dia 5 do mês seguinte
```

- Defaults iguais ao front: atividade principal, trilha `preferencia`/`principal` de
  `trilhas-empresa`; `cIndOp` automático quando `cindop` devolve só uma opção.
- `--dry-run` = `resumo-impostos` + corpo montado, sem POST. A emissão **exige `--yes` ou
  confirmação interativa** mostrando tomador, valor, atividade, retenções e líquido.
- `--aguardar`: ler a listagem do mês até `numero` ou `errosNotaFiscal`.
- Sem replicar ou cancelar em lote. Não implementar `cancelarAPI` nem `registro-de-notas` na 1.x.

---

## 2. Tomadores (clientes)

Serviço `0de2` (`app.9feb0a2a.js` ~1313–1580), tela `clients.65fc33d5.js`.

| Método/caminho | Corpo | Notas |
|---|---|---|
| `POST /api/plataforma/novo-emissor/clientes/salvar-cliente-nacional` | `{cpfCnpj: só dígitos (CNPJ alfanumérico em maiúsculas), razaoSocialOuNome, telefone: ""\|str, email: ""\|str, inscricaoMunicipal: str\|null (PF: sempre null), endereco:{bairro, cep: dígitos, codIbge, complemento, logradouro, estado: UF, numero, cepInvalido: bool}}` | **Cria e também edita**: não manda `id`, a chave é o documento (upsert). Resposta `{cpfCnpj, …}`; o front segue para `emitir-nota` com `documentoTomador = cpfCnpj`. |
| `POST /api/plataforma/novo-emissor/clientes/salvar-cliente-exterior` | `{razaoSocialOuNome, email, endereco:{complemento, logradouro, numero, cidade, pais, descricaoPais, simboloPais}, id? (só na edição)}` | Resposta com `cpfCnpj` (para estrangeiro é o identificador interno, incerto). A emissão usa `tomador.id` como `cpfCnpj`. |
| `DELETE /api/plataforma/autopilot/clientes/{documento\|id}` | — | Excluir cliente. `{documento}` = CNPJ ou CPF sem máscara; estrangeiro usa `id`. Bug no front: `delete(a).data` não espera a promise, mas a requisição sai. |

- Pré-leituras: `GET novo-emissor/novo-cliente/init`,
  `GET novo-emissor/cadastro-clientes/{clientId}` (204 = *"Cliente não encontrado."*; 200 →
  `{tipoCliente: NACIONAL|EXTERIOR, clienteNacionalDTO|clienteExteriorDTO}`),
  `GET novo-emissor/clientes/consulta/{cnpj}` (Receita; sem dados mostra *"Nenhum dado foi
  encontrado para este CNPJ. Por favor, preencha os campos manualmente."*),
  `GET novo-emissor/cep/logradouro?cep=`, `GET notafiscal/emitir/buscarPaisesParaEmissao/`.
- Textos: erro *"Cliente não cadastrado. Não foi possível salvar os dados do cliente. Tente
  novamente."*; sucesso, já na tela de emissão: *"Cadastramos o novo cliente. Os dados dele
  ficarão salvos para as próximas emissões."* / *"As informações deste cliente foram
  atualizadas e você já pode emitir a nota fiscal."*; exclusão: *"Deseja mesmo excluir este
  cliente?"*, *"Esta ação não pode ser desfeita."*, *"… Não faz mais parte da sua lista de
  clientes"*.
- **O que é "autopilot"**: nome antigo do serviço de clientes do BFF (há um front `/autopilot/#/`
  e um "Piloto Automático" na integração bancária). O `app` ainda tem uma store Vuex morta com
  `GET /autopilot/clientes` (lista), `GET /autopilot/clientes/{cnpj}`,
  `GET /autopilot/clientes/consulta/{cnpj}` e **`POST /autopilot/clientes`** com corpo
  `{cpfCnpj, razaoSocial, endereco:{uf, municipio, cep, logradouro, bairro, numero, complemento,
  codigoIbge, pais, cepInvalido}, contato:{email, telefone}, exterior}`. Nenhuma tela despacha
  `saveCustomer`; só o DELETE continua em uso.
- Reversibilidade: excluir não tem desfazer (recadastrar). Notas já emitidas não mudam (incerto).
  Risco **baixo/médio**: dados cadastrais que vão para notas futuras.

### Recomendação para a CLI (tomadores)

```text
ctbz tomadores criar   --cnpj 00000000000000 [--im ...] [--email ...]   # preenche pela consulta da Receita + CEP
ctbz tomadores criar   --exterior --nome "..." --pais US --cidade ... --logradouro ...
ctbz tomadores editar  <documento> --email ...      # GET cadastro-clientes + merge + POST salvar-cliente-*
ctbz tomadores excluir <documento|id> [--yes]
```

---

## 3. Notas de entrada (NF-e de compra), base `/api/emissor/`

Front `ne/app.f55c7140b16fca43db90.js`; serviços em ~3165–3345, telas em ~1760–2720.

### 3.1 Manifestação do destinatário

| | |
|---|---|
| **Caminho** | `POST /api/emissor/notasentrada/manifestar/` |
| **Corpo** | `{tipoManifestacao: "CIENCIA"\|"CONFIRMACAO"\|"DESCONHECIMENTO"\|"NAO_REALIZADO", origem: "PLATAFORMA", idNotas: [id, …], justificativa?: string}` (`justificativa` só em `NAO_REALIZADO`) |
| **Filtro do front** | só entram notas com `situacao.id` `PENDENTE` ou `CIENCIA`; sem nenhuma: *"Não é possível alterar o status das notas selecionadas."* |
| **Resposta** | array `[{id, situacao{id, …}}]` que o front aplica nas linhas; toast *"Estamos comunicando a Manifestação para a Receita Federal. Em alguns minutos a Nota Fiscal estará manifestada."* |
| **Pré** | `GET notasentrada/listar/0?mes&ano&empresa&qtdPagina&cursor` (ids e situação) |
| **Tela** | `#/manifestacao` (botões "Informar recebimento", "Não recebi", "Desconheço"). Diálogo "Informe a justificativa" com *"Por favor, escreva o que aconteceu para você não ter recebido os produtos…"* e exemplo *"Mercadoria foi extraviada antes da entrega"*. |
| **Ciência automática** | Baixar XML ou DANFE de nota `PENDENTE` pede *"Declara ciência da nota fiscal para baixar?"* e manda `CIENCIA` antes do download. Download permitido em `PENDENTE`, `CIENCIA` e `CONFIRMADA`; `GET notasentrada/download/{id}` responde `{xml: "<…>"}`. |
| **Risco** | **alto**: evento fiscal enviado à SEFAZ. Os eventos conclusivos (confirmação, desconhecimento, operação não realizada) não têm desfazer no front. Pela regra da SEFAZ (não do código), a justificativa de "operação não realizada" deve ter 15–255 caracteres; o front não valida, a CLI deve validar. |

Variante na aba de classificação ("Desconheço / Não recebi"):
`POST /api/emissor/notasentrada/manifestar` (**sem** barra final) com
`{chaves:[chave], idNotas:[id], justificativa, cnpj: <cnpj da empresa>, tipoManifestacao:
"DESCONHECIMENTO"|"NAO_REALIZADO", origem:"PLATAFORMA"}`. Sucesso: *"Desconhecimento efetuado com
sucesso!"*; erro: *"Erro ao efetuar o desconhecimento"*.

### 3.2 Classificação

| Método/caminho | Corpo | Semântica |
|---|---|---|
| `POST /api/emissor/classificacaonotas/salvarloteclassificacao/{tipo}` | **array das notas selecionadas** (os objetos da listagem `tipo=0`, como vieram) | Classifica a nota inteira. `tipo` ∈ `ESTOQUE`, `INSUMO`, `USO_CONSUMO`, `ATIVO_IMOBILIZADO`, `PRESTACAO_SERVICO`. Sucesso *"Nota classificada com sucesso!"*; erro *"Erro ao classificar nota: " + bodyText*. |
| `POST /api/emissor/classificacaonotas/salvarclassificacao/` | array de produtos de `GET classificacaonotas/listarprodutos/{idNota}`: `[{id, descricao, ncm, valor, quantidadeTotal, quantidadeEstoque, quantidadeInsumo, quantidadeAtivo, quantidadeConsumo, quantidadePrestacao}]` | "Por produto". Regras: nenhuma quantidade negativa (*"A classificação não pode ser negativa!"*) e soma = `quantidadeTotal` (*"Quantidade total deve ser totalmente distribuida entre as opções para cada produto!"*). Sucesso *"Classificação efetuada com sucesso!"*. |
| `POST /api/emissor/classificacaonotas/reclassificar?idNfe={idNota}` | mesmo array de produtos | Reclassifica nota já classificada. Atalho do front: tudo num balde (`ESTOQUE`, `INSUMO`, `ATIVO`, `CONSUMO`, `PRESTACAO_SERVICO` → zera os outros, põe `quantidadeTotal` num só). A nota passa a `tipoClassificacao: "PROCESSANDO"`; toast *"Reclassficação da nota para: X está sendo processada"*. |

- **Origem das opções**: fixas no front, sem endpoint. Note os nomes diferentes por nível:
  lote usa `USO_CONSUMO`/`ATIVO_IMOBILIZADO`, produto usa `quantidadeConsumo`/`quantidadeAtivo`
  e o atalho de reclassificação usa `CONSUMO`/`ATIVO`.
- Prazo (texto do painel): *"Você tem até o dia 05 do mês seguinte à emissão da Nota Fiscal de
  Entrada para classifica-la, após esse prazo, o sistema irá confirmar a pré-classificação
  sugerida."*.
- **Versão nova no painel** (chunk `nota-entrada.687b417a.js`, rota `#/nota-entrada/classificacao`):
  `GET /api/plataforma/nota-entrada/classificacao/init/`,
  `GET /api/plataforma/nota-entrada/classificacao/listar/?mes&ano&empresa&offset&limite=10&tipo=0`,
  `POST /api/plataforma/nota-entrada/classificacao/salvarloteclassificacao/{tipo}` (array das
  notas marcadas). No painel só o lote fica ativo; "Por Produto" e "Desconheço / Não Recebi" vêm
  desabilitados. Nenhum desses caminhos está no catálogo.
- Risco **médio** (contábil, reclassificável).

### 3.3 Outros com efeito

- `GET /api/emissor/classificacaonotas/agendarExportarArquivos/{mes}/{ano}?tipo=&email=`: agenda
  e-mail com os XMLs (GET com efeito colateral).
- `POST /api/plataforma/parametrosEmpresa/create` `{cnpj, chave, valor}` e `…/update`
  `{id, cnpj, chave, valor}`: parâmetros da empresa (ex.: tutorial visto). Baixo risco.

### Recomendação para a CLI (notas de entrada)

```text
ctbz entradas manifestar <id>... --ciencia|--confirmar|--desconhecer|--nao-realizada --justificativa "..." [--yes]
ctbz entradas classificar <id>... --como estoque|insumo|uso-consumo|ativo-imobilizado|prestacao-servico [--yes]
ctbz entradas produtos <id>                       # listarprodutos (leitura)
```

Validar a justificativa (15–255) e o filtro de situação (`PENDENTE`/`CIENCIA`) antes do POST.
Deixar classificação por produto e reclassificação para depois.

---

## 4. Pró-labore

Serviços: central `b7c5` (`app` ~23700–23930), assessor `7a0e` (`app` ~21680–22150), init e
motor `1ecf` (`chunk-d4950f1a`), dashboard (`chunk-b88c6b8a`), questionário
(`informe-de-rendimentos`).

### 4.1 Central do sócio (`#/socio/central`, `#/socio/editar-gestao/:socioId`)

| Método/caminho | Corpo | Semântica / tela | Risco |
|---|---|---|---|
| `PUT /api/plataforma/prolabore/central/gestao/{socioId}` | `{tipoGerenciamento, valorProlabore, valorProlaboreMinimo, modoEdicaoProlaboreMin, sairGestaoInteligente}` | **Alterar o pró-labore de um sócio** (`chunk-e3a42f36` ~1715–1790). | **alto** (INSS/IRRF e Fator R) |
| `PATCH /api/plataforma/prolabore/central/empresa/zerar-prolabore` | `{zerarProlabore: bool}` | Chave da empresa *"Não quero ter pró-labore cadastrado em meses sem faturamento."* Com mais de um sócio pede confirmação. Erro *"Não foi possível zerar o pró-labore. Tente novamente mais tarde."* Reversível (`false`). | alto |
| `PUT /api/plataforma/prolabore/central/empresa/gestao-inteligente` | — | Ativa a "Gestão Inteligente" (motor) para a empresa. Erro *"Não foi possível ativar a gestão inteligente…"* | alto |

Detalhes do `PUT gestao/{socioId}`:

- `tipoGerenciamento` ∈ `INTELIGENTE` | `SALARIO_MINIMO` | `TETO_INSS` | `PERSONALIZADO`
  (rótulos: "Gestão Inteligente", "Salário Mínimo", "Teto INSS", "Personalizado").
- `valorProlabore`: `TETO_INSS` → `valorMaximoProlabore` do GET; `SALARIO_MINIMO` →
  `salarioMinimo` do GET; `PERSONALIZADO` → valor digitado ("Pró-labore desejado", validação
  *"Valor deve ser maior ou igual ao salário mínimo vigente: …"*); `INTELIGENTE` → o valor
  passado (normalmente `null`).
- `valorProlaboreMinimo`: piso opcional na gestão inteligente ("Definir valor mínimo de
  pró-labore" / "Confirmar sem valor mínimo").
- `modoEdicaoProlaboreMin`: `true` se já havia mínimo.
- `sairGestaoInteligente`: `tipo != INTELIGENTE && qtdSocioGestaoInteligente <= 1`. Com mais
  sócios na GI aparece o modal *"Sua empresa permanecerá na Gestão Inteligente"* / "Confirmar
  alteração para Tradicional". Sem `qtdSocioGestaoInteligente` no GET, a CLI recusa antes de
  enviar em vez de assumir um valor (ADR-0021).
- Pré: `GET prolabore/central/init` (sócios e ids), `GET prolabore/central/gestao/{socioId}` →
  `{tipoGerenciamentoProlabore, valorProlaboreMinimo, valorMaximoProlabore, salarioMinimo,
  qtdSocioGestaoInteligente, elegivelNoMotor, …}`.
- Sucesso volta para `central-socios`; erro *"Não foi possível salvar as alterações. Tente
  novamente mais tarde."*.

### 4.2 Motor (cálculo automático), dashboard e questionário

| Método/caminho | Corpo | Semântica | Risco |
|---|---|---|---|
| `PUT /api/plataforma/prolabore/excluir-empresa-motor-fator-r` | — | "Sair do cálculo automático". Modal: *"Você não faz mais parte do cálculo automático de pró-labore — A partir de agora, você deve ajustar o valor do seu pró-labore todos os meses. O ajuste deve ser feito até o penúltimo dia do mês para ser aplicado no mês seguinte."* Desfazer: reativar gestão inteligente ou questionário (incerto). | alto |
| `POST /api/plataforma/prolabore/viu-dialog-informacao-prolabore` | — | "Não mostrar novamente" do aviso *"O ajuste no pró-labore deve ser feito até o dia 25 do mês para ser aplicado no mês seguinte…"* | baixo |
| `PATCH /api/plataforma/prolabore/central/empresa/zerar-prolabore-gt` | — | Painel lateral "OfertaZeramentoGTAside" (empresa em GI sem faturamento): *"Seu pró-labore será removido — Sua preferência já começa a valer a partir deste mês."* / *"Você escolheu não ter pró-labore a partir deste mês… Essa configuração não altera os meses passados."* Desfazer: provavelmente `zerar-prolabore {false}` (incerto). | alto |
| `POST /api/plataforma/dashboard/v2/prolabore/salvar-decisao` | `{acao: "MANTER"\|"REMOVER"}` | Card *"Sem faturamento este mês — Você pode manter o pró-labore para contribuir com o INSS, ou remover e economizar nos impostos."* Registra a decisão; se de fato remove é incerto (o front também liga `desejaAlterar`). | médio/alto |
| `POST /api/plataforma/questionario-preferencia-prolabore/salvar-resposta` | o item da pergunta vindo de `GET …/init` + `resposta` | Perguntas `PREVISAO_FATURAMENTO[_MULTIPLOS_SOCIOS]` (`PROXIMOS_MESES`, `NAO_SEI`, `NAO_VOU_MAIS_FATURAR`) e `CONTRIBUICAO_INSS[_MULTIPLOS_SOCIOS]` (`TODOS_OS_MESES`, `ECONOMIZAR_NOS_IMPOSTOS`, `NAO_QUERO_CONTRIBUIR`). O front não espera a resposta. | médio |
| `POST /api/plataforma/questionario-preferencia-prolabore/decisao-preferencia` | `{escolha: <cenario de GET decisao-preferencia> \| "SAIR_MOTOR_FATOR_R"}` | Escolhe o cenário recomendado ou desliga o "cálculo inteligente" (*"Tem certeza que deseja desativar o cálculo inteligente?"*). Texto: *"O pró-labore será ajustado de acordo com suas preferências a partir do próximo mês."* | alto |

### 4.3 Assessor de pró-labore (assistente `#/socio/assessor/*`)

Fluxo em passos (`ESCOLHA_SOCIO → ZERAMENTO|GERENCIAR_PROLABORE → DUPLO_VINCULO → DEPENDENTES →
SIMULAR_PROLABORE → sucesso`), estados `NAO_INICIADO|EM_ANDAMENTO|CONCLUIDO|EM_EDICAO|BLOQUEADO`.

| Método/caminho | Corpo |
|---|---|
| `PUT prolabore/assessor/escolher-socio/{cpf}` | — (o CPF do sócio vai no caminho) |
| `GET prolabore/assessor/passo/{cpf}/{PASSO}` | leitura |
| `PUT prolabore/assessor/passo` | `{cpf, passo, subpassoAtual\|null, emEdicao, passoConcluido, respostas:{tipo: <PASSO>, …}}`. Respostas por passo: ZERAMENTO `{zerarProlabore}`; GERENCIAR_PROLABORE `{tipoGerenciamentoProlabore: PERSONALIZADO\|TETO_INSS\|SALARIO_MINIMO\|OTIMIZACAO_INTELIGENTE, valorPersonalizado}`; DEPENDENTES `{possuiDependentes}`; DUPLO_VINCULO `{tipoVinculo (CLT\|PJ\|SERVIDOR_PUBLICO\|…), valoresFormulario:[{campo: cnpj\|tipo_remuneracao\|base_calculo\|inss, valor, status}], statusExtracao}`; SIMULAR_PROLABORE `{faturamentoEmpresa}` |
| `POST prolabore/assessor/socio/{cpf}/upload` | multipart `tipo` (`HOLERITE`\|`DEMONSTRATIVO_PAGAMENTO`), `formato`, `nome`, `documento` (arquivo). Também `GET …/upload/{id}`, `GET …/upload/status/{id}` (`SUCESSO`, `SUCESSO_PARCIAL`, `FALHA`, `ERRO`, `EM_PROCESSAMENTO`), `DELETE …/upload/{id}` |
| `POST prolabore/assessor/socio/{cpf}/dependente` / `DELETE …/dependente/{id}` | corpo do POST incerto |
| `GET prolabore/assessor/socio/{cpf}/simular` | leitura |
| `POST prolabore/assessor/socio/{cpf}/finalizar` | — (grava o resultado do assistente) |
| `POST prolabore/assessor/rollout` | — → `{destino}` |
| `POST prolabore/assessor/abandonar-fluxo` | — |

Mensagens: *"Selecione um sócio para continuar."*, *"Salvando sua preferência"*, *"Ocorreu um
erro ao salvar suas preferências. Tente novamente."*.

**Fora do catálogo** (caminho montado com `.concat`), base `prolabore/central/socios`:
`GET {socioId}/dependentes`, `GET|PUT|DELETE {socioId}/dependente/{id}`, `POST {socioId}/dependente`
(multipart `nome`, `cpf`, `dataNascimento`, `tipoDependenteESocial`, opcionais `tipoDocumento=
DECLARACAO_MATRICULA`, `nomeDocumento`, `declaracaoMatricula`), `GET {socioId}/duplo-vinculos`,
`POST {socioId}/duplo-vinculo`, `PUT|DELETE {socioId}/duplo-vinculo/{id}` (multipart). Tratam
dados pessoais de terceiros: **fora do escopo da CLI**.

### 4.4 Quando vale a mudança (competência)

Os textos do front não batem entre si; o que eles dizem:

- Gestão manual: *"Você pode escolher o valor do pró-labore **deste mês** até o dia 25."* /
  *"Você pode alterar este valor até o dia 25"*. Depois disso: *"O pró-labore do mês está
  fechado. Qualquer alteração de pró-labore terá efeito a partir do próximo mês."*
- Aviso geral: *"O ajuste no pró-labore deve ser feito até o dia 25 do mês para ser aplicado no
  **mês seguinte**."*
- Fora do motor: ajuste *"até o penúltimo dia do mês para ser aplicado no mês seguinte"*.
- Gestão inteligente: *"Iremos calcular o pró-labore até o dia 07 deste mês."*; preferências do
  questionário valem *"a partir do próximo mês"*.
- Zerar (GT) e remover: *"já começa a valer a partir deste mês"*; *"não altera os meses passados"*.

Nenhum PUT/PATCH manda competência: o servidor decide pela data. A CLI deve ler
`dashboard/prolabore` (`podeAlterar`, `gerado`) antes e mostrar "vale para MM/AAAA" vindo da
resposta ou de um GET posterior, sem calcular localmente.

### Recomendação para a CLI (pró-labore)

```text
ctbz prolabore definir --socio <id> --tipo personalizado --valor 1518.00 [--yes]
ctbz prolabore definir --socio <id> --tipo salario-minimo|teto-inss|inteligente [--minimo 1518] [--yes]
ctbz prolabore zerar-sem-faturamento on|off [--yes]       # PATCH zerar-prolabore
```

Ler `prolabore/central/gestao/{id}` antes (salário mínimo, teto, qtd. sócios em GI) e recusar
valor abaixo do mínimo. Não implementar o assessor, uploads, dependentes, saída do motor nem o
questionário (fluxos guiados com efeitos difíceis de prever).

---

## 5. Distribuição de lucros e informe de rendimentos

Chunk `informe-de-rendimentos.44e0d6a0.js` (serviços ~3300–3375).

| Método/caminho | Corpo | Semântica | Risco |
|---|---|---|---|
| `POST /api/plataforma/informerendimento/salvarconfiguracaocliente` | `[ {idSocio, percentual: "NN.NN" (string, 2 casas), valor: "NNNN.NN" (string, 2 casas)} … ]`, um item por sócio; com "não distribuir", `percentual` e `valor` = `0` | **Registrar a distribuição de lucros** do exercício aberto. Não manda ano nem datas: o exercício é o `ano` do GET. | **alto** (informe de rendimentos e IRPF dos sócios) |
| `POST /api/plataforma/informerendimento/historico` | `{tipoEvento: "VISUALIZOU_IR"\|"BAIXOU_IR", anoExercicio}` | Log de auditoria ao abrir ou baixar o informe | baixo |
| `POST /api/plataforma/informerendimento/aceitar-carta-responsabilidade/{ano}` | — | Aceite da carta de responsabilidade, exigido para ver o informe (*"Para visualizar o informe você precisa aceitar os termos da Carta de Responsabilidade."*) | médio (aceite legal) |
| `POST /api/plataforma/informerendimento/aceitar-termo-debitos/{ano}` | `{aceite: <fluxoAceiteTermoDeCiencia>}` (ex.: `ESTOU_CIENTE`, incerto) | Ciência de débitos federais | médio |
| `POST /api/plataforma/informerendimento/aceite/{ano}` | `{tipoAceite: "REGULARIZAR_PENDENCIA_DOCUMENTAL"\|"NAO_REGULARIZAR_PENDENCIA_DOCUMENTAL"}` | Escolha diante de pendência documental | médio |
| `POST /api/plataforma/informerendimento/reabrir-balanco/{ano}` | — | Reabre o exercício contábil (403 para admin: *"Usuário admin não pode fazer a reabertura do exercício contábil para um cliente."*). Pode levar a "contratar serviço adicional". | **alto** |
| `POST /api/plataforma/simulador-impostos-avancado/simular` | `{motorFatorR: bool (de GET …/init), notasFiscais:[{identificador, idCnaeEmpresa, valor}]}` | **Simulação só de leitura**. Resposta `{notasFiscais[{identificador, aliquota}], faturamento, impostosCalculados{valorDas, valorDarf}, prolabore, totalizadores{faturamentoLiquido, totalImpostos}, necessariaAnaliseManualProlabore, deveExibirAlertaDividendos, deveExibirAlertaIrrf}`. O front recusa soma acima de R$ 400.000 (texto com erro de digitação "R$ 400.00,00"). | nenhum |

Detalhes do `salvarconfiguracaocliente`:

- Pré: `GET informerendimento/recuperardadosdistribuicaocliente` → `ano`, `saldo`,
  `totalDistribuido`, `limitePermitidoDistribuicaoLucros`, `lucrosSocios[{id, porcentagem,
  valor}]`, `podeAlterar`, `motivoNaoPodeAlterar`, `dataLimite`, `perfil`. O front calcula
  `saldos.total = saldo + totalDistribuido`.
- Trava: só salva se `saldos.empresa == 0` e soma dos sócios == total (com 3 casas), senão
  *"Verifique se o lucro foi dividido por completo entre os sócios"*. Exceção:
  `desabilitarTravaDistLucroTotal`.
- Sucesso: *"Valores de lucro atualizado com sucesso! Seu Informe atualizado está pronto"*.
- Reversível: reenviar enquanto `podeAlterar` e antes de `dataLimite`.

### Recomendação para a CLI (lucros)

```text
ctbz lucros distribuir --socio <id>=<valor> [--socio <id>=<valor> ...] [--dry-run] [--yes]
ctbz lucros simular-impostos --nota <idCnaeEmpresa>=<valor> ...      # simulador-impostos-avancado/simular
```

`distribuir` lê o GET, calcula percentuais, exige soma = total distribuível e `podeAlterar`, e
mostra o antes e depois. `simular` pode sair já na 1.x porque é só leitura. Não automatizar
aceites, carta de responsabilidade nem reabertura de balanço.

---

## 6. Endpoints de escrita ausentes do `catalogo.md`

```text
POST /api/plataforma/novo-emissor/v2/emissao/emitir?origem=EMISSOR_SIMPLIFICADO_EMISSAO
POST /api/plataforma/novo-emissor/v2/emissao/emitir-async?origem=EMISSOR_SIMPLIFICADO_EMISSAO
POST /api/plataforma/novo-emissor/v2/emissao/replicar?origem=EMISSOR_SIMPLIFICADO_REPLICACAO
POST /api/plataforma/novo-emissor/v2/registro-de-notas/registrar
POST /api/plataforma/novo-emissor/v2/registro-de-notas/replicar
POST /api/plataforma/novo-emissor/v2/calculo/resumo-impostos          (só leitura)
POST /api/plataforma/novo-emissor/emissao/nova-nota?origem=…          (legado v1)
POST /api/plataforma/novo-emissor/emissao/replicar/{id}?origem=…      (legado v1)
POST /api/plataforma/novo-emissor/clientes/salvar-cliente-nacional
POST /api/plataforma/novo-emissor/clientes/salvar-cliente-exterior
POST /api/plataforma/autopilot/clientes                               (store morta)
POST /api/plataforma/nota-entrada/classificacao/salvarloteclassificacao/{tipo}
POST /api/legado/notafiscaltomada/{verificarNotaTomadaExiste,salvarnota}; DELETE …/removernota/{id}
POST|PUT|DELETE /api/plataforma/prolabore/central/socios/{id}/dependente[/{id}]
POST|PUT|DELETE /api/plataforma/prolabore/central/socios/{id}/duplo-vinculo[/{id}]
POST /api/plataforma/prolabore/assessor/socio/{cpf}/{upload,dependente,finalizar}; DELETE …/upload/{id}, …/dependente/{id}
POST /api/plataforma/questionario-preferencia-prolabore/{salvar-resposta,decisao-preferencia}  (já listados)
POST /api/emissor/notasentrada/manifestar        (variante sem barra, com chaves/cnpj)
POST /api/plataforma/parametrosEmpresa/{create,update}
```

Leituras novas úteis: `novo-emissor/v2/emissao/init[?idNotaReplicada=&isRegistro=]`,
`novo-emissor/v2/emissor-async`, `novo-emissor/tomador/{doc}`,
`novo-emissor/cadastro-clientes/{id}`, `novo-emissor/cep/logradouro?cep=`,
`banco-central/moeda[/{m}/data-referencia/{d}]`, `prolabore/central/gestao/{socioId}`,
`/api/legado/notafiscal/buscarNotaEmpresa?numero=`, `nota{perfil}/obtercodverificacao/{codigo}`,
`/api/plataforma/nota-entrada/classificacao/{init,listar}/`.

## 7. Observações gerais para o roadmap

1. Todo comando de escrita precisa de confirmação (`--yes`), de `--dry-run` quando houver
   cálculo e de idempotência mínima (antes de emitir, procurar na listagem do dia uma nota igual
   para o mesmo tomador e valor).
2. Nenhum endpoint de escrita devolve recibo estruturado conhecido: a CLI deve reler o GET
   correspondente para confirmar.
3. Erros vêm em formatos diferentes: `{identificador, mensagemAmigavel}` (emissor v2),
   `{error:{errorClassName, errorMessage}}` (cancelamento), `{errorMessage}` (cancelamento de
   nota importada) e `bodyText` (front de notas de entrada). Vale um decodificador tolerante.
4. GETs com efeito colateral, que nunca podem entrar em testes de contrato:
   `notafiscal/cancelarAPI` e `classificacaonotas/agendarExportarArquivos`.
5. Em notas de entrada, baixar o XML de nota `PENDENTE` pelo front **manifesta ciência**. A CLI
   deve baixar só notas já em `CIENCIA` ou `CONFIRMADA`, ou avisar.
