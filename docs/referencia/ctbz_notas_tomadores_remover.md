# ctbz notas tomadores remover

Exclui um tomador da lista de clientes

Exclui um tomador da lista de clientes do emissor (risco médio: não dá para desfazer, só
recadastrar com ctbz notas tomadores adicionar). Nacionais são identificados pelo documento;
tomadores do exterior, pelo id. As notas já emitidas para o tomador não mudam (a confirmar).

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz notas tomadores remover DOCUMENTO|ID [flags]
```

## Exemplos

```sh
  ctbz notas tomadores remover 00000000000191
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

Veja também: [`ctbz notas tomadores`](ctbz_notas_tomadores.md).
