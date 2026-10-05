# ADR-0021: Tipos e fixtures sintéticos para endpoints sem captura real

- **Status:** aceita
- **Data:** 2026-10-04

## Contexto

As escritas de v1.2 a v1.8 dependem de leituras que a CLI ainda não usava: classificações por
competência, lançamentos do extrato, `init` de modais etc. Os testes de contrato
([ADR-0010](0010-testes-de-contrato.md)) pedem uma fixture capturada da API real e podada
([ADR-0011](0011-fixtures-podadas-ao-contrato.md)), mas quem implementa nem sempre tem uma
sessão com dados para capturar (o login exige o código enviado por e-mail ao dono da conta).

O formato dessas respostas está descrito em [docs/api/escrita](../api/escrita/README.md),
pela leitura do código do painel que as consome.

## Decisão

- Quando não houver captura, o tipo em `internal/api` declara **só os campos que a
  documentação mostra que o front usa**. Os campos incertos ficam com `contract:"optional"`, e
  ids de formato incerto (número ou texto) ficam como `json.RawMessage`, que é reenviado
  igual nas escritas.
- A fixture é escrita à mão a partir da documentação, com dados fictícios, e o endpoint é
  registrado em `Endpoints()` com o comentário `// sintética (ADR-0021)` e um `LivePath`
  sempre que o caminho puder ser chamado sozinho.
- O monitoramento semanal roda os contratos ao vivo: se o formato real divergir, a quebra
  aparece na issue de monitoramento antes de alguém usar a escrita. Com uma sessão,
  `go run ./tools/capture NOME` substitui a fixture sintética pela real e o comentário sai.

## Consequências

- O desenvolvimento das escritas não depende de uma sessão real, e nenhuma chamada é feita
  durante a implementação.
- Até a primeira execução do monitoramento, o teste offline confirma só a coerência entre o
  tipo e a documentação, não com a API. As pendências ficam visíveis pelo comentário em
  `Endpoints()` (`grep -n "sintética" internal/api/api.go`).
- Um tipo errado pode deixar um comando quebrado até o monitoramento apontar. Por isso os
  campos incertos são opcionais, e a CLI valida antes de escrever (um campo ausente vira erro
  de leitura, nunca uma escrita com valor zero).

## Alternativas consideradas

- **Esperar a captura real para cada endpoint** — bloquearia todo o roadmap de escrita.
- **Sem contrato para esses endpoints** — o monitoramento não detectaria mudanças neles.
- **`map[string]any` em vez de tipos** — perde a verificação de contrato e espalha
  conversões pelos comandos.
