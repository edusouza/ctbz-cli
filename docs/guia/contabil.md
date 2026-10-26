# Contabilidade

Relatórios contábeis, caixa, extratos e plano de contas. Os relatórios usam o mês como
`AAAA-MM`; o painel mostra por padrão dezembro do ano anterior.

## Balancete

```sh
ctbz balancete 2026-08
ctbz balancete 2026-08 -o csv > balancete-2026-08.csv
```

```text
Conta          Descrição                          Nível  Saldo anterior      Débitos     Créditos       Saldo
1              ATIVO                                  1         R$ 0,00  R$ 1.815,00  R$ 1.419,83    R$ 395,17
1.01           CIRCULANTE                             2         R$ 0,00  R$ 1.815,00  R$ 1.419,83    R$ 395,17
1.01.01.01.00          Caixa Geral                    5         R$ 0,00  R$ 1.139,00  R$ 1.280,83  -R$ 141,83
```

- Cada linha é uma conta da árvore contábil; na tabela, a descrição é recuada pelo nível.
- Em CSV e JSON a descrição vem sem recuo e o nível fica na coluna `nivel`, o que facilita
  filtrar (ex.: só o nível 1 para ATIVO, PASSIVO e PATRIMÔNIO LÍQUIDO).
- O PDF do painel é gerado no navegador a partir dos mesmos dados; a API não oferece PDF.

## Balanço patrimonial

```sh
ctbz balanco 2025          # fechamento do exercício (dezembro)
ctbz balanco 2026-09       # posição em um mês
```

```text
Conta  Descrição            Nível  Grupo            Saldo  Exercício anterior
1      ATIVO                    1  ATIVO        R$ 395,17             R$ 0,00
1.01     CIRCULANTE             2  ATIVO        R$ 395,17             R$ 0,00
2      PASSIVO                  1  PASSIVO     -R$ 395,17             R$ 0,00
```

- Ativo, passivo e patrimônio líquido, com o saldo do exercício e o do exercício anterior.
- As contas de resultado (receitas e despesas) ficam de fora, como no painel.

### Reabrir o balanço

```sh
ctbz lucros informe restricoes --ano 2025   # confira a pendência e o caminho de regularização
ctbz balanco reabrir --ano 2025
```

- É o "Regularizar pendências" do informe de rendimentos: reabre o exercício contábil do ano
  para que as pendências documentais sejam acertadas.
- A CLI só envia quando o informe do ano tem pendência documental, a regularização indicada é
  a reabertura do balanço (`REABERTURA_BALANCO`) e não há outra reabertura em andamento.
- Quando o período já foi fechado e a regularização é um **serviço adicional pago**, a CLI
  recusa e mostra o custo: contrate pelo painel.
- Risco alto e **sem desfazer**: o informe de rendimentos fica indisponível até a análise da
  Contabilizei, e a distribuição de lucros do ano depende dela. Uma sessão de administrador é
  recusada pela Contabilizei.
- Depois do envio, a CLI relê o status da reabertura (`EM_ANDAMENTO`, `ANALISANDO`); o
  andamento aparece na Central de Rotinas (`ctbz pendencias`). Se a releitura falhar ou ainda
  não mostrar o status, a saída é `enviado` (com um aviso): a reabertura já foi pedida e não
  deve ser repetida.

## Razão

```sh
ctbz razao --de 2026-01 --ate 2026-09               # todas as contas no período
ctbz razao --conta 1.01.01.01.00 --de 2026-07       # uma conta
ctbz razao --conta 1.01 -o csv > circulante.csv     # prefixo: todas as contas de 1.01
```

```text
Data        Conta          Descrição da conta  Histórico       Contrapartida                                    Débito  Crédito        Saldo
21/07/2026  1.01.01.01.00  Caixa Geral         Capital social  2.07.01.01.00 Capital Social Realizado no País  R$ 1.000,00         R$ 1.000,00
```

