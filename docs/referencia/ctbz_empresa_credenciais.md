# ctbz empresa credenciais

Mostra os dados de acesso a órgãos públicos (código do Simples, prefeitura, Dataprev)

Mostra as credenciais que a Contabilizei usa nos órgãos públicos: código de acesso do
Simples Nacional, usuário e senha da prefeitura e do Dataprev, e se o painel pede para
redefinir a senha da prefeitura.

O código e as senhas aparecem só como "preenchido"; --mostrar exibe os valores, apenas num
terminal e na tabela. Para usuário administrador a Contabilizei recusa (403).

Para trocar: ctbz empresa credenciais atualizar; para responder ao alerta da prefeitura:
ctbz empresa credenciais confirmar-prefeitura.

## Uso

```
ctbz empresa credenciais [flags]
```

## Exemplos

```sh
  ctbz empresa credenciais
  ctbz empresa credenciais --mostrar
```

## Subcomandos

- [`ctbz empresa credenciais atualizar`](ctbz_empresa_credenciais_atualizar.md): Troca o código do Simples ou o usuário e a senha da prefeitura ou do Dataprev
- [`ctbz empresa credenciais confirmar-prefeitura`](ctbz_empresa_credenciais_confirmar-prefeitura.md): Responde "Sim, já atualizei a senha" ao alerta da senha da prefeitura

## Flags

```
      --mostrar   mostra o código e as senhas (só num terminal)
```

## Flags globais

```
  -h, --help            mostra a ajuda do comando
  -o, --output string   formato de saída: table, json ou csv (padrão: CTBZ_OUTPUT ou table)
```

Veja também: [`ctbz empresa`](ctbz_empresa.md).
