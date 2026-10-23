# ctbz empresa certificado

Mostra a situação do certificado digital da empresa e da renovação

Mostra a situação do certificado digital (e-CNPJ) usado pela Contabilizei, a data de
vencimento, os dias que faltam, se já é possível renovar, o alerta de vencimento do painel e
a etapa do processo de compra ou renovação feito pela Contabilizei (ex.: PRE_CHECKOUT antes
de começar, agendamentoVideoconferencia durante a validação).

A senha e o arquivo do certificado nunca são consultados.

## Uso

```
ctbz empresa certificado
```

## Exemplos

```sh
  ctbz empresa certificado
  ctbz empresa certificado -o json | jq .dias_para_vencer
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz empresa`](ctbz_empresa.md).
