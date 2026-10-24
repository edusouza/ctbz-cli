---
name: fechamento-anual
description: Fechamento do ano na Contabilizei pela CLI ctbz. Use para distribuição de lucros do exercício entre os sócios, restrições (débitos federais, pendência documental), carta de responsabilidade e termo de débitos, reabertura do balanço, balanço patrimonial e informe de rendimentos dos sócios para a declaração do IRPF, e para responder quanto cada sócio recebeu no ano.
---

# Fechamento anual: lucros, balanço e informe de rendimentos

Antes, siga a skill `assessor` (sessão, dossiê e protocolo de escrita). Quase tudo aqui é de
**risco alto** e sem desfazer: confirmação explícita para cada ação. `--ano` é o ano-calendário
(padrão: o ano anterior, o da declaração de IR).

Calendário típico: o exercício fecha em dezembro; a Contabilizei entrega a DEFIS e libera a
distribuição e o informe no começo do ano. A distribuição tem `data_limite` (em `ctbz lucros`).
A declaração do IRPF dos sócios vai até o fim de maio.

## 1. Diagnóstico (em paralelo)

```sh
ctbz lucros -o json                           # exercicio, saldo, total_distribuido, limite_distribuicao,
                                              # pode_alterar, data_limite, motivo, restrições, socios[]
ctbz lucros informe restricoes --ano A -o json
ctbz impostos debitos -o json
ctbz pendencias termos -o json
ctbz balanco A -o json
```

Resuma: lucro do exercício (`saldo` + `total_distribuido`), quanto já foi distribuído, até
quando dá para alterar e o que bloqueia.

## 2. Resolver restrições primeiro

| Restrição | O que significa | Caminhos |
|---|---|---|
| `debitos_federais: true` | débitos na Receita/PGFN; distribuir com débito não garantido gera multa de 50% do valor distribuído | regularizar (skill `impostos`: pagar ou parcelar); aceitar o termo de débitos (`lucros informe aceitar termo-debitos`); ou decidir não distribuir (`decidir --nao-distribuir-lucros`) |
| `divergencia_contabil_fiscal: true` (em `restricoes`) | contabilidade e fiscal não batem | falar com a Contabilizei antes de distribuir |
| `pendencia_documental: true` | faltam documentos do ano | `decidir --regularizar-pendencia` (mostra onde enviar) e, se o caminho for `REABERTURA_BALANCO`, `balanco reabrir --ano A`; ou `decidir --nao-regularizar-pendencia` (**sem distribuição de lucros** no informe) |
| carta de responsabilidade pendente | o painel exige o aceite para liberar o informe | ler e aceitar (seção 3) |
| `reabertura_balanco` em andamento | o balanço está sendo refeito | aguardar; acompanhar em `ctbz pendencias` |

`balanco reabrir`: a CLI recusa quando a regularização é um serviço pago (e mostra o custo:
contrate pelo painel). Depois de pedida, o informe fica indisponível até a análise.

## 3. Carta de responsabilidade e termo de débitos

```sh
ctbz lucros informe carta --ano A          # texto completo: mostre ao usuário
ctbz lucros informe aceitar carta-responsabilidade --ano A --dry-run
ctbz lucros informe aceitar termo-debitos --ano A --dry-run   # só quando há débitos
```

Mostre o texto, explique em duas linhas o que o sócio está declarando e só envie com
confirmação explícita. `situacao: "enviado"` = já enviado: não repita.

A Central de Rotinas tem os seus próprios aceites (`ctbz pendencias termos`, `termo CHAVE`,
`aceitar CHAVE`), com o mesmo cuidado. O termo de débitos tem aceite tácito
(`prazo_aceite_tacito_dias`).

## 4. Distribuir os lucros

1. Pergunte como dividir. O padrão é a proporção das quotas do contrato social; uma
   distribuição desproporcional precisa estar prevista no contrato, então avise.
2. A soma precisa ser o **lucro total** (`saldo` + `total_distribuido`); a CLI só acerta o
   centavo de arredondamento dos percentuais (nunca para quem pediu 0%) e recusa qualquer
   outra diferença. IDs dos sócios: `socios[].id` de `ctbz lucros`.
   ```sh
   ctbz lucros distribuir --socio ID1=60% --socio ID2=40% --dry-run
   ctbz lucros distribuir --socio ID1=30000,00 --socio ID2=20000,00 --dry-run
   ```
   Sócios não citados ficam com zero.
3. A CLI recusa sem exercício aberto, com `pode_alterar: false` (mostra o `motivo`), depois da
   `data_limite` ou com restrições. Resolva a causa; não contorne. Quando a recusa é por dado
   ausente da Contabilizei (saldo, total distribuído ou id de sócio), a distribuição é pelo
   painel.
4. **Risco alto:** são os lucros que cada sócio declara no IRPF. Dá para refazer enquanto
   `pode_alterar` for verdadeiro. O exercício de 2026 em diante cai inteiro na regra nova
   (Lei 15.270/2025, art. 6º-A da Lei 9.250):
   - quando a empresa paga mais de R$ 50 mil no mês ao mesmo sócio, retém 10% de IR sobre o
     **valor total** do mês (R$ 80 mil → R$ 8 mil retidos);
   - os lucros entram no imposto mínimo anual (IRPFM) de quem tem renda acima de R$ 600 mil
     por ano, chegando a 10% a partir de R$ 1,2 milhão;
   - só os lucros apurados até 2025 com distribuição aprovada até 31/12/2025 ficam na regra
     antiga.

   Mostre essas regras antes de confirmar valores altos e recomende planejar o calendário de
   pagamentos com o contador.

## 5. Informe de rendimentos (IRPF dos sócios)

```sh
ctbz lucros informe --ano A -o json
ctbz lucros informe --ano A -o csv > informe-A.csv     # se o usuário quiser guardar
```

Explique onde cada valor entra na declaração do sócio:

| Campo | Ficha do IRPF |
|---|---|
| `rendimentos` (pró-labore), `previdencia` (INSS), `irrf_retido` | Rendimentos Tributáveis Recebidos de Pessoa Jurídica (fonte pagadora: o CNPJ da empresa) |
| `decimo_terceiro`, `irrf_decimo_terceiro` | Rendimentos Sujeitos à Tributação Exclusiva (13º) |
| `lucros_isentos` | Rendimentos Isentos e Não Tributáveis: lucros e dividendos recebidos |

A API não gera o PDF do comprovante: são os números. Para o PDF, o painel. A participação na
empresa entra em Bens e Direitos (quotas de capital), pelo valor do capital social, que não vem
da CLI.

## 6. Balanço e relatórios

```sh
ctbz balanco A -o json            # ativo, passivo e PL em dezembro, comparado ao ano anterior
ctbz balanco AAAA-MM -o json      # posição num mês
ctbz balancete AAAA-MM -o json    # inclui contas de resultado
ctbz razao --conta 1.01 --de A-01 --ate A-12 -o json
```

Para leigos: ativo = o que a empresa tem; passivo = o que deve; patrimônio líquido = capital +
lucros acumulados. Filtre `nivel` 1 ou 2 para um resumo. Use o razão para explicar um saldo
estranho.

## 7. Registrar

No dossiê: lucro do exercício, distribuição feita (percentuais), aceites e decisões, com data.
