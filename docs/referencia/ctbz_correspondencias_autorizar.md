# ctbz correspondencias autorizar

Autoriza ou não o recebimento de correspondências no endereço informado

Liga (on) ou desliga (off) "Autorizo o recebimento de correspondências e documentos
entregues no endereço informado". O envio de correspondências pode ser cobrado (a lista mostra
o valor do envio).

Risco médio e reversível. Aceita --yes e --dry-run.

## Uso

```
ctbz correspondencias autorizar on|off [flags]
```

## Exemplos

```sh
  ctbz correspondencias autorizar on
```

## Flags

```
      --dry-run   mostra a requisição que seria enviada, sem enviar
  -y, --yes       envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz correspondencias`](ctbz_correspondencias.md).
