# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## O que é

`ctbz` é uma CLI em Go (cobra) que substitui o painel web da Contabilizei: faz o mesmo login
do navegador (usuário/senha → OTP por e-mail → escolha da empresa) e chama as URLs internas do
painel (`/api/plataforma/`, `/api/legado/`, `/api/multiusuario/`…) como se fossem uma API. O
repositório também é um plugin do Claude Code (`.claude-plugin/`, `skills/`) com skills que
operam a Contabilizei pelo `ctbz`.

## Comandos

```sh
go build -o ctbz ./cmd/ctbz
gofmt -l .                      # precisa sair vazio
go vet ./...
go test -race ./...             # o que o CI roda (Linux e macOS)
go test ./internal/cli -run TestCaixa -v          # um teste
go test ./internal/output -update                 # regrava os goldens de formato
go test ./internal/api -run Requisicoes -update   # regrava os goldens de requisição de escrita (revise o diff)
CTBZ_CONTRACT_LIVE=1 go test ./internal/api -run Live -v   # contratos contra a API real (precisa de ctbz login)

go run ./tools/gendocs                    # regenera docs/referencia (teste quebra se desatualizado)
python3 scripts/gerar-comandos-skill.py   # regenera skills/assessor/references/comandos.md depois do gendocs
go run ./tools/capture NOME               # captura, poda e anonimiza uma fixture em internal/api/testdata
go run ./tools/releasenotes vX.Y.0        # notas da release a partir do CHANGELOG
go run ./tools/notices                    # regenera THIRD_PARTY_NOTICES (teste quebra se faltar dependência do go.mod)

pip install -r docs/requirements.txt && mkdocs build --strict   # docs (também no CI)
python3 -m unittest discover -s scripts                          # testes dos scripts Python
```

No Windows com `core.autocrlf=true` (padrão do Git for Windows), o checkout fica com CRLF: o
`gofmt -l` lista todos os arquivos e falham os testes de golden (`internal/output`,
`internal/api -run Requisicoes`, `TestReferenceDocsUpToDate`). Alguns testes de sessão
concorrente (`TestSaveSessionConcorrente`, `TestReloginUnicoComChamadasEmParalelo`) também
falham no Windows. O CI roda em Linux e macOS: rode as verificações no WSL/Linux.

## Arquitetura

Camadas, de baixo para cima:

- `internal/ctbz` — HTTP e sessão: `Client` (base `https://app.contabilizei.com.br`), fluxo de
  login, `Store` em `$CTBZ_HOME` (padrão `~/.config/ctbz`) com `session.json` (0600) e o
  registro de escritas `acoes.jsonl` (sem corpo, query nem cookies). `ErrUnauthorized` e
  `WriteError` são os erros que as camadas de cima tratam.
- `internal/api` — contrato tipado da API: por contexto, uma constante `PathXxx`, o tipo da
  resposta **só com os campos usados** e `BuscarXxx(ctx, Getter)`; escritas recebem um `Sender`.
  Toda leitura é registrada em `Endpoints()` (`api.go`) e toda escrita em `Escritas()`
  (`escrita.go`). As tags `json` são o contrato verificado por `internal/contract`: campo com
  tag é obrigatório, a menos que tenha `contract:"optional"`; `json.RawMessage` = não verificado.
- `internal/cli` — árvore cobra (`root.go`, um `newXxxCmd` por arquivo). `session.go` implementa
  `Getter` com a sessão salva e **re-login automático** em 401 quando `CTBZ_OTP_CMD` está
  definido (só para leituras). `escrita.go`/`confirmacao.go` implementam `Sender`.
- `internal/output` — renderização `table|json|csv` a partir de `List`/`Record` e dos tipos
  `Money`, `Date`, `DateTime`, `CNPJ`. stdout recebe só dados; mensagens vão para stderr.
- `internal/otp` — executa `CTBZ_OTP_CMD` em polling até obter um código de 6 dígitos.
- `tools/` — geradores (`gendocs`, `capture`, `releasenotes`); `cmd/ctbz/main.go` só chama `cli.Execute`.
- `skills/` + `.claude-plugin/` — plugin do Claude Code. Cada skill tem `SKILL.md` (o que o
  agente segue), `README.md` (para pessoas), `references/` e `scripts/`. `assessor` é a base
  (login sem terminal, protocolo de escrita, dossiê); as outras a pressupõem.
  `skills/assessor/references/comandos.md` é gerado; o resto é escrito à mão e precisa
  acompanhar mudanças de comandos, flags e campos JSON. Versão do plugin = versão da CLI.

