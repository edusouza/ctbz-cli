# ctbz primeiros-passos

Lista as tarefas do checklist de primeiros passos do painel

Lista as tarefas do checklist "Primeiros passos" do painel com a situação de cada uma e se
dá para concluí-la pela CLI (ctbz primeiros-passos concluir ETAPA). As demais são concluídas
pelo sistema quando a ação é feita (abrir a conta PJ, cadastrar o pró-labore…).

## Uso

```
ctbz primeiros-passos
```

## Exemplos

```sh
  ctbz primeiros-passos
  ctbz primeiros-passos concluir REUNIAO_BOAS_VINDAS
```

## Subcomandos

- [`ctbz primeiros-passos concluir`](ctbz_primeiros-passos_concluir.md): Marca uma tarefa dos primeiros passos como concluída
- [`ctbz primeiros-passos dispensar`](ctbz_primeiros-passos_dispensar.md): Oculta o card de primeiros passos do painel
- [`ctbz primeiros-passos reativar`](ctbz_primeiros-passos_reativar.md): Volta a mostrar as tarefas dos primeiros passos

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
