# Skill `assessor`

A base de todas as outras skills do plugin `ctbz`. Ensina o agente a usar a CLI `ctbz` como o
assessor dedicado do plano Experts: conferir a sessão, fazer o login sem terminal, ler os dados
em JSON, seguir o protocolo obrigatório para qualquer ação que altera dados, manter o dossiê da
empresa e encaminhar cada pedido para a skill certa.

Você quase nunca a chama diretamente: o agente a carrega em qualquer pedido sobre a empresa
na Contabilizei, antes das outras skills.

Instalação e configuração do plugin: [README das skills](../README.md).

## Quando entra em ação

Em qualquer conversa sobre a empresa na Contabilizei: impostos, guias, pendências, rotinas,
pró-labore, Fator R, lucros, notas, extratos, caixa, documentos, certificado, mensalidade,
acessos. Pedidos que ela mesma resolve:

```text
> Entra na Contabilizei.
> Troca para a outra empresa.
> O que você já sabe sobre a minha empresa?
> O que você alterou na Contabilizei esta semana?
```

Para chamá-la pelo nome: `/ctbz:assessor`.

## O que ela faz

### 1. Começo de toda sessão

1. `ctbz version`: confere se a CLI existe (e mostra como instalar, se não existir). A skill foi
   escrita para a 1.8.
2. `ctbz status -o json`: sessão `válida` → segue; `expirada` ou "nenhuma sessão salva" → login.
3. Lê o dossiê da empresa; se não existir, cria.

### 2. Login sem terminal

O agente não tem um terminal interativo, então o login acontece em duas etapas:

```sh
ctbz login                     # envia o e-mail com o código; sai com código 3 (pendente)
ctbz login --otp 123456        # conclui com o código que você ditou
ctbz login --cnpj 00000000000100   # se você tiver mais de uma empresa
```

- Usuário e senha vêm de `CTBZ_USER` e `CTBZ_PASSWORD`. **O agente nunca pede a senha no chat**:
  se as variáveis não existirem, ele pede que você as exporte (ou rode `ctbz login` no seu
  terminal).
- Com `CTBZ_OTP_CMD`, a CLI busca o código sozinha e o agente não pergunta nada.
- O agente pede o código **uma vez** e não roda `ctbz login` de novo enquanto espera, porque
  cada login envia um e-mail novo e invalida o código anterior.
- `ctbz empresa usar CNPJ` troca de empresa (pede um código novo).

### 3. Leituras

Sempre com `-o json`, em paralelo quando são independentes, e com um resumo para você em vez do
JSON cru. O mapa de todos os comandos e flags fica em
[`references/comandos.md`](references/comandos.md) e os campos do JSON de cada um, em
[`references/campos-json.md`](references/campos-json.md): o agente não precisa explorar a CLI
com `--help` a cada conversa.

### 4. Protocolo de escrita

Toda ação que altera dados passa por: preparar → `--dry-run` → apresentar (o que muda, risco,
consequência, se dá para desfazer, custo) → sua confirmação → `--yes` → conferir → registrar
no dossiê. Nas ações de **risco alto**, você precisa confirmar por escrito aquela ação
específica, depois de ver a consequência.

Regras que valem para todas as skills:

- uma escrita que falhou por conexão nunca é repetida antes de conferir o estado;
- várias escritas de risco alto nunca vão num único "ok";
- segredos (senha, credenciais de órgãos públicos) são digitados por você, num terminal à parte;
- `ctbz api -X …` (escrita crua) não é usado por iniciativa do agente.

O risco de cada comando e o que não dá para desfazer estão em
[`references/escrita.md`](references/escrita.md).

### 5. Dossiê da empresa

Um arquivo por empresa em `${CTBZ_HOME:-~/.config/ctbz}/dossie/<CNPJ>.md` (permissão `0600`),
com perfil, IDs usados nos comandos (sócios, contas bancárias), decisões, histórico de checkups
e escritas feitas. Nunca guarda senhas, códigos, tokens, credenciais nem dados de cartão.
Modelo e regras em [`references/dossie.md`](references/dossie.md).

### 6. Encaminhamento

| Você quer… | Skill |
|---|---|
| saber como está a empresa, o que vence, as pendências | [`checkup`](../checkup/README.md) |
| fechar o mês: extratos, caixa, conciliação, notas de entrada, documentos | [`fechamento-mensal`](../fechamento-mensal/README.md) |
| guias, DAS/DARF, pagar, atraso, recálculo, parcelamento | [`impostos`](../impostos/README.md) |
| pagar menos imposto, Fator R, pró-labore ideal, limites do Simples | [`planejamento-tributario`](../planejamento-tributario/README.md) |
| lucros, balanço do ano, informe de rendimentos para o IRPF | [`fechamento-anual`](../fechamento-anual/README.md) |
| notas emitidas, faturamento por cliente, tomadores, emissor | [`notas-e-clientes`](../notas-e-clientes/README.md) |
| documentos, certificado, usuários, credenciais, e-mail/senha, mensalidade, plano | [`cadastro-e-acessos`](../cadastro-e-acessos/README.md) |

## Postura

O agente fala português, com números e datas concretos ("DAS de 09/2026, R$ 1.000,00, vence em
20/10/2026"), explica cada termo na primeira vez (glossário em
[`references/glossario.md`](references/glossario.md)), separa o que a empresa precisa fazer do
que a Contabilizei entrega e nunca inventa valores, prazos ou regras: se a CLI não mostra, ele
diz isso. Em dúvidas jurídicas ou casos fora do padrão, indica confirmar com o contador.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| [`SKILL.md`](SKILL.md) | instruções que o agente segue |
| [`references/comandos.md`](references/comandos.md) | mapa de comandos e flags (gerado de `docs/referencia`) |
| [`references/campos-json.md`](references/campos-json.md) | campos do JSON de cada comando |
| [`references/escrita.md`](references/escrita.md) | risco de cada escrita e como desfazer |
| [`references/dossie.md`](references/dossie.md) | onde fica, o que guarda e modelo do dossiê |
| [`references/glossario.md`](references/glossario.md) | termos contábeis e fiscais em linguagem simples |

## Fora do alcance

Emitir ou cancelar NFS-e, pagar guias, ativar débito automático, abrir chamado, folha de
pagamento de funcionários e troca de plano: o agente indica o painel, o app ou o atendimento da
Contabilizei.
