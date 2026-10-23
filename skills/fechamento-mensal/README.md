# Skill `fechamento-mensal`

O trabalho do analista do plano Experts: deixar a contabilidade do mês em dia. O agente descobre
o que falta, monta um roteiro e conduz você etapa por etapa: importar extratos, classificar e
desmembrar lançamentos, lançar no caixa, conciliar notas e recebimentos, manifestar e classificar
notas de entrada, enviar os documentos pedidos e resolver as rotinas do mês.

Instalação e configuração do plugin: [README das skills](../README.md).

## Quando usar

```text
> Fecha o mês de setembro.
> Importa o extrato do Itaú: ~/Downloads/extrato-set.ofx
> Classifica os lançamentos do extrato de setembro.
> Tem nota de compra para manifestar?
> A Contabilizei pediu um documento, me ajuda a enviar.
```

Pelo nome: `/ctbz:fechamento-mensal`.

A competência fechada costuma ser o **mês anterior** (os prazos caem no mês atual).

## Pré-requisitos

- O que a skill [`assessor`](../assessor/README.md) pede (sessão válida no `ctbz`).
- Os arquivos que você vai enviar (extrato em OFX ou PDF, documentos) no seu disco: o agente
  pede o caminho.

## Roteiro

O agente lê, em paralelo, rotinas, contas bancárias, extratos do ano, conciliação, documentos
pendentes e notas de entrada, e mostra só as etapas que têm algo a fazer, nesta ordem (o extrato
vem antes porque a conciliação e as classificações dependem dele):

### 1. Importar extratos

Para cada conta **não integrada** cujo mês falta (a Conta PJ Contabilizei é integrada e dispensa
importação):

1. pede o OFX (preferível) ou o PDF do mês. Se o banco só aceita outros formatos, ou se o PDF é
   recusado, usa o CSV ou o XLS do internet banking. Esses vão pela Central de Documentos, e a
   Contabilizei importa o arquivo manualmente, em até 45 dias. Até lá o mês continua `PENDENTE`
   em `ctbz extratos`, e o agente não reenvia o arquivo: a conferência é na Central de
   Documentos do painel. CSV/XLS não aceitam `--substituir` nem vão por cima de um extrato já
   importado;
2. num OFX, lê o saldo final do próprio arquivo e mostra para você conferir;
3. simula e importa, informando o saldo final para a Contabilizei conferir:

```sh
ctbz extratos importar ARQ --conta-bancaria ID --competencia 2026-09 --saldo-final 1.234,56 --dry-run
```

Mensagens como "outro banco", "conta não cadastrada" ou "período anterior pendente" vêm
traduzidas e o agente segue o que dizem (conta nova: skill
[`cadastro-e-acessos`](../cadastro-e-acessos/README.md)).

Competência já `IMPORTADO` ou `PROCESSANDO`: a CLI recusa, como o painel. O processamento de um
extrato pode deixar o mês seguinte `IMPORTADO` sem nenhum lançamento, e a mensagem avisa quando
o extrato está vazio. Para refazer a importação (extrato vazio ou arquivo errado), o agente usa
`--substituir`, que exclui o extrato atual e importa o arquivo numa confirmação só: mostra o
`--dry-run` (exclusão, envio e conclusão) e pede o seu "ok". Com o extrato atual vazio é risco
médio; com lançamentos é **risco alto**, porque as classificações se perdem sem desfazer. O
JSON traz `substituiu_extrato`.

Depois de importar, o agente confere `quantidade_lancamentos`: zero com um arquivo que tem
movimentações é problema, e a CLI avisa no stderr.

### 2. Classificar lançamentos do extrato

O agente lista os lançamentos sem classificação e propõe uma para cada um, **sempre escolhida
da lista que a Contabilizei aceita** (`ctbz extratos contas`), nunca inventada. Pergunta quando
não dá para saber (uma transferência do sócio é aporte, empréstimo ou devolução?). Você vê a
tabela inteira (data, descrição, valor, classificação) e um único "ok" cobre a tabela.

