# ctbz conta senha

Troca a senha de login (com código OTP)

Troca a senha de login. A nova senha é pedida duas vezes, sem eco, e nunca por argumento;
fora de um terminal, é lida da entrada padrão (duas linhas). Regras do painel: pelo menos 8
caracteres, com minúsculas, maiúsculas, números e um caractere especial.

A senha atual não é pedida: a prova é o código que a Contabilizei envia depois da confirmação
(--via email|sms; padrão: o método do 2FA), lido do --otp-cmd (ou CTBZ_OTP_CMD) ou digitado.

Risco alto: atualize CTBZ_PASSWORD; outras sessões podem ser encerradas. Aceita --yes e
--dry-run (que mostra *** no lugar da senha).

## Uso

```
ctbz conta senha [flags]
```

## Exemplos

```sh
  ctbz conta senha
  ctbz conta senha --via sms
```

## Flags

```
      --dry-run                mostra a requisição que seria enviada, sem enviar
      --otp-cmd string         comando de shell que imprime o código (env CTBZ_OTP_CMD)
      --otp-timeout duration   tempo máximo aguardando o --otp-cmd (env CTBZ_OTP_TIMEOUT) (default 3m0s)
      --via string             por onde receber o código: email ou sms (padrão: o método do 2FA)
  -y, --yes                    envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz conta`](ctbz_conta.md).
