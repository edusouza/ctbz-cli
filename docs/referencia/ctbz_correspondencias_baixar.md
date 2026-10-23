# ctbz correspondencias baixar

Baixa uma correspondência do escritório virtual

Baixa uma correspondência (o ID está em ctbz correspondencias) como correspondencia-ID.pdf.

## Uso

```
ctbz correspondencias baixar ID [flags]
```

## Exemplos

```sh
  ctbz correspondencias baixar 501 --dir ~/Documentos
```

## Flags

```
      --dir string   diretório onde salvar (default ".")
      --force        sobrescreve o arquivo se existir
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz correspondencias`](ctbz_correspondencias.md).
