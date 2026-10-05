# Impostos

## Guias a pagar

```sh
ctbz impostos
```

```text
Grupo     ID                Imposto         Competência  Vencimento        Valor  Situação     Tipo
este_mes  1000000000000001  DARF Unificado  07/2026      06/10/2026  R$ 1.000,00  RECALCULADA  guia
Totais — Em atraso: R$ 0,00 (0) · Este mês: R$ 1.000,00 (1) · Próximo mês: R$ 0,00 (0)
```

- Grupos: `em_atraso`, `este_mes` e `proximo_mes`, como na tela de impostos do painel.
- `--atrasadas` mostra só as guias em atraso.
- Os totais por grupo saem no stderr (só na tabela), para não misturar com os dados.
- O `ID` é o que os outros comandos de impostos recebem.

### Alerta de atraso em scripts

`--fail-on-atraso` faz o comando terminar com **código 4** quando há guias em atraso:

```sh
# cron diário
ctbz impostos --fail-on-atraso -o csv > /dev/null || notify-send "Há impostos em atraso"
```

## Detalhe de uma guia

```sh
ctbz impostos guia 1000000000000001
```

Mostra valor total, valor original, juros e multa, situação, as ações disponíveis no painel
(ex.: `PAGAR`, `INFORMAR_PAGAMENTO`) e a explicação do imposto (o que é, frequência e o
impacto do atraso).

## Como o imposto foi calculado

```sh
ctbz impostos calculo
ctbz impostos tabela-irrf
```

`calculo` mostra a memória de cálculo do mês mais recente (a competência é escolhida pela
Contabilizei): faturamento, DAS do Simples Nacional, DARF de INSS e IRRF sobre o pró-labore,
faturamento e pró-labore dos últimos 12 meses e o percentual do Fator R. Valores ainda não
calculados aparecem vazios (`null` em JSON). `tabela-irrf` mostra as faixas do IRRF usadas no
cálculo do pró-labore.

## Baixar as guias em PDF

```sh
ctbz impostos baixar --pendentes -d ~/guias     # todas as guias a pagar
ctbz impostos baixar 1000000000000001           # guias específicas
```

- Os arquivos são nomeados `COMPETÊNCIA-IMPOSTO-VENCIMENTO.pdf`, por exemplo
  `2026-07-darf-unificado-2026-10-06.pdf`.
- Um arquivo que já existe não é baixado de novo (situação `já existe`); use `--force` para
  sobrescrever.
- O PDF vem de um link temporário gerado pela Contabilizei; o download não envia os cookies
  da sessão.

## Informar que uma guia foi paga (ou não)

Se você já pagou uma guia e ela continua aparecendo como a pagar, marque como paga
("Já paguei" no painel). Se ela está como "a confirmar" mas não foi paga, informe com
`--nao-paguei`; isso também libera o recálculo.

```sh
ctbz impostos confirmar 1000000000000001                # já paguei
ctbz impostos confirmar 1000000000000001 --nao-paguei   # não paguei
ctbz impostos desmarcar 1000000000000001                # desfaz uma confirmação feita por engano
```

- É uma **declaração**, não um pagamento (risco médio). A confirmação não serve como
  comprovante oficial: a Contabilizei faz auditorias periódicas e pode mudar o status com os
  dados do governo.
- A CLI usa a mesma rota da tela de impostos da sua empresa (v5 ou v3, conforme
  `impostos/rollout`).
- Pendências críticas na Central de Rotinas bloqueiam a confirmação, como no painel.
- `ctbz impostos baixar` só baixa o PDF. No painel, baixar a guia (ou copiar o código de
  barras) também a marca como "a confirmar" (`registrar-a-confirmar`). A CLI **não** faz isso:
  baixar continua sendo só leitura.

## Recalcular uma guia vencida

