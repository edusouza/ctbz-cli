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
  acesso, ID e classificação (nas listas de classificação).
- Notas de serviço **tomadas** não estão disponíveis: a tela antiga foi removida do site.

### Manifestar notas de entrada

```sh
ctbz notas entrada manifestar 1000000000000701 --confirmar --mes 2026-09
ctbz notas entrada manifestar 1000000000000702 --nao-realizada \
  --justificativa "Mercadoria extraviada antes da entrega"
```

- É a manifestação do destinatário: o evento que diz à SEFAZ se a empresa reconhece a compra.
  Escolha uma: `--ciencia` (ciência da operação, preliminar), `--confirmar` ("recebi"),
  `--desconhecer` ("desconheço") ou `--nao-realizada` ("não recebi", com `--justificativa`
  de 15 a 255 caracteres, regra da SEFAZ).
- Os IDs vêm da coluna ID de `ctbz notas entrada`; `--mes` diz em que mês procurar (padrão:
  o mês atual). Vários IDs vão num único envio.
- Só notas **PENDENTE** ou **CIENCIA** são manifestadas, como no painel; as demais aparecem
  como ignoradas e não são enviadas.
- Confirmar, desconhecer e não realizada são eventos fiscais **sem desfazer**: a CLI mostra o
  resumo e pede que você digite `confirmo`. A ciência é de risco médio (`[s/N]`).
- A Contabilizei comunica o evento à Receita Federal; em alguns minutos a nota aparece em
  `ctbz notas entrada --lista manifestadas`. No painel, baixar o XML ou o DANFE de uma nota
  pendente manifesta ciência automaticamente.
- Aceita `--yes` e `--dry-run` (ver [Escrita](escrita.md)).

### Classificar notas de entrada

```sh
ctbz notas entrada --lista a-classificar --mes 2026-09
ctbz notas entrada produtos 1000000000000701          # itens da nota e a distribuição atual
ctbz notas entrada classificar 1000000000000701 1000000000000702 --como uso-consumo --mes 2026-09
```

- A classificação diz à contabilidade para que serve a compra: `estoque`, `insumo`,
  `uso-consumo`, `ativo-imobilizado` ou `prestacao-servico`.
- O prazo é o **dia 05 do mês seguinte** à emissão; depois dele a Contabilizei confirma a
  pré-classificação sugerida.
- `classificar --como` classifica as notas inteiras, num único envio (o lote do painel). Só
  notas da lista "a classificar" do mês (`--mes`, padrão: o mês atual) são aceitas.
- Depois do envio a CLI relê as listas e mostra a classificação de cada nota; `enviado` quer
  dizer que ela ainda não apareceu nas listas.
- Risco médio: é contábil e pode ser refeito. Aceita `--yes` e `--dry-run`.
- Por que a CLI usa a base `/api/emissor/` e reenvia as notas como vieram: [ADR-0022](../adr/0022-classificacao-de-notas-de-entrada.md).

## PDF e XML das notas

A API do painel não oferece o PDF nem o XML das NFS-e emitidas: a Contabilizei envia o
documento por e-mail, com o link da prefeitura, quando a nota é autorizada. Detalhes da
investigação em [Notas fiscais: o que a API oferece](../api/notas-fiscais.md).
