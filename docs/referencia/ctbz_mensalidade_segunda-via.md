# ctbz mensalidade segunda-via

Gera a segunda via da mensalidade e mostra o Pix copia e cola e a linha digitável

Gera a segunda via da mensalidade em atraso, como o botão "Gerar segunda via s/ juros" do
card da fatura no painel, e mostra o que pagar: itens, total, vencimento, situação, Pix copia e
cola, linha digitável, código de barras e o link do checkout.

Risco médio: registra um boleto novo, com vencimento novo, mas não cobra nada. O caminho vem
da ação do card (ctbz mensalidade -o json, botao_adicional_acao); sem ela (ex.: fatura em
dia), o comando para com erro. Com --competencia, o card precisa oferecer a segunda via
daquele mês.

--pix e --linha imprimem só o código, sem decoração, para usar num pipe:

  inter-pj pix enviar --copia-e-cola "$(ctbz mensalidade segunda-via --pix --yes)"

O link do checkout tem o identificador da fatura, que dá acesso a ela sem login: ele sai só no
resultado, nunca no registro de ações. Depois de pagar, confirme com ctbz mensalidade situacao:
a fatura pública demora a dar baixa.

Pede confirmação (--yes em scripts) e aceita --dry-run, que mostra o GET sem enviar.

## Uso

```
ctbz mensalidade segunda-via [flags]
```

## Exemplos

```sh
  ctbz mensalidade segunda-via
  ctbz mensalidade segunda-via --competencia 2026-10 --yes -o json
  ctbz mensalidade segunda-via --pix --yes
  ctbz mensalidade segunda-via --linha --yes
```

## Flags

```
      --competencia string   competência AAAA-MM esperada; o card precisa oferecer a segunda via dela
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --linha                imprime só a linha digitável do boleto (para pipe)
      --pix                  imprime só o Pix copia e cola (para pipe)
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz mensalidade`](ctbz_mensalidade.md).
