# ctbz impostos calculo

Mostra como o imposto do mês foi calculado

Mostra a memória de cálculo do mês mais recente, como a tela "Como meu imposto foi
calculado": faturamento, DAS do Simples Nacional, DARF de INSS e IRRF sobre o pró-labore,
faturamento e pró-labore dos últimos 12 meses e o Fator R.

A competência é escolhida pela Contabilizei (a API não recebe mês).

## Uso

```
ctbz impostos calculo
```

## Exemplos

```sh
  ctbz impostos calculo
  ctbz impostos calculo -o json | jq .fator_r
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
