# ctbz correspondencias endereco

Mostra ou troca o endereço de envio das correspondências

Sem flags, mostra o endereço para onde as correspondências são enviadas. Com --cep e
--numero (e --complemento, opcional), troca o endereço: logradouro, bairro e cidade vêm do CEP.
A mudança vale só para o envio das correspondências.

Risco médio: muda para onde vão documentos oficiais; reversível regravando o endereço
anterior. Aceita --yes e --dry-run.

## Uso

```
ctbz correspondencias endereco [flags]
```

## Exemplos

```sh
  ctbz correspondencias endereco
  ctbz correspondencias endereco --cep 01001-000 --numero 100 --complemento "sala 1"
```

## Flags

```
      --cep string           CEP do novo endereço
      --complemento string   complemento do novo endereço
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --numero string        número do novo endereço
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz correspondencias`](ctbz_correspondencias.md).
