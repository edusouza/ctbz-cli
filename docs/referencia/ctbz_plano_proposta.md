# ctbz plano proposta

Exporta a proposta do plano contratado, com a tabela de preços (HTML ou texto)

Imprime no stdout a proposta de prestação de serviços do plano contratado (tabela de
mensalidades por faixa de faturamento e o plano selecionado), em HTML ou, com --texto, em
texto simples.

## Uso

```
ctbz plano proposta [flags]
```

## Exemplos

```sh
  ctbz plano proposta --texto
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
