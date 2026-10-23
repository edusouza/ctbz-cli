# ctbz prolabore parametros

Mostra os valores usados no cálculo do pró-labore (INSS e IRRF)

Mostra os parâmetros que a Contabilizei usa para calcular o pró-labore: salário mínimo
(pró-labore mínimo), alíquota do INSS do sócio, contribuição máxima ao INSS, pró-labore a
partir do qual a contribuição atinge o teto e o valor a partir do qual incide IRRF.

## Uso

```
ctbz prolabore parametros
```

## Exemplos

```sh
  ctbz prolabore parametros
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz prolabore`](ctbz_prolabore.md).
