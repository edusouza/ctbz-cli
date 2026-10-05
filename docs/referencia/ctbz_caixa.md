# ctbz caixa

Lista os lançamentos do caixa de um mês

Lista os lançamentos do caixa do mês (entradas e saídas classificadas): data, descrição,
conta de classificação, tipo (receita, despesa…), valor (negativo nas saídas), situação, se
foi lançado pelo sistema e o id (usado por ctbz caixa editar e remover). Na tabela, o total do mês vai para o stderr.

O painel pede até 1000 lançamentos por mês; se houver mais, a CLI avisa no stderr.

## Uso

```
ctbz caixa AAAA-MM
```

## Exemplos

```sh
  ctbz caixa 2026-09
  ctbz caixa 2026-09 -o csv > caixa-2026-09.csv
```

## Subcomandos

- [`ctbz caixa adicionar`](ctbz_caixa_adicionar.md): Adiciona um recebimento ou pagamento no caixa de uma competência
- [`ctbz caixa contas`](ctbz_caixa_contas.md): Lista as classificações aceitas nos lançamentos do caixa de uma competência
- [`ctbz caixa editar`](ctbz_caixa_editar.md): Altera um lançamento manual do caixa
- [`ctbz caixa remover`](ctbz_caixa_remover.md): Exclui permanentemente um lançamento manual do caixa

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
