# ctbz pendencias conciliacao resolver

Concilia ou justifica pendências de conciliação fiscal

Resolve pendências de conciliação fiscal (risco alto: a resolução define a natureza fiscal
do recebimento, como receita tributável, empréstimo ou capital; pode ser refeita, mas não
apagada). Três formas, como no painel:

  --vincular ORIGEM:ID…      concilia com notas (NOTAFISCAL, para pendências de recebimento)
                             ou movimentações (MOVIMENTACAO, para pendências de nota)
  --motivo MOTIVO            justifica sem contraparte (ctbz pendencias conciliacao motivos)
  os dois                    concilia com diferença justificada

Motivos que envolvem sócio exigem --socio ID. --nota-ativacao informa o número da nota no
fluxo de ativação contábil. Os ids das pendências vêm de ctbz pendencias conciliacao --listar.

No painel, algumas opções só valem para itens abaixo de R$ 1.000,00; a CLI não conhece a lista
exata, e a Contabilizei recusa o pedido se a regra não for atendida.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz pendencias conciliacao resolver ID_PENDENCIA... [flags]
```

## Exemplos

```sh
  ctbz pendencias conciliacao resolver 123 --vincular NOTAFISCAL:987
  ctbz pendencias conciliacao resolver 123 124 --motivo CAIXA_OU_PESSOA_FISICA
  ctbz pendencias conciliacao resolver 123 --motivo EMPRESTIMO_DO_SOCIO_A_EMPRESA --socio 55 --dry-run
```

## Flags

```
      --dry-run                mostra a requisição que seria enviada, sem enviar
      --motivo string          motivo da resolução (ver ctbz pendencias conciliacao motivos)
      --nota-ativacao string   número da nota no fluxo de ativação contábil
      --socio string           sócio, nos motivos que exigem
      --vincular stringArray   contraparte ORIGEM:ID (NOTAFISCAL ou MOVIMENTACAO; repita para várias)
  -y, --yes                    envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz pendencias conciliacao`](ctbz_pendencias_conciliacao.md).
