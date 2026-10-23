# ctbz prolabore fator-r

Mostra a situação do Fator R e os anexos possíveis de cada atividade

Mostra se a Contabilizei ajusta o pró-labore para o Fator R (motor do Fator R), o Fator R
atual (pró-labore ÷ faturamento dos últimos 12 meses), os valores acumulados usados no
cálculo e, para cada atividade, os anexos do Simples Nacional em que pode ser tributada.

Com Fator R a partir de 28%, as atividades sujeitas a ele saem do Anexo V (alíquota inicial
de 15,5%) para o Anexo III (6%). Os dados vêm do simulador de impostos avançado do painel
(só leitura) e da memória de cálculo do mês.

## Uso

```
ctbz prolabore fator-r
```

## Exemplos

```sh
  ctbz prolabore fator-r
  ctbz prolabore fator-r -o json | jq .fator_r
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz prolabore`](ctbz_prolabore.md).
