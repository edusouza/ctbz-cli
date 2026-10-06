# Documentos e certificado

## Central de documentos

```sh
ctbz documentos                                           # tipos aceitos e quantos foram enviados
ctbz documentos --tipo EXTRATO_BANCARIO_MOVIMENTACOES     # documentos enviados de um tipo
```

```text
Tipo                            Nome              Enviados  Última modificação
EXTRATO_BANCARIO_MOVIMENTACOES  Extrato bancário        13  10/09/2026
```

- Sem `--tipo`, a lista mostra os tipos da central de documentos (extratos, contratos de
  empréstimo e financiamento, estoque, informes de investimentos…) e quantos já foram
  enviados.
- Com `--tipo`, aparecem os arquivos enviados: competência, nome do arquivo, data de envio,
  descrição, valor (quando o tipo tem) e o link do arquivo.
- Enviar documentos continua sendo feito pelo painel (a CLI só lê, ver
  [ADR-0002](../adr/0002-somente-leitura-ate-1-0.md)).

## Enviar documentos pedidos pelas pendências

```sh
ctbz documentos pendentes                                          # o que as pendências pedem
ctbz documentos enviar aplicacao-agosto.pdf --pendencia 1000000000000601
ctbz documentos enviar aplicacoes.pdf --pendencia 1000000000000601 --pendencia 1000000000000602
```

```text
Pendência         Tipo                          Competência  Conta                                 Investimento
1000000000000601  EXTRATO_APLICACAO_FINANCEIRA  08/2026      Banco Exemplo ag. 1234 conta 123456   RENDA_FIXA
```

- Sem `--pendencia`, `pendentes` usa as pendências das rotinas do painel (coluna `pendencias` de
  `ctbz rotinas`).
- O tipo do documento e os metadados (competência, conta, investimento) vêm da pendência, nunca
  são digitados.
- Com várias `--pendencia` do mesmo tipo, o arquivo vai uma vez só para todas ("em um arquivo
  único" no painel).
- Risco médio: o documento entra na contabilidade e resolve a pendência; não há exclusão pela API.
- Extratos de movimentação bancária usam `ctbz extratos importar`.

## Certificado digital

```sh
ctbz certificado            # o mesmo que "ctbz empresa certificado"
```

```text
Situação:            VALIDO
Válido:              sim
Vencimento:          22/07/2027
Dias para vencer:    292
Pode renovar:        não
Vencido (painel):    não
Prazo finalizado:    não
Etapa da renovação:  PRE_CHECKOUT
Fluxo da renovação:  BASICO
```

- Validade do e-CNPJ usado pela Contabilizei, dias até o vencimento e se já é possível
  renovar.
- **Etapa da renovação** é o andamento da compra ou renovação feita pela Contabilizei:
  `PRE_CHECKOUT` antes de começar; durante o processo aparecem etapas como `verificacaoCnh`,
  `agendamentoVideoconferencia` e `agendamentoPresencial`.
- A senha e o arquivo do certificado nunca são consultados.
