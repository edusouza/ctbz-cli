# ctbz api

Chama uma URL da plataforma com a sessão atual

Chama qualquer endpoint da plataforma com os cookies da sessão.

CAMINHO relativo é resolvido contra o BFF da plataforma (/api/plataforma/);
caminhos absolutos ("/api/legado/...") alcançam as demais APIs do app.

No Git Bash do Windows (e em agentes que rodam por ele), um argumento que começa com "/" vira
caminho do Windows ("C:/Program Files/Git/api/...") antes de chegar à CLI. Comece o caminho
com "//" ("//api/legado/..."), que funciona em qualquer shell, ou rode com MSYS_NO_PATHCONV=1.
A CLI recusa um caminho que já chegou convertido.

Com -X diferente de GET, a chamada é uma escrita: a CLI confere a sessão antes e envia uma
única vez, sem re-login automático nem retentativa depois do envio. Não há confirmação nem
--dry-run: ctbz api é a ferramenta de baixo nível (ADR-0018).

A resposta sai em JSON formatado (CTBZ_OUTPUT é ignorado). Com -o table ou -o csv,
listas de objetos viram tabelas. --raw imprime o corpo exatamente como veio.
Respostas HTTP 4xx/5xx terminam com código de saída 1.

## Uso

```
ctbz api CAMINHO [flags]
```

## Exemplos

```sh
  ctbz api dadosempresa/get
  ctbz api -o table menu/get
  ctbz api /api/legado/empresa/cnpj?cnpj=00000000000100
  ctbz api //api/legado/empresa/cnpj?cnpj=00000000000100   # no Git Bash do Windows
  ctbz api -X POST -d @corpo.json caminho/qualquer
```

## Flags

```
  -d, --data string     corpo JSON da requisição (@arquivo lê de um arquivo, @- da entrada padrão)
  -X, --method string   método HTTP (default "GET")
      --raw             imprime o corpo da resposta como veio, sem formatar
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
