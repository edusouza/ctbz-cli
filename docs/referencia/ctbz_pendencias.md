# ctbz pendencias

Lista as pendências da empresa (tipo, detalhe, prazo e situação)

Lista as pendências cadastradas para a empresa, como o card de pendências da tela
inicial do painel: tipo, detalhe, data de criação, prazo e situação.

Por padrão mostra só as abertas; --todas inclui as finalizadas. A coluna alerta marca
prazos vencidos ("vencida") e que vencem em até 7 dias ("próxima"). Com --fail-on-vencidas,
o comando termina com código 4 quando há pendência vencida.

Na tabela, o detalhe é cortado; use -o json para o texto completo.

## Uso

```
ctbz pendencias [flags]
```

## Exemplos

```sh
  ctbz pendencias
  ctbz pendencias --todas -o json
  ctbz pendencias --fail-on-vencidas || notify-send "Há pendências vencidas"
```

## Subcomandos

- [`ctbz pendencias aceitar`](ctbz_pendencias_aceitar.md): Aceita um termo ou carta pendente da Central de Rotinas
- [`ctbz pendencias conciliacao`](ctbz_pendencias_conciliacao.md): Mostra as pendências de conciliação fiscal (notas e recebimentos)
- [`ctbz pendencias termo`](ctbz_pendencias_termo.md): Mostra o texto completo de um termo ou carta da Central de Rotinas
- [`ctbz pendencias termos`](ctbz_pendencias_termos.md): Lista os termos e cartas com aceite pendente na Central de Rotinas

## Flags

```
      --fail-on-vencidas   termina com código 4 se houver pendência vencida
      --todas              inclui as pendências finalizadas
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