Um lançamento que mistura coisas (aluguel + condomínio) é **desmembrado** em partes que somam o
valor exato:

```sh
ctbz extratos desmembrar LANC --conta-bancaria ID --competencia 2026-09 \
  --parte "Aluguel:1.000,00:Classificação A" --parte "Condomínio:234,56:Classificação B"
```

Rotinas que pedem reclassificação também são resolvidas aqui (`ctbz rotinas reclassificar`).

### 3. Caixa (dinheiro em espécie)

Só se você recebeu ou pagou em dinheiro, fora do banco. O valor é sempre positivo e o sinal vem
de `--pagamento`/`--recebimento`; algumas classificações exigem a guia ou o sócio vinculado.

### 4. Conciliação de notas e recebimentos

Para cada crédito sem nota ou nota sem crédito, o agente pergunta o que é e propõe:

- recebimento que paga uma nota → vincular à nota;
- nota recebida noutra conta ou em dinheiro → vincular à movimentação ou informar o motivo;
- dinheiro que não é receita (aporte, empréstimo do sócio, resgate) → motivo correspondente;
- serviço recebido **sem nota emitida** → avisa que falta emitir a nota (pelo painel), porque
  justificar como outra coisa seria incorreto.

### 5. Notas de entrada (compras)

- **Manifestar** (você recebeu a compra?): confirmar, desconhecer ou "operação não realizada"
  são eventos na SEFAZ sem desfazer.
- **Classificar** até o dia 05 do mês seguinte à emissão (depois vale a pré-classificação da
  Contabilizei): estoque, insumo, uso e consumo, ativo imobilizado ou prestação de serviço, por
  nota inteira ou por item. Pistas para escolher estão em
  [`references/classificacao.md`](references/classificacao.md).

### 6. Documentos pedidos

- O arquivo existe → `ctbz documentos enviar ARQ --pendencia ID`.
- Não existe → declaração de inexistência (ex.: "não tive estoque", "sem aplicação
  financeira"), que fecha a pendência e não tem volta. O agente confirma com você que é verdade.
- Contrato de AFAC tem regra própria: a declaração ali significa "já enviei antes".

### 7. Conferência final

Relê rotinas e conciliação, mostra o que foi resolvido e o que ficou pendente (com o motivo) e
registra no dossiê a competência fechada e as escritas feitas.

## O que pode alterar

Tudo com `--dry-run` primeiro e a sua confirmação (protocolo da skill `assessor`).

| Ação | Risco | Desfazer |
|---|---|---|
| `extratos importar` | médio | `extratos excluir` (risco alto) |
| `extratos importar --substituir` | médio com o extrato atual vazio; **alto** com lançamentos | não: o extrato atual e as classificações se perdem |
| `extratos classificar` / `desmembrar` / `desfazer-desmembramento` | médio | sim, até os contadores fecharem o período |
| `rotinas reclassificar` | médio | a rotina não reabre |
| `caixa adicionar` / `editar` / `remover` | médio | editar de novo; remover mostra como recriar |
| `notas entrada classificar` / `reclassificar` | médio | `reclassificar` |
| `notas entrada manifestar --ciencia` | médio | evolui para confirmar/desconhecer |
| `documentos enviar` | médio | não há exclusão |
| `extratos excluir` | **alto** | perde as classificações; só reimportando |
| `pendencias conciliacao resolver` | **alto** | define a natureza fiscal do dinheiro; refaz, não apaga |
| `notas entrada manifestar --confirmar/--desconhecer/--nao-realizada` | **alto** | evento na SEFAZ, sem desfazer |
| `documentos sem-arquivo` / `sem-aplicacao` | **alto** | declaração sem desfazer |

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| [`SKILL.md`](SKILL.md) | instruções que o agente segue |
| [`references/classificacao.md`](references/classificacao.md) | pistas para classificar lançamentos e notas |
