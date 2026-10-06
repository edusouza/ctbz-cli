# ctbz notas entrada classificar

Classifica notas de entrada (estoque, insumo, uso e consumo…), inteiras ou por produto

Classifica notas da lista "a classificar" do mês: estoque, insumo, uso-consumo,
ativo-imobilizado ou prestacao-servico.

--como classifica as notas inteiras numa só opção; vários IDs vão num único envio, como o lote
do painel. --produto distribui a quantidade de cada item de uma nota (um ID só), no formato
ITEM:opcao=quantidade,opcao=quantidade (quantidades com ponto decimal); os itens não citados
ficam como estão. A soma de cada item precisa ser igual à quantidade total, sem negativos. Os
itens e a distribuição atual estão em ctbz notas entrada produtos.

O prazo é o dia 05 do mês seguinte à emissão; depois dele a Contabilizei confirma a
pré-classificação sugerida. --mes diz em que mês procurar as notas (padrão: o mês atual).
Para mudar uma nota já classificada: ctbz notas entrada reclassificar.

Risco médio: é contábil e pode ser refeito. Aceita --yes e --dry-run.

## Uso

```
ctbz notas entrada classificar ID... [flags]
```

## Exemplos

```sh
  ctbz notas entrada classificar 1000000000000701 --como uso-consumo
  ctbz notas entrada classificar 1000000000000701 1000000000000702 --como estoque --mes 2026-09
  ctbz notas entrada classificar 1000000000000701 --produto 11:estoque=2,uso-consumo=3 --produto 12:insumo=1
```

## Flags

```
      --como string           classifica as notas inteiras: estoque, insumo, uso-consumo, ativo-imobilizado, prestacao-servico
      --dry-run               mostra a requisição que seria enviada, sem enviar
      --mes string            mês das notas, AAAA-MM (padrão: o mês atual)
      --produto stringArray   distribui um item, ITEM:opcao=qtd,opcao=qtd (repetível)
  -y, --yes                   envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz notas entrada`](ctbz_notas_entrada.md).
