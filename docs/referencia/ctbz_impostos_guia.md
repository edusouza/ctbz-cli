# ctbz impostos guia

Mostra o detalhe de uma guia de imposto

Mostra o detalhe de uma guia: imposto, competência, vencimento, valor total, valor
original, juros e multa, situação e ações disponíveis no painel. O ID vem de "ctbz impostos".

Para ver como o imposto do mês foi calculado, use "ctbz impostos calculo".

## Uso

```
ctbz impostos guia ID
```

## Exemplos

```sh
  ctbz impostos guia 1000000000000001
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
