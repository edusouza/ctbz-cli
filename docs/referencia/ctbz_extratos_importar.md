# ctbz extratos importar

Importa o extrato bancário (OFX ou PDF) de uma conta num mês

Importa o extrato do mês de uma conta bancária, como a tela "Importar extrato" do painel
(risco médio: entra na contabilidade; desfaça com ctbz extratos excluir).

Antes de enviar, a CLI confere a extensão (.ofx ou .pdf), o tamanho (até 30 MB), os formatos
aceitos pelo banco, se a conta é integrada (essas não precisam de importação) e se há um
período anterior pendente. Depois:

  1. envia o arquivo;
  2. no OFX, lê o saldo do último dia que a Contabilizei encontrou e pede para confirmar
     (ou compara com --saldo-final; diferente, a importação para);
  3. conclui a importação.

Se a importação parar depois do envio (saldo recusado, extrato de outra competência), o
arquivo fica enviado mas o extrato não é efetivado; corrija e rode o comando de novo.

Pede confirmação (--yes em scripts) e aceita --dry-run, que mostra o envio e a conclusão sem
enviar nada.

## Uso

```
ctbz extratos importar ARQUIVO [flags]
```

## Exemplos

```sh
  ctbz extratos importar extrato-setembro.ofx --conta-bancaria 1000000000000001 --competencia 2026-09
  ctbz extratos importar extrato.ofx --conta-bancaria 1000000000000001 --competencia 2026-09 \
    --saldo-final 1.234,56 --yes
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --saldo-final string   saldo do último dia esperado (OFX); diferente do lido, a importação para
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz extratos`](ctbz_extratos.md).
