# Contabilidade

Relatórios contábeis, caixa, extratos e plano de contas. Os relatórios usam o mês como
`AAAA-MM`; o painel mostra por padrão dezembro do ano anterior.

## Balancete

```sh
ctbz balancete 2026-08
ctbz balancete 2026-08 -o csv > balancete-2026-08.csv
```

```text
Conta          Descrição                          Nível  Saldo anterior      Débitos     Créditos       Saldo
1              ATIVO                                  1         R$ 0,00  R$ 1.815,00  R$ 1.419,83    R$ 395,17
1.01           CIRCULANTE                             2         R$ 0,00  R$ 1.815,00  R$ 1.419,83    R$ 395,17
1.01.01.01.00          Caixa Geral                    5         R$ 0,00  R$ 1.139,00  R$ 1.280,83  -R$ 141,83
```

- Cada linha é uma conta da árvore contábil; na tabela, a descrição é recuada pelo nível.
- Em CSV e JSON a descrição vem sem recuo e o nível fica na coluna `nivel`, o que facilita
  filtrar (ex.: só o nível 1 para ATIVO, PASSIVO e PATRIMÔNIO LÍQUIDO).
- O PDF do painel é gerado no navegador a partir dos mesmos dados; a API não oferece PDF.

## Balanço patrimonial

```sh
ctbz balanco 2025          # fechamento do exercício (dezembro)
ctbz balanco 2026-09       # posição em um mês
```

```text
Conta  Descrição            Nível  Grupo            Saldo  Exercício anterior
1      ATIVO                    1  ATIVO        R$ 395,17             R$ 0,00
1.01     CIRCULANTE             2  ATIVO        R$ 395,17             R$ 0,00
2      PASSIVO                  1  PASSIVO     -R$ 395,17             R$ 0,00
```

- Ativo, passivo e patrimônio líquido, com o saldo do exercício e o do exercício anterior.
- As contas de resultado (receitas e despesas) ficam de fora, como no painel.

## Razão

```sh
ctbz razao --de 2026-01 --ate 2026-09               # todas as contas no período
ctbz razao --conta 1.01.01.01.00 --de 2026-07       # uma conta
ctbz razao --conta 1.01 -o csv > circulante.csv     # prefixo: todas as contas de 1.01
```

```text
Data        Conta          Descrição da conta  Histórico       Contrapartida                                    Débito  Crédito        Saldo
21/07/2026  1.01.01.01.00  Caixa Geral         Capital social  2.07.01.01.00 Capital Social Realizado no País  R$ 1.000,00         R$ 1.000,00
```

- Um lançamento por linha, com a conta de contrapartida e o saldo acumulado da conta no
  exercício depois do lançamento.
- A API devolve um mês por vez; a CLI consulta cada mês do período (até 24).

## Caixa do mês

```sh
ctbz caixa 2026-09
ctbz caixa 2026-09 -o csv > caixa-2026-09.csv
```

```text
Data        Descrição               Conta                         Classificação        Valor  Situação    Automático  ID
18/09/2026  Pagamento Contabilizei  Mensalidade de contabilidade  DESPESA         -R$ 141,83  CONFIRMADO  sim         1000000000000001
Total do mês: -R$ 141,83 em 1 lançamento(s)
```

- Entradas e saídas classificadas do mês, com a conta de classificação (de
  `ctbz contas`) e o tipo (receita, despesa…). Saídas têm valor negativo.
- `automatico` indica lançamentos feitos pelo sistema (ex.: a mensalidade da Contabilizei).
- Como no painel, a CLI pede até 1000 lançamentos por mês e avisa se houver mais.

## Extratos e contas bancárias

```sh
ctbz extratos --ano 2026      # situação do extrato por conta e mês
ctbz contas-bancarias         # contas cadastradas
```

```text
Competência  Banco                  Agência  Conta      Situação  Integração  ID da conta
10/2026      Contabilizei Conta PJ  0001     000000000  ABERTO                1000000000000001
```

