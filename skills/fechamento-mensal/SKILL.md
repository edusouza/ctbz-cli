---
name: fechamento-mensal
description: Fechamento do mês na Contabilizei pela CLI ctbz, o trabalho do analista do plano Experts. Use para importar extratos bancários (OFX, PDF, CSV/XLS), classificar ou desmembrar lançamentos, lançar no caixa, conciliar notas e recebimentos, manifestar e classificar notas de entrada, enviar documentos pedidos e resolver as rotinas do mês, deixando a contabilidade em dia.
---

# Fechamento do mês

Antes, siga a skill `assessor` (sessão, dossiê e protocolo de escrita). A competência
fechada costuma ser o **mês anterior** (`M` = `AAAA-MM`); os prazos caem no mês atual.

## 0. O que falta (leituras em paralelo)

```sh
ctbz rotinas -o json                    # responsavel=empresa e não realizadas = a fazer
ctbz contas-bancarias -o json           # id e integracao de cada conta
ctbz extratos --ano AAAA -o json        # situação do extrato por conta e mês
ctbz pendencias conciliacao -o json
ctbz documentos pendentes -o json
ctbz notas entrada --lista a-manifestar --mes M -o json
ctbz notas entrada --lista a-classificar --mes M -o json
```

Monte o roteiro com só as etapas que têm algo a fazer, na ordem abaixo (o extrato vem antes,
porque a conciliação e as classificações dependem dele). Mostre o roteiro ao usuário antes de
começar.

## 1. Importar extratos

Para cada conta **não integrada** (`integracao` vazio; a Conta PJ Contabilizei é integrada e
dispensa importação) cujo mês `M` falta:

1. Peça o arquivo OFX (preferível) ou PDF do mês ao usuário: o caminho no disco. O OFX passa
   mesmo quando o banco não o lista (a CLI avisa no stderr, como o Banco Inter, que lista "PDF,
   XLS ou CSV"): tente-o antes do CSV/XLS. Se o banco não aceita PDF (a CLI recusa e lista os
   aceitos), ou se o OFX e o PDF são recusados, use o CSV/XLS do internet banking. Esses formatos vão pela Central de Documentos e
   a Contabilizei os importa **manualmente, em até 45 dias** (`importacao: "manual"`,
   `situacao: "enviado"`). Avise o usuário e registre no dossiê que o mês aguarda essa importação.
   Até lá a competência continua `PENDENTE` em `ctbz extratos`, o que é o esperado: **nunca
   reenvie** um arquivo com `importacao: "manual"` (a fila manual recebe duplicados), nem depois
   de uma falha de conexão; peça ao usuário para conferir na Central de Documentos do painel.
   CSV/XLS não aceitam `--substituir` e são recusados sobre um extrato já importado.
2. Num OFX, leia o saldo final no próprio arquivo (`<LEDGERBAL>` → `<BALAMT>`) e mostre ao
   usuário para conferir.
3. Simule e importe (risco médio):
   ```sh
   ctbz extratos importar ARQ --conta-bancaria ID --competencia M --saldo-final 1.234,56 --dry-run
   ctbz extratos importar ARQ --conta-bancaria ID --competencia M --saldo-final 1.234,56 --yes
   ```
   `--saldo-final` só vale para OFX. Sempre informe: com `--yes` e sem ele, a CLI aceita sem
   conferir o saldo que a Contabilizei leu. Saldo diferente = a importação para; confira o
   arquivo com o usuário. Em PDF, omita a flag.
4. Mensagens de "outro banco", "outra conta", "conta não cadastrada" (cadastre com
   `ctbz contas-bancarias adicionar`, skill `cadastro-e-acessos`) e período anterior pendente
   vêm traduzidas: siga o que dizem.
5. Competência já `IMPORTADO` ou `PROCESSANDO`: a CLI recusa. Se o extrato atual está vazio
   (criado pelo processamento de outro mês) ou é o arquivo errado, refaça com `--substituir`
   (exclui o atual e importa): mostre o `--dry-run` e confirme com o usuário. Com lançamentos é
   risco alto (as classificações se perdem). Depois de importar, confira `quantidade_lancamentos`:
   zero com um arquivo que tem movimentações é problema.
