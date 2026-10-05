# ctbz impostos parcelamento

Mostra o detalhe de um parcelamento de impostos

Mostra o detalhe de um parcelamento (descrição, período da dívida, saldo devedor,
parcelas). O ID vem de "ctbz impostos parcelamentos".

O formato desta resposta não pôde ser verificado (a conta usada no desenvolvimento não tinha
parcelamentos), por isso os campos são mostrados como a API os devolve.

## Uso

```
ctbz impostos parcelamento ID
```

## Exemplos

```sh
  ctbz impostos parcelamento 1000000000000001
```

## Subcomandos

- [`ctbz impostos parcelamento simular`](ctbz_impostos_parcelamento_simular.md): Simula um parcelamento de débitos (opções de parcelas e custos)

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
