# ctbz contas

Lista o plano de contas usado para classificar os lançamentos

Lista as contas usadas para classificar entradas e saídas (as mesmas da tela de
classificação de extratos): descrição, conta contábil correspondente, classificação (receita,
despesa…) e situação.

--busca filtra por um trecho da descrição ou da conta contábil, sem diferenciar maiúsculas
(acentos contam: use um trecho como "alug" para achar "Aluguel" e "Aluguéis"). --situacao
filtra por ativo ou inativo; por padrão aparecem todas.

## Uso

```
ctbz contas [flags]
```

## Exemplos

```sh
  ctbz contas --busca alug
  ctbz contas --situacao ativo -o csv
```

## Flags

```
      --busca string      trecho da descrição ou da conta contábil
      --situacao string   ativo ou inativo (padrão: todas)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
