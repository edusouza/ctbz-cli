# ctbz impostos recorrente

Mostra a situação do pagamento recorrente (débito automático) de impostos

Mostra se o pagamento recorrente de impostos (débito automático no cartão ou na conta PJ)
está disponível e ativo, a competência atual, as datas do próximo pagamento e da próxima
tentativa e quantos pagamentos estão agendados, concluídos e recusados.

Dados de cartão e chaves de pagamento nunca são mostrados: só a quantidade de cartões salvos.
Para os meses anteriores, use "ctbz impostos recorrente historico".

## Uso

```
ctbz impostos recorrente
```

## Exemplos

```sh
  ctbz impostos recorrente
  ctbz impostos recorrente historico -o csv
```

## Subcomandos

- [`ctbz impostos recorrente historico`](ctbz_impostos_recorrente_historico.md): Lista os pagamentos recorrentes de impostos dos últimos meses

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
