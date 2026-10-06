# ctbz prolabore

Mostra o pró-labore vigente por sócio e o tipo de gerenciamento

Mostra o pró-labore da empresa como a central de pró-labore do painel: tipo de
gerenciamento (ex.: INTELIGENTE, quando a Contabilizei calcula o valor ideal), total, valor e
competências do card do painel e, para cada sócio, valor, se recebe pró-labore, se é o
responsável na Receita, data da última atualização e quantidade de dependentes.

O histórico mensal está em "ctbz prolabore historico"; para mudar a gestão de um sócio:
ctbz prolabore definir.

## Uso

```
ctbz prolabore
```

## Exemplos

```sh
  ctbz prolabore
  ctbz prolabore -o json | jq '.socios[] | {nome, valor}'
```

## Subcomandos

- [`ctbz prolabore definir`](ctbz_prolabore_definir.md): Muda como o pró-labore de um sócio é definido (salário mínimo, teto, valor ou automático)
- [`ctbz prolabore fator-r`](ctbz_prolabore_fator-r.md): Mostra a situação do Fator R e os anexos possíveis de cada atividade
- [`ctbz prolabore gestao-inteligente`](ctbz_prolabore_gestao-inteligente.md): Põe a empresa no cálculo automático do pró-labore ou a tira dele
- [`ctbz prolabore historico`](ctbz_prolabore_historico.md): Lista o histórico mensal de pró-labore dos sócios
- [`ctbz prolabore parametros`](ctbz_prolabore_parametros.md): Mostra os valores usados no cálculo do pró-labore (INSS e IRRF)
- [`ctbz prolabore zerar-sem-faturamento`](ctbz_prolabore_zerar-sem-faturamento.md): Liga ou desliga o pró-labore zerado nos meses sem faturamento

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz`](ctbz.md).
