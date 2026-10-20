---
name: checkup
description: Checkup do CNPJ na Contabilizei, como a reunião mensal com o assessor do plano Experts. Use quando o usuário perguntar como está a empresa, o que vence, o que precisa fazer, se há pendências, atrasos ou riscos, ou pedir um relatório ou resumo do mês. Coleta tudo com a CLI ctbz (só leituras), prioriza e entrega um relatório com próximos passos.
---

# Checkup da empresa

Antes, siga a skill `assessor` (sessão válida e dossiê lido). Esta skill **só lê**: ações
encontradas aqui são propostas e executadas pelas skills de cada assunto.

## 1. Coletar

Rode o script desta skill (todas as leituras em paralelo, numa chamada só):

```sh
bash "<pasta desta skill>/scripts/coletar.sh" [AAAA-MM]
```

Código 3 = sem sessão: faça o login (skill `assessor`) e rode de novo. Uma seção com erro não
invalida as outras: relate a lacuna ("não consegui ler X") e siga. Falhas conhecidas da 1.8:

- `rotinas` e `documentos pendentes` com "HTTP 206": a Central de Rotinas respondeu sem
  pendências no formato parcial. Leia o mesmo dado com `ctbz api dashboard/v2/central-rotinas`
  (JSON cru: `pendencias.pendenciasCriticas.*.possuiPendencia`).
- `impostos faturamento` com "HTTP 400 ... Ano inválido": a empresa é mais nova que a janela de
  12 meses. Use `faturamento_12_meses`, `prolabore_12_meses` e `fator_r` de `ctbz impostos calculo`.

Sem o script, rode em paralelo: `resumo`, `impostos`, `impostos debitos`,
`impostos faturamento`, `pendencias`, `pendencias termos`, `pendencias procuracao`,
`pendencias conciliacao`, `rotinas --mes M`, `documentos pendentes`,
`notas entrada --lista a-manifestar --mes M`, `notas entrada --lista a-classificar --mes M`,
`notas config`, `extratos --ano A`, `certificado`, `mensalidade`, `mensalidade situacao`,
`prolabore` e `prolabore fator-r`, todos com `-o json`.

## 2. Analisar

Classifique cada achado. A ordem das seções do relatório segue esta prioridade:

Guias com `situacao` `PAGA`/`PAGO` ainda aparecem em `impostos` e o `resumo` pode marcá-las
como `vencida`: não são pendências. Diga só que estão pagas (o `juros_e_multa` de
`impostos guia ID` mostra se foram pagas com atraso).

**🔴 Urgente (resolver hoje)**
- `impostos`: guia em `grupo: "em_atraso"` e não paga. Juros e multa correm todo dia.
- `impostos debitos`: `possui_debitos_federais: true`. Impede certidão negativa e a
  distribuição isenta de lucros.
- `resumo`: `alerta: "vencida"` ou `"crítica"`.
- `mensalidade`: `competencia_anterior_atrasada: true`, ou `mensalidade situacao` com
  `em_dia: false`.
- `certificado`: `valido: false` ou `dias_para_vencer < 0`. Sem ele não há emissão de notas
  nem entrega de obrigações.
- `notas config`: `certificado_vencido: true` ou `emissor_habilitado: false`.

**🟠 Esta semana (prazo em até 7 dias)**
- Guias do mês (`grupo: "este_mes"`) e pendências ou rotinas com `alerta: "próxima"`.
- `pendencias termos`: termo pendente; com `prazo_aceite_tacito_dias`, diga quantos dias
  faltam para o aceite tácito.
- `notas entrada` a manifestar ou a classificar (o prazo da classificação é o dia 05 do mês
  seguinte; depois disso vale a pré-classificação da Contabilizei).

**🟡 Este mês**
- `rotinas` com `responsavel: "empresa"` e status diferente de realizado (ex.: importar
  extrato bancário).
- `documentos pendentes` com itens.
- `pendencias conciliacao`: `notas_fiscais_pendentes` ou `recebimentos_pendentes` > 0.
- `pendencias procuracao`: `pendente: true`.
- `certificado`: `dias_para_vencer` ≤ 60 ou `apto_renovacao: true`.

**🟢 Oportunidades e acompanhamento**
- Fator R: atividade com `anexos_simples` contendo V e III (`empresa atividades`), `fator_r`
  < 28 e `motor_fator_r: false` → provável economia. Encaminhe para `planejamento-tributario`.
- `prolabore`: `gerenciamento: INTELIGENTE` com `total` 0 e uma pendência de cadastro (ex.: PIS)
  → a gestão inteligente não está funcionando; resolver a pendência é o que a ativa.
- `impostos faturamento`: RBT12 (some `faturamento`) acima de R$ 3 milhões (perto do sublimite
  de R$ 3,6 mi) ou de R$ 4 milhões (perto do teto de R$ 4,8 mi).
- `notas config`: `instabilidade: true` → avise que é melhor esperar para emitir notas.
- Rotinas da `contabilizei` (eSocial, DCTFWeb, DEFIS…) só como informação: não dependem do
  usuário.

Compare com o último checkup do dossiê: o que foi resolvido, o que é novo e o que se repete.

## 3. Relatório

Escreva em português, curto, com datas e valores. Modelo:

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
- Fator R em 24,1%: um pró-labore de R$ X levaria ao Anexo III (economia estimada de R$ Y/mês).

### Desde o último checkup (15/09)
- ✅ Resolvido: guia de INSS de agosto paga.
- 🔁 Repetindo: conciliação de 2 recebimentos sem nota.

### A Contabilizei entrega este mês
eSocial e DCTFWeb (15/10), sem ação sua.
```

Feche oferecendo os próximos passos concretos, um por linha, cada um ligado à skill que o
executa: `impostos`, `fechamento-mensal`, `planejamento-tributario`, `fechamento-anual`,
`notas-e-clientes` ou `cadastro-e-acessos`.

## 4. Registrar

Acrescente ao dossiê, em "Histórico de checkups", uma linha com a data, o status geral e os
itens principais. Se o usuário quiser o relatório num arquivo, grave-o onde ele indicar.

## Automatizar (se o usuário pedir)

Para um alerta diário sem agente:

```sh
ctbz resumo --fail-on-atencao -o csv > /dev/null || echo "Contabilizei precisa de atenção"
```

O código 4 indica algo vencido ou crítico. Também existem `impostos --fail-on-atraso`,
`pendencias --fail-on-vencidas`, `rotinas --fail-on-vencidas`, `impostos debitos
--fail-on-debitos`, `pendencias conciliacao --fail-on-pendencias`, `mensalidade
--fail-on-atraso` e `mensalidade situacao --fail-on-inadimplencia`. Para o login sem
intervenção no cron, é preciso `CTBZ_OTP_CMD`.
