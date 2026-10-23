# ctbz notas

Lista as notas fiscais de serviço (NFS-e) emitidas

Lista as NFS-e emitidas pelo emissor da Contabilizei: número, data de emissão, tomador,
documento, valor, status e situação. Na tabela, o total do período vai para o stderr.

O período vai de --de até --ate (meses AAAA-MM; padrão: o mês atual), no máximo 24 meses.
--tomador filtra pelo nome do tomador ou, se for um CPF/CNPJ, pelo documento; --numero
filtra pelo número da nota.

O PDF e o XML da nota não estão disponíveis pela API: a Contabilizei os envia por e-mail
quando a prefeitura autoriza a nota.

## Uso

```
ctbz notas [flags]
```

## Exemplos

```sh
  ctbz notas
  ctbz notas --de 2026-01 --ate 2026-09 -o csv > notas.csv
  ctbz notas --tomador 11.222.333/0001-81
  ctbz notas --tomador "ACME" --de 2026-07
```

## Subcomandos

- [`ctbz notas aliquotas`](ctbz_notas_aliquotas.md): Lista as alíquotas e os códigos de serviço por atividade
- [`ctbz notas config`](ctbz_notas_config.md): Mostra a configuração do emissor de notas
- [`ctbz notas entrada`](ctbz_notas_entrada.md): Lista as notas fiscais de entrada (NF-e recebidas pela empresa)
- [`ctbz notas tomadores`](ctbz_notas_tomadores.md): Lista os tomadores (clientes) cadastrados no emissor de notas

## Flags

```
      --ate string       último mês, AAAA-MM (padrão: --de ou o mês atual)
      --de string        primeiro mês, AAAA-MM (padrão: o mês atual)
      --numero string    número da nota
      --tomador string   nome do tomador ou CPF/CNPJ
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz`](ctbz.md).
