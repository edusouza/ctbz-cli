# ctbz balancete

Mostra o balancete de verificação de um mês

Mostra o balancete do mês: cada conta da árvore contábil com saldo anterior, débitos,
créditos e saldo do exercício. Na tabela, as contas aparecem recuadas por nível; em CSV e
JSON, a descrição vem sem recuo e o nível fica na coluna "nivel" (bom para planilhas).

O padrão do painel é dezembro do ano anterior; aqui o mês é obrigatório.

## Uso

```
ctbz balancete AAAA-MM
```

## Exemplos

```sh
  ctbz balancete 2026-08
  ctbz balancete 2026-08 -o csv > balancete-2026-08.csv
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
