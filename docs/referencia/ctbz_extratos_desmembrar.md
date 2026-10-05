# ctbz extratos desmembrar

Divide um lançamento do extrato em partes com classificações diferentes

Divide um lançamento do extrato em duas ou mais partes, cada uma com descrição, valor e
classificação próprios (risco médio, reversível com ctbz extratos desfazer-desmembramento).

Cada --parte é "descrição:valor:conta[:sócio]": o valor é positivo (o sinal é o do lançamento
original), a conta é o id ou a descrição exata de uma classificação aceita (ctbz extratos
contas) e o sócio é obrigatório nas classificações de sócio.

Regras do painel, conferidas antes de enviar: pelo menos 2 partes, nenhuma com valor zero, a
soma igual ao valor do lançamento (em centavos), e o lançamento não pode ser parte de outro
desmembramento nem ter vínculo.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz extratos desmembrar ID [flags]
```

## Exemplos

```sh
  ctbz extratos desmembrar 1000000000000102 --conta-bancaria 1000000000000001 --competencia 2026-09 \
    --parte "Aluguel:1.000,00:Pagamento de Fornecedores" --parte "Condomínio:234,56:1000000000000012"
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --parte stringArray    parte "descrição:valor:conta[:sócio]" (repita para cada parte)
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz extratos`](ctbz_extratos.md).
