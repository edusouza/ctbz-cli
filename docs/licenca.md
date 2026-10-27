# Licença

O `ctbz` não é open source no sentido da OSI: é **source-available**. O código, as skills e a
documentação ficam abertos para ler, usar e modificar, mas não para ganhar dinheiro com eles.
A licença é a [PolyForm Noncommercial 1.0.0](https://polyformproject.org/licenses/noncommercial/1.0.0/)
com uma permissão adicional para administrar a própria empresa. O texto que vale é o do
[`LICENSE`](https://github.com/edusouza/ctbz-cli/blob/main/LICENSE); este resumo não o substitui.

## Pode

- Usar para administrar a empresa de que você é titular, sócio ou administrador, mesmo que o
  uso dê lucro a essa empresa (consultar impostos, emitir notas, fechar o mês…). Sócios e
  empregados dela também podem, em benefício dela.
- Usar para estudo, pesquisa, testes e projetos pessoais sem fim comercial.
- Modificar e redistribuir, nos mesmos termos, mantendo o `LICENSE` e a linha `Required Notice`.

## Não pode

- Prestar serviço remunerado a terceiros com o `ctbz` ou com uma versão modificada dele, como
  um escritório ou uma assessoria operando a conta de clientes.
- Incluir o `ctbz`, as skills ou partes deles num produto ou serviço oferecido a terceiros.
- Usar, se você for o fornecedor do serviço que o `ctbz` acessa, uma empresa do mesmo grupo ou
  alguém agindo em nome deles. Isso não impede ninguém de fazer a sua própria ferramenta: a
  licença protege este código, não a ideia.

Qualquer outro uso comercial precisa de autorização por escrito do autor. Os motivos da escolha
estão no [ADR-0023](adr/0023-licenca-nao-comercial-com-uso-pela-propria-empresa.md).

## Material de terceiros

A licença não cobre o que não é do autor. Os trechos do JavaScript do painel, os nomes, as
marcas e os formatos de requisição e resposta da Contabilizei citados na documentação e nas
fixtures de teste pertencem à Contabilizei e aparecem só para documentar como a CLI conversa
com o painel.

## Software de terceiros

O binário inclui a biblioteca padrão do Go e algumas dependências (Apache-2.0 e BSD-3-Clause).
As licenças delas vão no [`THIRD_PARTY_NOTICES`](https://github.com/edusouza/ctbz-cli/blob/main/THIRD_PARTY_NOTICES),
também incluído em cada release.
