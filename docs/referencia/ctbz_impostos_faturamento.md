# ctbz impostos faturamento

Mostra o faturamento, o pró-labore e os impostos pagos nos últimos 12 meses

Mostra mês a mês o faturamento e o pró-labore considerados na apuração (últimos 12 meses,
do mais antigo ao mais recente) e o total pago em impostos em cada mês.

Na tabela, o resumo vai para stderr: faturamento acumulado em 12 meses (RBT12, que define
a alíquota do Simples Nacional), pró-labore acumulado e o percentual do Fator R.

## Uso

```
ctbz impostos faturamento
```

## Exemplos

```sh
  ctbz impostos faturamento
  ctbz impostos faturamento -o csv > faturamento.csv
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
