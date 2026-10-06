# ADR-0022: Classificar notas de entrada pela base `/api/emissor/`, reenviando os objetos como vieram

- **Status:** aceita
- **Data:** 2026-10-04

## Contexto

Há duas versões da tela de classificação de notas de entrada (issue #210):

- o front `/nota-entrada/`, com base `/api/emissor/classificacaonotas/`: lote
  (`salvarloteclassificacao/{tipo}`), por produto (`listarprodutos/{id}` e
  `salvarclassificacao/`) e reclassificação (`reclassificar?idNfe=`);
- a rota nova do painel, com base `/api/plataforma/nota-entrada/classificacao/`, onde só o lote
  está ativo ("Por Produto" vem desabilitado) e nenhum caminho está no catálogo.

A CLI já lê as listagens de `/api/emissor/` (verificadas, ver
[endpoints verificados](../api/endpoints-verificados.md)). Os corpos das escritas são os próprios
objetos que o front recebeu: o lote envia as notas da listagem `tipo=0` como vieram, e o por
produto envia os produtos de `listarprodutos` com as quantidades alteradas. Nenhum desses
objetos foi capturado com dados reais: só os campos citados no front são conhecidos.

## Decisão

1. Usar a base `/api/emissor/` para ler e escrever a classificação: é a única com lote, por
   produto e reclassificação, e é a que a CLI já lê.
2. Reenviar os objetos **como vieram** do servidor (`json.RawMessage`), mudando só as chaves de
   quantidade no por produto. Os tipos (`NotaEntrada`, `ProdutoNota`) servem para mostrar e
   validar, não para montar o corpo.
3. Validar no cliente as regras do front para o por produto: nenhuma quantidade negativa e a
   soma igual a `quantidadeTotal`, em todos os produtos da nota.

## Consequências

- Campos que a CLI não conhece chegam ao servidor intactos, como no front; um campo novo na
  listagem não quebra a escrita.
- Se a base antiga for desligada, os comandos de classificação param. O monitor de contratos
  (listagem) e o de catálogo (`CTBZ_CATALOGO_ESCRITAS=1`, ver [monitoramento](../monitoramento.md))
  avisam; trocar a base então é mudar as constantes de caminho do lote, cujo corpo é o mesmo.
- Os nomes das opções mudam por nível no front (lote `USO_CONSUMO`/`ATIVO_IMOBILIZADO`, produto
  `quantidadeConsumo`/`quantidadeAtivo`). A CLI expõe um só vocabulário (`estoque`, `insumo`,
  `uso-consumo`, `ativo-imobilizado`, `prestacao-servico`) e traduz para cada nível.

## Alternativas consideradas

- **Base nova `/api/plataforma/nota-entrada/classificacao/`** — só tem o lote, não está no
  catálogo e não foi lida nenhuma vez; adotá-la exigiria manter duas bases para o por produto.
- **Montar o corpo a partir dos tipos Go** — perderia campos desconhecidos dos objetos, que o
  servidor pode usar (o front os reenvia todos).
