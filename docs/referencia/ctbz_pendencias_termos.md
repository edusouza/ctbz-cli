# ctbz pendencias termos

Lista os termos e cartas com aceite pendente na Central de Rotinas

Lista os termos e cartas que a Central de Rotinas pede para aceitar: carta de
responsabilidade, termo de ciência sobre retiradas de lucros (com o prazo do aceite tácito, em
dias) e termo de adesão ao TotalPass. --todos inclui os que não estão pendentes.

Para ler um termo inteiro: ctbz pendencias termo CHAVE. Para aceitar: ctbz pendencias aceitar CHAVE.

## Uso

```
ctbz pendencias termos [flags]
```

## Exemplos

```sh
  ctbz pendencias termos
  ctbz pendencias termos --todos -o json
```

## Flags

```
      --todos   inclui os termos sem aceite pendente
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz pendencias`](ctbz_pendencias.md).
