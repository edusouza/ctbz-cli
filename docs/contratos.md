# Testes de contrato

Como a CLI detecta mudanças na API da Contabilizei. Decisões em
[ADR-0009](adr/0009-camada-api-tipada.md) e [ADR-0010](adr/0010-testes-de-contrato.md).

## Rodar

```sh
go test ./...                                          # contratos contra as fixtures
CTBZ_CONTRACT_LIVE=1 go test ./internal/api -run Live -v   # contra a API real (precisa de ctbz login)
```

Saída de uma quebra:

```text
--- FAIL: TestContractsLive/dadosempresa
    empresaAtual.cnpj: campo removido (esperado texto)
    empresaAtual.plano: tipo mudou (esperado texto, veio objeto)
```

## Adicionar um endpoint

1. Em `internal/api/<contexto>.go`: constante `PathXxx`, tipo da resposta só com os campos
   usados e função `BuscarXxx(ctx, g)`.
2. Registrar em `Endpoints()` (`internal/api/api.go`). `LivePath` vazio quando o caminho
   depende de um ID de outra resposta.
3. Gerar a fixture: `go run ./tools/capture NOME` (ou `-from resposta.json`, ou `-path CAMINHO`).
   A captura poda a resposta aos campos do tipo e depois anonimiza
   ([ADR-0011](adr/0011-fixtures-podadas-ao-contrato.md)); `-full` mantém tudo.
4. **Revisar a fixture** (checklist abaixo) e rodar `go test ./...`.

## Checklist de revisão de fixture

- [ ] Nenhum nome de pessoa ou empresa real, CPF, CNPJ, e-mail, telefone, endereço, conta bancária
- [ ] Nenhum valor monetário real (aparecem como `1000` ou `1234.56`)
- [ ] Nenhum texto livre real (assuntos, mensagens, observações viram `TEXTO EXEMPLO`)
- [ ] Nenhuma URL assinada ou token
- [ ] Se algo passou, acrescentar uma regra em `internal/contract/anon.go` (com teste) e capturar de novo

## Campos opcionais

Quando a API omite um campo em parte das respostas (ex.: itens de menu sem `children`), marque
`contract:"optional"` no campo. Campos que podem vir `null` não precisam de marcação.

## Campos não verificados

Quando a conta usada para capturar não tem exemplos de um campo (ex.: uma lista sempre vazia),
declare-o como `json.RawMessage`. O contrato aceita qualquer valor nele, a poda o mantém
inteiro (a anonimização continua valendo) e o comando o mostra como a API o devolve
(`output.FromJSON`). Quando houver dados reais, troque por um tipo.

## Requisições de escrita

Para escrita, o risco está na **requisição**: caminho errado, campo com nome errado ou valor
com o sinal trocado. Como escritas nunca são chamadas de verdade em teste
([ADR-0018](adr/0018-escrita-com-confirmacao.md)), a verificação é offline:

```sh
go test ./internal/api -run Requisicoes            # compara com os goldens
go test ./internal/api -run Requisicoes -update    # regrava (revise o diff!)
```

Cada escrita registrada em `Escritas()` (`internal/api/escrita.go`) tem uma chamada de
exemplo com valores fictícios. O teste a executa contra um `httptest.Server`, pela mesma pilha
da CLI (`api.EncodeBody`/`EncodeMultipart` e `ctbz.Client`), e compara com
`internal/api/testdata/requisicoes/<nome>.json`:

```json
[
  {
    "metodo": "POST",
    "caminho": "/api/plataforma/caixa/lancamentousuario/novo/",
    "content_type": "application/json",
    "corpo": {"data": "2026-09-15", "valor": -150.25}
  }
]
```

- Multipart é comparado por campo: texto com o valor, arquivo com nome, tipo e tamanho
  (sem o conteúdo).
- Corpos enviados como string JSON aparecem como string no golden.
- `TestRequisicoesCobertura` quebra se uma escrita não tem golden ou se sobrou golden sem
  escrita.
- Os goldens são montados a partir do que o front envia (ver [Escrita](api/escrita/README.md)),
  sem dados pessoais.
- O modo ao vivo (`CTBZ_CONTRACT_LIVE`) continua só com `GET`: percorre `Endpoints()`, nunca
  `Escritas()`.

### Adicionar uma escrita

1. Em `internal/api/<contexto>.go`: tipo da requisição (ex.: `NovoLancamentoCaixa`) e função
   que envia (ex.: `SalvarLancamentoCaixa(ctx, s Sender, req)`).
2. Registrar em `Escritas()` com um exemplo de valores fictícios.
3. `go test ./internal/api -run Requisicoes -update`, conferir o golden contra
   `docs/api/escrita/` e rodar `go test ./...`.

## Fixtures sintéticas

Endpoints que ainda não foram capturados de uma conta real têm tipo e fixture escritos a
partir de [docs/api/escrita](api/escrita/README.md), marcados com `// sintética (ADR-0021)`
em `Endpoints()` ([ADR-0021](adr/0021-tipos-e-fixtures-sinteticos.md)). Para trocar pela real:

```sh
grep -n "sintética" internal/api/api.go     # pendentes
go run ./tools/capture NOME                 # captura, poda e anonimiza; revise e tire o comentário
```

## Monitoramento

Os contratos ao vivo rodam toda semana no job de monitoramento, que abre uma issue quando
algum quebra. Ver [Monitoramento da API](monitoramento.md).