6. Conexão caiu ou o servidor respondeu com erro 5xx (502, 504…) no envio do arquivo: a CLI
   relê a competência e diz se é seguro repetir ("nada foi importado e é seguro rodar o mesmo
   comando de novo"). Só repita nesse caso, uma vez. Se ela mandar conferir (competência
   `IMPORTADO`/`PROCESSANDO`, ou queda/5xx ao concluir), rode `ctbz extratos` e
   `ctbz extratos lancamentos` e mostre ao usuário antes de qualquer nova tentativa. Depois de
   um `--substituir` que excluiu o extrato antigo e falhou, nunca repita com `--substituir` sem
   conferir: se o novo arquivo entrou, ele seria excluído. A mensagem diz a situação relida da
   competência: com `PROCESSANDO`, não importe de novo agora; espere e confira com
   `ctbz extratos`.

Importou o mês errado? `ctbz extratos excluir` é **risco alto** (perde as classificações).
Não use `excluir` para destravar um mês em `PROCESSANDO`: a Contabilizei aceita e nada muda, e a
CLI termina com erro. Peça ao usuário para importar o OFX (ou o PDF) pela tela "Importar extrato"
do painel, que funciona nesse estado, ou para abrir um chamado.

## 2. Classificar lançamentos do extrato

```sh
ctbz extratos lancamentos --conta-bancaria ID --competencia M -o json
ctbz extratos contas --competencia M -o json      # opções aceitas (RECEITA p/ entradas, DESPESA p/ saídas)
```

1. Separe os lançamentos sem classificação ou com classificação duvidosa.
2. Proponha uma classificação para cada um **escolhida da lista de `extratos contas`** (nunca
   invente nomes). Pistas em [references/classificacao.md](references/classificacao.md) e em
   `sugestao`/`probabilidade_sugestao`, a sugestão da própria Contabilizei (é só uma pista:
   confira com a descrição).
   Pergunte quando não der para saber (transferência de sócio, por exemplo: aporte, empréstimo
   ou devolução?).
3. Mostre a tabela inteira (data, descrição, valor, classificação proposta). Um "ok" do usuário
   para a tabela completa cobre esses lançamentos (risco médio). Depois rode um por um:
   ```sh
   ctbz extratos classificar LANC --conta-bancaria ID --competencia M --conta "Classificação" [--socio ID] --yes
   ```
4. Um lançamento que mistura coisas (aluguel + condomínio): desmembre (soma exata, 2+ partes):
   ```sh
   ctbz extratos desmembrar LANC --conta-bancaria ID --competencia M \
     --parte "Aluguel:1.000,00:Classificação A" --parte "Condomínio:234,56:Classificação B" --yes
   ```
   Para voltar: `ctbz extratos desfazer-desmembramento LANC ...`.
5. "Nossos contadores já classificaram…" = período fechado; não insista.

Rotinas que pedem reclassificação (coluna `pendencias` de `ctbz rotinas`):
`ctbz rotinas reclassificar ID` (lista as opções) → `--classificacao "Opção" [--socio ID]`
(risco médio; conclui a rotina).

## 3. Caixa (dinheiro em espécie)

Só se o usuário recebeu ou pagou em dinheiro, fora do banco:

```sh
ctbz caixa contas --competencia M [--recebimento|--pagamento] -o json   # "Exige": guia ou sócio
ctbz caixa contas --competencia M --vinculos -o json                    # ids de guia e de sócio
ctbz caixa adicionar --competencia M --data AAAA-MM-DD --pagamento|--recebimento \
  --valor 150,25 --conta "Classificação" --descricao "…" [--guia ID|--sem-guia] [--socio ID] --yes
```

O valor é sempre positivo; o sinal vem de `--pagamento`/`--recebimento`. Corrigir:
`caixa editar ID --competencia M --valor …` (só o que muda). Excluir: `caixa remover`
(a CLI mostra o comando que recria). Lançamentos `automatico: true` não se editam.

## 4. Conciliação de notas e recebimentos

