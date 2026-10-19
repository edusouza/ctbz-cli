# Pró-labore e lucros

## Pró-labore vigente

```sh
ctbz prolabore
```

```text
Gerenciamento:  INTELIGENTE
Total:          R$ 1.621,00
Calculando:     não
Indisponível:   não
Sócios:
  Nome           CPF                   Valor  Recebe pró-labore  Responsável na Receita  Gestão       Atualizado em  Dependentes
  FULANO DE TAL  000.000.000-00  R$ 1.621,00  sim                sim                     INTELIGENTE  01/09/2026               2
```

- **Gerenciamento** `INTELIGENTE` significa que a Contabilizei calcula todo mês o pró-labore
  que leva ao menor imposto (inclusive o Fator R).
- O valor e as competências do card do painel (`valor_card`, `competencia_atual`,
  `competencia_anterior`) aparecem quando a Contabilizei já calculou o mês.
- Em JSON, os sócios vêm em `socios`, com o `id` que os comandos `--socio` pedem; em CSV, a
  lista de sócios sai como JSON na célula.

## Mudar a gestão do pró-labore de um sócio

```sh
ctbz prolabore definir --socio ID --tipo salario-minimo
ctbz prolabore definir --socio ID --tipo teto-inss
ctbz prolabore definir --socio ID --tipo personalizado --valor 3000,00
ctbz prolabore definir --socio ID --tipo inteligente [--minimo 1518,00]
```

- É o "Editar gestão" da central do sócio. `salario-minimo` e `teto-inss` usam os valores
  vigentes que a Contabilizei informa; `personalizado` aceita qualquer valor a partir do
  salário mínimo; `inteligente` deixa a Contabilizei calcular o valor todo mês, com um piso
  opcional (`--minimo`).
- Se o sócio era o último na gestão inteligente, a empresa sai dela; com outros sócios nela,
  a empresa continua. A CLI avisa no resumo.
- **A partir de quando vale**: a Contabilizei decide pela data (a CLI não manda competência).
  Até o fechamento do mês (em geral o dia 25) a mudança vale para o mês atual; depois, para o
  seguinte. Quando o painel diz que o mês está fechado, a CLI avisa.
- Risco alto: muda o INSS, o IRRF e o Fator R. A CLI mostra o antes e o depois e pede que
  você digite `confirmo`. Dá para mudar de novo enquanto o mês estiver aberto.
- Depois do envio, a CLI relê a central e mostra a gestão e o valor; `enviado` quer dizer que
  a gestão relida ainda não é a pedida.
- Aceita `--yes` e `--dry-run` (ver [Escrita](escrita.md)).

## Pró-labore zerado nos meses sem faturamento

```sh
ctbz prolabore zerar-sem-faturamento on
ctbz prolabore zerar-sem-faturamento off
```

- É a chave "Não quero ter pró-labore cadastrado em meses sem faturamento" da central de
  pró-labore, que vale para todos os sócios. O estado atual aparece em `ctbz prolabore -o json`
  (`zerar` na resposta deste comando).
- Ligada, nos meses sem faturamento os sócios ficam sem pró-labore e **não contribuem para o
  INSS**, o que conta para aposentadoria e auxílios. Por isso `on` é de risco alto (`confirmo`)
  e `off`, de risco médio.
- Numa empresa na gestão inteligente, `on` usa a variante do painel que vale a partir deste
  mês e não altera os meses passados.
- Se a preferência já estiver no estado pedido, nada é enviado.
- Ainda não se sabe se `off` desfaz a variante da gestão inteligente; confira com
  `ctbz prolabore` depois.

## Gestão inteligente (cálculo automático)

```sh
ctbz prolabore gestao-inteligente ativar
ctbz prolabore gestao-inteligente sair
```

