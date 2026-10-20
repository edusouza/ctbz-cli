# Skills para agentes

O repositório traz um plugin do [Claude Code](https://code.claude.com) com skills que fazem um
agente operar a Contabilizei pelo `ctbz` como um assessor contábil, sem precisar explorar a CLI
a cada conversa. As skills simulam o atendimento do plano Experts (assessor e analista
dedicados): checkup do CNPJ, fechamento do mês, consultoria tributária, impostos e o fechamento
do ano.

## Instalar

No Claude Code:

```text
/plugin marketplace add edusouza/ctbz-cli
/plugin install ctbz@ctbz-cli
```

A partir do Claude Code 2.1.275, também em um passo: `/plugin install ctbz --marketplace edusouza/ctbz-cli`.

Para usar sem o marketplace, a partir de um clone: `claude --plugin-dir /caminho/do/ctbz-cli`.
O `ctbz` precisa estar instalado e no `PATH` (veja o [README](https://github.com/edusouza/ctbz-cli#instalação)).

## Skills

| Skill | Papel | O que faz |
|---|---|---|
| `ctbz:assessor` | base | login sem terminal, formatos e códigos de saída, protocolo de escrita, dossiê da empresa, para onde encaminhar cada pedido |
| `ctbz:checkup` | reunião mensal | lê tudo em paralelo e entrega um relatório priorizado (urgente, esta semana, este mês, oportunidades) |
| `ctbz:fechamento-mensal` | analista | extratos, classificação e desmembramento, caixa, conciliação, notas de entrada, documentos |
| `ctbz:impostos` | assessor | guias, PDFs, memória de cálculo, confirmação de pagamento, recálculo, parcelamento |
| `ctbz:planejamento-tributario` | consultoria tributária | Fator R e anexo III x V com simulação de cenários, pró-labore x lucros, limites do Simples |
| `ctbz:fechamento-anual` | assessor | restrições, carta de responsabilidade, distribuição de lucros, informe de rendimentos para o IRPF |
| `ctbz:notas-e-clientes` | assessor | faturamento por cliente, tomadores, emissor e alíquotas, ficha para emitir a nota no painel |
| `ctbz:cadastro-e-acessos` | atendimento | certificado, documentos, contas bancárias, usuários, credenciais, e-mail e senha, escritório virtual, mensalidade e contrato |

O agente escolhe a skill pelo pedido ("como está minha empresa?", "dá para pagar menos
imposto?", "fecha o mês de setembro"); também dá para chamar pelo nome, como `/ctbz:checkup`.

## Como o agente trabalha

- **Login:** usa `CTBZ_USER` e `CTBZ_PASSWORD` do ambiente e pede ao usuário só o código de 6
  dígitos (ou usa `CTBZ_OTP_CMD`). Nunca pede a senha no chat.
- **Leituras:** sempre com `-o json`, em paralelo quando possível.
- **Escritas:** sempre `--dry-run` primeiro, depois o resumo da ação, do risco e da
  consequência, e só então envia com `--yes`, depois da confirmação do usuário. Nas de risco alto, o
  usuário precisa confirmar aquela ação específica. Senhas e credenciais de órgãos públicos são
  digitadas pelo usuário no próprio terminal.
- **Dossiê:** um arquivo por empresa em `$CTBZ_HOME/dossie/<CNPJ>.md` (permissão `0600`) com
  o perfil, os IDs usados nos comandos, as decisões e o histórico de checkups. Nunca guarda
  senhas, códigos ou dados de cartão.

## Limites

As skills não fazem o que a CLI não faz: emitir NFS-e, pagar guias, abrir chamados, folha de
pagamento e troca de plano continuam no painel ou com o atendimento. As análises
(planejamento tributário, por exemplo) são estimativas de apoio e não substituem o contador
responsável.

## Manter

O mapa de comandos que o agente consulta (`skills/assessor/references/comandos.md`) é gerado
a partir de `docs/referencia`. Depois de mudar a CLI:

```sh
go run ./tools/gendocs
python3 scripts/gerar-comandos-skill.py
```

Os campos do JSON (`skills/assessor/references/campos-json.md`) e os passos de cada skill são
escritos à mão: atualize-os junto com o comando que mudou. A versão do plugin
(`.claude-plugin/plugin.json`) acompanha a da CLI.
