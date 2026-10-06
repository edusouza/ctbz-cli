# Notas fiscais

## Notas emitidas (NFS-e)

```sh
ctbz notas                                   # mês atual
ctbz notas --de 2026-01 --ate 2026-09        # um período (até 24 meses)
ctbz notas --tomador 11.222.333/0001-81      # por CPF/CNPJ do tomador
ctbz notas --tomador "ACME" --de 2026-07     # por nome do tomador
ctbz notas --numero 101
```

```text
Número  Emissão     Tomador    Documento                Valor  Status      Situação
101     04/09/2026  ACME LTDA  11.222.333/0001-81  R$ 1.000,00  AUTORIZADA  EMITIDA
Total: R$ 1.000,00 em 1 nota(s)
```

- A listagem é a do emissor de notas do painel (`novo-emissor/v2/listagem/notas/filtro`), uma
  consulta por mês e por página de 10 notas.
- `--tomador` decide sozinho: um CPF ou CNPJ (com ou sem pontuação) filtra pelo documento;
  qualquer outro texto filtra pelo nome. A API aceita um filtro por vez, por isso `--tomador`
  e `--numero` não podem ser usados juntos.
- Na tabela, o total do período vai para o stderr.
- Os campos das notas vêm do código do painel: a conta usada no desenvolvimento não tinha
  notas emitidas. Se algo vier diferente, abra uma issue com a saída de
  `ctbz api "novo-emissor/v2/listagem/notas/filtro?pagina=1&limite=10&ano=AAAA&mes=M"`.

## Tomadores (clientes)

```sh
ctbz notas tomadores                               # tomadores cadastrados no emissor
ctbz notas tomadores consulta 00.000.000/0001-91   # cadastro de um CNPJ na Receita
```

- `tomadores` lista nome, documento, e-mail, telefone, inscrição municipal, município, UF e
  se o tomador é do exterior. Os campos vêm do código do painel (a conta de desenvolvimento
  não tinha tomadores cadastrados).
- `consulta` usa a mesma busca que o emissor faz ao cadastrar um cliente: razão social, nome
  fantasia, abertura, atividade principal, natureza jurídica, situação cadastral, opção pelo
  Simples, endereço e contatos. Serve para conferir um cliente antes de emitir a nota.

### Cadastrar e editar tomadores

```sh
ctbz notas tomadores adicionar --documento 00.000.000/0001-91 --email financeiro@exemplo.com
ctbz notas tomadores adicionar --documento 529.982.247-25 --nome "Fulano de Tal" --cep 01001-000 --numero 10
ctbz notas tomadores adicionar --exterior --nome "Example Inc" --pais US --cidade Springfield \
  --logradouro "Main St" --numero 100
ctbz notas tomadores editar 00000000000191 --email novo@exemplo.com
```

- Com CNPJ, a razão social e o endereço vêm da consulta à Receita; as flags completam ou
  corrigem. Com CPF, informe nome, CEP e número (logradouro e bairro vêm do CEP).
- O documento é validado pelos dígitos verificadores, inclusive o CNPJ alfanumérico.
- O emissor salva o tomador nacional pelo documento: cadastrar um documento que já existe
  atualiza o cadastro.
- Tomadores do exterior são editados pelo id (coluna `id` de `ctbz notas tomadores`).
- Risco médio: os dados vão para as notas futuras; dá para editar de novo.

Para excluir um tomador (sem desfazer; recadastre se precisar):

```sh
ctbz notas tomadores remover 00000000000191
```

As notas já emitidas para o tomador não devem mudar; isso ainda não foi confirmado com a
Contabilizei.

## Configuração do emissor e alíquotas

```sh
ctbz notas config       # emissor habilitado, versão, instabilidade, certificado
ctbz notas aliquotas    # alíquota, ISS, Fator R e item de serviço por atividade
```

```text
Mercado  Tipo     CNAE       Atividade          Item   Descrição do item                         Alíquota (%)  ISS (%)  Fator R (%)  Anexo fixo
interno  servico  6204-0/00  Consultoria em TI  01.06  Assessoria e consultoria em informática.             6     2.01      33.3333  não
externo  servico  6204-0/00  Consultoria em TI  01.06  Assessoria e consultoria em informática.          3.05        0      33.3333  não
```

- `config` diz se o emissor está habilitado e se a Contabilizei reporta **instabilidade**
  com a prefeitura (quando é melhor esperar para emitir), além da validade do certificado
  digital usado na emissão.
- `aliquotas` traz, por atividade, o item da lista de serviços (LC 116) usado na nota, a
  alíquota do Simples, a parte que é ISS e o Fator R usado no cálculo. O mercado `externo`
  (tomador no exterior) não tem ISS.

## Notas de entrada (NF-e recebidas)

```sh
ctbz notas entrada                                    # a manifestar, mês atual
ctbz notas entrada --lista manifestadas --mes 2026-09
ctbz notas entrada --lista a-classificar --emitente "ACME"
```

- São as notas de compra emitidas contra o CNPJ da empresa, as mesmas da tela "Notas fiscais
  de entrada". `--lista` escolhe a aba: `a-manifestar` (padrão), `manifestadas`,
  `a-classificar` e `classificadas`.
- Colunas: emissão, emitente, CNPJ do emitente, valor, situação (ex.: Ciência), chave de
  acesso e ID.
- Manifestar (ciência, desconhecimento) e classificar (estoque, insumo, uso e consumo)
  continuam sendo feitos pelo painel.
- Notas de serviço **tomadas** não estão disponíveis: a tela antiga foi removida do site.

## PDF e XML das notas

A API do painel não oferece o PDF nem o XML das NFS-e emitidas: a Contabilizei envia o
documento por e-mail, com o link da prefeitura, quando a nota é autorizada. Detalhes da
investigação em [Notas fiscais: o que a API oferece](../api/notas-fiscais.md).
