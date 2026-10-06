# Ações que alteram dados

A partir da 1.1, o `ctbz` também altera dados na Contabilizei: adicionar, alterar, remover e
marcar como concluído. Todos esses comandos seguem as mesmas regras
([ADR-0018](../adr/0018-escrita-com-confirmacao.md)), descritas nesta página. Os comandos de
cada contexto chegam a partir da 1.2 (veja o [ROADMAP](https://github.com/edusouza/ctbz-cli/blob/main/ROADMAP.md)).

## Confirmação

Antes de enviar, o comando lê o que precisa, valida e mostra no stderr um resumo da ação e
o risco. No terminal, ele pede confirmação:

```text
$ ctbz caixa remover 123 --competencia 2026-09
Excluir o lançamento 123 de 15/09/2026, R$ 150,25 (risco médio)
Confirma? [s/N] s
```

Respostas aceitas: `s`, `sim`, `y` ou `yes`. Qualquer outra coisa (inclusive só Enter)
cancela.

## Níveis de risco

| Risco | Exemplos | Confirmação no terminal |
|---|---|---|
| **baixo** | estado de tela, tarefas de primeiros passos | `s` |
| **médio** | lançamentos do caixa, cadastros, classificações (dá para desfazer) | `s` |
| **alto** | aceites, manifestações à SEFAZ, parcelamentos, exclusões definitivas | digitar `confirmo` |

No risco alto, a CLI mostra também a consequência (texto do termo, custo, "não dá para
desfazer") antes de pedir a confirmação.

## Scripts e cron: `--yes`

Sem terminal (cron, pipes, CI), a CLI não tem a quem perguntar. Nesse caso, o comando só
envia com `--yes` (ou `-y`). Sem ele, termina com código `2` e nada é enviado:

```sh
ctbz caixa remover 123 --competencia 2026-09 --yes
```

O resumo continua indo para o stderr, para ficar no log do script.

## Simulação: `--dry-run`

`--dry-run` mostra a requisição que seria enviada (método, caminho e corpo) e termina com `0`,
sem pedir confirmação e **sem enviar nada**. Os `GET` de preparação (ler o lançamento, a
competência etc.) acontecem normalmente.

```text
$ ctbz caixa remover 123 --competencia 2026-09 --dry-run
DELETE /api/plataforma/caixa/lancamentousuario/remover/2026/9/123
```

- Senhas, códigos (inclusive o código de acesso do Simples) e tokens aparecem como `***`.
- Em uploads (multipart), aparecem os campos e o nome e o tamanho de cada arquivo, nunca o
  conteúdo.
- Com `-o json` ou `-o csv`, sai uma lista com `metodo`, `caminho`, `content_type`, `corpo` e
  `campos`.
- Numa ação com mais de uma requisição, valores que dependem da resposta de uma anterior
  (por exemplo, o id do arquivo enviado) aparecem vazios.

## Uma única tentativa

Uma escrita nunca é repetida automaticamente. Antes de enviar, a CLI confere se a sessão
está válida (e, com `CTBZ_OTP_CMD`, refaz o login nesse momento). Depois do envio:

- **sessão expirada (HTTP 401):** a escrita não foi aplicada; rode `ctbz login` e repita;
- **erro do servidor:** a mensagem da Contabilizei aparece traduzida, com o código HTTP;
- **falha de conexão:** a escrita **pode ou não** ter sido aplicada. Confira o estado (por
  exemplo, com `ctbz caixa 2026-09`) antes de repetir.

## Resultado

Depois de enviar, o comando relê o estado quando há como e mostra o resultado. Em `-o json`,
o resultado traz sempre as chaves `acao`, `situacao` e `id`, que fazem parte do
[contrato público](../adr/0017-contrato-publico-da-1-0.md).

## Registro das ações: `ctbz acoes`

Cada escrita enviada (inclusive as que falham, e as de `ctbz api -X`) vira uma linha JSON em
`$CTBZ_HOME/acoes.jsonl`, com permissão `0600`: data e hora, CNPJ da empresa, comando,
método, caminho (sem a query), status HTTP, resultado e id do objeto quando conhecido.
Simulações (`--dry-run`) não entram.

**Corpos de requisição e de resposta nunca são gravados**, porque podem ter dados pessoais ou
senhas. Se o registro não puder ser gravado, a CLI avisa no stderr e o resultado da escrita
continua valendo.

```sh
ctbz acoes                      # as 50 mais recentes
ctbz acoes --desde 2026-10-01   # a partir de uma data
ctbz acoes --limite 0 -o csv    # todas, para planilha
```

| Resultado | Significado |
|---|---|
| `enviada` | a Contabilizei aceitou (HTTP 2xx) |
| `recusada` | a Contabilizei respondeu com erro; nada foi alterado |
| `sem resposta` | a conexão falhou: a escrita pode ou não ter sido aplicada |

O arquivo só cresce; apague-o ou arquive-o quando quiser recomeçar.

## Códigos de saída

| Código | Quando |
|---|---|
| `0` | escrita enviada com sucesso, ou `--dry-run` |
| `1` | confirmação recusada ("operação cancelada") ou erro ao enviar |
| `2` | sem terminal e sem `--yes` (nada foi enviado), ou flag inválida |

## Baixo nível: `ctbz api -X`

`ctbz api -X POST|PUT|PATCH|DELETE` envia qualquer requisição sem confirmação, sem
`--dry-run` e sem validações. Ele também confere a sessão antes e envia uma única vez. Use só
quando souber exatamente o que está enviando.
