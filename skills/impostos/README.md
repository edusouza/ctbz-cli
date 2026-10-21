# Skill `impostos`

As guias de impostos da empresa: o que pagar e quando, o PDF de cada guia, como o valor foi
calculado, informar que você pagou (ou não), recalcular uma guia vencida, histórico, débitos
federais, parcelamento e pagamento recorrente.

Instalação e configuração do plugin: [README das skills](../README.md).

## Quando usar

```text
> O que tenho de imposto para pagar?
> Baixa o DAS deste mês em ~/Documentos/guias.
> Paguei a guia do INSS, pode marcar.
> Por que o DAS deu esse valor?
> Tenho uma guia vencida, o que faço?
> Quanto paguei de imposto este ano?
> Dá para parcelar meus débitos?
```

Pelo nome: `/ctbz:impostos`.

## Pré-requisitos

O que a skill [`assessor`](../assessor/README.md) pede (sessão válida no `ctbz`).

## O que ela faz

### O que pagar

`ctbz impostos` agrupa as guias em **em atraso**, **este mês** e **próximo mês**. O agente
mostra imposto, competência, vencimento, valor e situação; para cada guia em atraso, os juros e
a multa e como resolver. Guias já pagas continuam na lista e não são pendências.

### Pagar uma guia

1. Baixa o PDF (`ctbz impostos baixar ID -d DIR`, ou `--pendentes` para todas). O diretório é
   perguntado na primeira vez e guardado no dossiê. O arquivo se chama
   `COMPETÊNCIA-IMPOSTO-VENCIMENTO.pdf`.
2. **Você paga** (banco, Pix, código de barras). A CLI não paga.
3. Depois, o agente oferece informar o pagamento (`ctbz impostos confirmar ID`). É uma
   declaração, não um comprovante: a Contabilizei confere com a Receita.

Diferente do painel, baixar pela CLI **não** marca a guia como "a confirmar".

| Situação | O que o agente faz |
|---|---|
| Pagou e ainda aparece como a pagar | `impostos confirmar ID` |
| Aparece "a confirmar" e não pagou | `impostos confirmar ID --nao-paguei` |
| Confirmou por engano | `impostos desmarcar ID` |

### Guia vencida

Tributo federal pago com atraso tem multa de 0,33% ao dia (até 20%) mais juros pela Selic. O
agente:

1. marca como não paga, se estiver "a confirmar";
2. mostra o recálculo sem enviar nada: valor original, valor estimado, data recomendada, datas
   indisponíveis e se o serviço é cobrado na mensalidade;
3. com a data escolhida e a sua confirmação explícita, pede a nova guia, que sai em até 3 dias
   úteis, avisada por e-mail.

### Explicar o cálculo

A partir de `ctbz impostos calculo`, o agente explica o DAS (imposto bruto − retenções, e a
alíquota efetiva), o INSS do pró-labore (11% até o teto) e o IRRF (com o redutor de 2026, que
zera o IR até R$ 5.000 de pró-labore). Se o Fator R estiver abaixo de 28% numa atividade sujeita
a ele, sugere a skill [`planejamento-tributario`](../planejamento-tributario/README.md). Se a
Contabilizei marcar o DAS como inconsistente, pede que você fale com ela antes de pagar.

### Histórico

`ctbz impostos historico --ano 2026` e `ctbz impostos faturamento` respondem "quanto paguei de
imposto no ano" e acham guias antigas.

### Débitos e parcelamento

`ctbz impostos debitos` diz se há débitos federais. Para parcelar, o agente simula (só leitura)
o tipo adequado e monta um quadro com parcelas, entrada, demais parcelas, custo adicional na
mensalidade e total aproximado:

| Tipo | Quando |
|---|---|
| `simples-nacional` | DAS em atraso ainda na Receita |
| `pgfn-simples-nacional` | Simples já inscrito em dívida ativa |
| `pgfn-previdenciario` | INSS em dívida ativa |
| `pgfn-nao-previdenciario` | outros tributos em dívida ativa |
| `especializado` | negociação feita por um especialista (vira um pedido de atendimento) |

Contratar é **risco alto** (veja abaixo) e só acontece depois de você escolher uma das opções
simuladas.

### Pagamento recorrente

`ctbz impostos recorrente` mostra se o débito automático de impostos está disponível e ativo, o
próximo pagamento e os recusados. Ativar, desativar ou cadastrar cartão é feito só no painel.

## O que pode alterar

Tudo com `--dry-run` primeiro e a sua confirmação (protocolo da skill `assessor`).

| Ação | Risco | Consequência |
|---|---|---|
| `impostos confirmar ID [--nao-paguei]` | médio | declaração; `impostos desmarcar ID` desfaz |
| `impostos recalcular ID --vencimento D` | **alto** | serviço em geral cobrado na mensalidade; não cancela |
| `impostos parcelamento contratar TIPO` | **alto** | confissão de dívida; parcelas com Selic + 1% ao mês; cancelado se a 1ª não for paga; sem desfazer |

Pendências críticas na Central de Rotinas bloqueiam a confirmação de pagamento: resolva-as antes
(skill [`checkup`](../checkup/README.md)).

## Fora do alcance

Pagar a guia, ativar o débito automático e cadastrar cartão: painel ou app da Contabilizei.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| [`SKILL.md`](SKILL.md) | instruções que o agente segue |