- `extratos` mostra, para cada conta e mês, se o extrato está aberto ou fechado na
  Contabilizei. Mês sem extrato vira a rotina "Importar extrato bancário" (`ctbz rotinas`).
- `contas-bancarias` mostra banco, agência, conta, saldo inicial e a integração com o banco
  (ex.: `INTEGRADA` na conta PJ da Contabilizei, que dispensa importar extrato).
- Importar um extrato continua sendo feito pelo painel (`info-extrato` só existe depois de um
  upload).

## Lançar no caixa

Recebimentos e pagamentos em dinheiro entram no caixa da competência. É uma escrita de
risco médio: pede confirmação (`--yes` em scripts) e aceita `--dry-run`
([Ações que alteram dados](escrita.md)).

```sh
ctbz caixa adicionar --competencia 2026-09 --data 2026-09-15 --pagamento \
  --valor 150,25 --conta 1000000000000012 --descricao "Material de escritório"
```

```text
Adicionar pagamento de R$ 150,25 em 15/09/2026 no caixa de 09/2026: "Material de escritório" (Pagamento de Fornecedores) (risco médio)
Confirma? [s/N] s
Ação       adicionar
Situação   adicionado
ID         1000000000000099
…
```

- O valor é sempre positivo; `--recebimento` ou `--pagamento` define o sinal.
- `--conta` aceita o id ou a descrição exata de uma classificação aceita na competência
  (veja abaixo). Contas "Impostos - …" exigem `--guia ID` ou `--sem-guia`; "Sócios -
  Distribuição de Lucros Antecipados" exige `--socio ID`.
- Sem `--data`, vale o dia que o painel sugere: hoje na competência atual, senão o dia 1.
  Datas futuras são recusadas.
- Depois de salvar, a CLI relê o caixa e mostra o lançamento criado.

Para corrigir um lançamento manual, informe só o que muda; o resumo mostra o antes e o depois:

```sh
ctbz caixa editar 1000000000000002 --competencia 2026-09 --valor 900 --descricao "Venda balcão"
# Editar o lançamento 1000000000000002 do caixa de 09/2026: valor R$ 1.000,00 → R$ 900,00; descrição "Venda à vista" → "Venda balcão"
```

- Os ids aparecem em `ctbz caixa 2026-09 -o json`.
- Lançamentos feitos pelo sistema (coluna "Automático") não podem ser editados, como no painel.
- Trocar a classificação exige informar de novo o vínculo, quando a nova conta pede um.

Para excluir um lançamento manual:

```sh
ctbz caixa remover 1000000000000002 --competencia 2026-09
# Excluir permanentemente o lançamento 1000000000000002 do caixa de 09/2026: "Venda à vista" de 18/09/2026 no valor de R$ 1.000,00 (risco médio)
# Confirma? [s/N]
```

A API não tem "desfazer". Depois de remover, a CLI mostra os dados do lançamento e, no stderr,
o `ctbz caixa adicionar` que o recria.

## Classificar lançamentos do extrato

```sh
ctbz extratos lancamentos --conta-bancaria 1000000000000001 --competencia 2026-09
ctbz extratos classificar 1000000000000102 --conta-bancaria 1000000000000001 \
  --competencia 2026-09 --conta "Pagamento de Fornecedores"
```

```text
Classificar o lançamento 1000000000000102 ("PAGAMENTO BOLETO", -R$ 1.234,56): sem classificação → Pagamento de Fornecedores (risco médio)
Confirma? [s/N]
```

- O id da conta bancária vem de `ctbz contas-bancarias` ou `ctbz extratos`.
- `lancamentos` mostra a classificação atual e, nas partes de um lançamento desmembrado, o id
  do original (coluna "Parte de").
- Entradas aceitam classificações `RECEITA` e saídas, `DESPESA` (`ctbz extratos contas`).
  Classificações de sócio exigem `--socio` (ids dos sócios do extrato).
