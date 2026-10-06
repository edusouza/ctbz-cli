# ctbz notas tomadores editar

Corrige os dados de um tomador cadastrado

Corrige os dados de um tomador (risco médio: dá para editar de novo). A CLI lê o cadastro
atual e aplica só as flags informadas. Nacionais são identificados pelo documento; tomadores
do exterior, pelo id (coluna id de ctbz notas tomadores). O documento não muda.

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz notas tomadores editar DOCUMENTO|ID [flags]
```

## Exemplos

```sh
  ctbz notas tomadores editar 00000000000191 --email novo@exemplo.com
  ctbz notas tomadores editar 1000000000000002 --cidade Boston
```

## Flags

```
      --bairro string        bairro (nacional; padrão: o do CEP)
      --cep string           CEP (nacional)
      --cidade string        cidade (exterior)
      --complemento string   complemento
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --email string         e-mail
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
