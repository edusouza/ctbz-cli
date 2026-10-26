# ctbz api

Chama uma URL da plataforma com a sessão atual

Chama qualquer endpoint da plataforma com os cookies da sessão.

CAMINHO relativo é resolvido contra o BFF da plataforma (/api/plataforma/);
caminhos absolutos ("/api/legado/...") alcançam as demais APIs do app.

ctbz api só aceita caminhos relativos ou URLs do próprio painel (https://app.contabilizei.com.br/...):
uma URL de outro host (ou com outro esquema, outra porta ou usuário antes do host) é recusada
antes de qualquer envio, com código de saída 2, para que os cookies da sessão não saiam do painel.

No Git Bash do Windows (e em agentes que rodam por ele), um argumento que começa com "/" vira
caminho do Windows ("C:/Program Files/Git/api/...") antes de chegar à CLI. Comece o caminho
com "//" ("//api/legado/..."), que funciona em qualquer shell, ou rode com MSYS_NO_PATHCONV=1.
A CLI recusa um caminho que já chegou convertido.

Com -X diferente de GET, a chamada é uma escrita: a CLI confere a sessão antes e envia uma
única vez, sem re-login automático nem retentativa depois do envio. Não há confirmação nem
--dry-run: ctbz api é a ferramenta de baixo nível (ADR-0018).

-F monta um corpo multipart/form-data, como o curl: -F campo=valor para texto e
-F campo=@arquivo para um arquivo (o tipo vem da extensão; ;type=tipo escolhe outro). Sem -X,
-F envia com POST. -F e -d não podem ser usados juntos. Cada arquivo tem até 30 MB, e o envio
tem o mesmo prazo dos uploads da CLI (3 minutos). O registro de ações (ctbz acoes) guarda o
nome e o tamanho dos arquivos, nunca o conteúdo.

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
  ctbz api -F ano=2026 -F mes=9 -F documento=@extrato.csv caminho/de/upload
```

## Flags

```
  -d, --data string        corpo JSON da requisição (@arquivo lê de um arquivo, @- da entrada padrão)
  -F, --form stringArray   campo multipart/form-data: campo=valor ou campo=@arquivo[;type=tipo] (repita para mais campos)
  -X, --method string      método HTTP (default "GET")
      --raw                imprime o corpo da resposta como veio, sem formatar
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
