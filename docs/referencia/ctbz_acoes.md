# ctbz acoes

Lista as ações de escrita enviadas à Contabilizei por esta CLI

Lista as escritas que esta CLI enviou à Contabilizei, da mais antiga para a mais recente:
data e hora, CNPJ da empresa, comando, método, caminho, status HTTP, resultado, id do
objeto (quando conhecido) e, nos envios de arquivo, o nome e o tamanho de cada arquivo.

O registro fica em $CTBZ_HOME/acoes.jsonl (permissão 0600), uma linha JSON por escrita
enviada, inclusive as que falharam. Simulações (--dry-run) não entram. Corpos de
requisição e de resposta nunca são gravados, nem o conteúdo dos arquivos.

Resultado: "enviada" (HTTP 2xx), "recusada" (resposta de erro) ou "sem resposta" (falha de
conexão: a escrita pode ou não ter sido aplicada).

## Uso

```
ctbz acoes [flags]
```

## Exemplos

```sh
  ctbz acoes
  ctbz acoes --desde 2026-10-01
  ctbz acoes --limite 0 -o csv > acoes.csv
```

## Flags

```
      --desde string   só as ações a partir desta data (AAAA-MM-DD)
      --limite int     quantas ações mais recentes mostrar (0 mostra todas) (default 50)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
