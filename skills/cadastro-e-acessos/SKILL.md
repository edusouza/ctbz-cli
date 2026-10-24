---
name: cadastro-e-acessos
description: Cadastro, acessos e relação com a Contabilizei pela CLI ctbz. Use para dados da empresa e dos sócios, troca de empresa, certificado digital (validade e renovação), central de documentos, contas bancárias, usuários com acesso (convidar, desativar), credenciais de órgãos públicos (código do Simples, prefeitura, Dataprev), procuração do e-CAC, e-mail/telefone/senha de login, escritório virtual (correspondências), mensalidade, plano e contrato, chamados de atendimento e primeiros passos.
---

# Cadastro, acessos e relação com a Contabilizei

Antes, siga a skill `assessor` (sessão, dossiê e protocolo de escrita). Aqui ficam as
operações de maior impacto em **segurança**: não peça nem mostre senhas no chat.

## Empresa e sócios

```sh
ctbz empresa -o json              # regime, ramos, plano, certificado, inscrições
ctbz empresa socios -o json
ctbz empresa atividades -o json   # CNAEs e anexos do Simples
ctbz empresas -o json             # empresas do usuário (atual: true)
ctbz empresa usar CNPJ            # troca de empresa: refaz o login (novo código por e-mail)
```

Ao trocar de empresa, troque também de dossiê.

## Certificado digital (e-CNPJ)

`ctbz certificado -o json`: `valido`, `vencimento`, `dias_para_vencer`, `apto_renovacao`,
`renovacao_etapa`. Sem certificado válido não há emissão de notas nem entrega de obrigações.
Com 60 dias ou menos, avise; com `apto_renovacao: true`, oriente a renovar pelo painel (a CLI
não compra nem renova). Etapas como `agendamentoVideoconferencia` indicam renovação em
andamento.

## Central de documentos

```sh
ctbz documentos -o json                          # tipos aceitos e quantos foram enviados
ctbz documentos --tipo TIPO -o json              # arquivos enviados (com link)
ctbz documentos pendentes -o json                # o que as pendências pedem
```

Envio e declarações de inexistência: skill `fechamento-mensal`, seção 6.

## Contas bancárias

```sh
ctbz contas-bancarias -o json
ctbz contas-bancarias bancos -o json            # id e código dos bancos aceitos
ctbz contas-bancarias adicionar --banco 237 --agencia 1234 --conta 12345-6 \
  --abertura 2024-01-10 --declaro-conta-pj [--saldo-inicial 0,00] --dry-run
ctbz contas-bancarias editar ID --abertura 2024-02-01 --dry-run
ctbz contas-bancarias remover ID --dry-run      # risco alto: permanente
```

`--declaro-conta-pj` afirma que a conta é PJ e movimenta só este CNPJ: confirme com o usuário.

Confira o banco no resumo do `--dry-run` (nome e código) antes de confirmar: há códigos
repetidos e pares nacional/internacional (`077` Banco Inter S.A. e `077-1` Banco Inter -
Internacional). Se a CLI disser que o código corresponde a mais de um banco, pergunte qual é e
passe o **id** em `--banco`.
A data de abertura define desde quando a Contabilizei pede extratos. Se `editar` recusar
dizendo que a Contabilizei não informou a abertura ou o saldo inicial da conta, pergunte o
valor ao usuário e passe `--abertura` ou `--saldo-inicial`; não invente. Atualize o dossiê.

## Usuários com acesso

```sh
ctbz usuarios -o json                 # id, nome, email, tipo, status (CONVITE_ENVIADO…)
ctbz usuarios convidar EMAIL          # risco alto: acesso TOTAL, inclusive financeiro
ctbz usuarios desativar ID            # médio: única forma de revogar
ctbz usuarios ativar ID               # médio
```

Antes de convidar, deixe claro que não há perfil restrito e que o convite não pode ser
revogado antes do aceite (só desativando depois). Só usuários secundários podem ser
desativados (o administrador, nunca); usuário sem id, tipo ou status na lista é recusado
sem escrita: mande o usuário ao painel.

## Credenciais de órgãos públicos

`ctbz empresa credenciais -o json` mostra só se cada credencial está `preenchido` (código do
Simples, prefeitura, Dataprev). **Nunca use `--mostrar`** num contexto de agente, porque os
valores ficariam na conversa.

