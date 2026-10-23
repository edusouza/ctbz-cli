# ctbz impostos baixar

Baixa o PDF de guias de imposto

Baixa o PDF das guias informadas (IDs de "ctbz impostos") ou, com --pendentes, de todas
as guias a pagar (em atraso, do mês e do próximo mês).

Os arquivos são nomeados COMPETÊNCIA-IMPOSTO-VENCIMENTO.pdf (ex.:
2026-07-darf-unificado-2026-10-06.pdf). Arquivos existentes não são sobrescritos sem --force.
A lista de arquivos gravados sai em stdout (no formato de -o).

## Uso

```
ctbz impostos baixar [ID...] [flags]
```

## Exemplos

```sh
  ctbz impostos baixar 1000000000000001
  ctbz impostos baixar --pendentes -d ~/guias
```

## Flags

```
  -d, --dir string   diretório onde salvar os PDFs (default ".")
      --force        sobrescreve arquivos existentes
      --pendentes    baixa todas as guias a pagar
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
