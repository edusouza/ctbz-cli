# ADR-0023: Licenciar como PolyForm Noncommercial com uso pela própria empresa

- **Status:** aceita
- **Data:** 2026-10-26

## Contexto

O repositório não tinha licença, o que significa "todos os direitos reservados": ninguém podia
usar o `ctbz` legalmente. O público natural da CLI são clientes da Contabilizei que administram
a própria empresa, o que já é uso comercial. Ao mesmo tempo, o autor não quer que outros
ganhem dinheiro com este código, em especial o fornecedor do serviço que a CLI acessa
(oferecendo serviços em cima dele) e escritórios ou integradores que operem contas de terceiros
ou empacotem as skills num produto. Fazer uma ferramenta própria, sem este código, continua
livre para qualquer um. O autor é o único titular dos direitos do código (os commits feitos com
agentes são dele), e as dependências (cobra, pflag, mousetrap, x/term, x/sys) são Apache-2.0
ou BSD-3-Clause.

## Decisão

Licenciar o repositório inteiro (código, skills, scripts e documentação) pela PolyForm
Noncommercial 1.0.0, com uma permissão adicional, escrita em português e com tradução para o
inglês, que torna finalidade permitida administrar empresa da qual o usuário seja titular,
sócio ou administrador, desde que não preste serviço remunerado a terceiros nem inclua o
software em produto ou serviço oferecido a terceiros. A permissão não se aplica ao fornecedor do
serviço acessado, ao seu grupo nem a quem age em nome deles; ele é descrito pelo papel, não pelo
nome. Os avisos das dependências vão no `THIRD_PARTY_NOTICES`, gerado por `go run ./tools/notices`
e incluído nos arquivos da release com o `LICENSE`.

## Consequências

- O `ctbz` deixa de poder ser chamado de open source; a documentação usa "source-available".
- A permissão adicional é texto próprio, não revisado por advogado: o resumo em
  [Licença](../licenca.md) não substitui o `LICENSE`, e uma revisão jurídica é recomendável.
- Aplicar a licença depende do autor; ela dá base para notificar ou processar, não fiscaliza.
- Contribuições externas passam a ser aceitas sob os termos do `LICENSE`, com o autor podendo
  relicenciá-las ([Contribuindo](../contribuindo.md#licenca-das-contribuicoes)). Se elas se
  tornarem frequentes, adotar um CLA formal.
- Licenças comerciais pagas continuam possíveis caso a caso, porque a PolyForm não impede o
  licenciante de dar outras licenças.

## Alternativas consideradas

- **AGPL-3.0 (com ou sem licença comercial)** — é open source, mas uma empresa pode usar uma
  CLI internamente, ou prestar serviço com ela, sem descumprir a AGPL.
- **PolyForm Noncommercial sem permissão adicional** — barra o uso para administrar a própria
  empresa, que é o uso principal.
- **PolyForm Small Business** — libera empresas pequenas a prestar serviço remunerado com o
  `ctbz`.
- **PolyForm Internal Use** — permite que o fornecedor do serviço use a CLI nas próprias
  operações.
- **Business Source License 1.1** — pensada para bloquear uso em produção por um prazo e
  depois abrir; exigiria o mesmo texto próprio como "Additional Use Grant" e ainda converteria
  o código em licença permissiva no futuro.
- **Functional Source License** — só proíbe produtos concorrentes; prestar serviço com a CLI
  continuaria permitido.