```sh
ctbz pendencias conciliacao --listar recebimentos -o json   # créditos sem nota
ctbz pendencias conciliacao --listar notas -o json          # notas sem crédito
ctbz pendencias conciliacao candidatos --notas -o json      # notas que podem ser vinculadas
ctbz pendencias conciliacao motivos -o json
```

Para cada pendência, pergunte ao usuário o que ela é e proponha:

- recebimento que é pagamento de uma nota → `resolver ID --vincular NOTAFISCAL:ID_NOTA`;
- nota recebida noutra conta ou em dinheiro → `--vincular MOVIMENTACAO:ID` ou o motivo adequado;
- recebimento que não é receita (aporte, empréstimo do sócio, resgate…) → `--motivo M`
  (motivos com sócio exigem `--socio ID`);
- recebimento de serviço **sem nota emitida** → avise que falta emitir a nota (pelo painel) e
  que justificar como outra coisa seria incorreto.

`resolver` é **risco alto**: define a natureza fiscal do dinheiro (receita tributável ou não).
Confirmação explícita para cada um. Os formatos de `--listar`, `candidatos` e `detalhes` vêm
como a API devolve: leia os campos que vierem.

## 5. Notas de entrada (compras)

```sh
ctbz notas entrada --lista a-manifestar --mes M -o json
ctbz notas entrada --lista a-classificar --mes M -o json
ctbz notas entrada produtos ID -o json
```

As escritas procuram a nota no mês atual por padrão: passe sempre `--mes M`, o mesmo das
listagens (sem ele, a CLI responde "nota X não encontrada em AAAA-MM; use --mes" e não envia).

```sh
ctbz notas entrada manifestar ID... --confirmar --mes M --dry-run
ctbz notas entrada classificar ID... --como uso-consumo --mes M --dry-run
ctbz notas entrada reclassificar ID --como estoque --mes M --dry-run
```

- **Manifestar** (só `PENDENTE` ou `CIENCIA`): pergunte se a compra foi recebida.
  `--confirmar` (recebi), `--desconhecer` (não reconheço) e `--nao-realizada --justificativa
  "…"` (15 a 255 caracteres) são eventos na SEFAZ **sem desfazer** (risco alto). `--ciencia`
  é preliminar (médio). Vários IDs no mesmo envio.
- **Classificar** até o dia 05 do mês seguinte à emissão (depois vale a pré-classificação):
  `--como estoque|insumo|uso-consumo|ativo-imobilizado|prestacao-servico` para notas inteiras,
  ou `--produto ITEM:opcao=qtd,opcao=qtd` por item (a soma é a quantidade do item). Risco médio.
  Pistas: revenda → estoque; matéria-prima → insumo; material de escritório, limpeza → uso e
  consumo; bem durável acima de ~R$ 1.200 usado por mais de 1 ano → ativo imobilizado.
  Já classificada: `reclassificar`.

## 6. Documentos pedidos

```sh
ctbz documentos pendentes -o json     # id_pendencia, tipo, competencia, conta
```

- Arquivo existe → peça o caminho e `ctbz documentos enviar ARQ --pendencia ID` (médio; várias
  `--pendencia` do mesmo tipo vão num arquivo só). Extratos bancários vão por
  `extratos importar`.
- Não existe → `documentos sem-arquivo --pendencia ID` para `ESTOQUE` ("não tive estoque") e
  `CONTROLE_DE_INTERMEDIACOES` ("não tive intermediações"), ou `documentos sem-aplicacao
  --conta-bancaria ID --pendencia ID` (sem aplicação financeira). Declarações **de risco alto**
  e sem desfazer que fecham a pendência: confirme que é verdade.
- **AFAC** (`CONTRATO_DE_AFAC`) é diferente: ali o `sem-arquivo` declara que o contrato **já foi
  enviado antes** e que esse documento vale como comprovante. Use só nesse caso. Se o contrato
  não existe, não use `sem-arquivo`: oriente o usuário a falar com a Contabilizei.

## 7. Conferência final

Rode `ctbz rotinas -o json` e `ctbz pendencias conciliacao -o json` de novo e mostre o que
ficou resolvido e o que ainda falta (com o motivo). Registre no dossiê: competência fechada,
escritas feitas e pendências que sobraram.
