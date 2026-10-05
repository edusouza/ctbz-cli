# ctbz impostos confirmar

Informa que uma guia foi paga (ou, com --nao-paguei, que não foi)

Marca uma guia como paga ("Já paguei" no painel) para que ela não apareça mais como a
pagar. Com --nao-paguei, informa que a guia não foi paga: ela volta para contas a pagar como
disponível, e o recálculo (ctbz impostos recalcular) passa a ser possível.

Risco médio: é uma declaração, não um pagamento. Uma confirmação falsa esconde um imposto
não pago. Essa confirmação não serve como comprovante oficial: a Contabilizei faz auditorias periódicas e pode atualizar o status com dados oficiais do governo. Desfaça com ctbz impostos desmarcar.

A rota segue a tela de impostos da empresa (impostos/rollout: v5 ou v3). Pendências
críticas na Central de Rotinas bloqueiam a confirmação, como no painel.
Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz impostos confirmar ID [flags]
```

## Exemplos

```sh
  ctbz impostos confirmar 1000000000000001
  ctbz impostos confirmar 1000000000000001 --nao-paguei --yes
```

## Flags

```
      --dry-run      mostra a requisição que seria enviada, sem enviar
      --nao-paguei   informa que a guia não foi paga
  -y, --yes          envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz impostos`](ctbz_impostos.md).
