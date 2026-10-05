# ctbz rotinas

Lista as rotinas e obrigações do mês (da empresa e da Contabilizei)

Lista as rotinas do mês, como a Central de Rotinas do painel: as de responsabilidade da
empresa (importar extrato, pagar impostos e a mensalidade) e as obrigações entregues pela
Contabilizei (eSocial, DCTFWeb, EFD-Reinf…), com prazo, status e valor.

O mês é o do prazo (padrão: o mês atual). A API só devolve o mês anterior, o atual e o
próximo; outros meses saem vazios.

A coluna alerta marca prazos vencidos ("vencida") e que vencem em até 7 dias ("próxima").
A coluna pendencias traz os ids das pendências da rotina (ex.: lançamentos para reclassificar
com ctbz rotinas reclassificar).
Com --fail-on-vencidas, o comando termina com código 4 quando há rotina da empresa vencida.

## Uso

```
ctbz rotinas [flags]
```

## Exemplos

```sh
  ctbz rotinas
  ctbz rotinas --mes 2026-11
  ctbz rotinas --fail-on-vencidas -o json
```

## Subcomandos

- [`ctbz rotinas reclassificar`](ctbz_rotinas_reclassificar.md): Confirma ou altera a classificação de lançamentos pedida pela Central de Rotinas

## Flags

```
      --fail-on-vencidas   termina com código 4 se houver rotina da empresa vencida
      --mes string         mês do prazo, AAAA-MM (padrão: o mês atual)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