- Um lançamento por linha, com a conta de contrapartida e o saldo acumulado da conta no
  exercício depois do lançamento.
- A API devolve um mês por vez; a CLI consulta cada mês do período (até 24).

## Caixa do mês

```sh
ctbz caixa 2026-09
ctbz caixa 2026-09 -o csv > caixa-2026-09.csv
```

```text
Data        Descrição               Conta                         Classificação        Valor  Situação    Automático  ID
18/09/2026  Pagamento Contabilizei  Mensalidade de contabilidade  DESPESA         -R$ 141,83  CONFIRMADO  sim         1000000000000001
Total do mês: -R$ 141,83 em 1 lançamento(s)
```

- Entradas e saídas classificadas do mês, com a conta de classificação (de
  `ctbz contas`) e o tipo (receita, despesa…). Saídas têm valor negativo.
- `automatico` indica lançamentos feitos pelo sistema (ex.: a mensalidade da Contabilizei).
- Como no painel, a CLI pede até 1000 lançamentos por mês e avisa se houver mais.

## Extratos e contas bancárias

```sh
ctbz extratos --ano 2026      # situação do extrato por conta e mês
ctbz contas-bancarias         # contas cadastradas
```

```text
Competência  Banco                  Agência  Conta      Situação  Integração  ID da conta
10/2026      Contabilizei Conta PJ  0001     000000000  ABERTO                1000000000000001
```

- `extratos` mostra, para cada conta e mês, se o extrato está aberto ou fechado na
  Contabilizei. Mês sem extrato vira a rotina "Importar extrato bancário" (`ctbz rotinas`).
- `contas-bancarias` mostra banco, agência, conta, saldo inicial e a integração com o banco
  (ex.: `INTEGRADA` na conta PJ da Contabilizei, que dispensa importar extrato).

### Importar um extrato

```sh
ctbz extratos importar extrato-setembro.ofx --conta-bancaria 1000000000000001 --competencia 2026-09
```

```text
Importar o extrato extrato-setembro.ofx (OFX, 48 KB) da conta 1000000000000001 em 09/2026 (risco médio)
Confirma? [s/N] s
Saldo do último dia lido pela Contabilizei: R$ 1.234,56. Confere com o extrato? [s/N] s
```

