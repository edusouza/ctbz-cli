# ctbz plano

Mostra o plano contratado com a Contabilizei

Mostra o plano contratado com a Contabilizei: descrição, categoria, valor de tabela e ramos
de atividade cobertos. Os dados vêm do login (não fazem chamada à API); depois de uma
mudança de plano, rode "ctbz login" de novo.

Para o texto do contrato de serviço e da proposta do plano, use "ctbz plano contrato" e
"ctbz plano proposta".

## Uso

```
ctbz plano
```

## Exemplos

```sh
  ctbz plano
  ctbz plano contrato --texto > contrato.txt
  ctbz plano proposta > proposta.html
```

## Subcomandos

- [`ctbz plano contrato`](ctbz_plano_contrato.md): Exporta o contrato de prestação de serviços (HTML ou texto)
- [`ctbz plano proposta`](ctbz_plano_proposta.md): Exporta a proposta do plano contratado, com a tabela de preços (HTML ou texto)

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
