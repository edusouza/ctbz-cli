# ctbz plano contrato

Exporta o contrato de prestação de serviços (HTML ou texto)

Imprime no stdout o contrato de prestação de serviços aceito pela empresa, em HTML (como o
painel mostra) ou, com --texto, em texto simples. Redirecione para um arquivo para guardar.

## Uso

```
ctbz plano contrato [flags]
```

## Exemplos

```sh
  ctbz plano contrato > contrato.html
  ctbz plano contrato --texto | less
```

## Flags

```
      --texto   converte o HTML em texto simples
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz plano`](ctbz_plano.md).
