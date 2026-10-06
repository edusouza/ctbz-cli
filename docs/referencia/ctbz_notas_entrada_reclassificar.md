# ctbz notas entrada reclassificar

Muda a classificação de uma nota de entrada já classificada

Reclassifica uma nota da lista "classificadas" do mês. --como põe toda a quantidade de
cada item numa só opção (estoque, insumo, uso-consumo, ativo-imobilizado ou
prestacao-servico); --produto distribui item a item, como em ctbz notas entrada classificar.

A nota fica PROCESSANDO até a Contabilizei aplicar a nova classificação.
Risco médio: é contábil e pode ser refeito. Aceita --yes e --dry-run.

## Uso

```
ctbz notas entrada reclassificar ID [flags]
```

## Exemplos

```sh
  ctbz notas entrada reclassificar 1000000000000701 --como estoque --mes 2026-09
  ctbz notas entrada reclassificar 1000000000000701 --produto 11:estoque=5
```

## Flags

```
      --como string           toda a quantidade numa opção: estoque, insumo, uso-consumo, ativo-imobilizado, prestacao-servico
      --dry-run               mostra a requisição que seria enviada, sem enviar
      --mes string            mês da nota, AAAA-MM (padrão: o mês atual)
      --produto stringArray   distribui um item, ITEM:opcao=qtd,opcao=qtd (repetível)
  -y, --yes                   envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz notas entrada`](ctbz_notas_entrada.md).
