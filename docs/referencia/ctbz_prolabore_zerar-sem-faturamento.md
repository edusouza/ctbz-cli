# ctbz prolabore zerar-sem-faturamento

Liga ou desliga o pró-labore zerado nos meses sem faturamento

Liga (on) ou desliga (off) a preferência da empresa "Não quero ter pró-labore cadastrado
em meses sem faturamento". Vale para todos os sócios.

Com a preferência ligada, nos meses sem faturamento os sócios ficam sem pró-labore e não
contribuem para o INSS. Numa empresa na gestão inteligente, ligar vale a partir deste mês e
não altera os meses passados.

on é de risco alto (deixar de contribuir para o INSS) e off, de risco médio. Os dois são
reversíveis. Aceita --yes e --dry-run.

## Uso

```
ctbz prolabore zerar-sem-faturamento on|off [flags]
```

## Exemplos

```sh
  ctbz prolabore zerar-sem-faturamento on
  ctbz prolabore zerar-sem-faturamento off --yes
```

## Flags

```
      --dry-run   mostra a requisição que seria enviada, sem enviar
  -y, --yes       envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz prolabore`](ctbz_prolabore.md).
