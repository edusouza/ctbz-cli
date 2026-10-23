# ctbz mensalidade situacao

Indica se a empresa está em dia com a Contabilizei

Indica se a empresa está em dia com a Contabilizei (consulta de inadimplência do painel).
A API responde "OK" para empresas em dia; qualquer outra resposta é mostrada como veio e
tratada como "não está em dia".

Com --fail-on-inadimplencia, o comando termina com código 4 quando a empresa não está em dia.

## Uso

```
ctbz mensalidade situacao [flags]
```

## Exemplos

```sh
  ctbz mensalidade situacao
  ctbz mensalidade situacao --fail-on-inadimplencia || echo "Mensalidade pendente"
```

## Flags

```
      --fail-on-inadimplencia   termina com código 4 se a empresa não estiver em dia
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz mensalidade`](ctbz_mensalidade.md).
