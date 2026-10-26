# Contribuindo

Convenções do dia a dia. Decisões maiores estão nos [ADRs](adr/README.md).

## Fluxo

1. Uma issue por funcionalidade; um PR por issue, empilhado sobre o anterior
   ([ADR-0004](adr/0004-branches-empilhadas-e-prs-por-issue.md)).
2. Branch `vX.Y/NN-slug`; commits atômicos em [Conventional Commits](https://www.conventionalcommits.org/pt-br/),
   em português.
3. Cada commit compila e passa em `gofmt -l .`, `go vet ./...` e `go test ./...`.
4. O CI (`.github/workflows/ci.yml`) roda as mesmas verificações, com `-race`, em Linux e
   macOS, e valida a documentação com `mkdocs build --strict`
   ([ADR-0014](adr/0014-ci-no-github-actions.md)). Um PR só entra com o CI verde.

## Definição de pronto de um PR

- [ ] Testes cobrindo o comportamento novo (incluindo erro e caso vazio)
- [ ] Escrita nova registrada em `api.Escritas()` com golden de requisição revisado
  ([Requisições de escrita](contratos.md#requisicoes-de-escrita))
- [ ] Documentação de usuário atualizada (`README.md` e/ou página em `docs/`, com entrada no `mkdocs.yml`)
- [ ] Referência de comandos regenerada, se a CLI mudou (`go run ./tools/gendocs`), e o mapa das
  skills também (`python3 scripts/gerar-comandos-skill.py`); campos JSON, passos (`SKILL.md`) e
  README das skills em `skills/` atualizados à mão ([Skills para agentes](guia/agentes.md))
- [ ] Linha no `CHANGELOG.md` em `[Unreleased]` quando o usuário perceber a mudança ([ADR-0003](adr/0003-versionamento-e-changelog.md))
- [ ] ADR novo, se houve decisão com alternativas razoáveis
- [ ] Nenhum dado pessoal real (CPF, CNPJ, nomes, e-mails, endereços, valores) em código, testes ou docs
- [ ] `THIRD_PARTY_NOTICES` regenerado, se alguma dependência entrou ou mudou de versão (`go run ./tools/notices`)

## Licença das contribuições

O `ctbz` é licenciado pela PolyForm Noncommercial com uma permissão adicional ([Licença](licenca.md),
[ADR-0023](adr/0023-licenca-nao-comercial-com-uso-pela-propria-empresa.md)). Ao abrir um PR,
você declara que a contribuição é sua e a licencia nesses termos, e autoriza o autor do projeto
a licenciá-la também em outros termos, inclusive comerciais.

## Idioma e nomes

- Textos para o usuário (mensagens, ajuda, docs, CHANGELOG) e commits: **português do Brasil**.
- Código: termos do **domínio em português** (`Guia`, `Empresa`, `Competencia`), termos
  **técnicos em inglês** (`Client`, `Do`, `Write`). Sem acentos em identificadores.
- Chaves de saída JSON/CSV: `snake_case` em português sem acento ([ADR-0006](adr/0006-saida-padronizada.md)).
- Mensagens de erro: minúsculas, sem ponto final, com contexto (`fmt.Errorf("lendo guias: %w", err)`).

## Testes

- Tabelas de casos (`for _, tc := range ...`) e `t.Run` quando houver vários cenários.
- HTTP: `httptest.Server`; nunca chamar a Contabilizei real em `go test` comum.
- Respostas da API: fixtures **anonimizadas** em `internal/api/testdata/`, geradas e
  revisadas como descrito em [Testes de contrato](contratos.md).
- Formatos de saída: golden files (`go test ./internal/output -update` regrava).

## Segurança

- Comandos de escrita seguem a [ADR-0018](adr/0018-escrita-com-confirmacao.md): confirmação ou `--yes`,
  `--dry-run`, risco declarado, uma única tentativa e teste de requisição. Escritas nunca são
  chamadas de verdade em testes ou investigação.
- Sessão e credenciais nunca vão para logs, testes ou commits.

## Documentação publicada

A cada push na `main`, `.github/workflows/docs.yml` gera o site com MkDocs e o publica no
GitHub Pages. Configuração única no repositório: **Settings → Pages → Build and deployment →
Source: GitHub Actions**. Para ver localmente:

```sh
pip install -r docs/requirements.txt
mkdocs serve
```

## Lançar uma versão

Segue o [ADR-0003](adr/0003-versionamento-e-changelog.md). O último PR de um épico:

1. Move o conteúdo de `## [Unreleased]` do `CHANGELOG.md` para `## [X.Y.0] - AAAA-MM-DD`,
   deixando `## [Unreleased]` vazio.
2. Atualiza os links no fim do arquivo:
   `[Unreleased]: …/compare/vX.Y.0...HEAD` e `[X.Y.0]: …/releases/tag/vX.Y.0`.
3. Marca no `ROADMAP.md` as issues entregues.

Depois do merge em `main`, quem faz o merge cria a tag no commit de merge:

```sh
git tag -a vX.Y.0 -m "vX.Y.0" && git push origin vX.Y.0
```

A tag dispara `.github/workflows/release.yml` ([ADR-0015](adr/0015-releases-com-goreleaser.md)):
testes, binários do GoReleaser e uma release no GitHub cujas notas são a seção da versão no
CHANGELOG. Para conferir as notas antes da tag: `go run ./tools/releasenotes vX.Y.0`.
