# ctbz balanco

Mostra o balanço patrimonial (ativo, passivo e patrimônio líquido)

Mostra o balanço patrimonial: as contas de ativo, passivo e patrimônio líquido com o saldo
do exercício e o do exercício anterior, recuadas por nível na tabela. As contas de resultado
ficam de fora, como no painel.

Só com o ano, o balanço é o de dezembro (fechamento do exercício). Para reabrir o
exercício e regularizar a pendência documental do informe: ctbz balanco reabrir.

## Uso

```
ctbz balanco AAAA[-MM]
```

## Exemplos

```sh
  ctbz balanco 2025
  ctbz balanco 2026-09 -o csv
```

## Subcomandos

- [`ctbz balanco reabrir`](ctbz_balanco_reabrir.md): Reabre o balanço de um ano para regularizar a pendência documental do informe

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
