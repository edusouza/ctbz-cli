# ctbz extratos importar

Importa o extrato bancário (OFX ou PDF) de uma conta num mês

Importa o extrato do mês de uma conta bancária, como a tela "Importar extrato" do painel
(risco médio: entra na contabilidade; desfaça com ctbz extratos excluir).

Antes de enviar, a CLI confere a extensão (.ofx ou .pdf), o tamanho (até 30 MB), os formatos
aceitos pelo banco, se a conta é integrada (essas não precisam de importação), se há um
período anterior pendente e se a competência já foi importada ou está em processamento (o
painel também não deixa importar por cima). Depois:

  1. envia o arquivo;
  2. no OFX, lê o saldo do último dia que a Contabilizei encontrou e pede para confirmar
     (ou compara com --saldo-final; diferente, a importação para);
  3. conclui a importação.

Ao final, a CLI relê a competência: em processamento, avisa para conferir depois; importada
sem nenhum lançamento, avisa no stderr.

Para refazer a importação de um mês já importado (ex.: um extrato vazio criado pelo
processamento de outro mês, ou o arquivo errado), use --substituir: a CLI exclui o extrato
atual e importa o arquivo, com uma confirmação só, como excluir e importar no painel. Se o
extrato atual tem lançamentos, o risco é alto (as classificações se perdem; digite
"confirmo"); se está vazio, é médio. Se os contadores já concluíram a classificação, o painel
não deixa excluir, e a CLI também não. Se a exclusão der certo e a importação falhar, o mês
fica pendente: corrija e rode o comando de novo.

Se a importação parar depois do envio (saldo recusado, extrato de outra competência), o
arquivo fica enviado mas o extrato não é efetivado; corrija e rode o comando de novo.

Se a conexão cair ou o servidor responder com erro 5xx (502, 504…) no envio do arquivo, a CLI
relê a situação do extrato: se a competência continua como estava (ex.: PENDENTE), nada foi
efetivado e é seguro rodar o mesmo comando de novo (o arquivo vai com o mesmo nome e substitui
o anterior). Se isso acontecer ao concluir a importação, ou se a competência já aparecer como
IMPORTADO ou PROCESSANDO, a importação pode ter sido aplicada: confira com ctbz extratos e ctbz
extratos lancamentos antes de repetir, para não importar duas vezes.

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
  ctbz extratos importar extrato-setembro.pdf --conta-bancaria 1000000000000001 --competencia 2026-09 \
    --substituir --dry-run
```

## Flags

```
      --competencia string   competência AAAA-MM (obrigatória)
      --conta-bancaria int   id da conta bancária (ver ctbz contas-bancarias) (obrigatória)
      --dry-run              mostra a requisição que seria enviada, sem enviar
      --saldo-final string   saldo do último dia esperado (OFX); diferente do lido, a importação para
      --substituir           se o extrato do mês já foi importado, exclui-o antes e importa este arquivo (como excluir e importar no painel)
  -y, --yes                  envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz extratos`](ctbz_extratos.md).
