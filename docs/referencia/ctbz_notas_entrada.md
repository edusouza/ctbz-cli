# ctbz notas entrada

Lista as notas fiscais de entrada (NF-e recebidas pela empresa)

Lista as NF-e de compra recebidas pela empresa, como as abas da tela "Notas fiscais de
entrada": a manifestar (padrão), manifestadas, a classificar e classificadas. Mostra emissão,
emitente, CNPJ do emitente, valor, situação, chave de acesso e o ID.

--mes escolhe o mês (AAAA-MM; padrão: o mês atual) e --emitente filtra pela razão social do
emitente. Para manifestar: ctbz notas entrada manifestar; para classificar: ctbz notas
entrada classificar (itens em ctbz notas entrada produtos).

## Uso

```
ctbz notas entrada [flags]
```

## Exemplos

```sh
  ctbz notas entrada
  ctbz notas entrada --lista manifestadas --mes 2026-09 -o csv
  ctbz notas entrada --lista a-classificar --emitente "ACME"
```

## Subcomandos

- [`ctbz notas entrada classificar`](ctbz_notas_entrada_classificar.md): Classifica notas de entrada (estoque, insumo, uso e consumo…), inteiras ou por produto
- [`ctbz notas entrada manifestar`](ctbz_notas_entrada_manifestar.md): Envia a manifestação do destinatário das notas de entrada à SEFAZ
- [`ctbz notas entrada produtos`](ctbz_notas_entrada_produtos.md): Lista os itens de uma nota de entrada e como estão classificados
- [`ctbz notas entrada reclassificar`](ctbz_notas_entrada_reclassificar.md): Muda a classificação de uma nota de entrada já classificada

## Flags

```
      --emitente string   filtra pela razão social do emitente
      --lista string      a-manifestar, manifestadas, a-classificar ou classificadas (default "a-manifestar")
      --mes string        mês, AAAA-MM (padrão: o mês atual)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz notas`](ctbz_notas.md).
