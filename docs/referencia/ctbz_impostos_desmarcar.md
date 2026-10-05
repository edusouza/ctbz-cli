# ctbz impostos desmarcar

Desfaz a confirmação de pagamento de uma guia

Desfaz uma confirmação de pagamento feita por engano ("Desmarcar" no painel): a guia volta
a aparecer como a pagar. É a mesma requisição de ctbz impostos confirmar --nao-paguei.

Risco médio. Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz impostos desmarcar ID [flags]
```

## Exemplos

```sh
  ctbz impostos desmarcar 1000000000000001
```

## Flags

```
      --dry-run   mostra a requisição que seria enviada, sem enviar
  -y, --yes       envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
