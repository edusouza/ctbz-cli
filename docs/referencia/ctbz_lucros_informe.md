# ctbz lucros informe

Mostra os valores do informe de rendimentos dos sócios (para o IR)

Mostra, para cada sócio, os valores do comprovante de rendimentos do ano: rendimentos
tributáveis (pró-labore), contribuição previdenciária (INSS), IRRF retido, 13º salário e
seu IRRF, e lucros e dividendos isentos. São os números que vão para a declaração de IR.

O painel monta o PDF do comprovante no navegador a partir desses mesmos valores; a API não
oferece o PDF. O padrão de --ano é o ano anterior (o da declaração).

O que bloqueia o informe e os aceites: ctbz lucros informe restricoes, carta, aceitar e
decidir.

## Uso

```
ctbz lucros informe [flags]
```

## Exemplos

```sh
  ctbz lucros informe
  ctbz lucros informe --ano 2025 -o csv > informe-2025.csv
```

## Subcomandos

- [`ctbz lucros informe aceitar`](ctbz_lucros_informe_aceitar.md): Aceita a carta de responsabilidade ou o termo de débitos federais do informe
- [`ctbz lucros informe carta`](ctbz_lucros_informe_carta.md): Mostra a carta de responsabilidade do informe de rendimentos de um ano
- [`ctbz lucros informe decidir`](ctbz_lucros_informe_decidir.md): Decide o que fazer com a pendência documental ou os débitos do informe
- [`ctbz lucros informe restricoes`](ctbz_lucros_informe_restricoes.md): Mostra o que bloqueia o informe de rendimentos de um ano e o que falta aceitar

## Flags

```
      --ano int   ano-calendário (padrão: o ano anterior)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz lucros`](ctbz_lucros.md).
