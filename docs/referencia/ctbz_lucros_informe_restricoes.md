# ctbz lucros informe restricoes

Mostra o que bloqueia o informe de rendimentos de um ano e o que falta aceitar

Mostra as restrições do informe de rendimentos de um ano-calendário: débitos federais,
pendência documental (e como regularizá-la), reabertura do balanço e se a carta de
responsabilidade ainda precisa ser aceita. O padrão de --ano é o ano anterior.

Para aceitar: ctbz lucros informe aceitar; para decidir sobre as restrições:
ctbz lucros informe decidir.

## Uso

```
ctbz lucros informe restricoes [flags]
```

## Exemplos

```sh
  ctbz lucros informe restricoes --ano 2025
```

## Flags

```
      --ano int   ano-calendário (padrão: o ano anterior)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz lucros informe`](ctbz_lucros_informe.md).
