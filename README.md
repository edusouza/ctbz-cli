# ctbz — CLI para a Contabilizei

[![CI](https://github.com/edusouza/ctbz-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/edusouza/ctbz-cli/actions/workflows/ci.yml)

CLI que substitui a interface web da [Contabilizei](https://app.contabilizei.com.br):
faz o mesmo login que o navegador (usuário/senha → código por e-mail → escolha
da empresa) e chama as mesmas URLs internas que o painel usa, como se fossem uma API.

Próximas versões e funcionalidades: [ROADMAP.md](ROADMAP.md).

## Instalação

Baixe o binário do seu sistema na [página de releases](https://github.com/edusouza/ctbz-cli/releases)
(Linux, macOS e Windows, amd64 e arm64; confira com `checksums.txt`), ou instale com Go 1.24+:

```sh
go install github.com/edusouza/ctbz-cli/cmd/ctbz@latest
# ou, a partir do clone:
go build -o ctbz ./cmd/ctbz
```

## Configuração

| Variável           | Uso                                                                       |
|--------------------|---------------------------------------------------------------------------|
| `CTBZ_USER`        | e-mail ou CPF do login                                                    |
| `CTBZ_PASSWORD`    | senha                                                                     |
| `CTBZ_OTP_CMD`     | comando que imprime o código OTP (habilita login e re-login automáticos)  |
| `CTBZ_OTP_TIMEOUT` | quanto esperar pelo `CTBZ_OTP_CMD` (padrão `3m`)                          |
| `CTBZ_CNPJ`        | empresa a selecionar quando o usuário tem mais de uma                     |
| `CTBZ_HOME`        | onde guardar a sessão (padrão `~/.config/ctbz`)                           |
| `CTBZ_OUTPUT`      | formato de saída padrão: `table` (padrão), `json` ou `csv`                |
| `CTBZ_VERBOSE=1`   | mostra as tentativas do `CTBZ_OTP_CMD`                                     |

A sessão (cookies `__C` e `oauth-token`) fica em `$CTBZ_HOME/session.json` com
permissão `0600`. As credenciais nunca são gravadas em disco. As escritas enviadas ficam
registradas, sem corpo, em `$CTBZ_HOME/acoes.jsonl` (veja `ctbz acoes`).

## Login

A Contabilizei sempre envia um código de 6 dígitos por e-mail. Há três formas de entregá-lo à CLI:

### 1. Automático, via comando (`--otp-cmd` / `CTBZ_OTP_CMD`)

A CLI executa o comando a cada 5 s até ele imprimir um código de 6 dígitos (ou até
`CTBZ_OTP_TIMEOUT`). O comando recebe `CTBZ_OTP_SINCE` (epoch, em segundos) e
`CTBZ_OTP_SINCE_ISO` com o instante do pedido, para ignorar e-mails antigos.

O repositório traz [`scripts/otp-gmail-gws.sh`](scripts/otp-gmail-gws.sh), que
lê o Gmail com o [Google Workspace CLI](https://github.com/googleworkspace/cli) (`gws`) e `jq`:

```sh
export CTBZ_OTP_CMD="$PWD/scripts/otp-gmail-gws.sh"
# Se você encaminha o e-mail para outra caixa, ajuste a busca:
# export CTBZ_OTP_QUERY='subject:"código" to:bot@seudominio.com'
ctbz login --cnpj 00.000.000/0001-00
```

Qualquer comando serve: um script IMAP, uma chamada a um webhook etc. Basta imprimir o código.

Com `CTBZ_OTP_CMD` definido, `ctbz empresa` e `ctbz api` também **refazem o login
sozinhos** quando a sessão expira (HTTP 401) e repetem a chamada. Escritas (incluindo
`ctbz api -X POST`) nunca são repetidas: a sessão é conferida, e refeita se preciso, antes do
envio ([ADR-0019](docs/adr/0019-camada-de-escrita.md)).

### 2. Interativo

Sem `CTBZ_OTP_CMD` e rodando num terminal, a CLI pergunta o código (e a empresa, se houver mais de uma).

### 3. Em duas etapas (scripts, agentes, CI)

Sem terminal, o login fica salvo como pendente e o processo sai com status `3`:

```sh
$ ctbz login
Código de verificação enviado para fu***@gm***com.
Conclua com: ctbz login --otp NNNNNN
$ ctbz login --otp 123456 --cnpj 00000000000100
Login concluído: FULANO LTDA (CNPJ 00000000000100) — usuário fulano@exemplo.com
```

`ctbz login --restart` descarta um login pendente e pede um código novo. Atenção:
cada `ctbz login` gera um e-mail novo, e só o código mais recente vale.

## Comandos

```sh
ctbz resumo              # o que precisa de atenção (impostos, pendências, rotinas, mensalidade)
ctbz status              # dados da sessão e se ela ainda é válida
ctbz empresa             # resumo da empresa selecionada
ctbz empresas            # empresas do usuário, marcando a atual
ctbz empresa usar CNPJ   # troca de empresa (refaz o login: novo OTP)
ctbz empresa certificado # validade do certificado digital
ctbz empresa socios      # sócios e seus papéis
ctbz empresa atividades  # CNAEs e anexos do Simples Nacional
ctbz impostos            # guias a pagar (em atraso, do mês e do próximo mês)
ctbz impostos guia ID    # detalhe de uma guia
ctbz impostos calculo    # memória de cálculo do mês (DAS, INSS, IRRF, Fator R)
ctbz impostos baixar --pendentes -d ~/guias   # PDFs das guias a pagar
ctbz impostos confirmar ID [--nao-paguei]      # informa que a guia foi (ou não) paga; desmarcar desfaz
ctbz impostos recalcular ID [--vencimento D]   # dados do recálculo ou pede nova guia (serviço cobrado)
ctbz impostos parcelamento simular TIPO        # opções de parcelamento de débitos
ctbz impostos parcelamento contratar TIPO      # contrata (risco alto: confissão de dívida)
ctbz impostos historico --ano 2026             # guias de meses anteriores
ctbz impostos faturamento                      # faturamento, pró-labore e Fator R (12 meses)
ctbz impostos parcelamentos                    # parcelamentos de impostos
ctbz impostos debitos                          # débitos federais em aberto?
ctbz impostos recorrente [historico]           # pagamento recorrente (débito automático) de impostos
ctbz pendencias                                # pendências abertas, com alerta de prazo
ctbz pendencias conciliacao                    # notas e recebimentos a conciliar
ctbz pendencias termos | termo CHAVE           # termos e cartas para aceitar; texto completo
ctbz rotinas --mes 2026-10                     # rotinas e obrigações do mês (empresa e Contabilizei)
ctbz rotinas reclassificar ID [--classificacao ID]  # opções ou reclassifica e conclui a rotina
ctbz chamados [--finalizados]                  # chamados de atendimento
ctbz mensalidade                               # mensalidade atual da Contabilizei
ctbz mensalidade situacao                      # a empresa está em dia com a Contabilizei?
ctbz mensalidade historico                     # pagamentos anteriores e débito automático
ctbz plano [contrato|proposta] [--texto]       # plano contratado e contrato de serviço
ctbz notas --de 2026-01 --ate 2026-09          # NFS-e emitidas no período, com total
ctbz notas tomadores [consulta CNPJ]           # clientes do emissor; cadastro de um CNPJ
ctbz notas config | ctbz notas aliquotas       # emissor e alíquotas por atividade
ctbz notas entrada [--lista manifestadas]      # NF-e recebidas (notas de entrada)
ctbz prolabore [historico --ano 2026]          # pró-labore por sócio e histórico mensal
ctbz prolabore parametros                      # salário mínimo, INSS e IRRF usados no cálculo
ctbz prolabore fator-r                         # Fator R, motor do Fator R e anexos por atividade
ctbz lucros                                    # distribuição de lucros do exercício
ctbz lucros informe --ano 2025                 # valores do informe de rendimentos dos sócios (IR)
ctbz balancete 2026-08                         # balancete do mês (árvore de contas)
ctbz balanco 2025                              # balanço patrimonial do exercício
ctbz razao --conta 1.01 --de 2026-01           # lançamentos do razão por conta
ctbz caixa 2026-09                             # lançamentos do caixa do mês, com total
ctbz caixa contas --competencia 2026-09        # classificações aceitas no caixa (e vínculos)
ctbz caixa adicionar --competencia 2026-09 --pagamento --valor 150,25 --conta ID --descricao "…"
ctbz caixa editar ID --competencia 2026-09 --valor 900   # altera só o que for informado
ctbz caixa remover ID --competencia 2026-09              # exclui (mostra como recriar)
ctbz extratos contas --competencia 2026-09     # classificações aceitas no extrato
ctbz extratos lancamentos --conta-bancaria ID --competencia 2026-09   # lançamentos do extrato
ctbz extratos classificar ID --conta-bancaria ID --competencia 2026-09 --conta ID
ctbz extratos desmembrar ID ... --parte "Aluguel:1000:ID" --parte "Condomínio:200:ID"
ctbz extratos desfazer-desmembramento ID --conta-bancaria ID --competencia 2026-09
ctbz extratos --ano 2026 | ctbz contas-bancarias  # extratos por mês e contas cadastradas
ctbz contas --busca alug                       # plano de contas usado nas classificações
ctbz documentos [--tipo TIPO]                  # central de documentos: tipos e arquivos enviados
ctbz certificado                               # certificado digital: validade e renovação
ctbz acoes --desde 2026-10-01                  # escritas enviadas por esta CLI (registro local)
ctbz api menu/get        # qualquer endpoint; caminho relativo → /api/plataforma/
ctbz api -X POST -d @corpo.json /api/plataforma/algum/endpoint
ctbz logout              # apaga a sessão local
ctbz version             # versão do binário
```

Referência completa de cada comando: [`docs/referencia`](docs/referencia/README.md)
(também em `ctbz COMANDO --help`). Completion de shell: `source <(ctbz completion bash)`.

## Ações que alteram dados

Comandos que alteram dados na Contabilizei (a partir da 1.2) mostram um resumo e pedem
confirmação (`confirmo` nas ações de risco alto). Em scripts, use `--yes`; sem terminal e sem
`--yes`, nada é enviado (código `2`). `--dry-run` mostra a requisição sem enviar. Detalhes em
[docs/guia/escrita.md](docs/guia/escrita.md).

## Formatos de saída

Todo comando que imprime dados aceita `-o`/`--output` (antes ou depois do comando) ou a
variável `CTBZ_OUTPUT`:

| Formato | Para | Valores em reais | Datas | CNPJ |
|---|---|---|---|---|
| `table` (padrão) | pessoas | `R$ 1.234,56` | `20/10/2026` | `00.000.000/0000-00` |
| `json` | scripts (`jq`) | `1234.56` | `2026-10-20` (horários em RFC 3339) | só dígitos |
| `csv` | planilhas | `1234.56` | `2026-10-20` | só dígitos |

```sh
ctbz empresa -o json | jq -r .certificado_validade
ctbz -o csv status > sessao.csv
export CTBZ_OUTPUT=json        # muda o padrão
```

- **stdout recebe só dados**; mensagens, progresso e erros vão para stderr.
- Em JSON, os campos saem na ordem da tabela e campos sem valor saem como `null`.
- Em CSV, a primeira linha traz os nomes dos campos (os mesmos do JSON). Valores aninhados
  (ex.: `outras_empresas`) saem como JSON dentro da célula, e listas simples separadas por `; `.
- `ctbz empresa --json` é atalho para `-o json`. A resposta crua da API continua disponível em
  `ctbz api dadosempresa/get`.
- `ctbz api` sai em JSON por padrão e ignora `CTBZ_OUTPUT`. Com `-o table` ou `-o csv`, listas
  de objetos viram tabelas. `--raw` imprime o corpo exatamente como veio.

Códigos de saída: `0` sucesso, `1` erro (inclusive confirmação de escrita recusada), `2` uso
incorreto (comando, flag ou argumento inválido; escrita sem terminal e sem `--yes`),
`3` login pendente (aguardando OTP ou CNPJ), `4` atenção (ex.: impostos em atraso com
`--fail-on-atraso`, pendências vencidas com `--fail-on-vencidas`, algo vencido ou crítico com
`ctbz resumo --fail-on-atencao`).

Guia de uso por contexto: [`docs/guia`](docs/guia/README.md).

A partir da 1.0, comandos, flags, chaves de JSON/CSV e códigos de saída só mudam de forma
incompatível numa versão *major* ([ADR-0017](docs/adr/0017-contrato-publico-da-1-0.md)).

## Como funciona (engenharia reversa)

> Documentação detalhada, separada por contexto, em [`docs/`](docs/README.md):
> autenticação, OTP automático, API (endpoints verificados e catálogo), front-end, CLI e metodologia.

1. `GET /login` → formulário com um `token` anti-CSRF.
2. `POST /login` (`user`, `password`, `token`) → tela de OTP e cookie de pré-sessão `__C`.
   Erros voltam como redirecionamento para `/login#incorreto`, `#bloqueado`, `#limite-sessoes`…
3. `POST /login` com corpo `v=2&otp=NNNNNN` (text/plain, como o `fetch` do navegador):
   `206` = próxima tela (seleção de empresa), `200` = login concluído, `404` = código inválido.
4. `POST /selecionarempresa` (`cnpj`, `token`) → cookie `oauth-token` e um `<script>` que grava
   no `localStorage` as chaves `l` (login), `r` (usuário) e `e` (empresa), em base64 de JSON
   passado por `encodeURIComponent`. A CLI decodifica e guarda esses dados na sessão.
5. O painel (`/painel-de-controle/`) chama as APIs com os cookies (`withCredentials`).
   Bases encontradas no bundle do front:

   | Base                    | Uso                                   |
   |-------------------------|---------------------------------------|
   | `/api/plataforma/`      | BFF do painel (padrão do `ctbz api`)  |
   | `/api/legado/`          | monólito                              |
   | `/api/public/`          | endpoints públicos                    |
   | `/api/multiusuario/`    | usuários e convites                   |
   | `/api/pagamentos/`      | pagamentos                            |
   | `/api/fintech/`         | conta digital                         |

   Endpoints `GET` do BFF já testados: `dadosempresa/get`, `menu/get`, `appbar/get`,
   `dashboard/situacao-app`, `atendimento/chamados?em-andamento=true`.
   Outros vistos no front (não testados): `impostos/v5/impostos-a-pagar/guia/{id}`,
   `relatorios-ms/gerar-balancete/`, `prolabore/central/init`, `conciliacao-fiscal/pendencias`,
   `certificado/status`, `notafiscal/buscarNotaEmpresa`…

As sessões duram algumas horas (o cookie `__C` expira cerca de 2 h após o login).

## Desenvolvimento

```sh
go test ./...
```

Os testes simulam o fluxo de login com um servidor local, sem tocar na Contabilizei.
