# ctbz impostos historico

Lista o histórico de guias de impostos

Lista as guias de meses anteriores, com valor, valor pago, vencimento e situação,
lendo todas as páginas. Filtros por ano, mês (1 a 12) e situação (ex.: PAGO, PENDENTE).
Na tabela, o resumo (em dia / guias vencidas) vai para stderr.

Para baixar o PDF de uma guia do histórico, use "ctbz impostos baixar ID".

## Uso

```
ctbz impostos historico [flags]
```

## Exemplos

```sh
  ctbz impostos historico --ano 2026
  ctbz impostos historico --ano 2026 --mes 7 -o csv
```

## Flags

```
      --ano int         ano da competência
      --mes int         mês da competência (1 a 12)
      --status string   situação da guia (ex.: PAGO, PENDENTE)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
