# ctbz prolabore gestao-inteligente

Põe a empresa no cálculo automático do pró-labore ou a tira dele

Gestão inteligente é o cálculo automático do pró-labore: a Contabilizei define o valor
todo mês para manter o Fator R no melhor ponto (ver ctbz prolabore fator-r).

  ativar  põe a empresa na gestão inteligente (só se ela for elegível)
  sair    tira a empresa: os sócios passam a ajustar o pró-labore todo mês

Risco alto: muda quem decide o valor do pró-labore. Aceitam --yes e --dry-run.

## Uso

```
ctbz prolabore gestao-inteligente
```

## Subcomandos

- [`ctbz prolabore gestao-inteligente ativar`](ctbz_prolabore_gestao-inteligente_ativar.md): Põe a empresa no cálculo automático do pró-labore
- [`ctbz prolabore gestao-inteligente sair`](ctbz_prolabore_gestao-inteligente_sair.md): Tira a empresa do cálculo automático do pró-labore

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz prolabore`](ctbz_prolabore.md).
