# ctbz mensalidade

Mostra a mensalidade atual da Contabilizei (valor, vencimento e situação)

Mostra a fatura atual da Contabilizei: competência, valor, vencimento, situação e
observações (ex.: serviços adicionais), e se há mensalidade de competência anterior em atraso.

Com --fail-on-atraso, o comando termina com código 4 quando há competência anterior em atraso.

## Uso

```
ctbz mensalidade [flags]
```

## Exemplos

```sh
  ctbz mensalidade
  ctbz mensalidade -o json | jq .valor
  ctbz mensalidade --fail-on-atraso
```

## Subcomandos

- [`ctbz mensalidade historico`](ctbz_mensalidade_historico.md): Lista as faturas pagas da mensalidade
- [`ctbz mensalidade situacao`](ctbz_mensalidade_situacao.md): Indica se a empresa está em dia com a Contabilizei

## Flags

```
      --fail-on-atraso   termina com código 4 se houver competência anterior em atraso
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
