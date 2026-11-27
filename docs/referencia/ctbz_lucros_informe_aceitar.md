# ctbz lucros informe aceitar

Aceita a carta de responsabilidade ou o termo de débitos federais do informe

Dá os aceites que o painel pede para liberar o informe de rendimentos de um ano:

  carta-responsabilidade  a carta exigida para ver o informe (o texto aparece antes da
                          confirmação; também em ctbz lucros informe carta)
  termo-debitos           a ciência dos débitos federais da empresa, quando há débitos

Risco alto: são declarações sem desfazer, e a CLI pede "confirmo". Se o informe aponta débitos
mas a API não os lista, a CLI avisa e pede um segundo "confirmo" (ou --aceitar-sem-lista, que
é obrigatório com --yes). O padrão de --ano é o ano anterior. Aceita --yes e --dry-run.

## Uso

```
ctbz lucros informe aceitar carta-responsabilidade|termo-debitos [flags]
```

## Exemplos

```sh
  ctbz lucros informe aceitar carta-responsabilidade --ano 2025
  ctbz lucros informe aceitar termo-debitos --ano 2025
```

## Flags

```
      --aceitar-sem-lista   aceita o termo mesmo que a API não liste os débitos (obrigatório com --yes nesse caso)
      --ano int             ano-calendário (padrão: o ano anterior)
      --dry-run             mostra a requisição que seria enviada, sem enviar
  -y, --yes                 envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz lucros informe`](ctbz_lucros_informe.md).
