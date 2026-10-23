# ctbz primeiros-passos reativar

Volta a mostrar as tarefas dos primeiros passos

Volta a mostrar as tarefas dos primeiros passos (risco baixo: dispensar e reativar desfazem um ao outro).
Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz primeiros-passos reativar [flags]
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
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz primeiros-passos`](ctbz_primeiros-passos.md).
