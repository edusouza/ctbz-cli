# ctbz notas entrada manifestar

Envia a manifestação do destinatário das notas de entrada à SEFAZ

Manifesta as NF-e de compra como destinatário: --ciencia (ciência da operação),
--confirmar ("recebi"), --desconhecer ("desconheço") ou --nao-realizada ("não recebi", exige
--justificativa de 15 a 255 caracteres, regra da SEFAZ).

Só notas PENDENTE ou CIENCIA são manifestadas, como no painel; as demais são listadas como
ignoradas. --mes diz em que mês procurar as notas (padrão: o mês atual).

Risco alto nos eventos conclusivos (confirmar, desconhecer, não realizada): são eventos
fiscais enviados à SEFAZ, sem desfazer, e a CLI pede "confirmo". A ciência é de risco médio.
No painel, baixar o XML ou o DANFE de uma nota pendente manifesta ciência automaticamente.

Aceita --yes e --dry-run.

## Uso

```
ctbz notas entrada manifestar ID... [flags]
```

## Exemplos

```sh
  ctbz notas entrada manifestar 1000000000000701 --confirmar --mes 2026-09
  ctbz notas entrada manifestar 1000000000000702 --nao-realizada --justificativa "Mercadoria extraviada antes da entrega"
```

## Flags

```
      --ciencia                ciência da operação (preliminar)
      --confirmar              confirmação da operação (recebi)
      --desconhecer            desconhecimento da operação
      --dry-run                mostra a requisição que seria enviada, sem enviar
      --justificativa string   motivo da operação não realizada (15 a 255 caracteres)
      --mes string             mês das notas, AAAA-MM (padrão: o mês atual)
      --nao-realizada          operação não realizada (não recebi); exige --justificativa
  -y, --yes                    envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz notas entrada`](ctbz_notas_entrada.md).
