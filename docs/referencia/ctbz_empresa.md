# ctbz empresa

Mostra os dados da empresa selecionada

Mostra razão social, CNPJ, situação, regime tributário, plano, certificado digital
e as outras empresas do usuário.

A resposta crua da API está em "ctbz api dadosempresa/get".

## Uso

```
ctbz empresa [flags]
```

## Exemplos

```sh
  ctbz empresa
  ctbz empresa -o json | jq -r .certificado_validade
```

## Subcomandos

- [`ctbz empresa atividades`](ctbz_empresa_atividades.md): Lista os CNAEs da empresa e os anexos do Simples Nacional
- [`ctbz empresa certificado`](ctbz_empresa_certificado.md): Mostra a situação do certificado digital da empresa e da renovação
- [`ctbz empresa credenciais`](ctbz_empresa_credenciais.md): Mostra os dados de acesso a órgãos públicos (código do Simples, prefeitura, Dataprev)
- [`ctbz empresa socios`](ctbz_empresa_socios.md): Lista os sócios da empresa
- [`ctbz empresa usar`](ctbz_empresa_usar.md): Troca a empresa da sessão

## Flags

```
      --json   atalho para -o json
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
