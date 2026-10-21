# Skill `cadastro-e-acessos`

Cadastro, acessos e a relação com a Contabilizei: dados da empresa e dos sócios, troca de
empresa, certificado digital, central de documentos, contas bancárias, quem acessa a empresa,
credenciais de órgãos públicos, procuração do e-CAC, e-mail, telefone e senha de login,
escritório virtual, mensalidade, plano e contrato, chamados e primeiros passos.

É a skill com as operações de maior impacto em **segurança**: senhas e credenciais nunca passam
pelo chat.

Instalação e configuração do plugin: [README das skills](../README.md).

## Quando usar

```text
> Quando vence meu certificado digital?
> Cadastra minha conta do Itaú.
> Quem tem acesso à minha empresa? Convida meu sócio.
> Preciso trocar a senha da prefeitura.
> A procuração do e-CAC está pendente?
> Chegou alguma correspondência no escritório virtual?
> Estou em dia com a mensalidade? O que meu plano cobre?
> Troca para a outra empresa.
```

Pelo nome: `/ctbz:cadastro-e-acessos`.

## Pré-requisitos

- O que a skill [`assessor`](../assessor/README.md) pede (sessão válida no `ctbz`).
- Um terminal à parte, para os comandos que pedem segredos (veja abaixo).

## O que ela faz

### Empresa e sócios

`ctbz empresa`, `empresa socios`, `empresa atividades` e `empresas`. `ctbz empresa usar CNPJ`
troca de empresa: refaz o login (um código novo por e-mail) e o agente troca também de dossiê.

### Certificado digital (e-CNPJ)

`ctbz certificado` mostra validade, dias para vencer e se já dá para renovar. Sem certificado
válido não há emissão de notas nem entrega de obrigações. Com 60 dias ou menos, o agente avisa;
a renovação é feita pelo painel (a CLI não compra nem renova).

### Central de documentos

Tipos aceitos, arquivos enviados e o que as pendências pedem. Envio e declarações de
inexistência ficam na skill [`fechamento-mensal`](../fechamento-mensal/README.md).

### Contas bancárias

```sh
ctbz contas-bancarias -o json
ctbz contas-bancarias bancos -o json
ctbz contas-bancarias adicionar --banco 341 --agencia 1234 --conta 12345-6 \
  --abertura 2024-01-10 --declaro-conta-pj --dry-run
```

`--declaro-conta-pj` afirma que a conta é PJ e movimenta só este CNPJ: o agente confirma com
você. A data de abertura define desde quando a Contabilizei pede extratos.

### Usuários com acesso

Antes de convidar alguém, o agente deixa claro que **não há perfil restrito** (o convidado tem
acesso total, inclusive financeiro) e que o convite não pode ser revogado antes do aceite (só
desativando depois). O administrador não pode ser desativado.

### Credenciais de órgãos públicos

`ctbz empresa credenciais` mostra só se o código do Simples, a senha da prefeitura e a da
Dataprev estão preenchidos. O agente **nunca** usa `--mostrar`, porque os valores ficariam na
conversa.

Para trocar, **você** roda, num terminal à parte:

```sh
ctbz empresa credenciais atualizar --prefeitura      # ou --codigo-simples, --dataprev
```

Os valores são lidos sem eco. Credencial errada faz falhar a entrega de declarações; a
Contabilizei verifica o acesso e avisa por e-mail. Quando o painel pede "redefina a senha da
prefeitura" e você já atualizou, o agente pode rodar `empresa credenciais confirmar-prefeitura`.

### Procuração do e-CAC

`ctbz pendencias procuracao` mostra se ela está pendente, o CNPJ a quem outorgar e o link do
e-CAC. Você faz a procuração no gov.br; depois, o agente registra que foi feita
(`--ja-criei`).

### Login: e-mail, telefone e senha

`ctbz conta` mostra e-mail e telefone mascarados. Alterar e-mail/telefone (`conta alterar`) ou a
senha (`conta senha`) exige um código OTP e, no caso da senha, digitação sem eco: **você** roda
no seu terminal. Depois, atualize `CTBZ_USER`/`CTBZ_PASSWORD`; outras sessões podem cair.

### Escritório virtual

Correspondências recebidas (e o download de cada uma), endereço de envio e autorização de
recebimento. Sem o serviço contratado, a listagem só avisa.

### Mensalidade, plano e contrato

Mensalidade atual, se a empresa está em dia, histórico de pagamentos, plano contratado, contrato
e proposta. Para "o que meu plano cobre?" ou "posso cancelar?", o agente lê o contrato
(`ctbz plano contrato --texto`) e cita o trecho. Trocar de plano é feito só com a Contabilizei.

### Atendimento e primeiros passos

`ctbz chamados` lista os chamados (abrir é pelo painel ou pelo WhatsApp) e
`ctbz primeiros-passos` mostra o checklist de primeiros passos do painel.

### Auditoria

`ctbz acoes --desde AAAA-MM-DD` lista as escritas enviadas pela CLI (sem os corpos): responde
"o que foi alterado?" e evita repetir uma escrita.

## O que pode alterar

Tudo com `--dry-run` primeiro e a sua confirmação (protocolo da skill `assessor`).

| Ação | Quem roda | Risco | Consequência |
|---|---|---|---|
| `contas-bancarias adicionar` / `editar` | agente | médio | editar de novo |
| `contas-bancarias remover` | agente | **alto** | permanente; vínculos se perdem |
| `usuarios convidar EMAIL` | agente | **alto** | acesso total, sem revogar antes do aceite |
| `usuarios desativar` / `ativar` | agente | médio | um desfaz o outro |
| `pendencias procuracao --ja-criei` | agente | médio | a pendência volta se a procuração não existir |
| `empresa credenciais confirmar-prefeitura` | agente | médio | — |
| `correspondencias endereco` / `autorizar off` | agente | médio | regravar / `autorizar on` |
| `correspondencias autorizar on` | agente | **alto** | o envio pode ser cobrado |
| `primeiros-passos concluir ETAPA` | agente | baixo | sem volta |
| `empresa credenciais atualizar` | **você**, no terminal | **alto** | credencial errada faz falhar declarações |
| `conta alterar` / `conta senha` | **você**, no terminal | **alto** | pode tirar o acesso; exige OTP |

## Fora do alcance

Comprar ou renovar o certificado, abrir chamado, trocar de plano: painel, app ou atendimento da
Contabilizei.

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| [`SKILL.md`](SKILL.md) | instruções que o agente segue |
