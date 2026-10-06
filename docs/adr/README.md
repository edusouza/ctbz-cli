# Decisões de arquitetura (ADRs)

Registro das decisões do projeto. Antes de decidir algo, procure aqui; para registrar uma
decisão nova, copie o [modelo](template.md) e use o próximo número (ver [ADR-0001](0001-registrar-decisoes-em-adrs.md)).

| ADR | Decisão | Status |
|---|---|---|
| [0001](0001-registrar-decisoes-em-adrs.md) | Registrar decisões em ADRs | aceita |
| [0002](0002-somente-leitura-ate-1-0.md) | Somente leitura até a 1.0 | substituída por 0018 |
| [0003](0003-versionamento-e-changelog.md) | SemVer, Keep a Changelog e uma versão por épico | aceita |
| [0004](0004-branches-empilhadas-e-prs-por-issue.md) | Branches empilhadas, um PR por issue e commits atômicos | aceita |
| [0005](0005-documentacao-em-github-pages.md) | Documentação em Markdown publicada com MkDocs Material | aceita |
| [0006](0006-saida-padronizada.md) | Saída padronizada com List/Record e tipos de valor | aceita |
| [0007](0007-login-sessao-e-otp.md) | Login reproduzindo o navegador, sessão em arquivo e OTP por comando externo | aceita |
| [0008](0008-cobra-para-a-arvore-de-comandos.md) | Cobra para a árvore de comandos | aceita |
| [0009](0009-camada-api-tipada.md) | Camada `internal/api` com tipos de resposta e funções de leitura | aceita |
| [0010](0010-testes-de-contrato.md) | Testes de contrato derivados dos tipos, com fixtures anonimizadas | aceita |
| [0011](0011-fixtures-podadas-ao-contrato.md) | Fixtures podadas ao contrato | aceita |
| [0012](0012-prazos-e-alertas.md) | Alertas de prazo calculados pela CLI e código 4 | aceita |
| [0013](0013-sem-retentativa-com-fonte-alternativa.md) | Sem retentativa; fonte alternativa para erros conhecidos | aceita |
| [0014](0014-ci-no-github-actions.md) | CI no GitHub Actions com as mesmas verificações locais | aceita |
| [0015](0015-releases-com-goreleaser.md) | Releases com GoReleaser e notas tiradas do CHANGELOG | aceita |
| [0016](0016-monitoramento-agendado.md) | Monitorar a API com um job agendado que abre issue | aceita |
| [0017](0017-contrato-publico-da-1-0.md) | O que a 1.0 garante (contrato público da CLI) | aceita |
| [0018](0018-escrita-com-confirmacao.md) | Escrita com confirmação, simulação e sem retentativa | aceita |
| [0019](0019-camada-de-escrita.md) | Camada de escrita com `Sender`, conferência prévia da sessão e erro traduzido | aceita |
| [0020](0020-confirmacao-e-simulacao.md) | Confirmação e simulação num helper único, com o `--dry-run` como um `Sender` | aceita |
| [0021](0021-tipos-e-fixtures-sinteticos.md) | Tipos e fixtures sintéticos para endpoints sem captura real | aceita |
| [0022](0022-classificacao-de-notas-de-entrada.md) | Classificar notas de entrada pela base `/api/emissor/`, reenviando os objetos como vieram | aceita |
