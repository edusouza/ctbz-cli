# ctbz chamados

Lista os chamados de atendimento (em andamento ou finalizados)

Lista os chamados abertos com o atendimento da Contabilizei, com assunto, status, canal,
datas e o link para a central de ajuda.

--finalizados lista os já resolvidos. Em contas com muitos chamados, a Contabilizei não
consegue montar essa lista (erro do servidor); nesse caso a CLI mostra os finalizados
entre os 100 chamados mais recentes da empresa e avisa no stderr.

## Uso

```
ctbz chamados [flags]
```

## Exemplos

```sh
  ctbz chamados
  ctbz chamados --finalizados -o csv
```

## Flags

```
      --finalizados   lista os chamados finalizados
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
