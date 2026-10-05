# ctbz rotinas reclassificar

Confirma ou altera a classificação de lançamentos pedida pela Central de Rotinas

Resolve as rotinas que pedem para confirmar ou alterar a classificação de um lançamento
bancário (financiamento, investimento anjo e outros). Os ids das pendências aparecem na
coluna "pendencias" de ctbz rotinas.

Sem --classificacao, só lista as opções de cada pendência (nada é enviado). Com
--classificacao (id ou nome da opção), grava a classificação e conclui a rotina (risco médio:
a rotina não reabre, mas a classificação ainda pode ser trocada com ctbz extratos classificar
enquanto o período estiver aberto). Opções com sócio exigem --socio, exceto quando há um só
sócio, que é escolhido sozinho como no painel.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz rotinas reclassificar ID_PENDENCIA... [flags]
```

## Exemplos

```sh
  ctbz rotinas reclassificar 1000000000000501
  ctbz rotinas reclassificar 1000000000000501 --classificacao "Aporte de Capital"
```

## Flags

```
      --classificacao string   opção de classificação: id ou nome (sem ela, lista as opções)
      --dry-run                mostra a requisição que seria enviada, sem enviar
      --socio string           sócio, nas opções que exigem
  -y, --yes                    envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz rotinas`](ctbz_rotinas.md).
