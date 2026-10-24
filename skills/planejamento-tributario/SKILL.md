---
name: planejamento-tributario
description: Consultoria tributária do plano Experts pela CLI ctbz, para a empresa pagar o mínimo de impostos no Simples Nacional. Use quando o usuário quiser pagar menos imposto ou perguntar sobre Fator R, Anexo III x V, alíquota efetiva, quanto de pró-labore definir, gestão inteligente, pró-labore x distribuição de lucros, pró-labore zerado sem faturamento, sublimite ou teto do Simples, ou exportação de serviços. Analisa os números, simula cenários e, com o ok do usuário, aplica a decisão.
---

# Planejamento tributário

Antes, siga a skill `assessor` (sessão, dossiê e protocolo de escrita). O objetivo é uma
**recomendação com números**: quanto se paga hoje, quanto se pagaria em cada cenário e o que
muda para o sócio.

## 1. Diagnóstico (leituras em paralelo)

```sh
ctbz empresa -o json               # regime_tributario
ctbz empresa atividades -o json    # anexos_simples de cada CNAE (["V","III"] = depende do Fator R)
ctbz prolabore fator-r -o json     # fator_r, motor_fator_r, simulador
ctbz prolabore -o json             # gerenciamento, socios[] (id, valor, gestao)
ctbz prolabore parametros -o json  # salario_minimo, inss_*, prolabore_teto_inss, irrf_a_partir_de
ctbz impostos faturamento -o json  # 12 meses: faturamento, prolabore, impostos_pagos
ctbz impostos calculo -o json      # DAS, INSS e IRRF do mês, alíquota efetiva
ctbz impostos tabela-irrf -o json
ctbz notas aliquotas -o json       # alíquota por atividade e mercado (interno/externo)
```

Se `regime_tributario` não for Simples Nacional, diga que esta análise vale só para o Simples e
encaminhe ao contador.

Classifique a empresa:

| Caso | Sinal | Caminho |
|---|---|---|
| **A**: depende do Fator R | alguma atividade com `anexos_simples` contendo V e III | seção 2 |
| **B**: anexo único (I, II, III ou IV) | nenhuma atividade com V e III | pró-labore mínimo + lucros (seção 3) |
| **C**: já otimizada | `motor_fator_r: true` e `gerenciamento: INTELIGENTE` com pró-labore > 0 | confira se faz sentido manter (seção 4) |

Não use `anexo_fixo` nem `anexos` de `prolabore fator-r` para classificar: na prática eles
não batem com os anexos do CNAE (uma atividade "V ou III" pode vir com `anexo_fixo: true`).
`anexos_simples` é a fonte.

Empresa sem faturamento nos 12 meses (`faturamento_12_meses` 0, `fator_r` 100): não há o que
simular. Explique que o Fator R passa a contar com o primeiro faturamento e que, com a gestão
inteligente, a Contabilizei ajusta o pró-labore desde o primeiro mês. Se a gestão está
`INTELIGENTE` mas o pró-labore é 0, procure a pendência que a bloqueia (ex.: PIS do sócio).

## 2. Fator R: cenários

Rode o simulador desta skill com os dados reais (`--socios` = sócios que recebem pró-labore):

```sh
D=$(mktemp -d)
ctbz impostos faturamento -o json > "$D/fat.json"
ctbz impostos tabela-irrf -o json > "$D/irrf.json"
python3 "<pasta desta skill>/scripts/simular.py" --faturamento-json "$D/fat.json" \
  --irrf-json "$D/irrf.json" --socios N --salario-minimo S --teto-inss T --irrf-isento-ate I \
  [--dependentes K] [--meses-atividade M]
```

`S`, `T` e `I` vêm de `ctbz prolabore parametros` (`salario_minimo`, `prolabore_teto_inss`,
`irrf_a_partir_de`); `K` é o número de dependentes de cada sócio (`socios[].dependentes` de
`ctbz prolabore`).

**Empresa com menos de 12 meses:** a RBT12 é a média dos meses de atividade × 12. O script
anualiza sozinho, contando os meses desde o primeiro com faturamento ou pró-labore; passe
`--meses-atividade M` (meses desde a abertura, de `ctbz empresa`) quando a contagem automática
estiver errada. Se `impostos faturamento` falhar com "Ano inválido" (CLI até a 1.8.0), troque
`--faturamento-json` por `--rbt12`, `--folha12` e `--prolabore`, com os valores de
`ctbz impostos calculo` (`faturamento_12_meses`, `prolabore_12_meses`) e de `ctbz prolabore`
(`total`), e informe `--meses-atividade`. O script compara três cenários (atual, salário mínimo e o pró-labore que
dá Fator R de 28%) com DAS, INSS, IRRF e o total. Use `--json` para ler o resultado.

