# ctbz empresa credenciais confirmar-prefeitura

Responde "Sim, já atualizei a senha" ao alerta da senha da prefeitura

Declara que a senha da prefeitura cadastrada é válida e limpa o alerta "Redefina a senha
da prefeitura e atualize os dados de acesso". Se a senha mudou, use antes ctbz empresa
credenciais atualizar --prefeitura.

Risco médio: se a declaração for falsa, as obrigações municipais falham. Aceita --yes e
--dry-run.

## Uso

```
ctbz empresa credenciais confirmar-prefeitura [flags]
```

## Exemplos

```sh
  ctbz empresa credenciais confirmar-prefeitura
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

Veja também: [`ctbz empresa credenciais`](ctbz_empresa_credenciais.md).
