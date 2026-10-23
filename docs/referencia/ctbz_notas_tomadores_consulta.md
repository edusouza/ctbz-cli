# ctbz notas tomadores consulta

Consulta os dados cadastrais de um CNPJ na Receita

Consulta um CNPJ como o emissor de notas faz ao cadastrar um tomador: razão social, nome
fantasia, abertura, atividade principal, natureza jurídica, situação cadastral, opção pelo
Simples, endereço e contatos.

## Uso

```
ctbz notas tomadores consulta CNPJ
```

## Exemplos

```sh
  ctbz notas tomadores consulta 00.000.000/0001-91 -o json
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz notas tomadores`](ctbz_notas_tomadores.md).
