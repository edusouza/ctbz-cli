# ctbz documentos enviar

Envia o documento pedido por uma ou mais pendências

Envia o arquivo que uma pendência pede (risco médio: envia um documento contábil e resolve
a pendência; não há exclusão pela API). O tipo do documento e os metadados (competência, conta,
investimento) vêm da pendência, nunca são digitados.

Com várias --pendencia do mesmo tipo, o arquivo é enviado uma vez para todas ("em um arquivo
único" no painel). Extratos de movimentação usam ctbz extratos importar.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz documentos enviar ARQUIVO [flags]
```

## Exemplos

```sh
  ctbz documentos enviar aplicacao-agosto.pdf --pendencia 1000000000000601
  ctbz documentos enviar aplicacoes.pdf --pendencia 1000000000000601 --pendencia 1000000000000602
```

## Flags

```
      --dry-run             mostra a requisição que seria enviada, sem enviar
      --pendencia strings   id da pendência (repita para enviar um arquivo único para várias)
  -y, --yes                 envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz documentos`](ctbz_documentos.md).
