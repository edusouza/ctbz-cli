# ctbz impostos parcelamento simular

Simula um parcelamento de débitos (opções de parcelas e custos)

Mostra as opções de parcelamento de um tipo de débito: quantidade de parcelas, valor da
primeira e das demais, e os custos adicionais cobrados na mensalidade (serviço adicional e
emissão de guia). Só leitura: nada é contratado.

TIPO: pgfn-nao-previdenciario, pgfn-previdenciario, pgfn-simples-nacional, simples-nacional, especializado. O tipo especializado (negociação feita
por um especialista) exige --negociacao VENCIDOS|DIVIDA_ATIVA|DIVIDA_ATIVA_E_VENCIDOS.

Sem débitos para parcelar, a CLI avisa e termina com código 0.

## Uso

```
ctbz impostos parcelamento simular TIPO [flags]
```

## Exemplos

```sh
  ctbz impostos parcelamento simular simples-nacional
  ctbz impostos parcelamento simular especializado --negociacao DIVIDA_ATIVA -o json
```

## Flags

```
      --negociacao string   no tipo especializado: VENCIDOS, DIVIDA_ATIVA, DIVIDA_ATIVA_E_VENCIDOS
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz impostos parcelamento`](ctbz_impostos_parcelamento.md).
