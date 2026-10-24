---
name: notas-e-clientes
description: Notas fiscais de serviço e clientes na Contabilizei pela CLI ctbz. Use para listar NFS-e emitidas, somar o faturamento por período ou por cliente, ver a concentração de clientes, cadastrar, corrigir ou excluir tomadores, consultar um CNPJ antes de emitir, conferir o emissor (instabilidade, certificado) e as alíquotas por atividade, e preparar a emissão de uma nota (que é feita no painel, porque a CLI não emite).
---

# Notas fiscais e clientes

Antes, siga a skill `assessor`. A CLI **não emite, cancela, replica nem agenda NFS-e**, e a
API não fornece o PDF nem o XML (a Contabilizei os envia por e-mail). O agente prepara tudo e
o usuário emite no painel ou no app. No plano Experts, o assessor da Contabilizei também pode
emitir (confira com `ctbz plano`).

## Notas emitidas e faturamento

```sh
ctbz notas --de 2026-01 --ate 2026-09 -o json   # até 24 meses; total no stderr
ctbz notas --tomador 11222333000181 -o json     # CPF/CNPJ filtra pelo documento; outro texto, pelo nome
ctbz notas --numero 101 -o json                 # --tomador e --numero não andam juntos
```

Campos: `numero`, `emissao`, `tomador`, `documento`, `valor`, `status`, `situacao`. Some só as
notas válidas: deixe de fora as canceladas (o `status`/`situacao` diz) e diga quantas excluiu.

Análises úteis (com `jq`, se houver):

```sh
# faturamento por cliente, do maior para o menor
ctbz notas --de 2026-01 --ate 2026-09 -o json |
  jq 'group_by(.documento) | map({tomador: .[0].tomador, total: (map(.valor) | add), notas: length}) | sort_by(-.total)'
# por mês
ctbz notas --de 2026-01 --ate 2026-09 -o json | jq 'group_by(.emissao[0:7]) | map({mes: .[0].emissao[0:7], total: (map(.valor) | add)})'
```

- **Concentração:** um cliente com mais de ~50% do faturamento é um risco de negócio; vale
  mencionar.
- Compare o total por mês com o faturamento de `ctbz impostos faturamento` (base do DAS).
  Diferenças podem ser notas canceladas, emitidas noutro sistema ou de competência diferente:
  aponte e sugira confirmar com a Contabilizei.
- Recebimento sem nota aparece na conciliação (skill `fechamento-mensal`).

## Tomadores (clientes)

```sh
ctbz notas tomadores -o json                     # id, nome, documento, email, municipio, uf, exterior
ctbz notas tomadores consulta 00000000000191 -o json   # dados na Receita: situacao_cadastral, optante_simples…
```

Antes de cadastrar ou de emitir para um CNPJ novo, rode `consulta` e alerte se a
`situacao_cadastral` não for ativa.

Cadastrar e corrigir (risco médio, protocolo de escrita):

```sh
# CNPJ: razão social e endereço vêm da Receita
ctbz notas tomadores adicionar --documento 00000000000191 --email financeiro@cliente.com --dry-run
# CPF: informe nome, CEP e número
ctbz notas tomadores adicionar --documento 52998224725 --nome "Fulano de Tal" --cep 01001-000 --numero 10 --dry-run
# Exterior: país, cidade, logradouro e número
ctbz notas tomadores adicionar --exterior --nome "Example Inc" --pais US --cidade Springfield \
  --logradouro "Main St" --numero 100 --dry-run
ctbz notas tomadores editar 00000000000191 --email novo@cliente.com --dry-run   # exterior: pelo id
ctbz notas tomadores remover 00000000000191 --dry-run
```

`adicionar` recusa um documento já cadastrado (código 2) e indica `editar`: use `editar`, que
parte do cadastro atual, para corrigir um tomador existente. `--im` (inscrição municipal) só para
pessoa jurídica; `--telefone` só nacional.

## Emissor e alíquotas

```sh
ctbz notas config -o json      # emissor_habilitado, instabilidade, certificado_validade, certificado_vencido
ctbz notas aliquotas -o json   # por atividade: item_servico (LC 116), aliquota, aliquota_iss, fator_r, mercado
```

- `instabilidade: true` → a prefeitura está instável: melhor esperar para emitir.
- `certificado_vencido: true` ou `emissor_habilitado: false` → não dá para emitir; veja
  `ctbz certificado` (skill `cadastro-e-acessos`).
- `aliquota` é a do Simples para a atividade; `aliquota_iss` é a parte que é ISS. `mercado:
  externo` (cliente no exterior) não tem ISS.

## Preparar a emissão de uma nota

Monte para o usuário uma ficha pronta para copiar no emissor:

1. Tomador: cadastrado? (`notas tomadores`; senão, `adicionar`). CNPJ ativo? (`consulta`)
2. Atividade e item de serviço: de `notas aliquotas` (se houver mais de uma, pergunte qual).
3. Mercado: tomador do exterior → nota de exportação.
4. Valor, descrição do serviço e competência, confirmados com o usuário.
5. Emissor ok? (`notas config`: habilitado, sem instabilidade, certificado válido).
6. Imposto estimado: valor × `aliquota` (%). É uma estimativa: o DAS do mês considera a RBT12.

Depois da emissão, a nota aparece em `ctbz notas` (pode levar alguns minutos).

## Notas de entrada (compras)

Manifestação e classificação ficam na skill `fechamento-mensal` (seção 5).
