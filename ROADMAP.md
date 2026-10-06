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

Labels usadas: `épico`, `versão: vX.Y`, `contexto: …`, `tipo: funcionalidade|investigação|infra|documentação` e, nas
escritas, `risco: baixo|médio|alto`.

## Visão geral

| Versão | Contexto | Épico | Funcionalidades | Situação |
|---|---|---|---|---|
| **v0.1** | Login e dados da empresa | [#1](https://github.com/edusouza/ctbz-cli/issues/1) | 8 | em andamento (7/8) |
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
| ⬜ | [#9](https://github.com/edusouza/ctbz-cli/issues/9) | Validar OTP automático com Gmail real (gws) | investigação |

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

## Fora do roadmap de escrita

Endpoints documentados em [docs/api/escrita](docs/api/escrita/README.md), sem comando previsto:

- **Pagamentos e cartões:** pagar imposto ou fatura com cartão, cadastrar cartão, débito automático, troca de plano
  (o cartão é cifrado no navegador pelo Adyen ou pela Iugu)
- **Certificado digital:** upload do A1, emissão, agendamento e remoção
- **NFS-e:** emitir, replicar, registrar e cancelar
- **Chamados:** não há API para abrir ou responder
- Marcações de tela e telemetria (modais vistos, tours, NPS, timeline de ativação)
