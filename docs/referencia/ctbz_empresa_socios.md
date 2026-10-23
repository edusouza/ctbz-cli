# ctbz empresa socios

Lista os sócios da empresa

Lista os sócios com CPF, se são administradores ou responsáveis perante a Receita,
categoria (com ou sem pró-labore), salário-base e data de entrada.

## Uso

```
ctbz empresa socios
```

## Exemplos

```sh
  ctbz empresa socios
  ctbz empresa socios -o json
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz empresa`](ctbz_empresa.md).
