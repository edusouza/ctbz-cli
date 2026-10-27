# ADR-0024: Usar `--saldo-final` só no lugar de um saldo lido zero ou ausente

- **Status:** aceita
- **Data:** 2026-10-27

## Contexto

Ao concluir a importação de um OFX (`eventoUploadExtrato`), vai o saldo do último dia em
`SALDO_ULTIMO_MES`. O painel o preenche com o `saldoUltimoDia` que `info-extrato` leu do
arquivo, num campo editável. Desde a v1.5 (#202), `--saldo-final` era só uma conferência: igual
ao lido, seguia; diferente, a importação parava.

No uso real (Banco Inter, 09/2026), `info-extrato` devolveu `saldoUltimoDia: 0` para um OFX com
`LEDGERBAL` 719,17, e a CLI não tinha como enviar o saldo certo. A #1032 passou a enviar sempre o
valor da flag, com um aviso no stderr. O review (#1043) mostrou o custo: um `--saldo-final`
digitado errado, de outra conta ou de outro mês, era gravado no balancete com código 0, sem
pergunta, e o saldo lido não zero, que é o melhor indício do erro, era descartado. A divergência
só é conhecida depois do upload (e, no `--substituir`, depois da exclusão), quando a confirmação
já passou.

## Decisão

Com `--saldo-final`:

- igual ao lido: vai o lido;
- lido ausente ou R$ 0,00: vai o da flag, com aviso no stderr (o caso que motivou a mudança);
- lido diferente e não zero: no terminal, sem `--yes`, a CLI mostra os dois valores e pergunta
  se vai o da flag; com `--yes` ou sem terminal, a importação para antes da conclusão (código 1),
  como antes da #1032.

O resumo da confirmação mostra o saldo de `--saldo-final`, e o resultado ganha `saldo_lido` e
`saldo_divergente` (campos aditivos, ADR-0017), para quem lê o JSON saber que o valor enviado
não é o lido.

## Consequências

- O caso do Banco Inter (lido zero) importa por script com o saldo certo.
- Um `--saldo-final` errado continua parando a importação quando a Contabilizei leu um saldo,
  como antes; o comportamento público da flag só muda onde antes não havia saída (lido ausente,
  que parava sem alternativa, e lido zero).
- Para trocar um lido não zero sem terminal não há flag: o usuário confirma no terminal ou
  corrige o campo no painel. Um agente precisa pedir ao usuário.
- Com `--substituir`, a parada pela divergência acontece depois da exclusão do extrato antigo; a
  confirmação de risco alto avisa disso, e a mensagem diz em que situação a competência ficou.

## Alternativas consideradas

- **Sempre o valor da flag (a #1032)**: grava erros de digitação sem pergunta e muda uma flag
  pública numa versão de correção.
- **Flag `--forcar-saldo` para trocar um lido não zero sem terminal**: mais uma flag para um caso
  não observado; fica para quando aparecer.
- **Terminar com código 4 (atenção) depois de enviar o da flag**: a escrita já teria sido feita;
  o código só avisaria depois.
- **Ler o saldo antes da confirmação**: o `info-extrato` só existe depois do upload, que já é uma
  escrita.
