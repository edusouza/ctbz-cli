# ctbz logout

Apaga a sessão local

Apaga a sessão e qualquer login pendente salvos em CTBZ_HOME.

A Contabilizei não tem logout no servidor: a sessão continua válida lá até expirar.

## Uso

```
ctbz logout
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
