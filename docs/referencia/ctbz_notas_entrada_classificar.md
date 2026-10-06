# ctbz notas entrada classificar

Classifica notas de entrada inteiras (estoque, insumo, uso e consumo…)

Classifica as notas da lista "a classificar" numa só opção: estoque, insumo, uso-consumo,
ativo-imobilizado ou prestacao-servico. Vários IDs vão num único envio, como o lote do painel.

O prazo é o dia 05 do mês seguinte à emissão; depois dele a Contabilizei confirma a
pré-classificação sugerida. --mes diz em que mês procurar as notas (padrão: o mês atual).
Os itens de uma nota estão em ctbz notas entrada produtos.

Risco médio: é contábil e pode ser refeito. Aceita --yes e --dry-run.

## Uso

```
ctbz notas entrada classificar ID... [flags]
```

## Exemplos

```sh
  ctbz notas entrada classificar 1000000000000701 --como uso-consumo
  ctbz notas entrada classificar 1000000000000701 1000000000000702 --como estoque --mes 2026-09
```

## Flags

```
      --como string   classificação: estoque, insumo, uso-consumo, ativo-imobilizado, prestacao-servico
      --dry-run       mostra a requisição que seria enviada, sem enviar
      --mes string    mês das notas, AAAA-MM (padrão: o mês atual)
  -y, --yes           envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz notas entrada`](ctbz_notas_entrada.md).
