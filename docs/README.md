# Documentação — engenharia reversa da plataforma Contabilizei

Tudo o que foi descoberto para operar a Contabilizei (`app.contabilizei.com.br`)
pela linha de comando, sem a interface web. A Contabilizei não publica uma API: o
que está aqui foi obtido observando o próprio site (HTML das telas de login e
bundles JavaScript do painel) e confirmado com chamadas reais em outubro de 2026.

## Mapa

| Contexto | Conteúdo |
|---|---|
| [Autenticação](autenticacao/README.md) | Fluxo completo de login: credenciais, OTP, seleção de empresa, cookies e dados de sessão |
| [OTP automático](otp-automatico/README.md) | Como entregar o código do e-mail à CLI sem digitar (Gmail + `gws`, outros provedores) |
| [API](api/README.md) | Bases de API, autenticação das chamadas, erros, [endpoints verificados](api/endpoints-verificados.md) e [catálogo completo](api/catalogo.md) |
| [Front-end](frontend/README.md) | Como o painel é construído (Vue), onde ficam os bundles, configuração exposta e outros fronts |
| [CLI](cli/README.md) | Arquitetura do `ctbz`, estados persistidos, códigos de saída e decisões de projeto |
| [Metodologia](metodologia/README.md) | Passo a passo reproduzível da investigação, armadilhas encontradas e cuidados |
| [Decisões (ADRs)](adr/README.md) | Por que o projeto é como é: arquitetura, processo, versionamento |
| [Contribuindo](contribuindo.md) | Convenções de código, testes, commits e definição de pronto |
| [Licença](licenca.md) | O que a licença permite (administrar a própria empresa) e o que não permite (lucrar com o `ctbz`) |

## Resumo em uma tela

```text
GET  /login                         → HTML com <input name="token">            (anti-CSRF)
POST /login  user,password,token    → tela "Insira o código" + cookie __C     (e-mail com OTP é enviado)
POST /login  "v=2&otp=NNNNNN"       → 206 + tela "Selecione a empresa"        (404 = código inválido)
POST /selecionarempresa cnpj,token  → cookie oauth-token + <script> que grava localStorage
GET  /api/plataforma/...            → JSON, autenticado pelos cookies __C + oauth-token
```

- Os dois cookies são obrigatórios; sem eles qualquer API responde **401**.
- A sessão dura **2 horas** (validade do `__C`) e não é renovada pelo uso.
- Cada login gera um código novo, e na prática **só o OTP mais recente vale**.

## Convenções desta documentação

- Valores pessoais (CPF, CNPJ, e-mails, endereços) foram trocados por exemplos fictícios.
- "Verificado" = chamado de verdade com uma sessão válida; "visto no front" = encontrado
  no JavaScript, mas não executado.
- Só foram executadas chamadas de leitura (`GET`). Nenhuma ação (emitir nota, pagar,
  aceitar termo) foi disparada durante a investigação.
