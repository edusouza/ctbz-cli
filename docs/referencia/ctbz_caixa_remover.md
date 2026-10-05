# ctbz caixa remover

Exclui permanentemente um lançamento manual do caixa

Exclui um lançamento manual do caixa da competência (risco médio). A API não tem como
desfazer: para voltar atrás, é preciso recriar o lançamento. Por isso o resumo mostra
descrição, data e valor antes da confirmação, e a saída traz os dados do lançamento removido
e, no stderr, o comando ctbz caixa adicionar que o recria.

Lançamentos feitos pelo sistema não podem ser removidos, como no painel.
Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz caixa remover ID [flags]
```

## Exemplos

```sh
  ctbz caixa remover 1000000000000002 --competencia 2026-09
  ctbz caixa remover 1000000000000002 --competencia 2026-09 --yes -o json
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --dry-run              mostra a requisição que seria enviada, sem enviar
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz caixa`](ctbz_caixa.md).
