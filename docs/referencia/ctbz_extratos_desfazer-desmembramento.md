# ctbz extratos desfazer-desmembramento

Volta um lançamento desmembrado do extrato ao original

Apaga as partes de um lançamento desmembrado e restaura o lançamento original (risco
médio; dá para desmembrar de novo). ID pode ser o do lançamento original ou o de qualquer uma
das partes (coluna "Parte de" em ctbz extratos lancamentos).

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz extratos desfazer-desmembramento ID [flags]
```

## Exemplos

```sh
  ctbz extratos desfazer-desmembramento 1000000000000102 --conta-bancaria 1000000000000001 --competencia 2026-09
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
      --dry-run              mostra a requisição que seria enviada, sem enviar
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz extratos`](ctbz_extratos.md).
