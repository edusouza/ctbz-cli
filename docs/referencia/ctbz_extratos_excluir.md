# ctbz extratos excluir

Exclui o extrato importado de uma conta num mês

Exclui o extrato importado de uma conta na competência, por exemplo depois de enviar o
arquivo errado (risco alto). A importação volta a ficar pendente e as classificações e
informações adicionais das movimentações se perdem; não dá para desfazer, só reimportar e
reclassificar. Se os contadores já concluíram a classificação, o painel não deixa excluir, e a
CLI também não.

Antes de enviar, a CLI lê a situação da competência; se a leitura falhar, termina com erro
sem enviar nada (sem ela não dá para confirmar a exclusão depois).

Depois da exclusão, a competência fica PROCESSANDO com o botão "Importar" do painel e sem
lançamentos até um novo arquivo entrar: importe-o com ctbz extratos importar. Nesse estado
não há o que excluir, e o comando termina com erro sem enviar nada. O mesmo par (PROCESSANDO
com o botão) aparece nos minutos em que a Contabilizei processa um extrato recém-importado: se
extrato/info mostra lançamentos (ou não diz quantos), o comando também termina sem enviar
nada, mas manda esperar o processamento terminar e rodar de novo.

Com a competência em processamento sem o botão "Importar" (já visto depois de um arquivo
recusado), a Contabilizei pode aceitar a exclusão sem mudar nada. A confirmação avisa disso e,
se a competência estava e continua PROCESSANDO, o comando termina com erro em vez de informar a
exclusão: importe o arquivo pela tela "Importar extrato" do painel ou abra um chamado. Se ela
só passou a PROCESSANDO depois do envio, sem o botão, a exclusão vale: espere a competência sair
desse estado e confira com ctbz extratos antes de importar de novo.

Pede "confirmo" no terminal (--yes em scripts) e aceita --dry-run.

## Uso

```
ctbz extratos excluir [flags]
```

## Exemplos

```sh
  ctbz extratos excluir --conta-bancaria 1000000000000001 --competencia 2026-09
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
      --dry-run              mostra a requisição que seria enviada, sem enviar
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
      --verbose count   mostra as requisições no stderr, sem segredos (repita para incluir os corpos; ou CTBZ_DEBUG=1 ou 2)
```

Veja também: [`ctbz extratos`](ctbz_extratos.md).
