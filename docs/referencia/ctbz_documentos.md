# ctbz documentos

Lista os tipos de documento aceitos e os documentos enviados

Sem --tipo, lista os tipos de documento da central de documentos (extratos, contratos de
empréstimo, informes de investimentos…) com quantos já foram enviados e a última
modificação. Com --tipo, lista os documentos enviados daquele tipo: arquivo, competência,
data de envio, descrição e o link do arquivo.

Enviar documentos continua sendo feito pelo painel.

## Uso

```
ctbz documentos [flags]
```

## Exemplos

```sh
  ctbz documentos
  ctbz documentos --tipo EXTRATO_BANCARIO_MOVIMENTACOES -o csv
```

## Subcomandos

- [`ctbz documentos enviar`](ctbz_documentos_enviar.md): Envia o documento pedido por uma ou mais pendências
- [`ctbz documentos pendentes`](ctbz_documentos_pendentes.md): Lista os documentos pedidos pelas pendências da Central de Rotinas
- [`ctbz documentos sem-aplicacao`](ctbz_documentos_sem-aplicacao.md): Declara que não houve aplicação financeira numa conta nas competências pendentes
- [`ctbz documentos sem-arquivo`](ctbz_documentos_sem-arquivo.md): Declara que o documento pedido por pendências não existe

## Flags

```
      --tipo string   tipo de documento (veja a coluna tipo de "ctbz documentos")
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
