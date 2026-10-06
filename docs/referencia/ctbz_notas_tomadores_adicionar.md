# ctbz notas tomadores adicionar

Cadastra um tomador (cliente) nacional ou do exterior

Cadastra um tomador para as notas (risco médio: dá para editar depois).

Nacional (--documento): com CNPJ, a razão social e o endereço vêm da consulta à Receita, e as
flags só completam ou corrigem; com CPF, informe --nome, --cep e --numero (logradouro e bairro
vêm do CEP). O documento é validado pelos dígitos verificadores, inclusive o CNPJ
alfanumérico. Se o documento já estiver cadastrado, o cadastro é atualizado (o emissor salva
pelo documento).

Exterior (--exterior): informe --nome, --pais, --cidade, --logradouro e --numero.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz notas tomadores adicionar [flags]
```

## Exemplos

```sh
  ctbz notas tomadores adicionar --documento 00.000.000/0001-91 --email financeiro@exemplo.com
  ctbz notas tomadores adicionar --documento 529.982.247-25 --nome "Fulano de Tal" --cep 01001-000 --numero 10
  ctbz notas tomadores adicionar --exterior --nome "Example Inc" --pais US --cidade Springfield \
    --logradouro "Main St" --numero 100
```

## Flags

```
      --bairro string        bairro (nacional; padrão: o do CEP)
      --cep string           CEP (nacional)
      --cidade string        cidade (exterior)
      --complemento string   complemento
      --documento string     CPF ou CNPJ do tomador nacional
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --email string         e-mail
      --exterior             tomador do exterior (sem documento brasileiro)
      --im string            inscrição municipal (só pessoa jurídica)
      --logradouro string    logradouro (padrão: o do CEP)
      --nome string          razão social ou nome (CNPJ: padrão é o da Receita)
      --numero string        número
      --pais string          país: sigla (ex.: US), nome ou código (exterior)
      --telefone string      telefone (só nacional)
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz notas tomadores`](ctbz_notas_tomadores.md).
