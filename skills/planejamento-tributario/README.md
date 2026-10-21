# Skill `planejamento-tributario`

A consultoria tributária do plano Experts, para a empresa pagar o mínimo de impostos no Simples
Nacional. O agente lê os números reais, simula cenários de pró-labore (Fator R, Anexo III x V),
mostra quanto se paga hoje e quanto se pagaria em cada cenário, e, só com o seu "ok", aplica a
decisão.

Instalação e configuração do plugin: [README das skills](../README.md).

## Quando usar

```text
> Dá para pagar menos imposto?
> Como está meu Fator R? Vale a pena aumentar o pró-labore?
> Qual pró-labore devo definir?
> Gestão inteligente compensa?
> Pró-labore ou distribuição de lucros?
> Estou perto do teto do Simples?
> Tenho cliente no exterior, muda alguma coisa?
```

Pelo nome: `/ctbz:planejamento-tributario`.

## Pré-requisitos

- O que a skill [`assessor`](../assessor/README.md) pede (sessão válida no `ctbz`).
- `python3` (só a biblioteca padrão), para o simulador.
- A empresa no **Simples Nacional**. Em outro regime, o agente diz que a análise não se aplica e
  indica o contador.

## Como funciona

### 1. Diagnóstico

O agente lê, em paralelo, empresa, atividades e anexos, Fator R, pró-labore, parâmetros (salário
mínimo, teto do INSS, faixa isenta do IRRF), faturamento de 12 meses, cálculo do mês, tabela do
IRRF e alíquotas por atividade. Então classifica a empresa:

| Caso | Sinal | O que analisa |
|---|---|---|
| **A**: depende do Fator R | alguma atividade pode ser Anexo V ou III | cenários de pró-labore |
| **B**: anexo único | nenhuma atividade depende do Fator R | pró-labore mínimo + lucros |
| **C**: já otimizada | motor do Fator R e gestão inteligente com pró-labore > 0 | se faz sentido manter |

### 2. Cenários (caso A)

O simulador [`scripts/simular.py`](scripts/simular.py) compara três cenários com DAS, INSS,
IRRF e o total do mês:

- **atual**: o pró-labore de hoje;
- **salário mínimo**;
- **Fator R 28%**: o pró-labore que leva a atividade ao Anexo III.

```text
Cenário             Pró-labore  Fator R  Anexo          DAS       INSS       IRRF        Total        Dif.
atual                 3.000,00   10,00%      V     5.025,00     330,00       0,00     5.355,00        0,00
salario_minimo        1.621,00    5,40%      V     5.025,00     178,31       0,00     5.203,31     -151,69
fator_r_28            8.400,00   28,00%    III     2.580,00     924,00       0,00     3.504,00   -1.851,00
```

(Exemplo com números fictícios: RBT12 de R$ 360 mil, um sócio.)

Ao apresentar, o agente mostra a economia mensal e anual e explica que:

- o INSS e o IRRF a mais saem do sócio, mas o INSS conta para a aposentadoria;
- o pró-labore a mais sai do que seria distribuído como lucro;
- o Fator R é uma média de 12 meses: a mudança tem efeito gradual;
- se a economia for pequena (menos de ~R$ 100/mês) ou negativa, o melhor é **não** mudar.

Empresas com menos de 12 meses têm a receita anualizada (média dos meses × 12). Se o Fator R do
simulador diferir do da CLI, vale o da CLI.

Você pode rodar o simulador por conta própria:

```sh
ctbz impostos faturamento -o json > fat.json
ctbz impostos tabela-irrf -o json > irrf.json
python3 skills/planejamento-tributario/scripts/simular.py \
  --faturamento-json fat.json --irrf-json irrf.json --socios 1 \
  --salario-minimo 1621 --teto-inss 8475.55 --irrf-isento-ate 5000

# ou com os números à mão:
python3 skills/planejamento-tributario/scripts/simular.py --rbt12 360000 --folha12 36000 --prolabore 3000 --socios 1
```

`--json` imprime o resultado em JSON; `--help` lista todas as opções (dependentes, meses de
atividade, faturamento do mês etc.).

### 3. Recomendações

- **Caso A, quando compensa:** a gestão inteligente (a Contabilizei calcula o valor exato todo
  mês), se a empresa for elegível; senão, um pró-labore fixo igual ou acima do cenário de 28%.
- **Caso B:** pró-labore no salário mínimo e o restante como lucro, a menos que o sócio queira
  contribuir mais para o INSS. No Anexo IV, cada real de pró-labore custa 31%.
- **Todos:** zerar o pró-labore nos meses sem faturamento (economiza INSS, mas o sócio fica sem
  contribuir nesses meses); exportação de serviços (alíquota externa, nota como exportação);
  proximidade do sublimite (R$ 3,6 mi) e do teto (R$ 4,8 mi) no ano-calendário; ISS retido.

### 4. Aplicar

Todas as escritas desta skill são de **risco alto**: `--dry-run`, apresentação, sua confirmação
explícita, envio, conferência e registro no dossiê com o motivo e os números.

| Ação | Consequência |
|---|---|
| `prolabore definir --socio ID --tipo …` | muda INSS, IRRF e Fator R; dá para mudar de novo com o mês aberto |
| `prolabore gestao-inteligente ativar` / `sair` | muda quem define o pró-labore |
| `prolabore zerar-sem-faturamento on` | sócios sem INSS nos meses sem faturamento (aposentadoria, auxílios) |
| `prolabore zerar-sem-faturamento off` | volta ao normal (risco médio) |

Mudanças feitas até o fechamento do mês (em geral dia 25) valem para o mês atual; depois, para o
seguinte. A CLI avisa quando o mês está fechado.

### 5. Acompanhar

A decisão, o Fator R e a economia estimada ficam no dossiê. No próximo checkup, o agente compara
o Fator R e o DAS reais com o previsto.

## Limites

É uma estimativa de apoio: o cálculo oficial é o da Contabilizei (`ctbz impostos calculo`) e o
valor exato mês a mês é o da gestão inteligente. Mudança de regime e casos fora do padrão ficam
com o contador.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| [`SKILL.md`](SKILL.md) | instruções que o agente segue |
| [`scripts/simular.py`](scripts/simular.py) | simulador de cenários de pró-labore (Anexo III x V) |
| [`references/simples-nacional.md`](references/simples-nacional.md) | anexos, Fator R, limites e pró-labore x lucros |
