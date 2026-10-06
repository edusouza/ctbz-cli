# ctbz documentos sem-aplicacao

Declara que não houve aplicação financeira numa conta nas competências pendentes

Resolve pendências de extrato de aplicação financeira declarando que a conta não teve
aplicação no período ("Não tenho aplicação nesta conta" no painel; risco alto: é uma
declaração, sem desfazer pela API). As pendências precisam ser de extrato de aplicação da
conta informada.

Se a Contabilizei aceitar só parte das competências, a CLI lista as que falharam e termina
com código 1.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz documentos sem-aplicacao [flags]
```

## Exemplos

```sh
  ctbz documentos sem-aplicacao --conta-bancaria 7 --pendencia 1000000000000601 --pendencia 1000000000000602
```

## Flags

```
      --conta-bancaria int   id da conta bancária das pendências
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --pendencia strings    id da pendência de extrato de aplicação (repita para várias)
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz documentos`](ctbz_documentos.md).
