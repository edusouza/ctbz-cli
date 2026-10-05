# ctbz caixa contas

Lista as classificações aceitas nos lançamentos do caixa de uma competência

Lista as classificações (contas) que o caixa aceita na competência, com o mesmo filtro do
painel: só as exibidas no mês, do lado escolhido (recebimento ou pagamento). A coluna "Exige"
indica as classificações que pedem um vínculo: uma guia de imposto (contas "Impostos - …")
ou um sócio ("Sócios - Distribuição de Lucros Antecipados").

--vinculos lista as guias e os sócios aceitos como vínculo (use o id em --guia ou --socio
de ctbz caixa adicionar); "0" é SEM GUIA.

## Uso

```
ctbz caixa contas [flags]
```

## Exemplos

```sh
  ctbz caixa contas --competencia 2026-09
  ctbz caixa contas --competencia 2026-09 --pagamento
  ctbz caixa contas --competencia 2026-09 --vinculos
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --pagamento            só as classificações de pagamento
      --recebimento          só as classificações de recebimento
      --vinculos             lista as guias e os sócios aceitos como vínculo
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz caixa`](ctbz_caixa.md).
