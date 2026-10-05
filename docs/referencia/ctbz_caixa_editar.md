# ctbz caixa editar

Altera um lançamento manual do caixa

Altera data, valor, tipo (recebimento ou pagamento), classificação, vínculo ou descrição de
um lançamento manual do caixa (risco médio, reversível editando de novo). Só as flags
informadas mudam; o resto é mantido.

O painel não tem um "editar" separado: reenvia o lançamento inteiro. A CLI lê o lançamento
atual na competência, aplica as mudanças com as mesmas validações de ctbz caixa adicionar e
mostra o antes e o depois no resumo. Lançamentos feitos pelo sistema não podem ser editados.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz caixa editar ID [flags]
```

## Exemplos

```sh
  ctbz caixa editar 1000000000000002 --competencia 2026-09 --valor 900
  ctbz caixa editar 1000000000000002 --competencia 2026-09 --descricao "Venda balcão" --dry-run
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
