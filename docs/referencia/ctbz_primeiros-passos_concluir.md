# ctbz primeiros-passos concluir

Marca uma tarefa dos primeiros passos como concluída

Marca como concluída uma tarefa do checklist que o painel deixa concluir à mão
(LIVE_EMISSAO_NOTAS_FISCAIS, LIVE_VIDA_COM_CNPJ, REUNIAO_BOAS_VINDAS). Risco baixo, mas não há como desconcluir.
Se a tarefa já estava concluída, a CLI avisa e termina com código 0.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz primeiros-passos concluir ETAPA [flags]
```

## Exemplos

```sh
  ctbz primeiros-passos concluir LIVE_VIDA_COM_CNPJ
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

Veja também: [`ctbz primeiros-passos`](ctbz_primeiros-passos.md).