Para uma guia vencida e marcada como não paga, dá para pedir uma nova guia, com juros e
multa, para outra data. Sem `--vencimento`, a CLI só mostra os dados do recálculo:

```sh
ctbz impostos recalcular 1000000000000001
# Para pedir: ctbz impostos recalcular 1000000000000001 --vencimento 2026-10-09
ctbz impostos recalcular 1000000000000001 --vencimento 2026-10-20
```

- **Risco alto:** o recálculo é um serviço adicional, em geral cobrado na próxima
  mensalidade (a saída mostra `cobrar_recalculo`), e não dá para cancelar o pedido. No
  terminal, a CLI pede para digitar `confirmo`.
- A guia precisa estar marcada como não paga (`ctbz impostos confirmar ID --nao-paguei`) e o
  painel precisa oferecer o recálculo para ela. Datas indisponíveis e passadas são recusadas.
- A nova guia fica pronta em até 3 dias úteis; a Contabilizei avisa por e-mail.

## Histórico

```sh
ctbz impostos historico --ano 2026
ctbz impostos historico --ano 2026 --mes 7 --status pago -o csv
```

Lista as guias de meses anteriores com valor, valor pago, vencimento e situação, lendo todas
as páginas da API. Na tabela, o resumo "em dia / guias vencidas" sai no stderr.

## Faturamento e Fator R

```sh
ctbz impostos faturamento
```

```text
Competência  Faturamento    Pró-labore  Impostos pagos
10/2025       R$ 8.000,00  R$ 2.300,00     R$ 1.150,00
…
Faturamento 12 meses (RBT12): R$ 96.000,00 · Pró-labore 12 meses: R$ 27.600,00 · Fator R: 28.75%
```

Mostra os últimos 12 meses do mais antigo ao mais recente: faturamento e pró-labore usados na
apuração e o total pago em impostos em cada mês. O resumo traz o faturamento acumulado (RBT12,
que define a alíquota do Simples Nacional) e o Fator R (pró-labore ÷ faturamento; a partir de
28% algumas atividades saem do Anexo V para o Anexo III).

## Parcelamentos e débitos federais

```sh
ctbz impostos parcelamentos          # em andamento, ativos e encerrados
ctbz impostos parcelamento ID        # detalhe de um parcelamento
ctbz impostos debitos                # há débitos federais em aberto?
```

- A lista vem da aba "Parcelamentos" do painel (rota `impostos/v3/impostos-a-pagar/init`, que
  continua respondendo mesmo com a tela v5 ativa).
- O detalhe de um parcelamento é mostrado como a API devolve: o formato não pôde ser
  verificado, porque a conta usada no desenvolvimento não tinha parcelamentos.
- `debitos --fail-on-debitos` termina com código 4 quando há débitos federais.
- Simular ou contratar parcelamento continua sendo feito pelo painel (a CLI só lê,
  ver [ADR-0002](../adr/0002-somente-leitura-ate-1-0.md)).

## Pagamento recorrente (débito automático)

```sh
ctbz impostos recorrente             # disponível? ativo? próximo pagamento
ctbz impostos recorrente historico   # guias pagas pelo recorrente, por competência
```

- O pagamento recorrente paga as guias de impostos automaticamente no cartão de crédito ou na
  conta PJ. `recorrente` mostra se ele está disponível e ativo, a competência, as datas do
  próximo pagamento e da próxima tentativa, e quantos pagamentos estão agendados, concluídos e
  recusados.
- Dados de cartão e chaves de pagamento **nunca** aparecem na saída: só a quantidade de
  cartões salvos.
- O histórico traz uma linha por guia (nome, situação, valor) e uma linha `Custo de operação`
  por mês, quando houver. A Contabilizei só devolve um período recente; os meses anteriores
  estão em `ctbz impostos historico`.
- Ativar, desativar ou cadastrar cartão continua sendo feito pelo painel
  ([ADR-0002](../adr/0002-somente-leitura-ate-1-0.md)).
