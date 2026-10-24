# Skill `notas-e-clientes`

Notas fiscais de serviço (NFS-e) e clientes: listar as notas emitidas, somar o faturamento por
período ou por cliente, ver a concentração de clientes, cadastrar e corrigir tomadores, consultar
um CNPJ antes de emitir, conferir o emissor e as alíquotas, e preparar a emissão de uma nota.

**A CLI não emite, cancela, replica nem agenda NFS-e**, e a API não fornece o PDF nem o XML (a
Contabilizei os envia por e-mail). O agente prepara tudo e você emite no painel ou no app.

Instalação e configuração do plugin: [README das skills](../README.md).

## Quando usar

```text
> Quanto faturei de janeiro a setembro?
> Faturamento por cliente este ano.
> Algum cliente concentra muito do meu faturamento?
> Cadastra o cliente 00.000.000/0001-91 com o e-mail financeiro@cliente.com.
> Esse CNPJ está ativo?
> Prepara uma nota de R$ 8.000 para a ACME.
> O emissor está funcionando?
```

Pelo nome: `/ctbz:notas-e-clientes`.

## Pré-requisitos

- O que a skill [`assessor`](../assessor/README.md) pede (sessão válida no `ctbz`).
- `jq` ajuda nas análises por cliente e por mês (opcional).

## O que ela faz

### Notas emitidas e faturamento

```sh
ctbz notas --de 2026-01 --ate 2026-09 -o json   # até 24 meses; total no stderr
ctbz notas --tomador 11222333000181 -o json     # por CPF/CNPJ ou por nome
ctbz notas --numero 101 -o json
```

O agente soma só as notas válidas (deixa de fora as canceladas e diz quantas), agrupa por cliente
ou por mês e aponta:

- **concentração:** um cliente com mais de ~50% do faturamento é um risco de negócio;
- **diferenças** entre o total das notas e o faturamento usado no DAS (notas canceladas, emitidas
  noutro sistema, competência diferente), sugerindo confirmar com a Contabilizei.

Recebimentos sem nota aparecem na conciliação (skill
[`fechamento-mensal`](../fechamento-mensal/README.md)).

### Tomadores (clientes)

Antes de cadastrar ou de emitir para um CNPJ novo, o agente consulta a situação na Receita e
alerta se ele não estiver ativo:

```sh
ctbz notas tomadores consulta 00000000000191 -o json
```

Cadastro:

| Tipo | O que informar |
|---|---|
| CNPJ | só o documento (razão social e endereço vêm da Receita) e, opcionalmente, e-mail e inscrição municipal |
| CPF | nome, CEP e número |
| Exterior | `--exterior`, nome, país, cidade, logradouro e número |

```sh
ctbz notas tomadores adicionar --documento 00000000000191 --email financeiro@cliente.com --dry-run
ctbz notas tomadores editar 00000000000191 --email novo@cliente.com --dry-run
ctbz notas tomadores remover 00000000000191 --dry-run
```

Um documento já cadastrado não é cadastrado de novo: `adicionar` recusa e indica `editar`, que
corrige o cadastro atual sem apagar a inscrição municipal.

### Emissor e alíquotas

- `ctbz notas config`: emissor habilitado, instabilidade da prefeitura e certificado. Com
  instabilidade, o melhor é esperar; com certificado vencido ou emissor desabilitado, não dá para
  emitir (veja [`cadastro-e-acessos`](../cadastro-e-acessos/README.md)).
- `ctbz notas aliquotas`: por atividade, o item de serviço (LC 116), a alíquota do Simples, a
  parte de ISS e o mercado (cliente no exterior não tem ISS).

### Preparar a emissão

O agente monta uma ficha pronta para copiar no emissor:

1. tomador cadastrado e com CNPJ ativo;
2. atividade e item de serviço (pergunta, se houver mais de um);
3. mercado: cliente no exterior → nota de exportação;
4. valor, descrição e competência, confirmados com você;
5. emissor pronto (habilitado, sem instabilidade, certificado válido);
6. imposto estimado (valor × alíquota), lembrando que o DAS do mês considera a RBT12.

Depois da emissão, a nota aparece em `ctbz notas` em alguns minutos.

## O que pode alterar

| Ação | Risco | Desfazer |
|---|---|---|
| `notas tomadores adicionar` / `editar` / `remover` | médio | editar ou recadastrar |

Tudo com `--dry-run` primeiro e a sua confirmação (protocolo da skill `assessor`).

Notas de entrada (compras), manifestação e classificação ficam na skill
[`fechamento-mensal`](../fechamento-mensal/README.md).

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| [`SKILL.md`](SKILL.md) | instruções que o agente segue |
