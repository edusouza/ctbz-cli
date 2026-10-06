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