- Quando os contadores já fecharam o período, a classificação não pode mais ser trocada: o
  painel mostra "Nossos contadores já classificaram…", e a CLI recusa com a mesma mensagem.

## Desmembrar um lançamento do extrato

Um lançamento que mistura coisas (ex.: um boleto que paga aluguel e condomínio) pode ser
dividido em partes com classificações diferentes:

```sh
ctbz extratos desmembrar 1000000000000102 --conta-bancaria 1000000000000001 --competencia 2026-09 \
  --parte "Aluguel:1.000,00:Pagamento de Fornecedores" \
  --parte "Condomínio:234,56:1000000000000012"
```

- Cada `--parte` é `"descrição:valor:conta[:sócio]"`. O valor é positivo; o sinal é o do
  lançamento original. A descrição não pode ter `:`.
- São necessárias pelo menos 2 partes, nenhuma com valor zero, e a soma tem de ser igual ao
  valor do original (centavo por centavo).
- Não dá para desmembrar uma parte de outro desmembramento nem um lançamento com vínculo.

Para voltar ao lançamento original (o id pode ser o do original ou o de uma parte):

```sh
ctbz extratos desfazer-desmembramento 1000000000000102 --conta-bancaria 1000000000000001 --competencia 2026-09
```

## Classificações aceitas numa competência

Antes de lançar no caixa ou classificar um lançamento do extrato, veja quais classificações
a competência aceita (o mesmo filtro do painel):

```sh
ctbz caixa contas --competencia 2026-09 [--recebimento|--pagamento]
ctbz caixa contas --competencia 2026-09 --vinculos     # guias e sócios aceitos como vínculo
ctbz extratos contas --competencia 2026-09 [--receita|--despesa]
```

```text
ID                Classificação                                Tipo         Exige
1000000000000013  Impostos - Simples Nacional                  pagamento    guia
1000000000000014  Sócios - Distribuição de Lucros Antecipados  pagamento    sócio
```

- No caixa, "Exige" aponta as classificações que pedem um vínculo: uma guia de imposto
  (contas "Impostos - …", ou `0` para SEM GUIA) ou um sócio. Os ids vêm de `--vinculos`.
- No extrato, entradas aceitam classificações `RECEITA` e saídas, `DESPESA`. As marcadas em
  "Exige sócio" pedem `--socio`.

## Cadastrar e editar contas bancárias

```sh
ctbz contas-bancarias bancos                      # bancos aceitos (id e código)
ctbz contas-bancarias adicionar --banco 341 --agencia 1234 --conta 12345-6 \
  --abertura 2024-01-10 --declaro-conta-pj
ctbz contas-bancarias editar 1000000000000001 --abertura 2024-02-01
```

- `--declaro-conta-pj` é obrigatório, como o checkbox do painel: a conta é PJ e movimenta só
  este CNPJ.
- A data de abertura define desde quando a Contabilizei vai pedir os extratos da conta.
- A conta é digitada com dígito (`12345-6`); a CLI envia sem o hífen, como o painel.
- O painel pode bloquear a edição (por exemplo, com extratos já classificados); a CLI mostra
  o motivo.

Para excluir uma conta cadastrada por engano ou encerrada:

```sh
ctbz contas-bancarias remover 1000000000000001
```

É de **risco alto**: a exclusão é permanente, os vínculos da conta se perdem e um novo
cadastro não os recupera. No terminal, a CLI pede para digitar `confirmo`.

## Plano de contas (classificações)

```sh
ctbz contas --busca alug                # por trecho da descrição ou da conta contábil
ctbz contas --situacao ativo -o csv
```

```text
Descrição  Conta contábil            Classificação  Situação  ID
Aluguel    Aluguéis e Arrendamentos  DESPESA        ATIVO     1
```

- São as contas usadas para classificar entradas e saídas (as mesmas do `ctbz caixa`), cada
  uma ligada a uma conta contábil.
- `--busca` não diferencia maiúsculas, mas acentos contam: um trecho como `alug` acha
  "Aluguel" e "Aluguéis".
