# ctbz pendencias conciliacao candidatos

Lista as notas ou os recebimentos que podem ser vinculados numa conciliação

Lista os itens que podem ser vinculados a uma pendência: --notas (para pendências de
recebimento sem nota) ou --recebimentos (para notas sem recebimento). Use os ids em
ctbz pendencias conciliacao resolver --vincular.

O formato desta resposta não pôde ser verificado (a conta usada no desenvolvimento não tinha
pendências), por isso os campos saem como a API os devolve.

## Uso

```
ctbz pendencias conciliacao candidatos [flags]
```

## Exemplos

```sh
  ctbz pendencias conciliacao candidatos --notas -o json
```

## Flags

```
      --notas          notas fiscais para vincular a recebimentos
      --recebimentos   recebimentos para vincular a notas
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz pendencias conciliacao`](ctbz_pendencias_conciliacao.md).
