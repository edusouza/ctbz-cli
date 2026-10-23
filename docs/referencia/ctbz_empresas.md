# ctbz empresas

Lista as empresas do usuário, marcando a atual

Lista todas as empresas vinculadas ao usuário, inclusive as inativas, com a situação
de cada uma. A coluna "atual" marca a empresa da sessão; para trocar, use "ctbz empresa usar".

## Uso

```
ctbz empresas
```

## Exemplos

```sh
  ctbz empresas
  ctbz empresas -o csv
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
