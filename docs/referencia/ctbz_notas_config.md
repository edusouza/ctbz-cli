# ctbz notas config

Mostra a configuração do emissor de notas

Mostra se o emissor de notas está habilitado, a versão em uso, se a Contabilizei reporta
instabilidade com a prefeitura, a situação do certificado digital usado na emissão, se é
permitido emitir para o exterior e o município da empresa.

As alíquotas por atividade estão em "ctbz notas aliquotas".

## Uso

```
ctbz notas config
```

## Exemplos

```sh
  ctbz notas config
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz notas`](ctbz_notas.md).
