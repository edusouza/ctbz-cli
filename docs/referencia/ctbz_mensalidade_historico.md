# ctbz mensalidade historico

Lista as faturas pagas da mensalidade

Lista as faturas da Contabilizei já pagas, como a tela "Consultar histórico" do painel: por
competência, o vencimento, a data do pagamento, o valor, a multa e os juros cobrados pelo
atraso da fatura anterior, a forma de pagamento e o status. A mais recente vem primeiro.

A baixa da fatura pode demorar depois do pagamento: para saber se a mensalidade está em dia,
use ctbz mensalidade situacao.

Os campos debito_automatico_ativo, debito_automatico_habilitado, status, competencia e
proximo_pagamento são do débito automático de impostos (o mesmo de ctbz impostos recorrente),
não da mensalidade. Ficam por compatibilidade e estão obsoletos.

Só campos conhecidos são copiados para a saída: dados de cartão, chaves do gateway de
pagamento e os códigos de pagamento das faturas nunca são mostrados.

## Uso

```
ctbz mensalidade historico
```

## Exemplos

```sh
  ctbz mensalidade historico
  ctbz mensalidade historico -o json | jq '.pagamentos[] | {competencia, valor, multa_juros}'
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz mensalidade`](ctbz_mensalidade.md).
