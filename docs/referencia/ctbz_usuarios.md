# ctbz usuarios

Lista quem tem acesso à empresa no painel (multiusuário)

Lista os usuários e os convites da empresa: ID, nome, e-mail, tipo (principal ou
secundário) e situação (ativo, inativo ou convite enviado).

Para convidar: ctbz usuarios convidar; para revogar ou devolver o acesso: ctbz usuarios
desativar|ativar.

## Uso

```
ctbz usuarios
```

## Exemplos

```sh
  ctbz usuarios
  ctbz usuarios -o json
```

## Subcomandos

- [`ctbz usuarios ativar`](ctbz_usuarios_ativar.md): Devolve o acesso de um usuário desativado
- [`ctbz usuarios convidar`](ctbz_usuarios_convidar.md): Convida uma pessoa para acessar a empresa, com acesso total
- [`ctbz usuarios desativar`](ctbz_usuarios_desativar.md): Revoga o acesso de um usuário à empresa

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
