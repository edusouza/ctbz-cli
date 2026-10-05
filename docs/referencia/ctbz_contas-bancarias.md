# ctbz contas-bancarias

Lista as contas bancárias cadastradas da empresa

Lista as contas bancárias cadastradas na Contabilizei: banco, código do banco, agência,
conta, saldo inicial e sua data, e a situação da integração (ex.: INTEGRADA com a conta PJ
da Contabilizei).

## Uso

```
ctbz contas-bancarias
```

## Exemplos

```sh
  ctbz contas-bancarias
```

## Subcomandos

- [`ctbz contas-bancarias adicionar`](ctbz_contas-bancarias_adicionar.md): Cadastra uma conta bancária PJ
- [`ctbz contas-bancarias bancos`](ctbz_contas-bancarias_bancos.md): Lista os bancos aceitos no cadastro de contas bancárias
- [`ctbz contas-bancarias editar`](ctbz_contas-bancarias_editar.md): Corrige os dados de uma conta bancária cadastrada
- [`ctbz contas-bancarias remover`](ctbz_contas-bancarias_remover.md): Exclui permanentemente uma conta bancária

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