Confira o Fator R do script com o `fator_r` da CLI. Se forem diferentes, vale o da CLI: ajuste
`--folha12`/`--rbt12` e diga isso ao usuário.

Ao apresentar:

- Mostre o quadro e a **economia mensal e anual** (total do cenário − total atual).
- Explique que o INSS e o IRRF a mais são descontados do sócio, mas o INSS conta para a
  aposentadoria, e que o pró-labore a mais sai do que seria distribuído como lucro.
- Diga que o Fator R é uma média de 12 meses: a mudança tem efeito gradual.
- Se a economia for pequena (menos de ~R$ 100/mês) ou negativa, recomende **não** mudar.

## 3. Recomendações por caso

**A, quando compensa:**
1. Se elegível (`simulador` disponível; `ctbz prolabore gestao-inteligente ativar` só envia
   se a empresa estiver no motor), a recomendação padrão é a **gestão inteligente**: a
   Contabilizei calcula o valor exato todo mês.
2. Se não for elegível, ou se o usuário preferir um valor fixo:
   `ctbz prolabore definir --socio ID --tipo personalizado --valor V` para cada sócio, com
   `V` ≥ o valor do cenário `fator_r_28` dividido pelos sócios (arredonde para cima, com folga
   se o faturamento estiver subindo).

**B:** pró-labore no salário mínimo (`--tipo salario-minimo`) e o restante como distribuição
de lucros, a menos que o sócio queira contribuir mais para o INSS. No Anexo IV, cada real de
pró-labore custa 31% (11% + 20% patronal).

**Todos:**
- `ctbz prolabore zerar-sem-faturamento on` (risco alto): economiza o INSS nos meses sem
  faturamento, mas o sócio fica sem contribuir para a aposentadoria e os auxílios nesses meses.
  Só com a escolha consciente do usuário.
- **Exportação de serviços:** se houver clientes no exterior (`ctbz notas tomadores`,
  `exterior: true`), confira em `ctbz notas aliquotas` a alíquota `externo` e lembre que as
  notas precisam ser emitidas como exportação.
- **Limites:** sublimite (R$ 3,6 milhões) e teto (R$ 4,8 milhões) se medem pela receita do
  **ano-calendário** (proporcionais no ano de abertura), não pela RBT12. Some o faturamento do
  ano corrente em `impostos faturamento`, projete o resto do ano pela média dos últimos meses e,
  se a projeção passar perto, recomende conversar com o contador sobre a mudança de regime.
- **Retenções:** `das_deducao_retencao` > 0 indica ISS retido pelo cliente, o único que reduz o
  DAS. INSS retido (Anexo IV) é compensado fora do DAS. Se o cliente retém ISS e o valor não
  aparece, oriente a falar com a Contabilizei.

## 4. Aplicar (somente com o ok do usuário)

Protocolo de escrita da skill `assessor`: `--dry-run`, apresentar, confirmação explícita
(todas são de **risco alto**), `--yes`, conferir e registrar no dossiê com o motivo e os números.

```sh
ctbz prolabore gestao-inteligente ativar|sair
ctbz prolabore definir --socio ID --tipo salario-minimo|teto-inss|inteligente [--minimo V]
ctbz prolabore definir --socio ID --tipo personalizado --valor 3000,00
ctbz prolabore zerar-sem-faturamento on|off
```

- **Quando vale:** até o fechamento do mês (em geral dia 25), vale para o mês atual; depois,
  para o seguinte. A CLI avisa quando o mês está fechado.
- Com no máximo um sócio na gestão inteligente, mudar para outro tipo tira a empresa dela,
  mesmo que esse sócio seja outro (a CLI avisa no resumo). Se a CLI recusar por não saber
  quantos sócios estão nela, não contorne: oriente o usuário a usar o painel.
- `situacao: "enviado"` = aceito, mas a releitura ainda não mostra a mudança: não repita;
  confira com `ctbz prolabore` mais tarde.

## 5. Acompanhar

Grave no dossiê a decisão, o Fator R e a economia estimada. No próximo checkup, compare o
`fator_r` e o DAS reais com o previsto.

Tabelas dos anexos, regras do Fator R, limites e pró-labore x lucros:
[references/simples-nacional.md](references/simples-nacional.md).
