# ctbz usuarios desativar

Revoga o acesso de um usuário à empresa

Revoga o acesso de um usuário à empresa. O ID está em ctbz usuarios. Desativar é a única forma de revogar o
acesso, e o administrador da empresa (usuário principal) não pode ser desativado.

Risco médio: reversível com ativar ou desativar. Aceita --yes e --dry-run.

## Uso

```
ctbz usuarios desativar ID [flags]
```

## Exemplos

```sh
  ctbz usuarios desativar 102
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

Veja também: [`ctbz usuarios`](ctbz_usuarios.md).
