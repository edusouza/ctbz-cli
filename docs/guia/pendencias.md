# Pendências, rotinas e atendimento

## Resumo: o que precisa de atenção

```sh
ctbz resumo
```

```text
Seção        Item                                   Prazo       Valor        Situação     Alerta
impostos     DARF Unificado 07/2026                 06/10/2026  R$ 1.000,00  RECALCULADA  próxima
pendencias   Cadastre o PIS para ativar o pró-la…   21/07/2026               Pendente     vencida
rotinas      Importar extrato bancário de setembro  05/10/2026               EM_ABERTO    próxima
mensalidade  Mensalidade da Contabilizei            15/10/2026  R$ 15,90     EM_ABERTO
painel       Certificado digital                                             crítica      crítica
```

Uma lista só com o que exige ação, montada a partir de três consultas feitas em paralelo:

| Seção | De onde vem |
|---|---|
| `impostos` | guias em atraso (sempre `vencida`) e do mês, como em `ctbz impostos` |
| `pendencias` | pendências abertas, como em `ctbz pendencias` |
| `rotinas` | rotinas da empresa do mês ainda não realizadas (sem os impostos, que já aparecem acima) |
| `mensalidade` | a mensalidade do mês, paga ou não |
| `painel` | indicadores do painel: `crítica` (pendências críticas) ou `aviso` (outras pendências) |

Se nada precisa de atenção, a tabela sai vazia e o stderr diz `Nada precisa de atenção.`

Para um alerta diário em cron, `--fail-on-atencao` termina com **código 4** quando há algo
`vencida` ou `crítica`:

```sh
ctbz resumo --fail-on-atencao -o csv > /dev/null || notify-send "Contabilizei precisa de atenção"
```

Se qualquer uma das consultas falhar, o comando falha (código 1) em vez de mostrar um resumo
incompleto.

## Pendências da empresa

```sh
ctbz pendencias
```

```text
ID                Tipo                                                Detalhe                                                       Criada em   Prazo       Situação  Alerta
1000000000000001  Cadastre o PIS para ativar o pró-labore automático  O pró-labore influencia no valor do imposto, por isso assi…  21/07/2026  21/07/2026  Pendente  vencida
```

- São as pendências do card da tela inicial do painel: tipo, detalhe, criação, prazo e situação.
- Por padrão aparecem só as **abertas**; `--todas` inclui as finalizadas.
- O detalhe vem da API em HTML; a CLI tira as tags. Na tabela ele é cortado em 60
  caracteres; use `-o json` para ler o texto completo.

### Alertas de prazo

A coluna `alerta` só é preenchida para pendências abertas
([ADR-0012](../adr/0012-prazos-e-alertas.md)):

| `alerta` | Quando |
|---|---|
| `vencida` | o prazo já passou |
| `próxima` | o prazo é hoje ou nos próximos 7 dias |
| vazio (`null` no JSON) | prazo distante, sem prazo ou pendência finalizada |

`--fail-on-vencidas` faz o comando terminar com **código 4** quando há pendência vencida:

```sh
ctbz pendencias --fail-on-vencidas -o csv > /dev/null || notify-send "Há pendências vencidas"
ctbz pendencias -o json | jq -r '.[] | select(.alerta != null) | "\(.prazo) \(.tipo)"'
```

## Termos e cartas para aceitar

A Central de Rotinas pede aceites periódicos: a carta de responsabilidade da administração,
o termo de ciência sobre retiradas de lucros e o termo de adesão ao TotalPass. Leia antes de
aceitar:

```sh
ctbz pendencias termos                       # os pendentes (--todos inclui os demais)
ctbz pendencias termo carta-responsabilidade # texto completo
```

```text
Chave                   Termo                                                      Pendente  Aceite tácito em (dias)
carta-responsabilidade  Carta de Responsabilidade da Administração                 sim
termo-debitos           Termo de Ciência e Responsabilidade (retiradas de lucros)  sim       15
```

- O termo de débitos tem **aceite tácito**: se não houver resposta no prazo, a Contabilizei
  entende que não há impedimento para as retiradas de lucros.
- O texto vem em HTML e é convertido para texto, sem caracteres de controle.

Para aceitar:

```sh
ctbz pendencias aceitar carta-responsabilidade
ctbz pendencias aceitar termo-debitos
ctbz pendencias aceitar termo-totalpass
```

- **Risco alto:** são declarações legais e contábeis, sem como revogar. O texto completo é
  sempre impresso (no stderr) antes da confirmação, e no terminal é preciso digitar `confirmo`.
  Em scripts, use `--yes` (o texto continua indo para o stderr, para ficar no log).
- Sem aceite pendente, a CLI avisa "nada a aceitar" e termina com código 0.
- Se a sessão for de um administrador da Contabilizei, a API recusa (HTTP 403): os aceites só
  podem ser feitos pelo próprio cliente.

## Procuração do e-CAC

A Contabilizei precisa de uma procuração eletrônica no e-CAC da Receita. A CLI mostra se há
pendência e para quem outorgar:

