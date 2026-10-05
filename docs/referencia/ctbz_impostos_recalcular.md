# ctbz impostos recalcular

Mostra ou pede o recálculo de uma guia vencida

Sem --vencimento, mostra os dados do recálculo de uma guia vencida: valor original,
valor recalculado estimado, data recomendada, datas indisponíveis e se o serviço será cobrado.
Nada é enviado.

Com --vencimento AAAA-MM-DD, pede uma nova guia com juros e multa para essa data (risco alto:
o recálculo é um serviço adicional, em geral cobrado na próxima mensalidade, e não dá para
cancelar o pedido). A guia precisa estar vencida e marcada como não paga
(ctbz impostos confirmar ID --nao-paguei); datas indisponíveis e passadas são recusadas, e
pendências críticas na Central de Rotinas bloqueiam o pedido, como no painel.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz impostos recalcular ID [flags]
```

## Exemplos

```sh
  ctbz impostos recalcular 1000000000000001
  ctbz impostos recalcular 1000000000000001 --vencimento 2026-10-20
```

## Flags

```
      --dry-run             mostra a requisição que seria enviada, sem enviar
      --vencimento string   novo vencimento AAAA-MM-DD (sem ele, só mostra os dados do recálculo)
  -y, --yes                 envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