### Escritas (ADR-0018/0019/0020)

Todo comando que altera dados passa por `escrever(cmd, operacao{Risco, Resumo, Consequencia, ID}, enviar)`:
`--dry-run` troca o `Sender` por um que só imprime as requisições; senão mostra o resumo e pede
confirmação (`s/N`, ou digitar `confirmo` no risco alto), e sem terminal exige `--yes` (senão
sai com código 2). O envio confere a sessão com um GET antes e então envia **uma única vez**:
sem retentativa nem re-login depois do envio. Use `addWriteFlags` para `--yes`/`--dry-run`.
Escritas nunca são chamadas de verdade em testes nem em investigação.

A CLI não faz retentativa automática de leituras; fallback só para outra fonte dos mesmos
dados, com aviso no stderr (ADR-0013).

### Contrato público

Desde a 1.0, comandos, flags, chaves de JSON/CSV (`snake_case` em português sem acento) e
códigos de saída só mudam de forma incompatível numa major (ADR-0017). Códigos: `0` ok, `1`
erro/confirmação recusada, `2` uso incorreto, `3` login pendente, `4` atenção (`--fail-on-*`).

## Testes

- HTTP com `httptest.Server`; nunca chamar a Contabilizei real em `go test` comum. Nos testes
  da CLI: `execCLI` (`cli_test.go`), `withSession` (`output_test.go`), `writeServer`
  (`escrita_test.go`), `execWrite` (`confirmacao_test.go`) e `escritaFake`.
- Fixtures em `internal/api/testdata/` são **podadas ao contrato e anonimizadas**
  (`internal/contract/anon.go`); sintéticas ficam marcadas com `// sintética (ADR-0021)` em
  `Endpoints()`. Goldens de escrita em `testdata/requisicoes/`; `TestRequisicoesCobertura`
  exige um golden por escrita e vice-versa.
- `TestReferenceDocsUpToDate` falha se a árvore de comandos mudar sem `go run ./tools/gendocs`.
- Passos para adicionar endpoint ou escrita: `docs/contratos.md`.

## Convenções

- Português do Brasil em mensagens, ajuda, docs, CHANGELOG e commits (Conventional Commits:
  `feat:`, `fix:`, `docs:`…). Código: domínio em português sem acento (`Guia`, `Competencia`),
  termos técnicos em inglês (`Client`, `Send`). Erros em minúsculas, sem ponto, com contexto
  (`fmt.Errorf("lendo guias: %w", err)`).
- Uma issue por funcionalidade e um PR por issue, empilhados (`vX.Y/NN-slug`, corpo com
  `Closes #N`); correções em PR da base propagam com `git rebase --update-refs` (ADR-0004).
- Pronto de um PR (`docs/contribuindo.md`): testes (incluindo erro e caso vazio), escrita nova em
  `Escritas()` com golden, `docs/referencia` e mapa das skills regenerados, `README.md`/`docs/`
  (com entrada no `mkdocs.yml`), `skills/` (`SKILL.md` e `README.md`) atualizadas à mão quando
  comandos, campos JSON ou fluxos mudam,
  linha em `[Unreleased]` do `CHANGELOG.md`, ADR se houve decisão com alternativas.
- Nenhum dado pessoal real (CPF, CNPJ, nomes, e-mails, valores) em código, testes, fixtures ou docs.
- Licença (ADR-0023): PolyForm Noncommercial + permissão adicional para administrar a própria
  empresa (`LICENSE`); não chamar o projeto de open source.
- Release: mover `[Unreleased]` para `[X.Y.0]` no CHANGELOG, atualizar links e `ROADMAP.md`; após
  o merge, `git tag -a vX.Y.0` dispara GoReleaser (`.github/workflows/release.yml`).

## Documentação

`docs/` (MkDocs, publicado no GitHub Pages): `adr/` (decisões, ler antes de mudar algo
estrutural), `api/` (endpoints verificados, catálogo e `escrita/` com o que o front envia),
`autenticacao/` (fluxo de login), `guia/` (uso por contexto), `referencia/` (gerada).
`ROADMAP.md` lista versões e issues; `.github/workflows/monitor.yml` roda os contratos ao vivo
toda semana e abre issue quando algum quebra.
