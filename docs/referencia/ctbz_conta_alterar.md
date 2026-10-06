# ctbz conta alterar

Troca o e-mail e o telefone de login (com código OTP)

Troca o e-mail e o telefone de login. A API recebe os dois juntos e o painel os mostra
mascarados, então informe os dois (repita o atual no que não muda). Telefone no formato
(DD)NNNNN-NNNN ou só os 11 dígitos.

Depois da confirmação, a Contabilizei envia um código (--via email|sms; padrão: o método do
2FA). O código é lido do --otp-cmd (ou CTBZ_OTP_CMD), como no login, ou digitado no terminal.

Risco alto: os próximos códigos de login vão para o novo e-mail (atualize CTBZ_USER) e um
erro pode tirar seu acesso. Dá para alterar de novo. Aceita --yes e --dry-run.

## Uso

```
ctbz conta alterar [flags]
```

## Exemplos

```sh
  ctbz conta alterar --email novo@example.com --telefone "(11)98765-4321"
  CTBZ_OTP_CMD=./scripts/otp-gmail-gws.sh ctbz conta alterar --email novo@example.com --telefone 11987654321 --yes
```

## Flags

```
      --dry-run                mostra a requisição que seria enviada, sem enviar
      --email string           novo e-mail de login
      --otp-cmd string         comando de shell que imprime o código (env CTBZ_OTP_CMD)
      --otp-timeout duration   tempo máximo aguardando o --otp-cmd (env CTBZ_OTP_TIMEOUT) (default 3m0s)
      --telefone string        novo telefone, (DD)NNNNN-NNNN
      --via string             por onde receber o código: email ou sms (padrão: o método do 2FA)
  -y, --yes                    envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz conta`](ctbz_conta.md).
