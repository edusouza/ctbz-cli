# ctbz empresa usar

Troca a empresa da sessão

Troca a empresa da sessão refazendo o login com --cnpj: a Contabilizei escolhe a
empresa só no login, então um novo código OTP é enviado por e-mail. Com CTBZ_OTP_CMD
definido, a troca é automática; sem ele, o código é pedido no terminal ou a troca fica
pendente para "ctbz login --otp NNNNNN".

## Uso

```
ctbz empresa usar CNPJ
```

## Exemplos

```sh
  ctbz empresa usar 00.000.000/0001-00
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz empresa`](ctbz_empresa.md).
