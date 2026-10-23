# ctbz resumo

Mostra numa lista só o que precisa de atenção

Junta numa lista só o que precisa de atenção: impostos em atraso e do mês, pendências
abertas, rotinas do mês ainda não realizadas, a mensalidade do mês e as pendências que o
painel marca como críticas. As consultas são feitas em paralelo.

A coluna secao indica a origem (impostos, pendencias, rotinas, mensalidade, painel) e a
coluna alerta segue a regra dos outros comandos: "vencida", "próxima" (até 7 dias) ou
"crítica" (pendência crítica do painel). Com --fail-on-atencao, o comando termina com
código 4 quando há algo vencido ou crítico, para alertas em cron.

## Uso

```
ctbz resumo [flags]
```

## Exemplos

```sh
  ctbz resumo
  ctbz resumo -o json | jq '.[] | select(.alerta == "vencida")'
  ctbz resumo --fail-on-atencao -o csv > /dev/null || notify-send "Contabilizei precisa de atenção"
```

## Flags

```
      --fail-on-atencao   termina com código 4 se houver algo vencido ou crítico
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
