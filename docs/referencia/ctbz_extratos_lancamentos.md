# ctbz extratos lancamentos

Lista os lançamentos do extrato de uma conta bancária num mês

Lista os lançamentos do extrato importado de uma conta bancária na competência: id, data,
descrição, valor (negativo nas saídas), classificação atual e, nas partes de um lançamento
desmembrado, o id do lançamento original ("Parte de").

Os ids são usados por ctbz extratos classificar e desmembrar.

## Uso

```
ctbz extratos lancamentos [flags]
```

## Exemplos

```sh
  ctbz extratos lancamentos --conta-bancaria 1000000000000001 --competencia 2026-09
  ctbz extratos lancamentos --conta-bancaria 1000000000000001 --competencia 2026-09 -o csv
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz extratos`](ctbz_extratos.md).
