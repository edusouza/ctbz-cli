# ctbz impostos parcelamentos

Lista os parcelamentos de impostos (em andamento, ativos e encerrados)

Lista os parcelamentos de impostos da empresa, como a aba "Parcelamentos" do painel:
em negociação (em_andamento), ativos e encerrados (historico). Para o detalhe de um
parcelamento, use "ctbz impostos parcelamento ID"; para simular um novo, "ctbz impostos
parcelamento simular TIPO".

Para contratar: "ctbz impostos parcelamento contratar TIPO".

## Uso

```
ctbz impostos parcelamentos
```

## Exemplos

```sh
  ctbz impostos parcelamentos
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
