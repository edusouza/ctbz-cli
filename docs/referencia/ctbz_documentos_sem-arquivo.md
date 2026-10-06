# ctbz documentos sem-arquivo

Declara que o documento pedido por pendências não existe

Fecha pendências declarando que o documento não existe: "não tive estoque", "não tive
intermediações" ou "já enviei o contrato de AFAC" (risco alto: é uma declaração, sem
desfazer pela API). Só vale para os tipos ESTOQUE, CONTROLE_DE_INTERMEDIACOES e
CONTRATO_DE_AFAC; o tipo é lido da pendência.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz documentos sem-arquivo [flags]
```

## Exemplos

```sh
  ctbz documentos sem-arquivo --pendencia 1000000000000603
```

## Flags

```
      --dry-run             mostra a requisição que seria enviada, sem enviar
      --pendencia strings   id da pendência (repita para várias do mesmo tipo)
  -y, --yes                 envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz documentos`](ctbz_documentos.md).
