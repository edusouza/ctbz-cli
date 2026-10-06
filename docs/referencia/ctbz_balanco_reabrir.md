# ctbz balanco reabrir

Reabre o balanço de um ano para regularizar a pendência documental do informe

Reabre o exercício contábil (balanço) de um ano, o "Regularizar pendências" do informe de
rendimentos. Só é enviado quando o informe do ano tem pendência documental, o caminho de
regularização é a reabertura do balanço e não há outra reabertura em andamento.

Quando a regularização é um serviço pago (contratar serviço adicional), a CLI recusa e mostra
o custo: contrate pelo painel. Veja a situação em ctbz lucros informe restricoes.

Risco alto: afeta o balanço, o informe de rendimentos (indisponível até a análise) e a
distribuição de lucros, sem desfazer pela API. Aceita --yes e --dry-run.

## Uso

```
ctbz balanco reabrir [flags]
```

## Exemplos

```sh
  ctbz balanco reabrir --ano 2025
```

## Flags

```
      --ano int   ano do exercício a reabrir
      --dry-run   mostra a requisição que seria enviada, sem enviar
  -y, --yes       envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz balanco`](ctbz_balanco.md).
