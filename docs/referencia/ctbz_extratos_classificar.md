# ctbz extratos classificar

Troca a classificação de um lançamento do extrato

Troca a classificação de um lançamento do extrato importado (risco médio, reversível
classificando de novo enquanto o período estiver aberto).

--conta aceita o id ou a descrição exata de uma classificação aceita na competência: RECEITA
para entradas e DESPESA para saídas (ctbz extratos contas). As classificações de sócio exigem
--socio. Se os contadores já fecharam o período, o painel não deixa alterar, e a CLI também
não.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz extratos classificar ID [flags]
```

## Exemplos

```sh
  ctbz extratos classificar 1000000000000102 --conta-bancaria 1000000000000001 \
    --competencia 2026-09 --conta "Pagamento de Fornecedores"
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --conta string         classificação: id ou descrição exata (ver ctbz extratos contas)
      --conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --socio string         sócio, nas classificações de sócio
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz extratos`](ctbz_extratos.md).
