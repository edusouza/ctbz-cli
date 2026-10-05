# ctbz extratos contas

Lista as classificações aceitas nos lançamentos do extrato de uma competência

Lista as classificações que os lançamentos do extrato aceitam na competência, com o filtro
do painel: exibidas no mês, ativas, RECEITA (para entradas) ou DESPESA (para saídas). As
classificações com "Exige sócio" pedem --socio em ctbz extratos classificar.

## Uso

```
ctbz extratos contas [flags]
```

## Exemplos

```sh
  ctbz extratos contas --competencia 2026-09
  ctbz extratos contas --competencia 2026-09 --despesa -o csv
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --despesa              só as classificações de saída (DESPESA)
      --receita              só as classificações de entrada (RECEITA)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz extratos`](ctbz_extratos.md).
