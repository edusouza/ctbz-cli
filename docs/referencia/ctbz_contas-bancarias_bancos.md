# ctbz contas-bancarias bancos

Lista os bancos aceitos no cadastro de contas bancárias

Lista os bancos aceitos no cadastro de contas (id, código e nome); use o id ou o código em --banco.

Alguns códigos se repetem (ex.: dois bancos com 341) ou têm uma versão internacional (077 e
077-1). --banco escolhe o código exato; se mais de um banco tiver o código, a CLI lista os
candidatos e pede o id.

## Uso

```
ctbz contas-bancarias bancos
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz contas-bancarias`](ctbz_contas-bancarias.md).
