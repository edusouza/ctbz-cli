# ctbz login

Autentica na Contabilizei (usuário, senha e código por e-mail)

Autentica na Contabilizei usando CTBZ_USER e CTBZ_PASSWORD.

O código enviado por e-mail pode vir de:
  1. --otp-cmd / CTBZ_OTP_CMD: comando executado repetidamente até imprimir 6 dígitos;
  2. o terminal, se interativo;
  3. uma segunda chamada: "ctbz login --otp 123456" (o login fica salvo como pendente
     e o processo termina com código 3).

Com mais de uma empresa, escolha com --cnpj (ou CTBZ_CNPJ).

## Uso

```
ctbz login [flags]
```

## Exemplos

```sh
  ctbz login --cnpj 00.000.000/0001-00
  CTBZ_OTP_CMD=./scripts/otp-gmail-gws.sh ctbz login
  ctbz login --otp 123456
```

## Flags

```
      --cnpj string            CNPJ da empresa a selecionar (env CTBZ_CNPJ)
      --otp string             código OTP recebido por e-mail (retoma um login pendente)
      --otp-cmd string         comando de shell que imprime o OTP (env CTBZ_OTP_CMD)
      --otp-timeout duration   tempo máximo aguardando o --otp-cmd (env CTBZ_OTP_TIMEOUT) (default 3m0s)
      --restart                descarta um login pendente e recomeça
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
