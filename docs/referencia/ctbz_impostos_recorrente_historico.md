# ctbz impostos recorrente historico

Lista os pagamentos recorrentes de impostos dos últimos meses

Lista, por competência, as guias pagas pelo pagamento recorrente, com situação e valor,
e o custo de operação de cada mês. A Contabilizei só devolve um período recente; para os
meses anteriores, use "ctbz impostos historico".

## Uso

```
ctbz impostos recorrente historico
```

## Exemplos

```sh
  ctbz impostos recorrente historico
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz impostos recorrente`](ctbz_impostos_recorrente.md).
