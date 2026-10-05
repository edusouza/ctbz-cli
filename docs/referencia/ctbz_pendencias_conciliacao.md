# ctbz pendencias conciliacao

Mostra as pendências de conciliação fiscal (notas e recebimentos)

Mostra quantas notas fiscais estão sem recebimento e quantos recebimentos do extrato
estão sem nota, como a tela de conciliação fiscal do painel, e a competência de referência.

Com --listar notas ou --listar recebimentos, lista os itens pendentes dos últimos 12 meses
(o mesmo período do painel). O formato desses itens não pôde ser verificado (a conta usada
no desenvolvimento não tinha pendências), por isso os campos saem como a API os devolve.

Com --fail-on-pendencias, o comando termina com código 4 quando há algo a conciliar.

## Uso

```
ctbz pendencias conciliacao [flags]
```

## Exemplos

```sh
  ctbz pendencias conciliacao
  ctbz pendencias conciliacao --listar notas -o json
  ctbz pendencias conciliacao --fail-on-pendencias
```

## Subcomandos

- [`ctbz pendencias conciliacao candidatos`](ctbz_pendencias_conciliacao_candidatos.md): Lista as notas ou os recebimentos que podem ser vinculados numa conciliação
- [`ctbz pendencias conciliacao detalhes`](ctbz_pendencias_conciliacao_detalhes.md): Mostra os detalhes de uma pendência ou conciliação (justificativa, vínculos)
- [`ctbz pendencias conciliacao motivos`](ctbz_pendencias_conciliacao_motivos.md): Lista os motivos aceitos para justificar uma pendência de conciliação
- [`ctbz pendencias conciliacao resolver`](ctbz_pendencias_conciliacao_resolver.md): Concilia ou justifica pendências de conciliação fiscal

## Flags

```
      --fail-on-pendencias   termina com código 4 se houver algo a conciliar
      --listar string        lista os itens pendentes: notas ou recebimentos
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz pendencias`](ctbz_pendencias.md).
