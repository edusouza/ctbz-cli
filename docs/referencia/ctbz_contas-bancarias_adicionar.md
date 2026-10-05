# ctbz contas-bancarias adicionar

Cadastra uma conta bancária PJ

Cadastra uma conta bancária PJ da empresa, para poder importar os extratos dela (risco
médio: dá para editar depois). A data de abertura define desde quando a Contabilizei vai pedir os extratos desta conta.

--declaro-conta-pj é obrigatório, como o checkbox do painel: "a conta é uma Conta PJ e suas
movimentações refletem exclusivamente este CNPJ".

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz contas-bancarias adicionar [flags]
```

## Exemplos

```sh
  ctbz contas-bancarias adicionar --banco 341 --agencia 1234 --conta 12345-6 \
    --abertura 2024-01-10 --declaro-conta-pj
```

## Flags

```
      --abertura string        data de abertura da conta AAAA-MM-DD (define desde quando os extratos são pedidos)
      --agencia string         agência, sem dígito
      --banco string           banco: id ou código (ver ctbz contas-bancarias bancos)
      --conta string           conta com dígito (ex.: 12345-6)
      --declaro-conta-pj       declara que é uma conta PJ que movimenta só este CNPJ (obrigatório)
      --dry-run                mostra a requisição que seria enviada, sem enviar
      --saldo-inicial string   saldo na abertura, em reais (padrão: 0)
  -y, --yes                    envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz contas-bancarias`](ctbz_contas-bancarias.md).