```sh
ctbz pendencias procuracao
# Crie ou renove no e-CAC (https://cav.receita.fazenda.gov.br) uma procuração eletrônica ao CNPJ …
```

Depois de criar (ou renovar) a procuração no e-CAC, declare:

```sh
ctbz pendencias procuracao --ja-criei
```

- A procuração em si é feita no e-CAC, fora da Contabilizei; `--ja-criei` é só a declaração
  ("Já criei a procuração" no painel), de risco médio. Se a procuração não existir, a pendência
  deve voltar.
- A CLI usa o caminho da origem da pendência: a Central de Rotinas ou a etapa
  `GERAR_PROCURACAO_VIRTUAL` dos primeiros passos.

## Conciliação fiscal

```sh
ctbz pendencias conciliacao
```

```text
Competência de referência:        09/2026
Notas fiscais sem recebimento:    0
Recebimentos sem nota fiscal:     0
Conciliações automáticas no mês:  0
```

- **Notas sem recebimento**: notas fiscais emitidas que não foram ligadas a um crédito no extrato.
- **Recebimentos sem nota**: créditos no extrato sem nota fiscal correspondente.
- A competência de referência é o mês anterior, o mesmo que o painel mostra.
- `--listar notas` ou `--listar recebimentos` lista os itens pendentes dos últimos 12 meses.
  O formato desses itens ainda não foi verificado (a conta de desenvolvimento não tinha
  pendências), por isso os campos saem como a API os devolve.
- `--fail-on-pendencias` termina com código 4 quando há algo a conciliar.

## Rotinas e obrigações do mês

```sh
ctbz rotinas
ctbz rotinas --mes 2026-11
```

```text
Responsável   Rotina                                 Prazo       Status     Valor     Alerta   Pendências
empresa       Importar extrato bancário de setembro  05/10/2026  EM_ABERTO            próxima
empresa       DARF Unificada (Ativação do fator R)   06/10/2026  EM_ABERTO            próxima
empresa       Mensalidade da Contabilizei            15/10/2026  EM_ABERTO  R$ 15,90
contabilizei  eSocial                                15/10/2026  EM_ABERTO
contabilizei  DCTFWeb                                15/10/2026  EM_ABERTO
```

- `responsavel` separa o que a **empresa** precisa fazer do que a **Contabilizei** entrega
  (eSocial, DCTFWeb, EFD-Reinf…).
- O filtro é pelo mês do prazo. A API não aceita competência nem mês: devolve sempre o mês
  anterior, o atual e o próximo, e a CLI filtra. Fora dessa janela, a lista sai vazia.
- `alerta` segue a mesma regra das pendências: `vencida` ou `próxima` para rotinas não
  realizadas ([ADR-0012](../adr/0012-prazos-e-alertas.md)).
- `--fail-on-vencidas` (código 4) considera só as rotinas da **empresa**: atraso numa
  obrigação da Contabilizei não é algo que você resolve sozinho.
- `pendencias` traz os ids das pendências de uma rotina, como os lançamentos que ela pede para
  reclassificar (ver abaixo).

### Reclassificar lançamentos pedidos pela rotina

Algumas rotinas pedem para confirmar ou alterar a classificação de um lançamento bancário
(financiamento, investimento anjo e outros). Sem `--classificacao`, a CLI só mostra as opções:

```sh
ctbz rotinas reclassificar 1000000000000501
```

```text
Pendência         Lançamento          Valor        Classificação atual           ID                Opção                         Sócios
1000000000000501  TED RECEBIDA BANCO  R$ 50.000,00 Empréstimos e Financiamentos  1000000000000041  Empréstimos e Financiamentos
1000000000000501  TED RECEBIDA BANCO  R$ 50.000,00 Empréstimos e Financiamentos  1000000000000042  Aporte de Capital             1000000000000031 FULANO DE TAL
```

Com a opção escolhida (id ou nome), a classificação é gravada e a rotina é **concluída** (risco
médio: a rotina não reabre, mas a classificação ainda pode ser trocada com
`ctbz extratos classificar` enquanto o período estiver aberto):

```sh
ctbz rotinas reclassificar 1000000000000501 --classificacao "Aporte de Capital"
```

- Opções com sócio exigem `--socio`; com um único sócio, ele é escolhido sozinho, como no painel.
- Vários ids podem ir de uma vez, com a mesma classificação para todos.
- Pede confirmação (`--yes` em scripts) e aceita `--dry-run` ([Ações que alteram dados](escrita.md)).

## Chamados de atendimento

```sh
ctbz chamados                  # em andamento
ctbz chamados --finalizados    # já resolvidos
```

- Colunas: `id`, `assunto`, `status`, `canal`, `criado`, `atualizado`, `previsao_retorno` e
  `link` (a página do chamado na central de ajuda).
- Em contas com muitos chamados, a Contabilizei não consegue listar os finalizados (o
  servidor responde com erro sempre, não adianta repetir). A CLI então mostra os finalizados
  entre os **100 chamados mais recentes** da empresa e avisa no stderr
  ([ADR-0013](../adr/0013-sem-retentativa-com-fonte-alternativa.md)). Nessa fonte não há
  canal, data de atualização nem previsão de retorno.
