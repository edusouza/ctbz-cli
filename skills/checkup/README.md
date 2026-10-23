# Skill `checkup`

O checkup do CNPJ, como a reunião mensal com o assessor do plano Experts: o agente lê tudo o que
importa na Contabilizei de uma vez, classifica cada achado por urgência e entrega um relatório
curto com os próximos passos.

**Só lê.** Nada é alterado. As ações que o relatório sugere são executadas depois, pelas skills
de cada assunto, com a sua confirmação.

Instalação e configuração do plugin: [README das skills](../README.md).

## Quando usar

```text
> Como está minha empresa?
> O que vence esta semana?
> Tenho alguma pendência ou atraso na Contabilizei?
> Faz o relatório do mês.
```

Pelo nome: `/ctbz:checkup`.

## Pré-requisitos

- O que a skill [`assessor`](../assessor/README.md) pede (sessão válida no `ctbz`).
- `bash`, para o script de coleta. No Windows, use o WSL ou o Git Bash; sem `bash`, o agente
  roda as mesmas leituras uma a uma.

## Como funciona

### 1. Coleta

O agente roda o script [`scripts/coletar.sh`](scripts/coletar.sh), que confere a sessão e então
dispara em paralelo (quatro por vez) todas as leituras, com `-o json`:

| Assunto | Comandos |
|---|---|
| Visão geral | `resumo` |
| Impostos | `impostos`, `impostos debitos`, `impostos faturamento` |
| Pendências | `pendencias`, `pendencias termos`, `pendencias procuracao`, `pendencias conciliacao` |
| Rotinas e documentos | `rotinas --mes M`, `documentos pendentes` |
| Notas de entrada | `notas entrada --lista a-manifestar` e `a-classificar`, do mês e do mês anterior |
| Emissor de notas | `notas config` |
| Extratos | `extratos --ano A` |
| Cadastro | `certificado`, `mensalidade`, `mensalidade situacao` |
| Pró-labore | `prolabore`, `prolabore fator-r` |

Você também pode rodar o script por conta própria:

```sh
bash skills/checkup/scripts/coletar.sh            # mês atual
bash skills/checkup/scripts/coletar.sh 2026-09    # outro mês
CTBZ_BIN=/caminho/do/ctbz bash skills/checkup/scripts/coletar.sh
```

Ele imprime uma seção por comando (JSON, stderr e código de saída) e termina com:

| Código | Significado |
|---|---|
| 0 | coletou (uma seção pode ter falhado; a falha aparece nela) |
| 1 | não deu para conferir a sessão (rede ou servidor): tente de novo **sem** refazer o login |
| 3 | sem sessão ou sessão expirada: faça o login e rode de novo |

### 2. Análise

Cada achado vai para uma das quatro faixas:

| Faixa | Exemplos |
|---|---|
| 🔴 **Urgente** (hoje) | guia em atraso e não paga; débitos federais; pendência vencida ou crítica; mensalidade atrasada; certificado vencido; emissor de notas desabilitado |
| 🟠 **Esta semana** | guias do mês; pendências e rotinas com prazo próximo; termo com aceite tácito chegando; notas de entrada a manifestar ou classificar (as do mês anterior vencem no dia 05) |
| 🟡 **Este mês** | rotinas da empresa não realizadas (ex.: importar extrato); documentos pedidos; notas e recebimentos a conciliar; procuração do e-CAC; certificado vencendo em até 60 dias |
| 🟢 **Oportunidades** | Fator R abaixo de 28% numa atividade sujeita a ele; gestão inteligente parada por pendência de cadastro; faturamento do ano perto do sublimite ou do teto do Simples; instabilidade no emissor |

Guias já pagas continuam aparecendo na lista de impostos e não contam como pendência. O agente
compara com o último checkup do dossiê: o que foi resolvido, o que é novo e o que se repete.

### 3. Relatório

```markdown
## Checkup — RAZÃO SOCIAL — 20/10/2026

**Situação geral:** 🟠 2 itens para esta semana, nada em atraso.

### 🔴 Urgente
(nenhum)

### 🟠 Esta semana
- **DAS 09/2026: R$ 1.000,00, vence 20/10.** Baixe a guia: `ctbz impostos baixar ID`.
- **Termo de débitos pendente:** aceite tácito em 5 dias.

### 🟡 Este mês
- Importar o extrato de setembro do Banco X (prazo 05/11).

### 🟢 Oportunidades
- Fator R em 24,1%: um pró-labore de R$ X levaria ao Anexo III.

### Desde o último checkup (15/09)
- ✅ Resolvido: guia de INSS de agosto paga.
- 🔁 Repetindo: conciliação de 2 recebimentos sem nota.

### A Contabilizei entrega este mês
eSocial e DCTFWeb (15/10), sem ação sua.
```

No fim, o agente oferece os próximos passos, cada um ligado à skill que o executa, e registra uma
linha no histórico de checkups do dossiê. Se você pedir, grava o relatório num arquivo.

## Alerta diário sem agente

Os comandos da CLI têm flags que terminam com código 4 quando algo exige atenção. Para um cron:

```sh
ctbz resumo --fail-on-atencao -o csv > /dev/null || echo "Contabilizei precisa de atenção"
```

Também existem `impostos --fail-on-atraso`, `pendencias --fail-on-vencidas`,
`rotinas --fail-on-vencidas`, `impostos debitos --fail-on-debitos`,
`pendencias conciliacao --fail-on-pendencias`, `mensalidade --fail-on-atraso` e
`mensalidade situacao --fail-on-inadimplencia`. Para o cron fazer login sozinho, configure
`CTBZ_OTP_CMD`.

## Falhas conhecidas até a CLI 1.8.0

Corrigidas na 1.8.1; a skill contorna quando a CLI é mais velha:

- "HTTP 206" nos comandos que leem a Central de Rotinas (`rotinas`, `resumo`,
  `documentos pendentes`…): o agente lê por `ctbz api dashboard/v2/central-rotinas`.
- "Ano inválido" em `impostos faturamento` em empresas com menos de um ano: usa os números de
  `ctbz impostos calculo`.
- `resumo` marcando como vencida uma guia já paga.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| [`SKILL.md`](SKILL.md) | instruções que o agente segue |
| [`scripts/coletar.sh`](scripts/coletar.sh) | coleta em paralelo, só leituras |
