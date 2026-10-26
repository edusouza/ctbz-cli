# Roadmap

Funcionalidades planejadas para o `ctbz`, organizadas em versões. Cada versão cobre **um contexto completo** da
Contabilizei. Cada versão tem um épico no GitHub e uma sub-issue por funcionalidade.

Até a 1.0 a CLI só lê. A partir da 1.1 entram as ações de escrita (adicionar, alterar, remover e marcar como
concluído), uma versão por contexto, em ordem crescente de risco.

Princípios:

- **Somente leitura até a 1.0** ([ADR-0002](docs/adr/0002-somente-leitura-ate-1-0.md)). As ações de escrita
  começam na v1.1 e seguem a [ADR-0018](docs/adr/0018-escrita-com-confirmacao.md): confirmação ou `--yes`,
  `--dry-run`, risco declarado, uma única tentativa, releitura do estado e teste de requisição.
- **Saída padronizada e testes de contrato desde a v0.1**: todo comando novo aceita `--output table|json|csv` e
  ganha um contrato que detecta mudanças no formato das respostas.
- Endpoints de cada funcionalidade vêm de [docs/api](docs/api/README.md); os de escrita, de
  [docs/api/escrita](docs/api/escrita/README.md). Os itens marcados como *investigação* dependem de descobrir
  parâmetros ainda não testados.

Labels usadas: `épico`, `versão: vX.Y`, `contexto: …`, `tipo: funcionalidade|investigação|infra|documentação`,
nas escritas, `risco: baixo|médio|alto` e, nas correções, `prioridade: crítica|alta|média|baixa`.

Depois da v1.9, as versões deixam de ser por contexto: a v1.9.3 (planejada como v1.9.1) corrige só os achados
críticos e cada versão seguinte fecha **um tema** de correção, com os achados de review que continuam válidos na
`main` (triagem de 23/10/2026).

## Visão geral

