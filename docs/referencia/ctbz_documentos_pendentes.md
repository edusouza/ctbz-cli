# ctbz documentos pendentes

Lista os documentos pedidos pelas pendências da Central de Rotinas

Lista os documentos que as pendências pedem (extrato de aplicação, informe de
investimentos, contratos, estoque, intermediações…): id da pendência, tipo, competência e
conta. Sem --pendencia, usa as pendências das rotinas do painel (coluna pendencias de
ctbz rotinas).

Envie com ctbz documentos enviar ARQUIVO --pendencia ID, ou declare que o documento não
existe com ctbz documentos sem-arquivo.

## Uso

```
ctbz documentos pendentes [flags]
```

## Exemplos

```sh
  ctbz documentos pendentes
  ctbz documentos pendentes --pendencia 1000000000000601 -o json
```

## Flags

```
      --pendencia strings   id da pendência (repita ou separe por vírgula)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz documentos`](ctbz_documentos.md).
