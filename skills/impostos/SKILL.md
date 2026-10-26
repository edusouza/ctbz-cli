---
name: impostos
description: Guias de impostos da empresa na Contabilizei pela CLI ctbz. Use para ver o que pagar (DAS, DARF de INSS/IRRF), baixar guias em PDF, explicar como o imposto foi calculado, informar que uma guia foi paga ou não, recalcular guia vencida, ver histórico, débitos federais, simular ou contratar parcelamento e consultar o pagamento recorrente.
---

# Impostos

Antes, siga a skill `assessor` (sessão, dossiê e protocolo de escrita). IDs de guia vêm da
coluna `id` de `ctbz impostos -o json` ou de `ctbz impostos historico -o json`.

## O que pagar

```sh
ctbz impostos -o json            # grupo: em_atraso | este_mes | proximo_mes
ctbz impostos guia ID -o json    # valor_total, juros_e_multa, situacao, acoes, descricao
```

Apresente por grupo: imposto, competência, vencimento, valor e situação. Guias com
`situacao` `PAGA` continuam na lista: não são pendências (no detalhe, `situacao` é uma lista,
ex. `["PAGA"]`, e o `vencimento` pode vir `null`). Para cada guia em atraso e não paga, mostre
`juros_e_multa` e diga como resolver (abaixo).

## Pagar uma guia

1. Baixe o PDF: `ctbz impostos baixar ID -d DIR` (ou `--pendentes` para todas). Pergunte o
   diretório na primeira vez e guarde no dossiê. O nome do arquivo é
   `COMPETÊNCIA-IMPOSTO-VENCIMENTO.pdf`; com `situacao: "já existe"` nada foi baixado
   (`--force` sobrescreve).
2. O pagamento é feito pelo usuário (banco, Pix, código de barras). A CLI não paga.
3. Depois que ele pagar, ofereça informar o pagamento: `ctbz impostos confirmar ID`
   (risco médio, protocolo de escrita). É uma declaração, não um comprovante; a Contabilizei
   confere com a Receita depois.

Diferente do painel, baixar pela CLI **não** marca a guia como "a confirmar".

## Corrigir o status

| Situação | Comando (risco médio) |
|---|---|
| Pagou e ainda aparece como a pagar | `ctbz impostos confirmar ID` |
| Aparece "a confirmar" e não pagou | `ctbz impostos confirmar ID --nao-paguei` |
| Confirmou por engano | `ctbz impostos desmarcar ID` |

Pendências críticas na Central de Rotinas bloqueiam a confirmação: resolva-as antes (skill
`checkup`).

## Guia vencida

Tributo federal pago com atraso tem multa de 0,33% ao dia (até 20%) mais juros pela Selic.

1. Se ela estiver "a confirmar", marque como não paga: `ctbz impostos confirmar ID --nao-paguei`.
2. Veja o recálculo, sem enviar nada: `ctbz impostos recalcular ID -o json`. Mostre
   `valor_original`, `valor_recalculo` (estimado), `data_recomendada`, `datas_indisponiveis` e
   `cobrar_recalculo` (se o serviço entra na mensalidade).
3. Escolhida a data (não passada, fora de `datas_indisponiveis`; sugira `data_recomendada`):
   `ctbz impostos recalcular ID --vencimento AAAA-MM-DD`. **Risco alto**: serviço em geral
   cobrado e sem cancelamento. A nova guia sai em até 3 dias úteis, avisada por e-mail.

## Explicar o cálculo

`ctbz impostos calculo -o json` (a Contabilizei escolhe a competência; o nome está em `mes`):

- **DAS** = `das_imposto_bruto` − `das_deducao_retencao` = `das_total`. Alíquota efetiva =
  `das_total ÷ faturamento`. `das_inconsistente: true` → peça ao usuário que fale com a
  Contabilizei antes de pagar.
- **INSS** = `inss_aliquota` (11%) × `inss_prolabore`, até o teto (`ctbz prolabore parametros`).
- **IRRF** = (`irrf_base_calculo` × `irrf_aliquota`) − `irrf_deducao`, menos o redutor de 2026
  (Lei 15.270/2025): IR zero até R$ 5.000 de pró-labore e, de R$ 5.000,01 a R$ 7.350, redução
  de R$ 978,62 − 0,133145 × pró-labore. A base é o pró-labore menos o INSS e os dependentes, ou
  menos o desconto simplificado (R$ 607,20), o que for maior. Por isso, nessa faixa, a conta
  pela tabela (`ctbz impostos tabela-irrf`) dá mais que `irrf_total`: vale o `irrf_total`.
- `fator_r`, `faturamento_12_meses`, `prolabore_12_meses`: se o Fator R estiver abaixo de 28%
  numa atividade sujeita a ele, sugira a skill `planejamento-tributario`.

Valores `null` = ainda não calculados pela Contabilizei.

## Histórico e comprovação

```sh
ctbz impostos historico --ano 2026 [--mes 7] [--status PAGO] -o json
ctbz impostos faturamento -o json    # faturamento, pró-labore e impostos pagos, 12 meses
```

Use para responder "quanto paguei de imposto no ano" (some `valor_pago`) ou para achar uma
guia antiga. "Em dia / vencidas" sai no stderr.

## Débitos e parcelamento

```sh
ctbz impostos debitos -o json         # possui_debitos_federais
ctbz impostos parcelamentos -o json   # parcelamentos existentes
ctbz impostos parcelamento ID -o json # detalhe (como a API devolve)
```

Para simular (só leitura), escolha o tipo:

| Tipo | Quando |
|---|---|
| `simples-nacional` | DAS em atraso ainda na Receita (parcelamento digital) |
| `pgfn-simples-nacional` | Simples já inscrito em dívida ativa |
| `pgfn-previdenciario` | INSS em dívida ativa |
| `pgfn-nao-previdenciario` | outros tributos em dívida ativa |
| `especializado --negociacao VENCIDOS\|DIVIDA_ATIVA\|DIVIDA_ATIVA_E_VENCIDOS` | negociação feita por um especialista (vira um pedido de atendimento) |

```sh
ctbz impostos parcelamento simular TIPO -o json   # parcelas, entrada, demais_parcelas, servico_adicional, emissao_guia
```

Monte um quadro com cada opção: nº de parcelas, primeira, demais, custo adicional na
mensalidade e total aproximado. "Nenhum débito para parcelar" (código 0) encerra o assunto.

Contratar (`ctbz impostos parcelamento contratar TIPO [--parcelas N]`) é **risco alto**: é
confissão de dívida, as parcelas são corrigidas pela Selic + 1% ao mês, o parcelamento cai se
a 1ª parcela não for paga e não há como desfazer. Só com confirmação explícita, depois de
mostrar a simulação escolhida. Acompanhe com `ctbz impostos parcelamentos`. Se a CLI recusar
porque a simulação veio sem quantidade ou sem valores ("contrate pelo painel"), não insista:
oriente o usuário a contratar pelo painel.

## Pagamento recorrente

`ctbz impostos recorrente -o json` (`disponivel`, `ativo`, `proximo_pagamento`,
`pagamentos_recusados`…) e `ctbz impostos recorrente historico`. Ativar, desativar ou
cadastrar cartão é feito só no painel. Pagamentos recusados são um alerta: a guia
continua em aberto.
