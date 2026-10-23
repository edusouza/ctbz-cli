# API

As APIs ficam no mesmo host do site (`https://app.contabilizei.com.br`) e são as mesmas
que o painel Vue usa. Não há documentação oficial, versionamento público nem garantia de
estabilidade: os caminhos podem mudar quando o front for atualizado.

- [endpoints-verificados.md](endpoints-verificados.md): chamados de verdade, com status e formato da resposta
- [catalogo.md](catalogo.md): todos os endpoints encontrados no JavaScript (gerado por script)
- [notas-fiscais.md](notas-fiscais.md): o que existe (e o que não existe) para baixar notas
- [escrita/](escrita/README.md): endpoints que alteram dados (corpos, riscos e como desfazer)

## Bases

O front cria uma instância axios por base (todas com `withCredentials: true`, exceto HubSpot):

| Base | Uso | Situação |
|---|---|---|
| `/api/plataforma/` | BFF do painel: dashboard, impostos, pró-labore, notas, certificado, documentos… | ✅ quase tudo passa por aqui |
| `/api/legado/` | monólito antigo (notas, consulta de CNPJ, artigos de ajuda) | visto no front |
| `/api/multiusuario/` | usuários e convites da conta | visto no front |
| `/api/fintech/` | conta digital | visto no front |
| `/api/emissor/` | notas de entrada: manifestação, classificação, XML e DANFE (front `/nota-entrada/`) | ✅ listagens chamadas |
| `/api/pagamentos/` | pagamentos | declarado, sem chamadas encontradas |
| `/api/public/` | endpoints públicos | declarado, sem chamadas encontradas |
| `/api/leads/hubspot/` | envio de eventos de marketing (sem cookies) | visto no front |

A mesma rota não existe em bases diferentes: `menu/get` responde em `/api/plataforma/`
e dá 404 em `/api/legado/`.

## Autenticação das chamadas

Só cookies (`__C` + `oauth-token`). Não há cabeçalho `Authorization`, CSRF ou assinatura.

```sh
curl -sS --compressed \
  -H 'Cookie: __C=…; oauth-token=…' \
  -H 'Accept: application/json' \
  https://app.contabilizei.com.br/api/plataforma/dadosempresa/get
```

Com a CLI:

```sh
ctbz api dadosempresa/get                 # relativo → /api/plataforma/dadosempresa/get
ctbz api /api/legado/empresa/cnpj?cnpj=…  # absoluto → qualquer base
ctbz api //api/legado/empresa/cnpj?cnpj=… # o mesmo, em qualquer shell (inclusive Git Bash)
ctbz api -X POST -d @corpo.json caminho   # corpo JSON (Content-Type: application/json)
```

No Git Bash do Windows (MSYS2), um argumento que começa com `/` é convertido em caminho do
Windows antes de chegar à CLI: `/api/legado/…` vira `C:/Program Files/Git/api/legado/…`, e a
Contabilizei responderia 404. A CLI recusa esse caminho e explica. Use `//api/…` (o Git Bash
troca `//` por `/`, e a CLI aceita `//` nos outros shells) ou rode com `MSYS_NO_PATHCONV=1`.

A empresa da sessão é a escolhida no login: **nenhuma rota recebe o CNPJ como parâmetro**
para dados da própria empresa.

## Parâmetros

O front passa parâmetros de dois jeitos:

- **no caminho**, por concatenação. Por exemplo
  `relatorios-ms/gerar-balancete/{ano}/{mes}`, `caixa/listpaginada/{ano}/{mes}/{porPagina}/{pagina}`,
  `impostos/v5/impostos-a-pagar/guia/{id}`. No catálogo, esses caminhos terminam em `/`.
- **na query** (`params` do axios). Por exemplo
  `impostos/v2/historico-impostos/guias?pagina=1&status=&mes=&ano=`.

Para descobrir os parâmetros de um endpoint, procure o caminho no JavaScript e leia a
função ao redor (ver [metodologia](../metodologia/README.md)).

## Respostas e erros

| Status | Corpo | Significado |
|---|---|---|
| 200 | JSON | sucesso; algumas rotas devolvem texto puro (ex.: `"OK"`) |
| 400 | texto (`timeout`, `Lista de tipos…`) | parâmetro faltando ou erro de validação |
| 401 | — | sem sessão ou sessão expirada |
| 403 | `{"message": …}` | rota existe, mas não está liberada para o perfil |
| 404 | **HTML** do SPA | caminho inexistente ou faltam segmentos no caminho |
| 560 | `{"identificador", "detalhe", "dataHora"}` | erro de negócio/serviço da aplicação (código próprio) |

Um 404 com HTML (`<meta charset="utf-8">…`) quase sempre indica que faltam parâmetros
de caminho, e não que a rota não existe.

## Valores e datas

- Datas aparecem como `"dd/mm/aaaa"` (strings) ou epoch em milissegundos, conforme a rota.
- Valores monetários aparecem como número ou dentro de objetos de exibição
  (`{"label": 1234.56, "descricao": null}`), formato comum nas rotas `v5`, que já vêm
  prontas para a interface.
- CNPJ às vezes vem formatado (`dadosempresa/get → empresaAtual.cnpj`), às vezes só com dígitos.
