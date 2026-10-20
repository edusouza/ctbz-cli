# ctbz lucros informe decidir

Decide o que fazer com a pendência documental ou os débitos do informe

Registra a decisão que o painel pede diante das restrições do informe de rendimentos:

  --regularizar-pendencia      regularizar a pendência documental (a CLI mostra o formulário
                               de atendimento onde enviar os documentos)
  --nao-regularizar-pendencia  não regularizar: não haverá distribuição de lucros
  --nao-distribuir-lucros      não distribuir lucros no ano (diante de débitos federais)

As duas primeiras exigem pendência documental e a última, débitos federais (ver ctbz lucros
informe restricoes). Risco alto: decide se haverá distribuição de lucros isenta e não dá para
desfazer. O padrão de --ano é o ano anterior. Aceita --yes e --dry-run.

## Uso

```
ctbz lucros informe decidir [flags]
```

## Exemplos

```sh
  ctbz lucros informe decidir --ano 2025 --regularizar-pendencia
```

## Flags

```
      --ano int                     ano-calendário (padrão: o ano anterior)
      --dry-run                     mostra a requisição que seria enviada, sem enviar
      --nao-distribuir-lucros       não distribuir lucros
      --nao-regularizar-pendencia   não regularizar a pendência documental
      --regularizar-pendencia       regularizar a pendência documental
  -y, --yes                         envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz lucros informe`](ctbz_lucros_informe.md).
