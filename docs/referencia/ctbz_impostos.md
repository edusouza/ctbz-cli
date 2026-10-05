# ctbz impostos

Lista as guias de impostos a pagar (em atraso, do mês e do próximo mês)

Lista as guias de impostos a pagar, agrupadas em atraso, deste mês e do próximo mês,
com competência, vencimento, valor e situação. Na tabela, os totais por grupo vão para stderr.

Com --fail-on-atraso, o comando termina com código 4 quando há guias em atraso (útil em
scripts e alertas).

## Uso

```
ctbz impostos [flags]
```

## Exemplos

```sh
  ctbz impostos
  ctbz impostos --atrasadas
  ctbz impostos -o csv > guias.csv
  ctbz impostos --fail-on-atraso || notify-send "Há impostos em atraso"
```

## Subcomandos

- [`ctbz impostos baixar`](ctbz_impostos_baixar.md): Baixa o PDF de guias de imposto
- [`ctbz impostos calculo`](ctbz_impostos_calculo.md): Mostra como o imposto do mês foi calculado
- [`ctbz impostos confirmar`](ctbz_impostos_confirmar.md): Informa que uma guia foi paga (ou, com --nao-paguei, que não foi)
- [`ctbz impostos debitos`](ctbz_impostos_debitos.md): Indica se a empresa tem débitos federais em aberto
- [`ctbz impostos desmarcar`](ctbz_impostos_desmarcar.md): Desfaz a confirmação de pagamento de uma guia
- [`ctbz impostos faturamento`](ctbz_impostos_faturamento.md): Mostra o faturamento, o pró-labore e os impostos pagos nos últimos 12 meses
- [`ctbz impostos guia`](ctbz_impostos_guia.md): Mostra o detalhe de uma guia de imposto
- [`ctbz impostos historico`](ctbz_impostos_historico.md): Lista o histórico de guias de impostos
- [`ctbz impostos parcelamento`](ctbz_impostos_parcelamento.md): Mostra o detalhe de um parcelamento de impostos
- [`ctbz impostos parcelamentos`](ctbz_impostos_parcelamentos.md): Lista os parcelamentos de impostos (em andamento, ativos e encerrados)
- [`ctbz impostos recorrente`](ctbz_impostos_recorrente.md): Mostra a situação do pagamento recorrente (débito automático) de impostos
- [`ctbz impostos tabela-irrf`](ctbz_impostos_tabela-irrf.md): Mostra a tabela progressiva do IRRF usada no cálculo do pró-labore

## Flags

```
      --atrasadas        mostra só as guias em atraso
      --fail-on-atraso   termina com código 4 se houver guias em atraso
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
