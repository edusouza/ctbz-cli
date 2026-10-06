# ctbz correspondencias

Lista as correspondências recebidas no escritório virtual

Lista as correspondências recebidas no escritório virtual: ID, data, remetente, descrição,
situação e valor do envio. Sem o serviço contratado, avisa e termina com código 0.

Para baixar: ctbz correspondencias baixar ID; endereço de envio e autorização:
ctbz correspondencias endereco e autorizar.

## Uso

```
ctbz correspondencias
```

## Exemplos

```sh
  ctbz correspondencias
  ctbz correspondencias -o json
```

## Subcomandos

- [`ctbz correspondencias autorizar`](ctbz_correspondencias_autorizar.md): Autoriza ou não o recebimento de correspondências no endereço informado
- [`ctbz correspondencias baixar`](ctbz_correspondencias_baixar.md): Baixa uma correspondência do escritório virtual
- [`ctbz correspondencias endereco`](ctbz_correspondencias_endereco.md): Mostra ou troca o endereço de envio das correspondências

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
