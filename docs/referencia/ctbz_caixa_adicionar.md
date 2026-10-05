# ctbz caixa adicionar

Adiciona um recebimento ou pagamento no caixa de uma competência

Registra um recebimento ou pagamento em dinheiro no caixa da competência (risco médio:
entra na contabilidade do mês; dá para editar ou remover depois).

O valor é sempre positivo: --recebimento ou --pagamento define o sinal. A classificação
(--conta) precisa estar entre as aceitas na competência (ctbz caixa contas). Contas
"Impostos - …" exigem --guia ou --sem-guia, e "Sócios - Distribuição de Lucros
Antecipados" exige --socio (ids em ctbz caixa contas --vinculos). Sem --data, usa o dia
que o painel sugere: hoje na competência atual, senão o dia 1.

Pede confirmação (--yes em scripts) e aceita --dry-run (ver docs/guia/escrita.md).

## Uso

```
ctbz caixa adicionar [flags]
```

## Exemplos

```sh
  ctbz caixa adicionar --competencia 2026-09 --data 2026-09-15 --pagamento \
    --valor 150,25 --conta 1000000000000012 --descricao "Material de escritório"
  ctbz caixa adicionar --competencia 2026-09 --pagamento --valor 80 \
    --conta "Impostos - Simples Nacional" --sem-guia --descricao "DAS" --dry-run
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --conta string         classificação: id ou descrição exata (ver ctbz caixa contas)
      --data string          data do lançamento AAAA-MM-DD (não pode ser futura)
      --descricao string     descrição do lançamento
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --guia string          guia de imposto vinculada (contas "Impostos - …"; ver --vinculos)
      --pagamento            saída de dinheiro do caixa
      --recebimento          entrada de dinheiro no caixa
      --sem-guia             classificação de imposto sem guia vinculada
      --socio string         sócio vinculado (Distribuição de Lucros Antecipados; ver --vinculos)
      --valor string         valor em reais, sempre positivo (ex.: 150,25)
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz caixa`](ctbz_caixa.md).
