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
Data        Descrição               Conta                         Classificação        Valor  Situação    Automático
18/09/2026  Pagamento Contabilizei  Mensalidade de contabilidade  DESPESA         -R$ 141,83  CONFIRMADO  sim
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
