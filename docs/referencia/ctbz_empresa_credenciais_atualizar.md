# ctbz empresa credenciais atualizar

Troca o código do Simples ou o usuário e a senha da prefeitura ou do Dataprev

Troca as credenciais escolhidas, pedindo os valores sem eco no terminal (nunca por
argumento): --codigo-simples (12 dígitos), --prefeitura e --dataprev (usuário e senha, com
pelo menos 3 caracteres; Enter no usuário mantém o atual). As demais credenciais seguem como
estão. Fora de um terminal, os valores são lidos da entrada padrão, uma linha por pergunta.

Depois de salvar, a Contabilizei verifica o acesso e avisa por e-mail se for negado.

Risco alto: credencial errada faz falhar a entrega de declarações (multas e juros). Aceita
--yes e --dry-run (que mostra *** no lugar dos segredos).

## Uso

```
ctbz empresa credenciais atualizar [flags]
```

## Exemplos

```sh
  ctbz empresa credenciais atualizar --prefeitura
  ctbz empresa credenciais atualizar --codigo-simples --dataprev
```

## Flags

```
      --codigo-simples   troca o código de acesso do Simples Nacional
      --dataprev         troca o usuário e a senha do Dataprev
      --dry-run          mostra a requisição que seria enviada, sem enviar
      --prefeitura       troca o usuário e a senha da prefeitura
  -y, --yes              envia sem pedir confirmação (obrigatório sem terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz empresa credenciais`](ctbz_empresa_credenciais.md).
