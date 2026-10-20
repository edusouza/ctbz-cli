# ctbz conta

Mostra os dados de login do usuário (e-mail, telefone e método do 2FA)

Mostra o e-mail e o telefone de login, mascarados como no painel, e o método do segundo
fator (EMAIL, SMS ou APP).

Para trocar: ctbz conta alterar (e-mail e telefone) e ctbz conta senha.

## Uso

```
ctbz conta
```

## Exemplos

```sh
  ctbz conta
```

## Subcomandos

- [`ctbz conta alterar`](ctbz_conta_alterar.md): Troca o e-mail e o telefone de login (com código OTP)
- [`ctbz conta senha`](ctbz_conta_senha.md): Troca a senha de login (com código OTP)

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