- Na gestão inteligente, a Contabilizei define o pró-labore de todos os sócios todo mês para
  manter o Fator R no melhor ponto (ver [Fator R](#fator-r)).
- `ativar` só envia se a empresa for elegível (`elegivelNoMotor` da central de pró-labore).
- `sair` tira a empresa do cálculo automático: a partir daí os sócios ajustam o pró-labore
  todo mês, até o penúltimo dia do mês, para valer no mês seguinte (`ctbz prolabore definir`).
- A empresa conta como "na gestão inteligente" quando a central diz `INTELIGENTE` ou os
  parâmetros dizem que ela está no motor do Fator R. Se ela já estiver no estado pedido, nada é
  enviado.
- Risco alto: muda quem decide o valor do pró-labore. Ainda não se sabe se `ativar` desfaz um
  `sair` sem passar pelo questionário de preferências do painel; a CLI relê o estado e mostra
  `enviado` quando a mudança ainda não aparece.

## Histórico

```sh
ctbz prolabore historico               # todos os sócios, todo o histórico
ctbz prolabore historico --ano 2026
ctbz prolabore historico --socio ID    # ID de "ctbz prolabore -o json"
```

```text
Competência  Sócio          Pró-labore  Descontos
07/2026      FULANO DE TAL  R$ 1.621,00  R$ 178,31
```

A API devolve o histórico inteiro de um sócio de uma vez (não aceita ano nem página); a CLI
consulta cada sócio e filtra por `--ano`. Os descontos são o INSS e o IRRF retidos.

## Parâmetros de cálculo

```sh
ctbz prolabore parametros
```

```text
Salário mínimo:               R$ 1.621,00
Alíquota do INSS (%):         11
Contribuição máxima ao INSS:  R$ 932,31
Pró-labore no teto do INSS:   R$ 8.475,55
IRRF a partir de:             R$ 5.000,00
```

- O pró-labore não pode ser menor que o salário mínimo.
- O sócio contribui com 11% de INSS até o teto; acima do pró-labore no teto, a contribuição
  não aumenta.
- Abaixo do valor de incidência, não há IRRF retido.

## Fator R

```sh
ctbz prolabore fator-r
```

```text
Motor do Fator R:      sim
Fator R (%):           28.75
Pró-labore 12 meses:   R$ 27.600,00
Faturamento 12 meses:  R$ 96.000,00
Simulador:             DISPONIVEL
Atividades:
  CNAE       Atividade                                       Anexos    Alíquotas                                     Anexo fixo
  6311-90/0  Tratamento de dados, provedores de serviços d…  V ou III  Alíquota inicial de 6% ou 15,5%; Alíquota 6%  sim
```

- **Fator R** = pró-labore ÷ faturamento dos últimos 12 meses. A partir de 28%, as atividades
  sujeitas a ele saem do Anexo V (alíquota inicial de 15,5%) para o Anexo III (6%).
- **Motor do Fator R** indica que a Contabilizei ajusta o pró-labore todo mês para manter o
  Fator R no melhor ponto.
- `anexos` mostra em quais anexos cada atividade pode ser tributada; `anexo_fixo` diz se
  alguma delas não depende do Fator R.
- O simulador do painel é só consultado: simular cenários continua sendo feito por lá.

## Distribuição de lucros

```sh
ctbz lucros
```

```text
Exercício:               2025
Saldo na empresa:        R$ 5.000,00
Total distribuído:       R$ 3.000,00
Adiantamentos:           R$ 0,00
Limite para distribuir:  R$ 8.000,00
Exercício fechado:       sim
Pode alterar:            não
Data limite:             31/03/2026
Pendência documental:    não
Débitos federais:        não
Reabertura do balanço:   NENHUM
Por sócio:
  Sócio          Valor        Percentual  ID
  FULANO DE TAL  R$ 3.000,00  100         1000000000000001
```

- O exercício é o que a Contabilizei tem aberto para distribuição; a API não aceita outro
  ano. Sem exercício aberto, as restrições mostradas são as do ano anterior (como no painel).
- **Restrições**: pendências documentais e débitos federais impedem o informe de
  rendimentos; `reabertura_balanco` mostra se há um pedido de reabertura do balanço em
  andamento.

### Registrar a distribuição

```sh
ctbz lucros distribuir --socio 1000000000000001=60% --socio 1000000000000002=40%
ctbz lucros distribuir --socio 1000000000000001=30000,00 --socio 1000000000000002=20000,00
```

- Diz quanto do lucro do exercício cabe a cada sócio no informe de rendimentos. Os IDs são o
  `id` de cada sócio em `ctbz lucros`; cada `--socio` aceita um percentual (`60%`, `33,33%`)
  ou um valor em reais. Os sócios não citados ficam com zero.
- A soma precisa ser o **lucro total** do exercício: o saldo na empresa mais o já distribuído
  (a regra do painel). A CLI calcula em centavos; quando os percentuais não fecham por
  arredondamento, o centavo de diferença vai para o último sócio citado em percentual.
- A CLI recusa sem exercício aberto, quando o painel não deixa alterar (com o motivo), depois
  da data limite e com restrições no informe (pendência documental ou débitos federais).
- Risco alto: são os rendimentos isentos que os sócios declaram no IRPF. A CLI mostra a parte
  de cada um e pede `confirmo`. Dá para refazer enquanto a distribuição puder ser alterada.
- Depois do envio, a CLI relê a distribuição: `registrado` quando o valor relido é o enviado,
  `enviado` quando ainda não.
- O formato dos sócios na distribuição (`lucrosSocios`) não foi visto com dados reais
  ([#147](https://github.com/edusouza/ctbz-cli/issues/147)); sem sócios listados, a CLI pede
  para distribuir pelo painel.

## Informe de rendimentos dos sócios

```sh
ctbz lucros informe              # ano anterior (o da declaração de IR)
ctbz lucros informe --ano 2025 -o csv > informe-2025.csv
```

```text
Ano   Sócio          CPF             Rendimentos tributáveis  Previdência (INSS)  IRRF retido  13º salário  IRRF sobre o 13º  Lucros isentos
2025  FULANO DE TAL  000.000.000-00              R$ 19.452,00          R$ 2.139,72      R$ 0,00      R$ 0,00           R$ 0,00    R$ 50.000,00
```

- São os valores do comprovante de rendimentos que cada sócio usa na declaração de IR:
  rendimentos tributáveis (pró-labore), INSS, IRRF, 13º e lucros isentos.
- O painel monta o PDF do comprovante no navegador a partir desses valores; a API não
  oferece o PDF, por isso a CLI entrega os números (use `-o csv` ou `-o json` para guardar).
- O formato vem do código do painel: a conta usada no desenvolvimento não tinha informes.
