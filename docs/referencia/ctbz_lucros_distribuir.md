# ctbz lucros distribuir

Registra quanto do lucro do exercício cabe a cada sócio

Registra a distribuição de lucros do exercício aberto, que vai para o informe de
rendimentos dos sócios. Cada --socio ID=parte diz a parte de um sócio, em percentual (60%) ou
em reais (30000,00); os sócios não citados ficam com zero. A soma precisa ser o lucro total
(saldo na empresa + o já distribuído), como no painel.

Os IDs dos sócios são o id de cada sócio em ctbz lucros. A distribuição só é aceita
enquanto o painel deixa alterar (até a data limite) e sem restrições no informe (pendência
documental ou débitos federais, também em ctbz lucros).

Risco alto: define os rendimentos isentos que os sócios declaram no IRPF. Dá para refazer
enquanto a distribuição puder ser alterada. Aceita --yes e --dry-run.

## Uso

```
ctbz lucros distribuir [flags]
```

## Exemplos

```sh
  ctbz lucros distribuir --socio 1000000000000001=60% --socio 1000000000000002=40%
  ctbz lucros distribuir --socio 1000000000000001=30000,00 --socio 1000000000000002=20000,00
```

## Flags

```
      --dry-run             mostra a requisição que seria enviada, sem enviar
      --socio stringArray   parte de um sócio: ID=60% ou ID=30000,00 (repetível)
  -y, --yes                 envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz lucros`](ctbz_lucros.md).
