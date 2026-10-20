# Risco de cada escrita

O risco é o que a CLI declara. No terminal, o médio pede `[s/N]` e o alto pede que se digite
`confirmo`; o agente, sem terminal, segue o protocolo da `SKILL.md` e envia com `--yes`.
"Desfazer" diz se há como voltar atrás pela própria CLI.

## Baixo

| Comando | Desfazer |
|---|---|
| `primeiros-passos concluir ETAPA` | não |
| `primeiros-passos dispensar` / `reativar` | um desfaz o outro |

## Médio

| Comando | Desfazer / observação |
|---|---|
| `caixa adicionar` / `editar` | editar de novo ou `caixa remover` |
| `caixa remover ID` | não (a CLI mostra o `caixa adicionar` que recria) |
| `extratos importar ARQ` | `extratos excluir` (risco alto) |
| `extratos classificar` / `desmembrar` / `desfazer-desmembramento` | sim, enquanto os contadores não fecharem o período |
| `rotinas reclassificar --classificacao` | a rotina não reabre; a classificação ainda muda com `extratos classificar` |
| `impostos confirmar ID [--nao-paguei]` | `impostos desmarcar ID`. É declaração, não pagamento |
| `pendencias procuracao --ja-criei` | não; a pendência volta se a procuração não existir |
| `contas-bancarias adicionar` / `editar` | editar de novo |
| `notas tomadores adicionar` / `editar` / `remover` | editar ou recadastrar |
| `notas entrada manifestar --ciencia` | evolui para confirmar/desconhecer |
| `notas entrada classificar` / `reclassificar` | `reclassificar` |
| `documentos enviar ARQ --pendencia ID` | não há exclusão pela API |
| `usuarios desativar` / `ativar` | um desfaz o outro |
| `empresa credenciais confirmar-prefeitura` | não |
| `correspondencias endereco --cep …` / `autorizar off` | regravar / `autorizar on` |
| `prolabore zerar-sem-faturamento off` | `on` |

## Alto (confirmação explícita do usuário para cada um)

| Comando | Consequência |
|---|---|
| `impostos recalcular ID --vencimento D` | serviço em geral cobrado na mensalidade (`cobrar_recalculo`); não cancela |
| `impostos parcelamento contratar TIPO` | confissão de dívida; parcelas com Selic + 1%; cancelado se a 1ª não for paga; sem desfazer |
| `pendencias aceitar CHAVE` | declaração legal sem revogação (texto impresso antes) |
| `pendencias conciliacao resolver` | define a natureza fiscal do recebimento; pode ser refeita, não apagada |
| `extratos excluir` | perde classificações; só reimportando |
| `contas-bancarias remover ID` | permanente; vínculos se perdem |
| `notas entrada manifestar --confirmar/--desconhecer/--nao-realizada` | evento fiscal na SEFAZ, sem desfazer |
| `documentos sem-arquivo` / `sem-aplicacao` | declaração que fecha pendências, sem desfazer |
| `prolabore definir` | muda INSS, IRRF e Fator R; dá para mudar de novo com o mês aberto |
| `prolabore zerar-sem-faturamento on` | sócios sem INSS nos meses sem faturamento (aposentadoria, auxílios) |
| `prolabore gestao-inteligente ativar` / `sair` | muda quem define o pró-labore |
| `lucros distribuir` | rendimentos isentos declarados no IRPF dos sócios |
| `lucros informe aceitar` / `decidir` | sem desfazer; `--nao-regularizar-pendencia` zera a distribuição |
| `balanco reabrir --ano A` | informe indisponível até a análise; sem desfazer |
| `usuarios convidar EMAIL` | acesso total, sem revogar antes do aceite |
| `empresa credenciais atualizar` | credencial errada faz falhar declarações (multa); pede segredos sem eco: usuário roda no terminal |
| `conta alterar` / `conta senha` | pode tirar o acesso; exige OTP; senha só no terminal do usuário |
| `correspondencias autorizar on` | envio de correspondências pode ser cobrado |

## Fora da CLI

`ctbz api -X POST|PUT|PATCH|DELETE` envia sem confirmação, sem `--dry-run` e sem validação.
Não use por iniciativa própria.
