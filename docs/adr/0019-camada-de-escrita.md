# ADR-0019: Camada de escrita com `Sender`, conferência prévia da sessão e erro traduzido

- **Status:** aceita
- **Data:** 2026-10-04

## Contexto

A [ADR-0018](0018-escrita-com-confirmacao.md) exige uma única tentativa por escrita e a
sessão conferida antes do envio. A leitura usa `api.Getter` ([ADR-0009](0009-camada-api-tipada.md))
e refaz o login sozinha quando recebe `401`, repetindo a chamada; numa escrita, repetir
poderia duplicar um lançamento ou um aceite.

As escritas do painel (ver [Escrita](../api/escrita/README.md)) usam três formatos de
corpo: JSON, um objeto serializado como string JSON (`JSON.stringify`) e `multipart/form-data`
para uploads. (Correção, #1051: o `JSON.stringify` do front vira o próprio corpo, porque o axios
envia texto como está; pela rede vai o objeto, não uma string JSON. O `JSONString` foi removido.) Algumas respostas são texto puro. Os erros chegam em formatos variados:
texto (`400`), `{"message"}` (`403`), HTML (`404` de caminho incompleto),
`{"detalhes":[{"detalhe"}]}` (`400`, `406`, `560`).

## Decisão

- **`api.Sender`** ao lado do `Getter`: `Send` (JSON) e `SendMultipart` (campos e arquivos).
  A codificação do corpo (`EncodeBody`, `EncodeMultipart`, `JSONString`) e a decodificação da
  resposta (`DecodeResponse`: struct, `*string` ou `nil`) ficam em `internal/api`, para que
  qualquer implementação (sessão real, `--dry-run`, testes de requisição) envie exatamente
  os mesmos bytes.
- **Conferência prévia:** antes de cada escrita, a CLI faz `GET appbar/get` pela camada de
  leitura. Se a sessão expirou, o re-login automático (com `CTBZ_OTP_CMD`) acontece aqui,
  quando nada foi enviado. Sem re-login possível, a escrita não sai.
- **Uma tentativa:** a escrita é enviada uma vez. `401` na escrita vira "a escrita não foi
  aplicada; rode `ctbz login` e repita"; erro de rede vira "pode ou não ter sido aplicada;
  confira antes de repetir"; `5xx` é mostrado como erro. Nada é repetido.
- **Erro traduzido:** `ctbz.WriteError` monta a mensagem a partir do status (em português) e
  do corpo (`detalhes[0].detalhe`, `message`, `mensagem`, `erro`, `error`, string JSON ou
  texto), resumida em 300 caracteres. HTML vira uma explicação curta. Cabeçalhos e cookies
  nunca entram na mensagem.
- `ctbz api -X` (métodos diferentes de `GET`) usa a mesma conferência e a tentativa única.

## Consequências

- Cada escrita custa um `GET` a mais (~100 ms). Em troca, o caso mais comum de falha
  (sessão vencida num cron) é resolvido antes do envio.
- A janela entre o `GET` e a escrita é curta, mas existe: se a sessão vencer nela, o usuário
  recebe o erro de `401` e repete o comando.
- O Go não repete sozinho `POST`, `PUT`, `PATCH` nem `DELETE` sem chave de idempotência, então
  a tentativa única vale também na camada de transporte (há teste com conexão derrubada).

## Alternativas consideradas

- **Reaproveitar `authedAPI` com re-login depois do `401`** — repetiria a escrita.
- **Conferir a sessão pela data de expiração salva** — sem custo de rede, mas a sessão
  pode ser invalidada no servidor antes do prazo; o `GET` é a única confirmação real.
- **Mostrar o corpo inteiro do erro** — pode ser uma página HTML longa ou ecoar dados
  enviados; um resumo basta para o usuário agir.
