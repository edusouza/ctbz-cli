# ctbz usuarios convidar

Convida uma pessoa para acessar a empresa, com acesso total

Envia um convite por e-mail para uma pessoa acessar a empresa no painel. O acesso é
total: todas as ações e todas as informações, inclusive financeiras.

A CLI recusa e-mails da Contabilizei, e-mails que já têm acesso ou convite, e convites quando o
serviço está indisponível ou o limite de usuários secundários ativos foi atingido.

Risco alto: não há como revogar o convite antes do aceite; depois, só com ctbz usuarios
desativar. Aceita --yes e --dry-run.

## Uso

```
ctbz usuarios convidar EMAIL [flags]
```

## Exemplos

```sh
  ctbz usuarios convidar pessoa@example.com
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
