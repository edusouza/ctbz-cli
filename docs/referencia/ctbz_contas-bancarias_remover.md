# ctbz contas-bancarias remover

Exclui permanentemente uma conta bancária

Exclui uma conta bancária cadastrada por engano ou encerrada (risco alto: "Excluir
permanentemente" no painel; os vínculos da conta se perdem e um novo cadastro não os
recupera). O painel pode bloquear a exclusão; nesse caso a CLI mostra o motivo.

Mostra banco, agência e conta e pede "confirmo" no terminal (--yes em scripts); aceita --dry-run.

## Uso

```
ctbz contas-bancarias remover ID [flags]
```

## Exemplos

```sh
  ctbz contas-bancarias remover 1000000000000001
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

Veja também: [`ctbz contas-bancarias`](ctbz_contas-bancarias.md).
