# ctbz impostos debitos

Indica se a empresa tem débitos federais em aberto

Indica se a Contabilizei identificou débitos federais em aberto para a empresa.
Com --fail-on-debitos, termina com código 4 quando há débitos.

## Uso

```
ctbz impostos debitos [flags]
```

## Exemplos

```sh
  ctbz impostos debitos
  ctbz impostos debitos --fail-on-debitos -o json
```

## Flags

```
      --fail-on-debitos   termina com código 4 se houver débitos federais
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
