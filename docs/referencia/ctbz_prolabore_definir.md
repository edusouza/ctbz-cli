# ctbz prolabore definir

Muda como o pró-labore de um sócio é definido (salário mínimo, teto, valor ou automático)

Muda a gestão do pró-labore de um sócio, como "Editar gestão" na central do sócio:

  salario-minimo  o salário mínimo vigente
  teto-inss       o teto do INSS
  personalizado   o valor de --valor, de pelo menos um salário mínimo
  inteligente     a Contabilizei calcula o valor todo mês (motor do Fator R); --minimo
                  define um piso opcional

O ID do sócio está em ctbz prolabore -o json. A Contabilizei decide a partir de que mês a
mudança vale: até o fechamento (em geral dia 25) vale para o mês atual; depois, para o
seguinte. A CLI avisa quando o painel diz que o mês está fechado.

Risco alto: muda INSS, IRRF e o Fator R; dá para alterar de novo enquanto o mês estiver
aberto. Aceita --yes e --dry-run.

## Uso

```
ctbz prolabore definir [flags]
```

## Exemplos

```sh
  ctbz prolabore definir --socio 1000000000000001 --tipo salario-minimo
  ctbz prolabore definir --socio 1000000000000001 --tipo personalizado --valor 3000,00
  ctbz prolabore definir --socio 1000000000000001 --tipo inteligente --minimo 1518,00
```

## Flags

```
      --dry-run         mostra a requisição que seria enviada, sem enviar
      --minimo string   piso da gestão inteligente, em reais (opcional)
      --socio int       ID do sócio (de ctbz prolabore -o json)
      --tipo string     salario-minimo, teto-inss, personalizado ou inteligente
      --valor string    pró-labore personalizado, em reais (3000,00)
  -y, --yes             envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz prolabore`](ctbz_prolabore.md).