| Versão | Contexto | Épico | Funcionalidades | Situação |
|---|---|---|---|---|
| **v0.1** | Login e dados da empresa | [#1](https://github.com/edusouza/ctbz-cli/issues/1) | 8 | concluída (#9 passou para a v1.17) |
| **v0.2** | Impostos | [#10](https://github.com/edusouza/ctbz-cli/issues/10) | 6 | concluída |
| **v0.3** | Pendências, rotinas e atendimento | [#17](https://github.com/edusouza/ctbz-cli/issues/17) | 5 | concluída |
| **v0.4** | Mensalidade e pagamentos da Contabilizei | [#23](https://github.com/edusouza/ctbz-cli/issues/23) | 4 | concluída |
| **v0.5** | Notas fiscais | [#28](https://github.com/edusouza/ctbz-cli/issues/28) | 5 | concluída |
| **v0.6** | Pró-labore e distribuição de lucros | [#34](https://github.com/edusouza/ctbz-cli/issues/34) | 5 | concluída |
| **v0.7** | Contabilidade: relatórios, caixa e extratos | [#40](https://github.com/edusouza/ctbz-cli/issues/40) | 6 | concluída |
| **v0.8** | Documentos e certificado digital | [#47](https://github.com/edusouza/ctbz-cli/issues/47) | 2 | concluída |
| **v1.0** | Estabilidade e distribuição | [#50](https://github.com/edusouza/ctbz-cli/issues/50) | 3 | concluída |
| **v1.1** | Fundação da escrita | [#173](https://github.com/edusouza/ctbz-cli/issues/173) | 6 | concluída |
| **v1.2** | Caixa e classificação de lançamentos | [#180](https://github.com/edusouza/ctbz-cli/issues/180) | 7 | concluída |
| **v1.3** | Impostos: confirmar pagamento, recálculo e parcelamento | [#188](https://github.com/edusouza/ctbz-cli/issues/188) | 4 | concluída |
| **v1.4** | Pendências, termos e conciliação | [#193](https://github.com/edusouza/ctbz-cli/issues/193) | 5 | concluída |
| **v1.5** | Extratos, contas bancárias e documentos | [#199](https://github.com/edusouza/ctbz-cli/issues/199) | 6 | concluída |
| **v1.6** | Notas fiscais: tomadores e notas de entrada | [#206](https://github.com/edusouza/ctbz-cli/issues/206) | 4 | concluída |
| **v1.7** | Pró-labore, distribuição de lucros e informe de rendimentos | [#211](https://github.com/edusouza/ctbz-cli/issues/211) | 6 | concluída |
| **v1.8** | Conta, usuários e acessos | [#218](https://github.com/edusouza/ctbz-cli/issues/218) | 4 | concluída |
| **v1.8.1** | Correções do primeiro uso real de ponta a ponta | — | 6 | concluída |
| **v1.9** | Segunda via da mensalidade, extratos em CSV/XLS e diagnóstico | — | 5 | concluída |
| **v1.9.2** | Extratos do Banco Inter: OFX e mês preso em processamento | — | 2 | concluída |
| **v1.9.3** | Correções críticas (planejada como v1.9.1) | [#943](https://github.com/edusouza/ctbz-cli/issues/943) | 16 | concluída |
| **v1.10** | Saída segura: caracteres de controle, CSV, segredos e dados pessoais | [#944](https://github.com/edusouza/ctbz-cli/issues/944) | 50 | planejada |
| **v1.11** | Escrita segura e `--dry-run` | [#945](https://github.com/edusouza/ctbz-cli/issues/945) | 78 | planejada |
| **v1.12** | Depois do envio: releitura, situação e orientação | [#946](https://github.com/edusouza/ctbz-cli/issues/946) | 66 | planejada |
| **v1.13** | Validação de entrada | [#947](https://github.com/edusouza/ctbz-cli/issues/947) | 54 | planejada |
| **v1.14** | Leitura de dados e contratos da API | [#948](https://github.com/edusouza/ctbz-cli/issues/948) | 96 | planejada |
| **v1.15** | Testes | [#949](https://github.com/edusouza/ctbz-cli/issues/949) | 75 | planejada |
| **v1.16** | Documentação e skills | [#950](https://github.com/edusouza/ctbz-cli/issues/950) | 75 | planejada |
| **v1.17** | Infra, manutenção e notas de release | [#951](https://github.com/edusouza/ctbz-cli/issues/951) | 85 | planejada |

Nas versões de correção, a coluna Funcionalidades conta as sub-issues do épico.

## v0.1 — Login e dados da empresa

Épico: [#1](https://github.com/edusouza/ctbz-cli/issues/1) · contexto `autenticação`

Base da CLI: autenticar na Contabilizei (usuário/senha + OTP por e-mail + seleção de empresa), manter a sessão e consultar os dados cadastrais da empresa. Inclui a fundação usada por todas as versões seguintes: formato de saída padronizado e testes de contrato contra a API.

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#2](https://github.com/edusouza/ctbz-cli/issues/2) | Login com usuário, senha e OTP por e-mail | funcionalidade |
| ✅ | [#3](https://github.com/edusouza/ctbz-cli/issues/3) | Comandos básicos: status, empresa, api e logout | funcionalidade |
| ✅ | [#4](https://github.com/edusouza/ctbz-cli/issues/4) | Documentação da engenharia reversa em docs/ | documentação |
| ✅ | [#5](https://github.com/edusouza/ctbz-cli/issues/5) | Saída padronizada: --output table\|json\|csv | infra |
| ✅ | [#6](https://github.com/edusouza/ctbz-cli/issues/6) | Testes de contrato da API | infra |
| ✅ | [#7](https://github.com/edusouza/ctbz-cli/issues/7) | Listar empresas do usuário e trocar de empresa | funcionalidade |
| ✅ | [#8](https://github.com/edusouza/ctbz-cli/issues/8) | Dados completos da empresa: sócios, endereço, atividades e certificado | funcionalidade |
| ➡️ | [#9](https://github.com/edusouza/ctbz-cli/issues/9) | Validar OTP automático com Gmail real (gws) (passou para a v1.17) | investigação |

Fora de escopo: CI e releases (v1.0); Ações que alteram dados.

## v0.2 — Impostos

Épico: [#10](https://github.com/edusouza/ctbz-cli/issues/10) · contexto `impostos`

Tudo sobre os impostos da empresa: guias a pagar (atrasadas, do mês e do próximo mês), detalhes e memória de cálculo, download das guias em PDF, histórico de pagamentos, faturamento usado na apuração e parcelamentos.

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#11](https://github.com/edusouza/ctbz-cli/issues/11) | Listar guias a pagar | funcionalidade |
| ✅ | [#12](https://github.com/edusouza/ctbz-cli/issues/12) | Detalhes de uma guia e memória de cálculo | funcionalidade |
| ✅ | [#13](https://github.com/edusouza/ctbz-cli/issues/13) | Baixar guias em PDF | funcionalidade |
| ✅ | [#14](https://github.com/edusouza/ctbz-cli/issues/14) | Histórico de impostos | funcionalidade |
| ✅ | [#15](https://github.com/edusouza/ctbz-cli/issues/15) | Faturamento mensal usado na apuração | funcionalidade |
| ✅ | [#16](https://github.com/edusouza/ctbz-cli/issues/16) | Parcelamentos e débitos federais (leitura) | investigação |

Fora de escopo: Confirmar pagamento de guia, recálculo e contratação de parcelamento (escrita).

## v0.3 — Pendências, rotinas e atendimento

Épico: [#17](https://github.com/edusouza/ctbz-cli/issues/17) · contexto `pendências`

O que a empresa precisa resolver: pendências abertas, rotinas e obrigações do mês, conciliações pendentes e chamados com o atendimento. Culmina num resumo de uma tela com tudo que exige atenção.

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#18](https://github.com/edusouza/ctbz-cli/issues/18) | Listar pendências da empresa | funcionalidade |
| ✅ | [#19](https://github.com/edusouza/ctbz-cli/issues/19) | Central de rotinas e obrigações do mês | funcionalidade |
| ✅ | [#20](https://github.com/edusouza/ctbz-cli/issues/20) | Pendências de conciliação fiscal | funcionalidade |
| ✅ | [#21](https://github.com/edusouza/ctbz-cli/issues/21) | Chamados de atendimento | funcionalidade |
| ✅ | [#22](https://github.com/edusouza/ctbz-cli/issues/22) | Resumo geral: o que precisa de atenção | funcionalidade |

Fora de escopo: Resolver pendência, aceitar termos, abrir chamado (escrita).

## v0.4 — Mensalidade e pagamentos da Contabilizei

Épico: [#23](https://github.com/edusouza/ctbz-cli/issues/23) · contexto `mensalidade`

A relação financeira com a própria Contabilizei: mensalidade atual, faturas, histórico de pagamentos, débito automático, plano e contrato.

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#24](https://github.com/edusouza/ctbz-cli/issues/24) | Mensalidade e fatura atual | funcionalidade |
| ✅ | [#25](https://github.com/edusouza/ctbz-cli/issues/25) | Histórico de pagamentos e débito automático | funcionalidade |
| ✅ | [#26](https://github.com/edusouza/ctbz-cli/issues/26) | Situação de inadimplência | funcionalidade |
| ✅ | [#27](https://github.com/edusouza/ctbz-cli/issues/27) | Plano e contrato de serviço | funcionalidade |

Fora de escopo: Pagar fatura, cadastrar cartão, trocar plano (escrita).

## v0.5 — Notas fiscais

Épico: [#28](https://github.com/edusouza/ctbz-cli/issues/28) · contexto `notas fiscais`

Consulta das notas fiscais de serviço emitidas, dos tomadores e da configuração do emissor, além das notas tomadas/importadas.

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#29](https://github.com/edusouza/ctbz-cli/issues/29) | Listar notas fiscais emitidas | funcionalidade |
| ✅ | [#30](https://github.com/edusouza/ctbz-cli/issues/30) | Baixar PDF e XML das notas (inviável: a API não oferece; ver docs/api/notas-fiscais.md) | investigação |
| ✅ | [#31](https://github.com/edusouza/ctbz-cli/issues/31) | Tomadores (clientes) | funcionalidade |
| ✅ | [#32](https://github.com/edusouza/ctbz-cli/issues/32) | Configuração do emissor e alíquotas | funcionalidade |
| ✅ | [#33](https://github.com/edusouza/ctbz-cli/issues/33) | Notas tomadas e notas de entrada | investigação |

Fora de escopo: Emitir, cancelar, replicar ou agendar notas; cadastrar tomador (escrita).

## v0.6 — Pró-labore e distribuição de lucros

Épico: [#34](https://github.com/edusouza/ctbz-cli/issues/34) · contexto `pró-labore`

Remuneração dos sócios: pró-labore atual e histórico, parâmetros (INSS, IRRF, teto), distribuição de lucros e informes de rendimentos.

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#35](https://github.com/edusouza/ctbz-cli/issues/35) | Pró-labore atual e histórico | funcionalidade |
| ✅ | [#36](https://github.com/edusouza/ctbz-cli/issues/36) | Parâmetros de cálculo do pró-labore | funcionalidade |
| ✅ | [#37](https://github.com/edusouza/ctbz-cli/issues/37) | Distribuição de lucros | funcionalidade |
| ✅ | [#38](https://github.com/edusouza/ctbz-cli/issues/38) | Informe e comprovante de rendimentos dos sócios | investigação |
| ✅ | [#39](https://github.com/edusouza/ctbz-cli/issues/39) | Fator R e simulador de impostos (leitura) | funcionalidade |

Fora de escopo: Alterar ou zerar pró-labore, ativar gestão inteligente, registrar distribuição (escrita).

## v0.7 — Contabilidade: relatórios, caixa e extratos

Épico: [#40](https://github.com/edusouza/ctbz-cli/issues/40) · contexto `contabilidade`

Relatórios contábeis (balancete, balanço, razão), lançamentos do caixa, extratos bancários e plano de contas.

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#41](https://github.com/edusouza/ctbz-cli/issues/41) | Balancete mensal | funcionalidade |
| ✅ | [#42](https://github.com/edusouza/ctbz-cli/issues/42) | Balanço patrimonial | funcionalidade |
| ✅ | [#43](https://github.com/edusouza/ctbz-cli/issues/43) | Razão contábil | funcionalidade |
| ✅ | [#44](https://github.com/edusouza/ctbz-cli/issues/44) | Caixa: lançamentos do mês | funcionalidade |
| ✅ | [#45](https://github.com/edusouza/ctbz-cli/issues/45) | Extratos e contas bancárias | funcionalidade |
| ✅ | [#46](https://github.com/edusouza/ctbz-cli/issues/46) | Plano de contas e classificações | funcionalidade |

Fora de escopo: Importar extrato, classificar/desmembrar lançamentos, reabrir balanço (escrita).

## v0.8 — Documentos e certificado digital

Épico: [#47](https://github.com/edusouza/ctbz-cli/issues/47) · contexto `documentos`

Documentos enviados à Contabilizei e situação do certificado digital (candidata: pode ser reordenada ou removida).

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#48](https://github.com/edusouza/ctbz-cli/issues/48) | Documentos enviados | funcionalidade |
| ✅ | [#49](https://github.com/edusouza/ctbz-cli/issues/49) | Certificado digital: situação e renovação (leitura) | funcionalidade |

Fora de escopo: Enviar documentos, emitir/renovar/remover certificado (escrita).

## v1.0 — Estabilidade e distribuição

Épico: [#50](https://github.com/edusouza/ctbz-cli/issues/50) · contexto `infra`

Itens transversais adiados: integração contínua, releases com binários e monitoramento de mudanças no site da Contabilizei.

| | Issue | Funcionalidade | Tipo |
|---|---|---|---|
| ✅ | [#51](https://github.com/edusouza/ctbz-cli/issues/51) | CI: testes, vet e lint no GitHub Actions | infra |
| ✅ | [#52](https://github.com/edusouza/ctbz-cli/issues/52) | Releases com binários | infra |
| ✅ | [#53](https://github.com/edusouza/ctbz-cli/issues/53) | Monitorar mudanças no front e na API | infra |

Fora de escopo: Ações de escrita (avaliadas depois da 1.0).

## v1.1 — Fundação da escrita

Épico: [#173](https://github.com/edusouza/ctbz-cli/issues/173) · contexto `infra`

Base comum para todas as escritas: política registrada em ADR, camada HTTP de escrita (JSON e multipart) sem retentativa, confirmação e `--dry-run`, testes de requisição, registro local das ações e catálogo que enxerga as escritas.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#174](https://github.com/edusouza/ctbz-cli/issues/174) | Documentar endpoints de escrita e decidir a política de escrita (ADR-0018) | documentação | — |
| ✅ | [#175](https://github.com/edusouza/ctbz-cli/issues/175) | Camada de escrita: JSON e multipart, sem retentativa nem re-login no meio | infra | — |
| ✅ | [#176](https://github.com/edusouza/ctbz-cli/issues/176) | Confirmação, --dry-run, --yes e níveis de risco nos comandos de escrita | infra | — |
| ✅ | [#177](https://github.com/edusouza/ctbz-cli/issues/177) | Testes de requisição para as escritas | infra | — |
| ✅ | [#178](https://github.com/edusouza/ctbz-cli/issues/178) | Registro local das ações de escrita (ctbz acoes) | funcionalidade | — |
| ✅ | [#179](https://github.com/edusouza/ctbz-cli/issues/179) | Catálogo: capturar escritas em axios({method,url}) e caminhos em variáveis | infra | — |

Fora de escopo: Comandos de domínio que escrevem (a partir da v1.2).

## v1.2 — Caixa e classificação de lançamentos

Épico: [#180](https://github.com/edusouza/ctbz-cli/issues/180) · contexto `contabilidade`

Primeiro contexto de escrita: lançamentos do caixa e classificação dos lançamentos do extrato. Risco médio e reversível enquanto o período está aberto, ideal para estrear a fundação.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#181](https://github.com/edusouza/ctbz-cli/issues/181) | Classificações disponíveis para lançamentos do caixa e do extrato | funcionalidade | baixo |
| ✅ | [#182](https://github.com/edusouza/ctbz-cli/issues/182) | Adicionar lançamento no caixa | funcionalidade | médio |
| ✅ | [#183](https://github.com/edusouza/ctbz-cli/issues/183) | Editar lançamento do caixa | funcionalidade | médio |
| ✅ | [#184](https://github.com/edusouza/ctbz-cli/issues/184) | Remover lançamento do caixa | funcionalidade | médio |
| ✅ | [#185](https://github.com/edusouza/ctbz-cli/issues/185) | Listar e classificar lançamentos do extrato | funcionalidade | médio |
| ✅ | [#186](https://github.com/edusouza/ctbz-cli/issues/186) | Desmembrar lançamento do extrato e desfazer o desmembramento | funcionalidade | médio |
| ✅ | [#187](https://github.com/edusouza/ctbz-cli/issues/187) | Reclassificar lançamento pendente e concluir a rotina | funcionalidade | médio |

Fora de escopo: Importar ou excluir extrato (v1.5); reabrir balanço (v1.7).

## v1.3 — Impostos: confirmar pagamento, recálculo e parcelamento

Épico: [#188](https://github.com/edusouza/ctbz-cli/issues/188) · contexto `impostos`

Marcar guias como pagas (ou desmarcar), pedir recálculo de guia vencida e simular ou contratar parcelamentos.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#189](https://github.com/edusouza/ctbz-cli/issues/189) | Confirmar pagamento de guia, "não paguei" e desmarcar | funcionalidade | médio |
| ✅ | [#190](https://github.com/edusouza/ctbz-cli/issues/190) | Pedir recálculo de guia vencida | funcionalidade | alto |
| ✅ | [#191](https://github.com/edusouza/ctbz-cli/issues/191) | Simular parcelamento de débitos | investigação | baixo |
| ✅ | [#192](https://github.com/edusouza/ctbz-cli/issues/192) | Contratar parcelamento de débitos | funcionalidade | alto |

Fora de escopo: Pagar imposto com cartão e débito automático (pagamentos fora do roadmap de escrita).

## v1.4 — Pendências, termos e conciliação

Épico: [#193](https://github.com/edusouza/ctbz-cli/issues/193) · contexto `pendências`

Resolver o que a Central de Rotinas aponta: ler e aceitar termos, declarar a procuração do e-CAC, resolver conciliação fiscal e concluir tarefas de primeiros passos.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#194](https://github.com/edusouza/ctbz-cli/issues/194) | Ler termos e cartas pendentes da Central de Rotinas | funcionalidade | baixo |
| ✅ | [#195](https://github.com/edusouza/ctbz-cli/issues/195) | Aceitar carta de responsabilidade e termos | funcionalidade | alto |
| ✅ | [#196](https://github.com/edusouza/ctbz-cli/issues/196) | Declarar procuração do e-CAC criada | funcionalidade | médio |
| ✅ | [#197](https://github.com/edusouza/ctbz-cli/issues/197) | Resolver pendência de conciliação fiscal | funcionalidade | alto |
| ✅ | [#198](https://github.com/edusouza/ctbz-cli/issues/198) | Primeiros passos: listar, concluir tarefa, dispensar e reativar | funcionalidade | baixo |

Fora de escopo: Abrir ou responder chamados (não há API: o atendimento é no Zendesk); aceites do informe de rendimentos (v1.7).

## v1.5 — Extratos, contas bancárias e documentos

Épico: [#199](https://github.com/edusouza/ctbz-cli/issues/199) · contexto `documentos`

Enviar o que a Contabilizei pede todo mês (extratos, documentos e declarações de "não tenho") e manter as contas bancárias. Primeiras escritas com upload multipart.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#200](https://github.com/edusouza/ctbz-cli/issues/200) | Cadastrar e editar conta bancária | funcionalidade | médio |
| ✅ | [#201](https://github.com/edusouza/ctbz-cli/issues/201) | Remover conta bancária | funcionalidade | alto |
| ✅ | [#202](https://github.com/edusouza/ctbz-cli/issues/202) | Importar extrato bancário (OFX ou PDF) | funcionalidade | médio |
| ✅ | [#203](https://github.com/edusouza/ctbz-cli/issues/203) | Excluir extrato importado | funcionalidade | alto |
| ✅ | [#204](https://github.com/edusouza/ctbz-cli/issues/204) | Enviar documento para uma pendência | funcionalidade | médio |
| ✅ | [#205](https://github.com/edusouza/ctbz-cli/issues/205) | Declarar ausência de documento ou de aplicação financeira | funcionalidade | alto |

Fora de escopo: Open Finance e BS2; certificado digital (fora do roadmap de escrita).

## v1.6 — Notas fiscais: tomadores e notas de entrada

Épico: [#206](https://github.com/edusouza/ctbz-cli/issues/206) · contexto `notas fiscais`

Cadastro de tomadores e as obrigações sobre as NF-e de compra: manifestação à SEFAZ e classificação.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#207](https://github.com/edusouza/ctbz-cli/issues/207) | Cadastrar e editar tomador | funcionalidade | médio |
| ✅ | [#208](https://github.com/edusouza/ctbz-cli/issues/208) | Excluir tomador | funcionalidade | médio |
| ✅ | [#209](https://github.com/edusouza/ctbz-cli/issues/209) | Manifestar notas de entrada | funcionalidade | alto |
| ✅ | [#210](https://github.com/edusouza/ctbz-cli/issues/210) | Classificar notas de entrada | funcionalidade | médio |

Fora de escopo: Emitir, replicar, registrar e cancelar NFS-e (fora do roadmap de escrita); anexo principal do Simples.

## v1.7 — Pró-labore, distribuição de lucros e informe de rendimentos

Épico: [#211](https://github.com/edusouza/ctbz-cli/issues/211) · contexto `pró-labore`

Remuneração dos sócios: alterar o pró-labore, zeramento em meses sem faturamento, gestão inteligente, distribuição de lucros, aceites do informe de rendimentos e reabertura de balanço.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#212](https://github.com/edusouza/ctbz-cli/issues/212) | Alterar o pró-labore de um sócio | funcionalidade | alto |
| ✅ | [#213](https://github.com/edusouza/ctbz-cli/issues/213) | Zerar pró-labore em meses sem faturamento | funcionalidade | alto |
| ✅ | [#214](https://github.com/edusouza/ctbz-cli/issues/214) | Gestão inteligente: ativar e sair do cálculo automático | investigação | alto |
| ✅ | [#215](https://github.com/edusouza/ctbz-cli/issues/215) | Registrar distribuição de lucros entre os sócios | funcionalidade | alto |
| ✅ | [#216](https://github.com/edusouza/ctbz-cli/issues/216) | Aceites do informe de rendimentos por ano | funcionalidade | alto |
| ✅ | [#217](https://github.com/edusouza/ctbz-cli/issues/217) | Reabrir balanço para regularizar pendência documental | funcionalidade | alto |

Fora de escopo: Assistente de pró-labore (holerite, dependentes, duplo vínculo); reabertura como serviço pago.

## v1.8 — Conta, usuários e acessos

Épico: [#218](https://github.com/edusouza/ctbz-cli/issues/218) · contexto `conta`

Usuários da empresa, credenciais de órgãos públicos, dados de login e escritório virtual. Senhas só por prompt oculto.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#219](https://github.com/edusouza/ctbz-cli/issues/219) | Usuários da empresa: listar, convidar, ativar e desativar | funcionalidade | alto |
| ✅ | [#220](https://github.com/edusouza/ctbz-cli/issues/220) | Dados de acesso a órgãos públicos: ver, atualizar e confirmar senha da prefeitura | funcionalidade | alto |
| ✅ | [#221](https://github.com/edusouza/ctbz-cli/issues/221) | Alterar e-mail, telefone e senha da conta com código OTP | funcionalidade | alto |
| ✅ | [#222](https://github.com/edusouza/ctbz-cli/issues/222) | Escritório virtual: correspondências, endereço de envio e autorização | investigação | médio |

Fora de escopo: Segundo fator por app autenticador; revogar convite pendente (sem endpoint).

## v1.8.1 — Correções do primeiro uso real de ponta a ponta

Label `uso real: primeiro ponta-a-ponta` · versão de correção

Achados do primeiro uso real (23/10/2026): conta bancária, extratos e mensalidade.

| | Issue | Correção | Tipo |
|---|---|---|---|
| ✅ | [#852](https://github.com/edusouza/ctbz-cli/issues/852) | `extratos lancamentos` aceita a resposta em array e pagina com limite | bug |
| ✅ | [#853](https://github.com/edusouza/ctbz-cli/issues/853) | `--banco` escolhe o código exato e recusa código ambíguo | bug |
| ✅ | [#854](https://github.com/edusouza/ctbz-cli/issues/854) | `extratos importar` orienta quando a conexão cai no envio | melhoria |
| ✅ | [#855](https://github.com/edusouza/ctbz-cli/issues/855) | Competência `IMPORTADO` sem lançamentos: causa documentada, recusa, aviso e `--substituir` | investigação |
| ✅ | [#856](https://github.com/edusouza/ctbz-cli/issues/856) | `mensalidade historico` lista as faturas pagas | investigação |
| ✅ | [#857](https://github.com/edusouza/ctbz-cli/issues/857) | `ctbz api` e a conversão de caminho do Git Bash | documentação |

A funcionalidade nova que surgiu desse uso, [#851](https://github.com/edusouza/ctbz-cli/issues/851)
(`ctbz mensalidade segunda-via`), ficou para a v1.9.

## v1.9 — Segunda via da mensalidade, extratos em CSV/XLS e diagnóstico

Label `uso real: primeiro ponta-a-ponta` · versão: `v1.9`

O que faltou no primeiro uso real: pagar a mensalidade atrasada, enviar o extrato no formato
que o banco aceita e diagnosticar uma chamada que falha.

| | Issue | Funcionalidade | Tipo | Risco |
|---|---|---|---|---|
| ✅ | [#902](https://github.com/edusouza/ctbz-cli/issues/902) | Modo verbose (`--verbose`/`CTBZ_DEBUG`) com segredos mascarados | infra | — |
| ✅ | [#900](https://github.com/edusouza/ctbz-cli/issues/900) | `ctbz api -F`: envio `multipart/form-data` | infra | — |
| ✅ | [#899](https://github.com/edusouza/ctbz-cli/issues/899) | `extratos importar` aceita CSV/XLS quando o banco declara o formato | funcionalidade | médio |
| ✅ | [#901](https://github.com/edusouza/ctbz-cli/issues/901) | `--substituir` relê a competência depois de uma falha | bug | médio |
| ✅ | [#851](https://github.com/edusouza/ctbz-cli/issues/851) | `ctbz mensalidade segunda-via`: Pix copia e cola e linha digitável | funcionalidade | médio |

## v1.9.2 — Extratos do Banco Inter: OFX e mês preso em processamento

Label `uso real: primeiro ponta-a-ponta` · versão: `v1.9.2`

O que travou o extrato de setembro do Banco Inter no uso real (26/10/2026): o PDF foi recusado
e deixou o mês em `PROCESSANDO`, a exclusão não mudou nada, e o OFX, que o painel importou, a CLI
recusava. Correção curta, fora das versões por tema.

| | Issue | Correção | Tipo |
|---|---|---|---|
| ✅ | [#1011](https://github.com/edusouza/ctbz-cli/issues/1011) | `extratos importar` aceita OFX fora da lista de formatos do banco, com aviso | bug |
| ✅ | [#1012](https://github.com/edusouza/ctbz-cli/issues/1012) | `extratos excluir` não informa "excluido" quando a competência continua `PROCESSANDO` | bug |

## Versões de correção (triagem de 23/10/2026)

Os code reviews da v0.x à v1.9 deixaram 671 issues de achados abertas. Cada uma foi conferida contra a `main`
(`758b36f`): as 78 já corrigidas, obsoletas ou duplicadas foram fechadas com comentário, e as 593 válidas (mais
duas novas, #941 e #942) receberam `prioridade: …` reavaliada e entraram no épico do seu tema. A `severity:`
original do review ficou só como histórico.

Critério de prioridade:

- **crítica** — pode enviar escrita errada ou irreversível, perder dados, vazar segredo ou dado pessoal, ou
  quebrar o contrato público (ADR-0017).
- **alta** — leitura que entrega dado errado em silêncio, orientação que leva a repetir uma escrita, caracteres
  de controle da API no terminal, injeção de fórmula no CSV.
- **média** — caso de borda, erro confuso, `--dry-run` que aborta, validação faltando, teste faltando em escrita.
- **baixa** — testes de leitura, documentação, estilo, manutenção.

Cada épico agrupa as sub-issues em frentes; cada frente tende a virar um PR que fecha várias de uma vez.

| Versão | Épico | Crítica | Alta | Média | Baixa |
|---|---|---|---|---|---|
| v1.9.3 | [#943](https://github.com/edusouza/ctbz-cli/issues/943) | 16 | — | — | — |
| v1.10 | [#944](https://github.com/edusouza/ctbz-cli/issues/944) | — | 43 | 5 | 2 |
| v1.11 | [#945](https://github.com/edusouza/ctbz-cli/issues/945) | — | 24 | 43 | 11 |
| v1.12 | [#946](https://github.com/edusouza/ctbz-cli/issues/946) | — | 19 | 41 | 6 |
| v1.13 | [#947](https://github.com/edusouza/ctbz-cli/issues/947) | — | 4 | 38 | 12 |
| v1.14 | [#948](https://github.com/edusouza/ctbz-cli/issues/948) | — | 5 | 40 | 51 |
| v1.15 | [#949](https://github.com/edusouza/ctbz-cli/issues/949) | — | — | 25 | 50 |
| v1.16 | [#950](https://github.com/edusouza/ctbz-cli/issues/950) | — | — | 2 | 73 |
| v1.17 | [#951](https://github.com/edusouza/ctbz-cli/issues/951) | — | — | 4 | 81 |

## v1.9.3 — Correções críticas

Épico: [#943](https://github.com/edusouza/ctbz-cli/issues/943) · versão de correção · label `versão: v1.9.1`

Planejada como v1.9.1; saiu como v1.9.3 porque a v1.9.2 foi lançada antes, e uma tag menor deixaria estas
correções fora do `go install ...@latest`.

Patch só com o que pode enviar uma escrita errada ou irreversível à Contabilizei ou vazar a sessão. Quase tudo é
o mesmo padrão: um campo ausente da API, ou um argumento repetido ou ambíguo, vira zero, `null` ou o primeiro
candidato numa escrita, contra a ADR-0021. As correções são guardas pequenas antes de `escrever`.

| | Issue | Correção | Contexto |
|---|---|---|---|
| ✅ | [#941](https://github.com/edusouza/ctbz-cli/issues/941) | `ctbz api` com URL absoluta envia os cookies da sessão para qualquer host | autenticação |
| ✅ | [#489](https://github.com/edusouza/ctbz-cli/issues/489) | `parcelamento contratar`: quantidade ausente vira `quantidadeParcelas: 0` | impostos |
| ✅ | [#428](https://github.com/edusouza/ctbz-cli/issues/428) | `rotinas reclassificar`: `idLancamento` 0 / `idPendencia` null ao concluir a rotina | contabilidade |
| ✅ | [#435](https://github.com/edusouza/ctbz-cli/issues/435) | `--classificacao` por nome ignora maiúsculas e escolhe a primeira opção | contabilidade |
| ✅ | [#543](https://github.com/edusouza/ctbz-cli/issues/543) | `conciliacao resolver` não confere se os ids são pendências abertas | pendências |
| ✅ | [#573](https://github.com/edusouza/ctbz-cli/issues/573) | `contas-bancarias editar` envia `dataSaldoInicial` 1969-12-31 sem `--abertura` | documentos |
| ✅ | [#619](https://github.com/edusouza/ctbz-cli/issues/619) | `--pendencia` repetido envia o mesmo documento duas vezes | documentos |
| ✅ | [#628](https://github.com/edusouza/ctbz-cli/issues/628) | `documentos enviar`: pendência sem competência vai como `{"mes":0,"ano":0}` | documentos |
| ✅ | [#652](https://github.com/edusouza/ctbz-cli/issues/652) | `sem-aplicacao`: pendência sem competência vai como `{"mes":0,"ano":0}` | documentos |
| ✅ | [#660](https://github.com/edusouza/ctbz-cli/issues/660) | `tomadores adicionar` de documento já cadastrado sobrescreve o cadastro | notas fiscais |
| ✅ | [#716](https://github.com/edusouza/ctbz-cli/issues/716) | `notas entrada reclassificar --como` envia as cinco quantidades zeradas | notas fiscais |
| ✅ | [#729](https://github.com/edusouza/ctbz-cli/issues/729) | `prolabore definir`: campo ausente tira a empresa da gestão inteligente | pró-labore |
| ✅ | [#770](https://github.com/edusouza/ctbz-cli/issues/770) | `lucros distribuir`: total que valida a soma trata `nil` como 0 | pró-labore |
| ✅ | [#777](https://github.com/edusouza/ctbz-cli/issues/777) | `lucros distribuir`: sócio sem `id` vai como `idSocio: null` | pró-labore |
| ✅ | [#779](https://github.com/edusouza/ctbz-cli/issues/779) | `lucros distribuir`: ajuste de centavo absorve diferenças que não são arredondamento | pró-labore |
| ✅ | [#818](https://github.com/edusouza/ctbz-cli/issues/818) | `usuarios`: id como `any` e `tipo` ausente desligam a trava do administrador | conta |

Fora de escopo: achados de prioridade alta que não enviam escrita errada (v1.10 em diante).

## v1.10 — Saída segura: caracteres de controle, CSV, segredos e dados pessoais

Épico: [#944](https://github.com/edusouza/ctbz-cli/issues/944) · 50 sub-issues

Texto vindo da API chega ao terminal, ao CSV, ao prompt de confirmação e às mensagens de erro sem filtro de
caracteres de controle, ANSI e bidi, e o CSV não neutraliza fórmulas. Segredos e dados pessoais aparecem no
`--dry-run`, no `--verbose` e no `acoes.jsonl`. A maior parte se resolve na camada, não comando a comando.

Frentes: filtro central de caracteres de controle em `internal/output`, no prompt e em `WriteError`; injeção de
fórmula no CSV; segredos no `--dry-run`, no `--verbose` e nos erros (ADR da política do verbose primeiro, #914);
dados pessoais no `acoes.jsonl` (ADR).

## v1.11 — Escrita segura e `--dry-run`

Épico: [#945](https://github.com/edusouza/ctbz-cli/issues/945) · 78 sub-issues

O que a CLI mostra e confere antes de enviar: confirmação de risco alto que omite custo ou promete reversão,
alvo mal identificado, risco declarado abaixo da ADR-0018, nulo exibido como R$ 0,00, estado já aplicado,
escolha ambígua do alvo, ordem da exclusão no `--substituir` e o `--dry-run` que aborta nas pré-leituras em vez
de simular (decisão em ADR, aplicada a todos os comandos).

## v1.12 — Depois do envio: releitura, situação e orientação

Épico: [#946](https://github.com/edusouza/ctbz-cli/issues/946) · 66 sub-issues

O que a CLI diz e devolve depois de enviar: uma regra comum de "resultado incerto" para 5xx, 2xx ilegível e
falha na segunda requisição; releitura que falha sai com `situacao: enviado` e aviso, em vez de erro; `situacao`
conferida no relido; "pode repetir" só numa lista de permissão; chave ausente na Central não vira "nada a fazer".

## v1.13 — Validação de entrada

Épico: [#947](https://github.com/edusouza/ctbz-cli/issues/947) · 54 sub-issues

Resolução por nome ou código sem ambiguidade, `parseValor`, datas com fuso da máquina, ids repetidos ou crus,
flags ignoradas em silêncio e validação de arquivos e downloads.

## v1.14 — Leitura de dados e contratos da API

Épico: [#948](https://github.com/edusouza/ctbz-cli/issues/948) · 96 sub-issues

Começa por religar o monitor semanal dos contratos ao vivo ([#942](https://github.com/edusouza/ctbz-cli/issues/942)).
Depois: nulo lido como R$ 0,00 ou `false`, erro HTTP tratado como "vazio", paginação e truncamento, parse que
descarta erro em silêncio, contratos nunca verificados numa conta real e chaves de saída já publicadas
(documentar ou criar alias, sem quebrar a ADR-0017).

## v1.15 — Testes

Épico: [#949](https://github.com/edusouza/ctbz-cli/issues/949) · 75 sub-issues

Uma matriz de testes para cada escrita (`--dry-run`, recusa do `confirmo`, sem terminal, 4xx, 5xx e releitura
com falha), o mecanismo de goldens e o gravador multipart, fixtures irreais e asserções fracas.

## v1.16 — Documentação e skills

Épico: [#950](https://github.com/edusouza/ctbz-cli/issues/950) · 75 sub-issues

"Recomendação para a CLI" desatualizada em `docs/api/escrita`, guias sem chaves de saída, `situacao`,
`--yes`/`--dry-run`/`confirmo`, ajuda e README, ADRs, critérios de investigação fechados sem registro e as
skills.

## v1.17 — Infra, manutenção e notas de release

Épico: [#951](https://github.com/edusouza/ctbz-cli/issues/951) · 85 sub-issues

Actions fixadas por SHA e `persist-credentials: false`, catálogo do monitor, ordem de `Escritas()`
([#412](https://github.com/edusouza/ctbz-cli/issues/412), que absorveu as repetições por PR), `operacao.ID` com
argumento cru, comentários errados e código morto, o OTP com Gmail real (#9) e o que faltou nas notas de release
e nas seções já publicadas deste roadmap.

## Fora do roadmap de escrita

Endpoints documentados em [docs/api/escrita](docs/api/escrita/README.md), sem comando previsto:

- **Pagamentos e cartões:** pagar imposto ou fatura com cartão, cadastrar cartão, débito automático, troca de plano
  (o cartão é cifrado no navegador pelo Adyen ou pela Iugu)
- **Certificado digital:** upload do A1, emissão, agendamento e remoção
- **NFS-e:** emitir, replicar, registrar e cancelar
- **Chamados:** não há API para abrir ou responder
- Marcações de tela e telemetria (modais vistos, tours, NPS, timeline de ativação)
