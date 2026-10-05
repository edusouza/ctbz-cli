# ctbz pendencias conciliacao detalhes

Mostra os detalhes de uma pendência ou conciliação (justificativa, vínculos)

Mostra os detalhes de pendências ou conciliações de recebimentos (--recebimento) e de notas
(--nota). A API usa POST, mas é só leitura: nada muda e nada é registrado em ctbz acoes.
Os campos saem como a API os devolve (formato não verificado).

## Uso

```
ctbz pendencias conciliacao detalhes [flags]
```

## Exemplos

```sh
  ctbz pendencias conciliacao detalhes --recebimento 123
```

## Flags

```
      --nota strings          id de nota fiscal (repita ou separe por vírgula)
      --recebimento strings   id de recebimento (repita ou separe por vírgula)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz pendencias conciliacao`](ctbz_pendencias_conciliacao.md).
