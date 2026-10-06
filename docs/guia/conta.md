# Conta e acessos

## Usuários da empresa

```sh
ctbz usuarios
```

```text
ID   Nome             E-mail                 Tipo                Situação
101  FULANO DE TAL    fulano@example.com     USUARIO_PRINCIPAL   ATIVO
102  BELTRANO DE TAL  beltrano@example.com   USUARIO_SECUNDARIO  INATIVO
```

- São as pessoas com acesso à empresa no painel (multiusuário) e os convites ainda não
  aceitos (`CONVITE_ENVIADO`). O usuário principal é o administrador da empresa.

### Convidar, desativar e ativar

```sh
ctbz usuarios convidar pessoa@example.com
ctbz usuarios desativar 102
ctbz usuarios ativar 102
```

- O convite dá **acesso total**: a pessoa pode fazer todas as ações e ver todas as informações
  da empresa, inclusive financeiras. Não há como revogar o convite antes do aceite; depois, só
  desativando o usuário. Por isso o convite é de risco alto (`confirmo`).
- A CLI recusa e-mails da Contabilizei, e-mails que já estão na lista e convites quando o
  serviço está indisponível ou o limite de usuários secundários ativos foi atingido.
- `desativar` é a única forma de revogar um acesso; o administrador não pode ser desativado.
  `ativar` devolve o acesso. Os dois são de risco médio e reversíveis.
- Não se sabe se desativar um convite ainda não aceito o cancela; a CLI avisa no resumo.
- O e-mail convidado não vai para o registro de ações (`ctbz acoes`), que só guarda o comando
  e o resultado.

## Dados de acesso a órgãos públicos

```sh
ctbz empresa credenciais
ctbz empresa credenciais atualizar --prefeitura
ctbz empresa credenciais atualizar --codigo-simples --dataprev
ctbz empresa credenciais confirmar-prefeitura
```

- São as credenciais que a Contabilizei usa para entregar declarações: o código de acesso do
  Simples Nacional e o usuário e a senha da prefeitura e do Dataprev.
- `ctbz empresa credenciais` mostra o código e as senhas só como `preenchido`; `--mostrar`
  exibe os valores, apenas num terminal e na tabela. A API devolve as senhas em texto claro, e
  a CLI nunca as grava nem as registra.
- `atualizar` pede os novos valores **sem eco** no terminal, nunca por argumento: o código do
  Simples tem 12 dígitos; usuário e senha precisam de pelo menos 3 caracteres (Enter no
  usuário mantém o atual). As credenciais não escolhidas seguem como estão. Em scripts, sem
  terminal, os valores são lidos da entrada padrão, uma linha por pergunta, e é preciso `--yes`.
- Risco alto: credencial errada faz falhar a entrega de declarações (multas e juros). Depois
  de salvar, a Contabilizei verifica o acesso e avisa por e-mail se ele for negado.
- O `--dry-run` mostra `***` no lugar das senhas e do código. O registro de ações guarda só o
  comando e o resultado.
- `confirmar-prefeitura` é o "Sim, já atualizei a senha" do alerta "Redefina a senha da
  prefeitura": declara que a senha cadastrada está correta (risco médio). Só é enviado quando o
  painel mostra o alerta.
- Usuários administradores não acessam os dados de acesso (a Contabilizei responde 403).
- A situação da verificação (Pendente, Verificando acesso, Vinculada, Acesso negado) ainda não
  aparece: o campo não foi visto numa resposta real.

## Dados de login: e-mail, telefone e senha

```sh
ctbz conta                                                     # e-mail e telefone mascarados, método do 2FA
ctbz conta alterar --email novo@example.com --telefone "(11)98765-4321"
ctbz conta senha [--via sms]
```

- Toda troca exige um **código OTP**: depois da confirmação, a Contabilizei envia o código por
  e-mail ou SMS (`--via`; padrão: o método do 2FA). Com o app autenticador (`APP`), nenhum
  código é enviado: digite o do app.
- O código é lido do `--otp-cmd` (ou `CTBZ_OTP_CMD`), como no [login](https://github.com/edusouza/ctbz-cli#login),
  ou digitado no terminal. Sem terminal, `--otp-cmd` é obrigatório e é conferido antes de
  qualquer envio. Se um código foi pedido há pouco, a CLI diz quantos segundos esperar.
- `alterar` pede **e-mail e telefone juntos**: a API recebe os dois e o painel mostra os atuais
  mascarados, então repita o que não muda. Telefone com DDD e 9 dígitos. Os próximos códigos
  de login vão para o novo e-mail: atualize `CTBZ_USER`.
- `senha` pede a nova senha duas vezes, sem eco (em scripts, duas linhas na entrada padrão).
  Regras do painel: pelo menos 8 caracteres, com minúsculas, maiúsculas, números e um caractere
  especial. A senha atual não é pedida: a prova é o código. Atualize `CTBZ_PASSWORD`; outras
  sessões podem ser encerradas.
- Risco alto: um erro pode tirar seu acesso. Dá para trocar de novo. Código errado não altera
  nada (a CLI avisa). O `--dry-run` mostra `***` no lugar da senha e do código.

## Escritório virtual (correspondências)

```sh
ctbz correspondencias                                   # correspondências recebidas
ctbz correspondencias baixar 501 --dir ~/Documentos     # salva correspondencia-501.pdf
ctbz correspondencias endereco                          # endereço de envio
ctbz correspondencias endereco --cep 01310-100 --numero 1000 [--complemento "sala 1"]
ctbz correspondencias autorizar on|off
```

- Para quem contrata o escritório virtual. Sem o serviço, a listagem avisa e termina com
  código 0, e os demais comandos explicam que ele não foi contratado.
- A lista mostra data, remetente, descrição, situação e o **valor do envio**: o envio de
  correspondências pode ser cobrado.
- `endereco` com `--cep` e `--numero` troca o endereço de envio: logradouro, bairro e cidade
  vêm do CEP, e a CLI mostra o endereço de antes e o novo. A mudança vale só para o envio das
  correspondências.
- `autorizar` liga ou desliga "Autorizo o recebimento de correspondências e documentos
  entregues no endereço informado".
- Trocar o endereço e a autorização são de risco médio e reversíveis.
- Os formatos vêm do código do painel: a conta usada no desenvolvimento não tinha o serviço.

