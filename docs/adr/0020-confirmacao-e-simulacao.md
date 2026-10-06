# ADR-0020: Confirmação e simulação num helper único, com o `--dry-run` como um `Sender`

- **Status:** aceita
- **Data:** 2026-10-04

## Contexto

A [ADR-0018](0018-escrita-com-confirmacao.md) pede, para todo comando de escrita, resumo,
confirmação por risco, `--yes` obrigatório sem terminal e `--dry-run` que mostra método,
caminho e corpo sem enviar. São dezenas de comandos previstos de v1.2 a v1.8. Se cada um
montasse a requisição duas vezes (uma para mostrar, outra para enviar), as duas poderiam
divergir, e o `--dry-run` mostraria algo diferente do que é enviado.

## Decisão

- **Helper único** `escrever(cmd, operacao, enviar)` em `internal/cli/confirmacao.go`. A
  `operacao` traz risco, resumo e consequência; `enviar` recebe um `api.Sender` e faz as
  chamadas de escrita. `addWriteFlags` acrescenta `--yes`/`-y` e `--dry-run`.
- **`--dry-run` é um `Sender`** (`dryRunSender`) que monta o corpo com os mesmos
  `api.EncodeBody`/`EncodeMultipart` e o guarda em vez de enviar. O código do comando é o
  mesmo nos dois modos. As respostas simuladas são vazias.
- **Confirmação:** o resumo vai sempre para o stderr, inclusive com `--yes` (fica no log).
  Riscos baixo e médio aceitam `s`, `sim`, `y` ou `yes`; o risco alto exige `confirmo`.
  Qualquer outra resposta cancela (código 1). Sem terminal e sem `--yes`: erro de uso
  (código 2). O `--dry-run` não pede confirmação.
- **Segredos no `--dry-run`:** campos cujo nome contém `senha`, `password`, `secret`, `token`,
  `otp` ou `chaveAcesso` (código do Simples, acrescentado na v1.8), ou se chama
  `codigo`/`code`, aparecem como `***`. Arquivos aparecem por nome e tamanho.
- **Saída do `--dry-run`:** texto no formato `table` (parecido com uma requisição HTTP);
  lista com `metodo`, `caminho`, `content_type`, `corpo` e `campos` em JSON e CSV.
- **Resultado:** `resultadoEscrita(acao, situacao, id)` começa o registro de saída com as
  chaves estáveis `acao`, `situacao` e `id`, que entram no contrato da
  [ADR-0017](0017-contrato-publico-da-1-0.md).

## Consequências

- Um comando de escrita novo só declara a operação e a função de envio; confirmação,
  simulação e (na #178) registro vêm de graça.
- Numa ação com várias requisições encadeadas, o `--dry-run` mostra valores vazios onde a
  requisição depende da resposta da anterior. Fica documentado no guia.
- O mascaramento por nome de campo pode esconder um campo inofensivo (ex.: `tokenizado`);
  esconder demais é preferível a mostrar uma senha.

## Alternativas consideradas

- **Cada comando descreve a requisição para o `--dry-run`** — duplicação e risco de o
  `--dry-run` mentir.
- **`--dry-run` interrompendo na primeira requisição** — mais simples, mas esconderia as
  requisições seguintes de ações em duas etapas (ex.: upload do extrato e evento de importação).
- **Confirmação por `s` também no risco alto** — fácil de responder por reflexo; digitar
  `confirmo` obriga a ler.
