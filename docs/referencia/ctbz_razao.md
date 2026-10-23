# ctbz razao

Lista os lançamentos do razão contábil por conta

Lista os lançamentos do razão de cada conta no período: data, conta, histórico,
contrapartida, débito, crédito e o saldo acumulado no exercício depois do lançamento.

O período vai de --de até --ate (meses AAAA-MM; padrão: o mês atual), no máximo 24 meses;
a API devolve um mês por vez. --conta filtra pelo código da conta (ex.: 1.01.01.01.00) ou
por um prefixo dele (ex.: 1.01 para todo o ativo circulante).

## Uso

```
ctbz razao [flags]
```

## Exemplos

```sh
  ctbz razao --de 2026-01 --ate 2026-09
  ctbz razao --conta 1.01.01.01.00 --de 2026-07 -o csv
```

## Flags

```
      --ate string     último mês, AAAA-MM (padrão: --de ou o mês atual)
      --conta string   código da conta ou prefixo (ex.: 1.01)
      --de string      primeiro mês, AAAA-MM (padrão: o mês atual)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
