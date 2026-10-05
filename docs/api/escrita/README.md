# Escrita

Endpoints que alteram dados na Contabilizei: criar, alterar, remover e marcar como
concluído. Foram levantados por **análise estática** do JavaScript do painel
(`/painel-de-controle/`, versão 2.0.88) e do front de notas de entrada (`/nota-entrada/`)
em 2026-10-03, lendo o código ao redor de cada chamada (ver [metodologia](../../metodologia/README.md)).
**Nenhuma escrita foi chamada**: os corpos refletem o que o front envia, e o que não pôde ser
confirmado está marcado "(incerto)".

A política para os comandos de escrita da CLI está na
[ADR-0018](../../adr/0018-escrita-com-confirmacao.md). O plano de implementação está no
[ROADMAP](https://github.com/edusouza/ctbz-cli/blob/main/ROADMAP.md) (v1.1 a v1.8).

| Página | Contextos |
|---|---|
| [Pendências e conta](pendencias-e-conta.md) | Central de Rotinas, termos e cartas, informe de rendimentos (aceites), conciliação fiscal, primeiros passos, chamados, conta do usuário, multiusuário, dados de acesso, escritório virtual |
| [Impostos e pagamentos](impostos-e-pagamentos.md) | confirmar pagamento de guia, recálculo, parcelamento, pagamento com cartão, débito automático, mensalidade e cartões |
| [Contabilidade e documentos](contabilidade-e-documentos.md) | caixa, classificação e desmembramento de lançamentos, extratos, contas bancárias, reabertura de balanço, envio de documentos, certificado digital |
| [Notas e pró-labore](notas-e-prolabore.md) | emissão e cancelamento de NFS-e, tomadores, notas de entrada (manifestação e classificação), pró-labore, distribuição de lucros |

## Resumo

Base `/api/plataforma/`, salvo indicação. Risco: **baixo** (estado de tela, reversível),
**médio** (cadastro ou contabilidade, reversível ou declaração simples), **alto** (efeito
fiscal, legal ou financeiro, em geral sem desfazer).

| Ação | Método e caminho | Desfazer | Risco | Versão |
|---|---|---|---|---|
| Lançamento do caixa: criar e editar | `POST caixa/lancamentousuario/novo/` | editar ou remover | médio | v1.2 |
| Lançamento do caixa: remover | `DELETE caixa/lancamentousuario/remover/{ano}/{mes}/{id}` | — | médio | v1.2 |
| Classificar lançamento do extrato | `PUT movimentacao-financeira/classificar` | reclassificar | médio | v1.2 |
| Desmembrar lançamento | `PUT movimentacao-financeira/desmembrar` | `DELETE movimentacao-financeira/desmembrar/desfazer/{id}` | médio | v1.2 |
| Reclassificar pendência da rotina | `POST documentos/reclassificar` | reclassificar | médio | v1.2 |
| Confirmar ou negar pagamento de guia | `PUT impostos/v5/impostos-a-pagar/guia/{id}/confirmar-pagamento` | mesmo endpoint, `false` | médio | v1.3 |
| Recálculo de guia | `PUT impostos/v2/impostos-a-pagar/guia/{id}/v2/recalcular` | — | alto | v1.3 |
| Contratar parcelamento | `POST impostos/parcelamento/…/contratar` | — | alto | v1.3 |
| Aceitar carta e termos | `POST central-rotinas/aceitar-*` | — | alto | v1.4 |
| Declarar procuração e-CAC | `POST central-rotinas/resolver-pendencia-procuracao-ecac` | — | médio | v1.4 |
| Resolver conciliação fiscal | `POST conciliacao-fiscal/conciliar/resolver-pendencia` | refazer (mesmo endpoint) | alto | v1.4 |
| Concluir tarefa de primeiros passos | `POST checklist-onboarding/aside/concluir-etapa` | — | baixo | v1.4 |
| Conta bancária: salvar e excluir | `POST contabancaria/salvar`, `DELETE contabancaria/excluir/{id}` | editar / — | médio / alto | v1.5 |
| Importar extrato | multipart `upload-documentos/extrato/enviar/bucket` + `POST movimentacao-financeira/eventoUploadExtrato` | `DELETE movimentacao-financeira/extrato` | médio | v1.5 |
| Enviar documento | multipart `documentos/envio-documento/enviar[/consolidado]` | — | médio | v1.5 |
| Declarar ausência de documento | `POST documentos/envio-documento/enviar/sem-arquivo` | — | alto | v1.5 |
| Tomador: salvar e excluir | `POST novo-emissor/clientes/salvar-cliente-{nacional,exterior}`, `DELETE autopilot/clientes/{doc}` | editar / — | médio | v1.6 |
| Manifestar nota de entrada | `POST /api/emissor/notasentrada/manifestar/` | — | alto | v1.6 |
| Classificar nota de entrada | `POST /api/emissor/classificacaonotas/salvarloteclassificacao/{tipo}` | reclassificar | médio | v1.6 |
| Alterar pró-labore | `PUT prolabore/central/gestao/{socioId}` | alterar de novo | alto | v1.7 |
| Distribuir lucros | `POST informerendimento/salvarconfiguracaocliente` | salvar de novo (incerto) | alto | v1.7 |
| Aceites do informe de rendimentos | `POST informerendimento/{aceitar-carta-responsabilidade,aceitar-termo-debitos,aceite}/{ano}` | — | alto | v1.7 |
| Reabrir balanço | `POST informerendimento/reabrir-balanco/{ano}` | — | alto | v1.7 |
| Usuários: convidar, ativar e desativar | `POST /api/multiusuario/invites/enviar`, `PUT /api/multiusuario/usuarios/ativar` | desativar | alto / médio | v1.8 |
| Dados de acesso (prefeitura, Simples) | `POST empresa/dadosacesso/atualizarDadosAcesso` | regravar | alto | v1.8 |
| E-mail, telefone e senha | `POST conta-usuario/{alterar-dados,alterar-senha}` (com OTP) | alterar de novo | alto | v1.8 |

## Fora do roadmap de escrita

Documentados nas páginas acima, mas sem comando previsto:

- **Pagamentos e cartões**: pagar imposto ou fatura com cartão, cadastrar cartão, débito
  automático, troca de plano. O cartão é cifrado no navegador (Adyen ou Iugu), e uma CLI não
  consegue montar esses corpos.
- **Certificado digital**: upload do A1, emissão, agendamento e remoção.
- **NFS-e**: emitir, replicar, registrar e cancelar.
- **Chamados**: não há API para abrir ou responder; o atendimento é no Zendesk.
- Marcações de tela e telemetria: modais vistos, tours, NPS, timeline de ativação.

## Ausentes do catálogo

Até a v1.0, o gerador do [catálogo](../catalogo.md) não capturava chamadas feitas com
`axios({method, url})`, sufixos concatenados e caminhos guardados em variáveis, nem o front
de notas de entrada. A v1.1 passou a reconhecê-las (#179); as entradas aparecem na próxima
regeneração do catálogo (`ctbz login && scripts/extrair-endpoints.py > docs/api/catalogo.md`).
Entre as escritas acima, as que faltavam no catálogo da v1.0 eram:
`…/confirmar-pagamento`, `…/v2/recalcular`, `desmembrar/desfazer`, `DELETE
movimentacao-financeira/extrato`, `documentos/envio-documento/enviar[/consolidado]`,
`novo-emissor/clientes/salvar-cliente-*`, `novo-emissor/v2/emissao/*` e a base `/api/emissor/`.
Cada página traz a lista completa da sua área.
