# ctbz impostos parcelamento contratar

Contrata um parcelamento de débitos

Contrata um parcelamento de débitos (risco alto). A CLI roda a simulação antes e mostra a
opção escolhida e os custos adicionais cobrados na mensalidade. Contratar é uma confissão de dívida perante a Receita ou a PGFN. Se a primeira parcela não for paga, o parcelamento é cancelado automaticamente. As parcelas são corrigidas todo mês pela Selic + 1% de juros. Não dá para desfazer pela CLI nem pelo painel.

TIPO: pgfn-nao-previdenciario, pgfn-previdenciario, pgfn-simples-nacional, simples-nacional, especializado.
- pgfn-*: --parcelas escolhe a quantidade entre as opções da simulação (padrão: a primeira,
  como no painel).
- simples-nacional: a quantidade não é escolhida aqui (o pedido vai sem corpo).
- especializado: exige --negociacao; vira um pedido de atendimento com idTicket.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz impostos parcelamento contratar TIPO [flags]
```

## Exemplos

```sh
  ctbz impostos parcelamento contratar pgfn-previdenciario --parcelas 24
  ctbz impostos parcelamento contratar especializado --negociacao VENCIDOS --dry-run
```

## Flags

```
      --dry-run             mostra a requisição que seria enviada, sem enviar
      --negociacao string   no tipo especializado: VENCIDOS, DIVIDA_ATIVA, DIVIDA_ATIVA_E_VENCIDOS
      --parcelas int        quantidade de parcelas (tipos pgfn-*; padrão: a primeira opção da simulação)
  -y, --yes                 envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz impostos parcelamento`](ctbz_impostos_parcelamento.md).
