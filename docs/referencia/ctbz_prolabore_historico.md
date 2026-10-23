# ctbz prolabore historico

Lista o histórico mensal de pró-labore dos sócios

Lista, por competência, o pró-labore e os descontos (INSS e IRRF) de cada sócio, como a
tabela de histórico da central de pró-labore. A API devolve todo o histórico de um sócio de
uma vez; --ano filtra o resultado e --socio escolhe um sócio pelo ID (o padrão é todos).

## Uso

```
ctbz prolabore historico [flags]
```

## Exemplos

```sh
  ctbz prolabore historico
  ctbz prolabore historico --ano 2026 -o csv
```

## Flags

```
      --ano int     só as competências deste ano
      --socio int   ID do sócio (de "ctbz prolabore -o json"); padrão: todos
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz prolabore`](ctbz_prolabore.md).
