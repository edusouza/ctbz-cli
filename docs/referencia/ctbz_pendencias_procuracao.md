# ctbz pendencias procuracao

Mostra a pendência de procuração do e-CAC ou declara que ela foi criada

Sem --ja-criei, mostra se há pendência de procuração eletrônica (SEM_PROCURACAO,
EM_EXPIRACAO…), o CNPJ a quem outorgar e o link do e-CAC. Nada é enviado.

A procuração é criada (ou renovada) no e-CAC da Receita, fora da Contabilizei. Depois disso,
--ja-criei declara que ela existe ("Já criei a procuração" no painel), pelo caminho da origem
da pendência: Central de Rotinas ou checklist de primeiros passos (risco médio: é uma
declaração; se a procuração não existir, a pendência deve voltar).

Pede confirmação (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz pendencias procuracao [flags]
```

## Exemplos

```sh
  ctbz pendencias procuracao
  ctbz pendencias procuracao --ja-criei
```

## Flags

```
      --dry-run    mostra a requisição que seria enviada, sem enviar
      --ja-criei   declara que a procuração já foi criada no e-CAC
  -y, --yes        envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz pendencias`](ctbz_pendencias.md).
