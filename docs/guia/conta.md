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

