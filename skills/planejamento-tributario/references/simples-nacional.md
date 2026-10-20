# Simples Nacional: tabelas e regras de apoio

Fonte: LC 123/2006, na redação da LC 155/2016 (em vigor desde 2018). Use para explicar e
estimar. O valor oficial é sempre o da Contabilizei (`ctbz impostos calculo`, `ctbz notas
aliquotas`).

**Alíquota efetiva** = (RBT12 × alíquota nominal − parcela a deduzir) ÷ RBT12.
**DAS do mês** ≈ faturamento do mês × alíquota efetiva.

Empresa com menos de 12 meses: a RBT12 é proporcional (média dos meses de atividade × 12).

## Faixas (RBT12)

| Faixa | Até | Anexo I (comércio) | Anexo II (indústria) | Anexo III (serviços) | Anexo IV (serviços*) | Anexo V (serviços) |
|---|---|---|---|---|---|---|
| 1 | R$ 180.000,00 | 4,00% / 0 | 4,50% / 0 | 6,00% / 0 | 4,50% / 0 | 15,50% / 0 |
| 2 | R$ 360.000,00 | 7,30% / 5.940 | 7,80% / 5.940 | 11,20% / 9.360 | 9,00% / 8.100 | 18,00% / 4.500 |
| 3 | R$ 720.000,00 | 9,50% / 13.860 | 10,00% / 13.860 | 13,50% / 17.640 | 10,20% / 12.420 | 19,50% / 9.900 |
| 4 | R$ 1.800.000,00 | 10,70% / 22.500 | 11,20% / 22.500 | 16,00% / 35.640 | 14,00% / 39.780 | 20,50% / 17.100 |
| 5 | R$ 3.600.000,00 | 14,30% / 87.300 | 14,70% / 85.500 | 21,00% / 125.640 | 22,00% / 183.780 | 23,00% / 62.100 |
| 6 | R$ 4.800.000,00 | 19,00% / 378.000 | 30,00% / 720.000 | 33,00% / 648.000 | 33,00% / 828.000 | 30,50% / 540.000 |

Cada célula: alíquota nominal / parcela a deduzir (R$).

\* Anexo IV (ex.: construção, vigilância, limpeza, advocacia): a contribuição patronal (CPP,
20% sobre o pró-labore e salários) **não** está no DAS e é paga à parte. Aumentar o pró-labore
custa 11% + 20%.

## Fator R

- Fator R = folha dos 12 meses anteriores (para sócios sem funcionários, o pró-labore) ÷ RBT12.
- Atividades "V ou III" (veja `ctbz prolabore fator-r`, coluna `anexos`): Fator R ≥ 28% →
  Anexo III; abaixo → Anexo V.
- Atividade com `anexo_fixo: sim` não depende do Fator R: aumentar o pró-labore só gera INSS e
  IRRF a mais.
- Pró-labore que mantém 28% em regime ≈ 0,28 × RBT12 ÷ 12. Como a média é móvel, sair do V
  para o III pode exigir valores maiores por alguns meses, ou levar alguns meses.

## Limites

| RBT12 / receita do ano | Efeito |
|---|---|
| acima de R$ 3,6 milhões (sublimite) | ISS e ICMS saem do DAS e passam a ser pagos à parte, no ano seguinte (ou de imediato se passar 20% do sublimite) |
| acima de R$ 4,8 milhões (teto) | exclusão do Simples: no ano seguinte se o excesso for de até 20%; senão, a partir do mês seguinte |

A repartição dos tributos dentro do DAS por faixa e os casos de transição ficam com o contador.

## Pró-labore x distribuição de lucros

| | Pró-labore | Distribuição de lucros |
|---|---|---|
| INSS | 11% do sócio (até o teto) | não |
| IRRF | tabela progressiva (isenção até R$ 5.000 em 2026) | isenta até o lucro apurado; desde 2026, retenção de 10% sobre o que passar de R$ 50 mil/mês por sócio (Lei 15.270/2025) |
| Conta para aposentadoria | sim | não |
| Fator R | entra | não entra |

Regra prática: pró-labore no mínimo que faz sentido (salário mínimo, ou o que leva ao Fator R
de 28% quando isso compensa) e o resto como distribuição de lucros. Valores e regras da
reforma da renda de 2026 devem ser confirmados com o contador.

## Mercado externo

Serviço prestado a tomador no exterior (exportação de serviços) não paga ISS nem PIS/Cofins
dentro do DAS. Em `ctbz notas aliquotas`, as linhas `mercado: externo` mostram a alíquota
menor. Notas para clientes do exterior precisam ser emitidas como exportação.
