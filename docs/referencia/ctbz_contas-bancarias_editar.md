# ctbz contas-bancarias editar

Corrige os dados de uma conta bancária cadastrada

Corrige banco, agência, conta, data de abertura ou saldo inicial de uma conta (risco médio:
dá para editar de novo). Só as flags informadas mudam. O painel pode bloquear a edição (por
exemplo, com extratos já classificados); nesse caso a CLI mostra o motivo.
A data de abertura define desde quando a Contabilizei vai pedir os extratos desta conta.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz contas-bancarias editar ID [flags]
```

## Exemplos

```sh
  ctbz contas-bancarias editar 1000000000000001 --abertura 2024-02-01
```

## Flags

```
      --abertura string        data de abertura da conta AAAA-MM-DD (define desde quando os extratos são pedidos)
      --agencia string         agência, sem dígito
      --banco string           banco: id ou código (ver ctbz contas-bancarias bancos)
      --conta string           conta com dígito (ex.: 12345-6)
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
