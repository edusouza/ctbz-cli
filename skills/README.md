# Skills do ctbz para o Claude Code

Este diretório é o plugin `ctbz` do [Claude Code](https://code.claude.com): oito skills que
fazem o agente operar a Contabilizei pela CLI `ctbz` como um assessor contábil. Você pergunta
em português ("como está minha empresa?", "fecha o mês de setembro", "dá para pagar menos
imposto?") e o agente lê os dados pela CLI, explica, propõe o que fazer e só altera algo na
Contabilizei depois do seu "ok".

As skills simulam o atendimento do plano Experts (assessor e analista dedicados): checkup do
CNPJ, fechamento do mês, impostos, consultoria tributária, fechamento do ano, notas e clientes,
cadastro e acessos.

## Pré-requisitos

| O quê | Para quê | Como conferir |
|---|---|---|
| Claude Code | rodar o agente e o plugin | `claude --version` |
| `ctbz` 1.8 ou mais novo, no `PATH` | todas as skills | `ctbz version` |
| Conta na Contabilizei (e-mail ou CPF e senha) | login | — |
| Acesso ao e-mail da conta | o código de 6 dígitos de cada login | — |
| `bash` | script de coleta do `checkup` (Linux, macOS, WSL ou Git Bash) | `bash --version` |
| `python3` (só a biblioteca padrão) | simulador do `planejamento-tributario` | `python3 --version` |
| `jq` (opcional) | filtros que o agente usa em algumas análises | `jq --version` |

Instalar o `ctbz`: baixe o binário do seu sistema na
[página de releases](https://github.com/edusouza/ctbz-cli/releases) e coloque-o no `PATH`, ou,
com Go 1.24+:

```sh
go install github.com/edusouza/ctbz-cli/cmd/ctbz@latest
```

Se o `ctbz` não existir, a skill `assessor` mostra ao agente como instalá-lo, mas é melhor
deixar pronto antes.

## Instalar

### Pelo marketplace (recomendado)

Dentro do Claude Code:

```text
/plugin marketplace add edusouza/ctbz-cli
/plugin install ctbz@ctbz-cli
```

A partir do Claude Code 2.1.275, também em um passo:

```text
/plugin install ctbz --marketplace edusouza/ctbz-cli
```

Confira em `/plugin` (aba de plugins instalados) que o `ctbz` aparece habilitado. As skills
ficam com o prefixo do plugin: `ctbz:assessor`, `ctbz:checkup` etc.

### A partir de um clone (desenvolvimento ou teste)

```sh
git clone https://github.com/edusouza/ctbz-cli
claude --plugin-dir /caminho/do/ctbz-cli
```

Assim o Claude Code carrega as skills direto do diretório, sem o marketplace. Útil para testar
mudanças nas skills antes de publicá-las.

### Atualizar e remover

```text
/plugin marketplace update ctbz-cli     # busca a versão nova do marketplace
/plugin uninstall ctbz@ctbz-cli         # remove o plugin
```

A versão do plugin (`.claude-plugin/plugin.json`) acompanha a da CLI. Atualize os dois juntos:
uma skill nova pode usar um comando que só existe no `ctbz` mais novo.

## Configurar o login

O login da Contabilizei é usuário e senha → código de 6 dígitos por e-mail → escolha da
empresa. O agente **nunca pede a senha no chat**: ela vem do ambiente. Exporte as variáveis
**no shell onde você abre o Claude Code**, antes de abri-lo (o agente roda o `ctbz` com o
ambiente do processo `claude`):

```sh
export CTBZ_USER="voce@exemplo.com"        # e-mail ou CPF do login
export CTBZ_PASSWORD="…"                   # use um gerenciador de senhas ou um arquivo fora do repositório
export CTBZ_CNPJ="00000000000100"          # opcional: empresa, se o usuário tiver mais de uma
claude
```

O código de 6 dígitos pode chegar de duas formas:

1. **Você dita ao agente.** Ele roda `ctbz login`, a Contabilizei envia o e-mail e o agente pede
   o código uma única vez. Cada `ctbz login` gera um e-mail novo e só o último código vale.
2. **Automático, com `CTBZ_OTP_CMD`.** Um comando que imprime o código (por exemplo
   [`scripts/otp-gmail-gws.sh`](../scripts/otp-gmail-gws.sh), que lê o Gmail). Com ele, o
   agente faz o login sem perguntar nada e a CLI refaz o login sozinha quando a sessão expira.
   Detalhes na seção "Login" do [README da CLI](../README.md#login).

Também dá para fazer o login você mesmo, num terminal, antes de abrir o Claude Code
(`ctbz login`): a sessão fica salva em `~/.config/ctbz/session.json` (ou `$CTBZ_HOME`) e dura
algumas horas. O agente começa toda conversa conferindo-a com `ctbz status`.

## Primeiro uso

Abra o Claude Code em qualquer diretório e peça, por exemplo:

```text
> Faz um checkup da minha empresa na Contabilizei.
```

O que acontece:

1. A skill `assessor` confere o `ctbz` e a sessão. Se não houver sessão, o agente faz o login e
   pede o código que chegou no seu e-mail.
2. Na primeira vez, monta o **dossiê** da empresa (perfil, sócios, contas bancárias e seus IDs)
   e o grava em `~/.config/ctbz/dossie/<CNPJ>.md`.
3. A skill `checkup` lê tudo em paralelo (só leituras) e entrega um relatório com o que é
   urgente, o que vence nesta semana, neste mês e as oportunidades.
4. No fim, oferece os próximos passos, cada um ligado à skill que o executa.

O Claude Code pede permissão antes de cada comando de shell. Leia o comando antes de aprovar:
as escritas sempre aparecem primeiro com `--dry-run` e só depois, com a sua confirmação, com
`--yes`.

## As skills

O agente escolhe a skill pelo pedido. Também dá para chamar uma pelo nome, como
`/ctbz:checkup` ou `/ctbz:impostos`.

| Skill | Papel | Exemplos de pedido | O que pode alterar |
|---|---|---|---|
| [`assessor`](assessor/README.md) | base de todas | "entra na Contabilizei", "troca de empresa" | nada sozinha; define o protocolo de escrita |
| [`checkup`](checkup/README.md) | reunião mensal | "como está minha empresa?", "o que vence?" | nada (só lê) |
| [`fechamento-mensal`](fechamento-mensal/README.md) | analista | "fecha setembro", "importa meu extrato", "classifica os lançamentos" | extratos, classificações, caixa, conciliação, notas de entrada, documentos |
| [`impostos`](impostos/README.md) | assessor | "o que tenho para pagar?", "baixa o DAS", "paguei a guia" | status de pagamento, recálculo, parcelamento |
| [`planejamento-tributario`](planejamento-tributario/README.md) | consultoria | "dá para pagar menos imposto?", "como fica o Fator R?" | pró-labore e gestão inteligente |
| [`fechamento-anual`](fechamento-anual/README.md) | assessor | "distribui os lucros de 2025", "informe de rendimentos para o IR" | distribuição de lucros, aceites, reabertura do balanço |
| [`notas-e-clientes`](notas-e-clientes/README.md) | assessor | "quanto faturei por cliente?", "cadastra este CNPJ" | cadastro de tomadores |
| [`cadastro-e-acessos`](cadastro-e-acessos/README.md) | atendimento | "quando vence meu certificado?", "convida meu sócio" | contas bancárias, usuários, procuração, escritório virtual; credenciais e login só por você, no terminal |

Cada link leva ao README da skill: quando usar, o que ela lê, o que pode alterar (com o risco de
cada ação), exemplos e limites.

## Segurança: o que o agente pode alterar

Toda ação que altera dados na Contabilizei segue o mesmo protocolo, definido na skill
`assessor`:

1. **Prepara:** lê o estado atual e os IDs.
2. **Simula** com `--dry-run`: nada é enviado.
3. **Apresenta:** o que muda, o risco (baixo, médio ou alto), a consequência, se dá para desfazer
   e o custo.
4. **Pede a sua confirmação:**
   - risco baixo ou médio: um "sim" ou "ok" claro. Um único "ok" pode cobrir uma lista de ações
     iguais, se a lista inteira tiver sido mostrada (ex.: classificar 12 lançamentos);
   - risco **alto** (efeito fiscal, legal ou financeiro, em geral sem desfazer): você confirma
     por escrito **aquela** ação, depois de ver a consequência. Um "ok" genérico não vale.
5. **Envia** com `--yes`, uma única vez (a CLI nunca repete uma escrita).
6. **Confere** o resultado e relê o estado.
7. **Registra** no dossiê, com o motivo.

Outras garantias:

- Senhas, a senha da prefeitura, o código do Simples e o da Dataprev **nunca passam pelo chat**:
  os comandos que pedem segredos (`ctbz conta senha`, `ctbz empresa credenciais atualizar`) são
  digitados por você, num terminal à parte.
- O agente não usa `ctbz api -X …` (escrita crua, sem confirmação) por iniciativa própria.
- `ctbz acoes` lista tudo o que a CLI enviou (sem os corpos), com data, comando e resultado:
  pergunte "o que você alterou na Contabilizei esta semana?".
- Permissões do Claude Code: as leituras e as escritas usam o mesmo binário. Se você liberar
  `Bash(ctbz:*)` nas permissões, o Claude Code deixa de pedir aprovação também para as escritas
  com `--yes`; o protocolo de confirmação continua valendo, mas você perde a segunda barreira.

## O dossiê da empresa

É a memória do assessor entre conversas: um arquivo Markdown por empresa em
`${CTBZ_HOME:-~/.config/ctbz}/dossie/<CNPJ>.md`, com permissão `0600`. Guarda:

- **Perfil:** regime, atividades e anexos, sócios e seus IDs, contas bancárias e seus IDs, plano,
  certificado.
- **Decisões e preferências:** "pró-labore na gestão inteligente desde 09/2026", "guias em
  ~/Documentos/guias".
- **Histórico de checkups** e **escritas feitas pelo agente**, com o motivo.

Nunca guarda senhas, códigos, tokens, credenciais de órgãos públicos, dados de cartão nem CPF
completo. Você pode ler, editar ou apagar o arquivo a qualquer momento; o agente o recria na
próxima conversa.

## Limites

As skills não fazem o que a CLI não faz. Continuam no painel, no app ou com o atendimento da
Contabilizei:

- emitir, cancelar ou agendar NFS-e (a skill `notas-e-clientes` prepara a ficha para emitir);
- pagar guias (a CLI baixa o PDF e registra que você pagou);
- ativar o débito automático de impostos ou cadastrar cartão;
- abrir chamado (a CLI só lista);
- folha de pagamento de funcionários e troca de plano;
- comprar ou renovar o certificado digital.

As análises (Fator R, distribuição de lucros, limites do Simples) são **estimativas de apoio** e
não substituem o contador responsável. Em caso de dúvida jurídica ou fora do padrão, o agente
diz que é preciso confirmar com a Contabilizei.

## Solução de problemas

| Sintoma | Causa e o que fazer |
|---|---|
| "ctbz: command not found" | O `ctbz` não está no `PATH` do processo `claude`. Instale-o e reabra o Claude Code. |
| O agente pede o código e o e-mail não chega | Confira o spam. Não peça para o agente rodar `ctbz login` de novo enquanto espera: cada login invalida o código anterior. Para recomeçar do zero: `ctbz login --restart`. |
| Muitos e-mails de código | A sessão expirou várias vezes, ou o login foi repetido. Configure `CTBZ_OTP_CMD`, ou faça o login no terminal antes de abrir o Claude Code. |
| "nenhuma sessão salva" / código 3 | Sem login ou login pendente. O agente conclui com `ctbz login --otp NNNNNN`. |
| O agente pede `CTBZ_USER`/`CTBZ_PASSWORD` | As variáveis não estavam no ambiente quando o Claude Code abriu. Exporte-as e reabra. |
| "HTTP 206" em `rotinas`, `resumo`, `documentos pendentes` (CLI até a 1.8.0) | Falha conhecida, corrigida na 1.8.1. Atualize o `ctbz`; até lá, o `checkup` lê a Central de Rotinas por `ctbz api`. |
| "Ano inválido" em `impostos faturamento` (CLI até a 1.8.0) | Empresa com menos de um ano. Corrigido na 1.8.1; as skills usam os números de `ctbz impostos calculo` como alternativa. |
| "cannot unmarshal array" em `extratos lancamentos`, `classificar` ou `desmembrar` (CLI até a 1.8.0) | A Contabilizei devolve os lançamentos num formato que a CLI antiga não lia. Corrigido na 1.8.1: atualize o `ctbz`. |
| `mensalidade historico` com `pagamentos` vazio (CLI até a 1.8.0) | A CLI antiga lia o histórico do débito automático de impostos. Corrigido na 1.8.1: atualize o `ctbz`. |
| "unknown command" ou "unknown flag" | O `ctbz` é mais velho que o plugin. Atualize a CLI. |
| "convertido pelo Git Bash do Windows" em `ctbz api` | O Git Bash troca um argumento que começa com `/` por um caminho do Windows. Escreva o caminho com `//` (`ctbz api //api/legado/…`) ou use `MSYS_NO_PATHCONV=1`. |
| O `checkup` falha no Windows | O script de coleta precisa de `bash`. Use o WSL ou o Git Bash, ou deixe o agente rodar as leituras uma a uma. |

## Para quem mantém

Estrutura de cada skill:

```text
skills/<nome>/
├── SKILL.md          # o que o agente lê: descrição (gatilho) e instruções
├── README.md         # este tipo de página, para pessoas
├── references/       # material que o agente consulta sob demanda
└── scripts/          # scripts que o agente executa
```

- `skills/assessor/references/comandos.md` é **gerado** a partir de `docs/referencia`. Depois de
  mudar a CLI:
  ```sh
  go run ./tools/gendocs
  python3 scripts/gerar-comandos-skill.py
  ```
- `skills/assessor/references/campos-json.md` e os passos de cada `SKILL.md` são escritos à mão:
  atualize-os junto com o comando que mudou, e o README da skill quando o comportamento visível
  mudar.
- Teste as mudanças com `claude --plugin-dir .` a partir do clone.
- Mais contexto: [Skills para agentes](../docs/guia/agentes.md) e
  [Contribuindo](../docs/contribuindo.md).
