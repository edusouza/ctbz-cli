# Mensalidade e pagamentos

## Mensalidade atual

```sh
ctbz mensalidade
```

```text
Competência:                     10/2026
Valor:                           R$ 15,90
Situação:                        Em aberto
Observação:                      + 1 Serviço adicional
Competência anterior em atraso:  não
```

- Os dados vêm do card de mensalidade do painel (`dashboard/fatura`) e do indicador de
  atraso (`dashboard/v1/mensalidade`).
- O vencimento aparece quando a fatura já foi gerada.
- `--fail-on-atraso` termina com **código 4** quando há mensalidade de competência anterior
  em atraso.

## Segunda via e o que pagar

Com a mensalidade atrasada, o card do painel oferece "Gerar segunda via s/ juros". A CLI faz o
mesmo e mostra o que pagar:

```sh
ctbz mensalidade segunda-via
```

```text
Gerar a segunda via da mensalidade de 10/2026 (registra um boleto novo, com vencimento novo; não cobra nada) (risco médio)
Confirma? [s/N] s
```

O resultado traz competência, vencimento, total, situação da fatura, itens (mensalidade, multa e
juros, serviços adicionais), Pix copia e cola, linha digitável, código de barras e o link do
checkout.

- Risco **médio**: registra um boleto novo, com vencimento novo, mas não cobra nada. Pede
  confirmação (`--yes` em scripts) e aceita `--dry-run`.
- O caminho vem da ação do card (`ctbz mensalidade -o json`, `botao_adicional_acao`). Com a
  fatura em dia, o card não oferece segunda via e o comando para com erro.
- `--competencia AAAA-MM` confere se a segunda via oferecida é do mês esperado.
- `--pix` e `--linha` imprimem só o código, para usar num pipe:

```sh
inter-pj pix enviar --copia-e-cola "$(ctbz mensalidade segunda-via --pix --yes)"
```

- O link tem o identificador da fatura, que dá acesso a ela sem login. Ele sai só no resultado,
  nunca no registro de ações (`ctbz acoes`), e o `--verbose` o mascara.
- Depois de pagar, confirme com `ctbz mensalidade situacao`. A fatura pública demora a dar
  baixa e continua "Pendente" por um tempo.

## Situação com a Contabilizei

```sh
ctbz mensalidade situacao
```

```text
Em dia:    sim
Situação:  OK
```

- Usa a mesma consulta de inadimplência do painel, que responde texto puro: `OK` quando a
  empresa está em dia. Outras respostas não foram vistas; qualquer valor diferente de `OK` é
  mostrado como veio, com `em_dia` falso.
- `--fail-on-inadimplencia` termina com **código 4** quando a empresa não está em dia.

## Histórico de pagamentos

```sh
ctbz mensalidade historico
```

```text
Débito automático de impostos ativo:           não
Débito automático de impostos habilitado:      sim
Competência do débito automático de impostos:  09-2026
Faturas pagas:
  Data        Competência  Vencimento      Valor  Multa e juros  Forma   Status
  18/09/2026  09/2026      15/09/2026  R$ 141,83        R$ 2,83  PIX     Paga
  19/08/2026  08/2026      15/08/2026  R$ 139,00        R$ 0,00  BOLETO  Paga
              07/2026      26/07/2026    R$ 0,00        R$ 0,00          Paga
```

- As faturas pagas vêm da mesma fonte da tela "Consultar histórico" do painel
  (`GET /api/pagamentos/faturas?cnpj=…&confirmadas=true`), da mais recente para a mais antiga.
  "Multa e juros" é o que a fatura cobrou pelo atraso da anterior. A fatura zerada (ex.: o
  crédito de abertura da empresa) aparece sem data de pagamento.
- A baixa pode demorar depois do pagamento. Para saber se a mensalidade está em dia, use
  `ctbz mensalidade situacao`.
- Os campos do topo (`debito_automatico_*`, `status`, `competencia`, `proximo_pagamento`) vêm de
  `payments/recorrencia/init`, que é o débito automático de **impostos** (o mesmo de
  `ctbz impostos recorrente`), não o da mensalidade. Ficam por compatibilidade e estão obsoletos.
  Até a 1.8.0, `pagamentos` também vinha dessa fonte, por isso saía vazio para quem paga a
  mensalidade por boleto ou Pix.
- Só campos conhecidos são copiados para a saída: dados de cartão, chaves do gateway de
  pagamento e os códigos de pagamento das faturas (linha digitável, Pix) nunca aparecem.

## Plano e contrato

```sh
ctbz plano                              # plano contratado
ctbz plano contrato > contrato.html     # contrato de prestação de serviços (HTML)
ctbz plano contrato --texto | less      # o mesmo, em texto simples
ctbz plano proposta --texto             # proposta do plano, com a tabela de preços
```

```text
Plano:               2025 - Simples - Serviço - Básico [139]
Categoria:           BASICO
Valor de tabela:     R$ 139,00
Ramos de atividade:  SERVICO
```

- `ctbz plano` usa os dados gravados no login (não chama a API). Depois de mudar de plano,
  faça o login de novo.
- O valor de tabela é o do plano; a mensalidade cobrada (com serviços adicionais) está em
  `ctbz mensalidade`.
- `contrato` e `proposta` imprimem o documento no stdout. Com `--texto`, a CLI descarta
  estilos e tags, quebra linha por parágrafo e separa as células das tabelas.
