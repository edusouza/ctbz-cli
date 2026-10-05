# ctbz pendencias aceitar

Aceita um termo ou carta pendente da Central de Rotinas

Assina um aceite cobrado pela Central de Rotinas (risco alto: são declarações legais e
contábeis, sem como revogar). O texto completo do termo é sempre impresso no stderr antes da
confirmação, que no terminal exige digitar "confirmo" (--yes em scripts).

Se não houver aceite pendente, a CLI avisa "nada a aceitar" e termina com código 0.
Depois do envio, relê a Central de Rotinas para confirmar que a pendência sumiu.

CHAVE: carta-responsabilidade, termo-debitos, termo-totalpass.

## Uso

```
ctbz pendencias aceitar CHAVE [flags]
```

## Exemplos

```sh
  ctbz pendencias aceitar carta-responsabilidade
  ctbz pendencias aceitar termo-debitos --dry-run
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

Veja também: [`ctbz pendencias`](ctbz_pendencias.md).
