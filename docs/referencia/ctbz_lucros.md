# ctbz lucros

Mostra a distribuição de lucros do exercício e o que a impede

Mostra a distribuição de lucros como a tela de informe de rendimentos do painel: exercício,
saldo disponível na empresa, total distribuído aos sócios, adiantamentos, limite permitido,
se o exercício está fechado, se a distribuição ainda pode ser alterada (e até quando) e o
valor por sócio, com o id que ctbz lucros distribuir pede em --socio.

Também mostra as restrições do exercício: pendências documentais, débitos federais e
reabertura do balanço. A API não recebe ano: o exercício é o que a Contabilizei tem aberto
(sem exercício aberto, as restrições são as do ano anterior).

## Uso

```
ctbz lucros
```

## Exemplos

```sh
  ctbz lucros
  ctbz lucros -o json | jq .saldo
```

## Subcomandos

- [`ctbz lucros distribuir`](ctbz_lucros_distribuir.md): Registra quanto do lucro do exercício cabe a cada sócio
- [`ctbz lucros informe`](ctbz_lucros_informe.md): Mostra os valores do informe de rendimentos dos sócios (para o IR)

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