- Antes de enviar, a CLI confere o tamanho (até 30 MB), se o banco aceita o formato do arquivo
  e se há um período anterior com importação pendente. Contas integradas não precisam de
  importação. Formato que o banco não aceita: a mensagem mostra a lista dele (ex.: "o banco
  desta conta aceita extratos em OFX ou PDF; este arquivo é .csv").
- **OFX fora da lista do banco** passa com um aviso no stderr, porque a tela "Importar extrato"
  do painel aceita OFX de qualquer banco. O Banco Inter, por exemplo, lista só PDF, XLS e CSV, e
  o painel importou o OFX dele (26/10/2026). Se o arquivo não servir, a recusa vem da
  Contabilizei no envio, com o motivo.
- **CSV, XLS e outros formatos que o banco aceita** (fora OFX e PDF) vão pela Central de
  Documentos, como no painel, porque a tela "Importar extrato" só aceita OFX e PDF. É um envio só,
  sem conferência de saldo. A Contabilizei importa o arquivo **manualmente, em até 45 dias**. O
  resultado sai com `situacao: "enviado"` e `importacao: "manual"`. Até lá a competência
  continua `PENDENTE` em `ctbz extratos`, o que é o esperado: não envie o arquivo de novo (a
  Central não substitui por nome) e acompanhe na Central de Documentos do painel. Esses formatos
  não aceitam `--substituir` (a exclusão deixaria o mês sem extrato até a importação manual) e
  são recusados sobre um extrato já importado, para não duplicar os lançamentos.
- A CLI também recusa importar uma competência que já está `IMPORTADO` ou `PROCESSANDO`, como o
  painel. Atenção: o processamento de um extrato pode deixar o mês **seguinte** como
  `IMPORTADO` sem nenhum lançamento (visto numa conta real); a mensagem avisa quando o extrato
  está vazio.
- Para refazer a importação de um mês já importado, use `--substituir`: a CLI exclui o extrato
  atual e importa o arquivo, com uma confirmação só (o mesmo que excluir e importar no painel,
  que não tem outro jeito de reimportar). Extrato atual vazio: risco médio. Com lançamentos:
  risco alto, porque as classificações se perdem. O `--dry-run` mostra a exclusão, o envio e a
  conclusão. Se a exclusão der certo e a importação falhar, a CLI relê a competência e diz em
  que situação ela ficou. Pendente: corrija e rode de novo, sem `--substituir`. `PROCESSANDO`
  (já visto por minutos depois de um arquivo recusado): espere e confira com `ctbz extratos`
  antes de importar de novo. Se a conexão cair ou o servidor responder com erro 5xx ao
  concluir, não se sabe se o novo arquivo entrou: confira antes de qualquer nova tentativa e não
  use `--substituir`, que excluiria o extrato novo.
- Depois de importar, a CLI relê a competência. Se ela ainda está `PROCESSANDO`, confira em
  alguns minutos com `ctbz extratos lancamentos`; se terminou `IMPORTADO` sem nenhum lançamento,
  a CLI avisa no stderr. O JSON traz `quantidade_lancamentos`.
- "Não reconheceu o arquivo como extrato bancário": tente o OFX do banco ou o PDF baixado
  direto do internet banking (não um PDF gerado por outro programa); se continuar, abra um
  chamado no painel.
- No OFX, a Contabilizei lê o saldo do último dia do arquivo e a CLI pede para conferir. Em
  scripts, `--saldo-final 1.234,56` compara com o valor esperado: se for diferente, a importação
  para.
- Se a importação parar depois do envio (saldo recusado, extrato de outra competência), o
  arquivo fica enviado mas o extrato não é efetivado; corrija e rode de novo.
- Se a conexão cair ou o servidor responder com erro 5xx (502, 504…) no envio do arquivo (o
  envio tem prazo de 3 minutos), a CLI relê a situação da competência. Se ela continua como
  estava (ex.: `PENDENTE`), nada foi importado e é seguro rodar o mesmo comando de novo: o
  arquivo vai com o mesmo nome e substitui o anterior. Se ela já aparece como `IMPORTADO` ou
  `PROCESSANDO`, ou se a queda ou o 5xx aconteceu ao concluir a importação, a importação pode
  ter sido aplicada: confira com `ctbz extratos` e `ctbz extratos lancamentos` antes de repetir,
  para não importar duas vezes.
- Erros conhecidos viram mensagens claras: extrato de outro banco, outra conta ou agência, conta
  não cadastrada (cadastre com `ctbz contas-bancarias adicionar`).
- Para desfazer, use `ctbz extratos excluir`:

```sh
ctbz extratos excluir --conta-bancaria 1000000000000001 --competencia 2026-09
```

`excluir` é de **risco alto**: a importação volta a ficar pendente, as classificações e
informações adicionais se perdem e não dá para desfazer (só reimportar e reclassificar). Se os
contadores já concluíram a classificação do extrato, ele não pode mais ser excluído.

Uma competência presa em `PROCESSANDO` (visto com o Banco Inter depois de um PDF recusado) não
sai desse estado com `excluir`: a Contabilizei responde 200 e nada muda. A confirmação avisa
disso e, se a releitura ainda mostra `PROCESSANDO`, o comando termina com erro (código 1), sem
informar a exclusão. O que destravou o mês foi importar o OFX pela tela "Importar extrato" do
painel, que mostra o botão de importar nesse estado; a outra saída é um chamado.

## Lançar no caixa

Recebimentos e pagamentos em dinheiro entram no caixa da competência. É uma escrita de
risco médio: pede confirmação (`--yes` em scripts) e aceita `--dry-run`
([Ações que alteram dados](escrita.md)).

```sh
ctbz caixa adicionar --competencia 2026-09 --data 2026-09-15 --pagamento \
  --valor 150,25 --conta 1000000000000012 --descricao "Material de escritório"
```

```text
Adicionar pagamento de R$ 150,25 em 15/09/2026 no caixa de 09/2026: "Material de escritório" (Pagamento de Fornecedores) (risco médio)
Confirma? [s/N] s
Ação       adicionar
Situação   adicionado
ID         1000000000000099
…
```

- O valor é sempre positivo; `--recebimento` ou `--pagamento` define o sinal.
- `--conta` aceita o id ou a descrição exata de uma classificação aceita na competência
  (veja abaixo). Contas "Impostos - …" exigem `--guia ID` ou `--sem-guia`; "Sócios -
  Distribuição de Lucros Antecipados" exige `--socio ID`.
- Sem `--data`, vale o dia que o painel sugere: hoje na competência atual, senão o dia 1.
  Datas futuras são recusadas.
- Depois de salvar, a CLI relê o caixa e mostra o lançamento criado.

Para corrigir um lançamento manual, informe só o que muda; o resumo mostra o antes e o depois:

```sh
ctbz caixa editar 1000000000000002 --competencia 2026-09 --valor 900 --descricao "Venda balcão"
# Editar o lançamento 1000000000000002 do caixa de 09/2026: valor R$ 1.000,00 → R$ 900,00; descrição "Venda à vista" → "Venda balcão"
```

- Os ids aparecem em `ctbz caixa 2026-09 -o json`.
- Lançamentos feitos pelo sistema (coluna "Automático") não podem ser editados, como no painel.
- Trocar a classificação exige informar de novo o vínculo, quando a nova conta pede um.

Para excluir um lançamento manual:

```sh
ctbz caixa remover 1000000000000002 --competencia 2026-09
# Excluir permanentemente o lançamento 1000000000000002 do caixa de 09/2026: "Venda à vista" de 18/09/2026 no valor de R$ 1.000,00 (risco médio)
# Confirma? [s/N]
```

A API não tem "desfazer". Depois de remover, a CLI mostra os dados do lançamento e, no stderr,
o `ctbz caixa adicionar` que o recria.

## Classificar lançamentos do extrato

```sh
ctbz extratos lancamentos --conta-bancaria 1000000000000001 --competencia 2026-09
ctbz extratos classificar 1000000000000102 --conta-bancaria 1000000000000001 \
  --competencia 2026-09 --conta "Pagamento de Fornecedores"
```

```text
Classificar o lançamento 1000000000000102 ("PAGAMENTO BOLETO", -R$ 1.234,56): sem classificação → Pagamento de Fornecedores (risco médio)
Confirma? [s/N]
```

- O id da conta bancária vem de `ctbz contas-bancarias` ou `ctbz extratos`.
- `lancamentos` mostra a classificação atual (pelo nome), a situação, a classificação que a
  Contabilizei sugere com a probabilidade em % e, nas partes de um lançamento desmembrado, o id
  do original (coluna "Parte de"). Mês sem lançamentos imprime uma lista vazia, avisa no stderr
  e sai com código 0.
- Entradas aceitam classificações `RECEITA` e saídas, `DESPESA` (`ctbz extratos contas`).
  Classificações de sócio exigem `--socio` (ids dos sócios do extrato).
- Quando os contadores já fecharam o período, a classificação não pode mais ser trocada: o
  painel mostra "Nossos contadores já classificaram…", e a CLI recusa com a mesma mensagem.

## Desmembrar um lançamento do extrato

Um lançamento que mistura coisas (ex.: um boleto que paga aluguel e condomínio) pode ser
dividido em partes com classificações diferentes:

```sh
ctbz extratos desmembrar 1000000000000102 --conta-bancaria 1000000000000001 --competencia 2026-09 \
  --parte "Aluguel:1.000,00:Pagamento de Fornecedores" \
  --parte "Condomínio:234,56:1000000000000012"
```

- Cada `--parte` é `"descrição:valor:conta[:sócio]"`. O valor é positivo; o sinal é o do
  lançamento original. A descrição não pode ter `:`.
- São necessárias pelo menos 2 partes, nenhuma com valor zero, e a soma tem de ser igual ao
  valor do original (centavo por centavo).
- Não dá para desmembrar uma parte de outro desmembramento nem um lançamento com vínculo.

Para voltar ao lançamento original (o id pode ser o do original ou o de uma parte):

```sh
ctbz extratos desfazer-desmembramento 1000000000000102 --conta-bancaria 1000000000000001 --competencia 2026-09
```

## Classificações aceitas numa competência

Antes de lançar no caixa ou classificar um lançamento do extrato, veja quais classificações
a competência aceita (o mesmo filtro do painel):

```sh
ctbz caixa contas --competencia 2026-09 [--recebimento|--pagamento]
ctbz caixa contas --competencia 2026-09 --vinculos     # guias e sócios aceitos como vínculo
ctbz extratos contas --competencia 2026-09 [--receita|--despesa]
```

```text
ID                Classificação                                Tipo         Exige
1000000000000013  Impostos - Simples Nacional                  pagamento    guia
1000000000000014  Sócios - Distribuição de Lucros Antecipados  pagamento    sócio
```

- No caixa, "Exige" aponta as classificações que pedem um vínculo: uma guia de imposto
  (contas "Impostos - …", ou `0` para SEM GUIA) ou um sócio. Os ids vêm de `--vinculos`.
- No extrato, entradas aceitam classificações `RECEITA` e saídas, `DESPESA`. As marcadas em
  "Exige sócio" pedem `--socio`.

## Cadastrar e editar contas bancárias

```sh
ctbz contas-bancarias bancos                      # bancos aceitos (id e código)
ctbz contas-bancarias adicionar --banco 237 --agencia 1234 --conta 12345-6 \
  --abertura 2024-01-10 --declaro-conta-pj
ctbz contas-bancarias editar 1000000000000001 --abertura 2024-02-01
```

- `--banco` aceita o id ou o código. O código vale pelo valor exato: `077` é o Banco Inter
  S.A., e o Banco Inter - Internacional só é escolhido com `077-1`. Sem código exato, vale o
  código sem o sufixo (`104` → `104-0`) se houver um só banco. Se o código for de mais de um
  banco (há códigos repetidos na lista), a CLI lista id, código e nome e pede o id. O resumo da
  confirmação e o `--dry-run` mostram o nome e o código do banco escolhido.
- `--declaro-conta-pj` é obrigatório, como o checkbox do painel: a conta é PJ e movimenta só
  este CNPJ.
- A data de abertura define desde quando a Contabilizei vai pedir os extratos da conta.
- A conta é digitada com dígito (`12345-6`); a CLI envia sem o hífen, como o painel.
- O painel pode bloquear a edição (por exemplo, com extratos já classificados); a CLI mostra
  o motivo.

Para excluir uma conta cadastrada por engano ou encerrada:

```sh
ctbz contas-bancarias remover 1000000000000001
```

É de **risco alto**: a exclusão é permanente, os vínculos da conta se perdem e um novo
cadastro não os recupera. No terminal, a CLI pede para digitar `confirmo`.

## Plano de contas (classificações)

```sh
ctbz contas --busca alug                # por trecho da descrição ou da conta contábil
ctbz contas --situacao ativo -o csv
```

```text
Descrição  Conta contábil            Classificação  Situação  ID
Aluguel    Aluguéis e Arrendamentos  DESPESA        ATIVO     1
```

- São as contas usadas para classificar entradas e saídas (as mesmas do `ctbz caixa`), cada
  uma ligada a uma conta contábil.
- `--busca` não diferencia maiúsculas, mas acentos contam: um trecho como `alug` acha
  "Aluguel" e "Aluguéis".