Para trocar (`ctbz empresa credenciais atualizar --prefeitura|--codigo-simples|--dataprev`),
os valores são lidos sem eco. Peça ao usuário que rode **num terminal à parte**, fora da
conversa, para que os valores não passem pelo agente. Risco alto: credencial errada faz falhar
a entrega de declarações. Depois, a Contabilizei verifica o acesso e avisa por e-mail.

O alerta "Redefina a senha da prefeitura", quando o usuário já atualizou:
`ctbz empresa credenciais confirmar-prefeitura` (médio).

## Procuração do e-CAC

`ctbz pendencias procuracao -o json` → `pendente`, `cnpj_outorgado`, `link_ecac`. A procuração
é feita pelo usuário no e-CAC (gov.br), outorgando ao `cnpj_outorgado`. Depois de feita:
`ctbz pendencias procuracao --ja-criei` (médio; a pendência volta se ela não existir).

## Login: e-mail, telefone e senha

```sh
ctbz conta -o json    # e-mail e telefone mascarados, método do 2FA
```

`ctbz conta alterar --email … --telefone …` (os dois juntos; repita o que não muda) e
`ctbz conta senha` são de risco alto e exigem um código OTP; a senha é digitada sem eco. Peça
ao usuário que rode no próprio terminal. Depois: atualizar `CTBZ_USER`/`CTBZ_PASSWORD`; outras
sessões podem cair.

## Escritório virtual (correspondências)

```sh
ctbz correspondencias -o json                 # data, remetente, descricao, situacao, valor_envio
ctbz correspondencias baixar ID --dir DIR
ctbz correspondencias endereco -o json        # endereço de envio
ctbz correspondencias endereco --cep 01310-100 --numero 1000 [--complemento "sala 1"] --dry-run
ctbz correspondencias autorizar on|off --dry-run   # on: risco alto (envio pode ser cobrado)
```

Sem o serviço contratado, a listagem avisa e termina com código 0.

## Mensalidade, plano e contrato

```sh
ctbz mensalidade -o json             # competencia, valor, vencimento, situacao, competencia_anterior_atrasada
ctbz mensalidade situacao -o json    # em_dia
ctbz mensalidade historico -o json   # faturas pagas (pagamentos[]), da mais recente
ctbz mensalidade segunda-via --dry-run          # risco médio: registra boleto novo, não cobra
ctbz mensalidade segunda-via --yes -o json      # itens, total, vencimento, pix_copia_e_cola, linha_digitavel, link
ctbz mensalidade segunda-via --pix --yes        # só o copia e cola, para pipe
ctbz plano -o json                   # plano e categoria (lidos no login)
ctbz plano contrato --texto          # contrato de prestação de serviços
ctbz plano proposta --texto          # proposta com a tabela de preços
```

"Pague a mensalidade": confira `ctbz mensalidade -o json`. Se `botao_adicional_acao` começa com
`request:`, gere a segunda via (protocolo de escrita, risco médio) e entregue ao usuário o Pix
copia e cola ou a linha digitável. O pagamento é feito fora da Contabilizei (ex.: pelo app do
banco, ou `inter-pj pix enviar --copia-e-cola "$(ctbz mensalidade segunda-via --pix --yes)"`
com a confirmação dele). Não gere outra segunda via sem necessidade: cada uma registra um boleto.
Depois do pagamento, confirme com `ctbz mensalidade situacao` (`em_dia`). A fatura pública demora
a dar baixa. O `link` dá acesso à fatura sem login: não o grave no dossiê.

Para perguntas como "o que meu plano cobre?" ou "posso cancelar?", leia o contrato
(`--texto`) e cite o trecho. Trocar de plano é feito só com a Contabilizei.

## Atendimento e primeiros passos

```sh
ctbz chamados -o json [--finalizados]   # a CLI só lista; abrir chamado é pelo painel ou WhatsApp
ctbz primeiros-passos -o json           # checklist; conclusao_manual diz o que dá para concluir
ctbz primeiros-passos concluir ETAPA    # risco baixo, sem volta
```

## Auditoria

`ctbz acoes --desde AAAA-MM-DD -o json` lista as escritas enviadas por esta CLI (sem os corpos).
Use para responder "o que foi alterado?" e para não repetir uma escrita.
