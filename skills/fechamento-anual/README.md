# Skill `fechamento-anual`

O fechamento do ano: distribuição dos lucros do exercício entre os sócios, restrições que a
bloqueiam (débitos federais, pendência documental), carta de responsabilidade e termo de
débitos, reabertura do balanço, balanço patrimonial e o informe de rendimentos que cada sócio
usa na declaração do IRPF.

Quase tudo aqui é de **risco alto** e sem desfazer: o agente pede a sua confirmação explícita
para cada ação.

Instalação e configuração do plugin: [README das skills](../README.md).

## Quando usar

```text
> Distribui os lucros de 2025 meio a meio.
> O que falta para liberar o informe de rendimentos?
> Me passa os números do informe para a minha declaração do IR.
> Quanto cada sócio recebeu no ano?
> Explica o balanço de 2025.
```

Pelo nome: `/ctbz:fechamento-anual`.

`--ano` é o ano-calendário; o padrão é o ano anterior (o da declaração de IR). Calendário
típico: o exercício fecha em dezembro, a Contabilizei entrega a DEFIS e libera a distribuição e
o informe no começo do ano, a distribuição tem data-limite e a declaração do IRPF vai até o fim
de maio.

## Pré-requisitos

O que a skill [`assessor`](../assessor/README.md) pede (sessão válida no `ctbz`).

## Como funciona

### 1. Diagnóstico

O agente lê lucros do exercício, restrições do informe, débitos federais, termos pendentes e o
balanço do ano, e resume: lucro do exercício, quanto já foi distribuído, até quando dá para
alterar e o que bloqueia.

### 2. Resolver restrições primeiro

| Restrição | O que significa | Caminhos |
|---|---|---|
| Débitos federais | distribuir com débito não garantido gera multa de 50% do valor distribuído | regularizar (skill [`impostos`](../impostos/README.md)); aceitar o termo de débitos; ou decidir não distribuir |
| Divergência contábil-fiscal | contabilidade e fiscal não batem | falar com a Contabilizei antes de distribuir |
| Pendência documental | faltam documentos do ano | regularizar (às vezes reabrindo o balanço) ou não regularizar, o que **zera a distribuição** no informe |
| Carta de responsabilidade | o painel exige o aceite para liberar o informe | ler e aceitar |
| Reabertura do balanço em andamento | o balanço está sendo refeito | aguardar |

A CLI recusa reabrir o balanço quando a regularização é um serviço pago (mostra o custo e
indica contratar pelo painel).

### 3. Carta de responsabilidade e termo de débitos

O agente mostra o texto completo (`ctbz lucros informe carta --ano A`), explica em duas linhas o
que o sócio está declarando e só envia o aceite com a sua confirmação explícita.

### 4. Distribuir os lucros

1. Pergunta como dividir. O padrão é a proporção das quotas do contrato social; uma divisão
   desproporcional precisa estar prevista no contrato.
2. A soma precisa ser o lucro total do exercício; a CLI só acerta o centavo de arredondamento
   dos percentuais e recusa outras diferenças. Por percentual ou por valor:
   ```sh
   ctbz lucros distribuir --socio ID1=60% --socio ID2=40% --dry-run
   ctbz lucros distribuir --socio ID1=30000,00 --socio ID2=20000,00 --dry-run
   ```
3. Antes de confirmar valores altos, mostra as regras de 2026 em diante (Lei 15.270/2025):
   - mais de R$ 50 mil no mês para o mesmo sócio → retenção de 10% de IR sobre o valor total do
     mês;
   - lucros entram no imposto mínimo anual (IRPFM) de quem ganha mais de R$ 600 mil por ano;
   - só lucros apurados até 2025, com distribuição aprovada até 31/12/2025, ficam na regra
     antiga.

A CLI recusa sem exercício aberto, fora do prazo ou com restrições. O agente resolve a causa,
não contorna.

### 5. Informe de rendimentos (IRPF dos sócios)

```sh
ctbz lucros informe --ano 2025 -o json
ctbz lucros informe --ano 2025 -o csv > informe-2025.csv
```

O agente explica onde cada valor entra na declaração:

| Valor | Ficha do IRPF |
|---|---|
| pró-labore, INSS, IRRF retido | Rendimentos Tributáveis Recebidos de Pessoa Jurídica |
| 13º e IRRF do 13º | Rendimentos Sujeitos à Tributação Exclusiva |
| lucros isentos | Rendimentos Isentos e Não Tributáveis |

A API não gera o PDF do comprovante (é pelo painel). A participação na empresa (quotas) entra em
Bens e Direitos, pelo valor do capital social, que não vem da CLI.

### 6. Balanço e relatórios

`ctbz balanco`, `ctbz balancete` e `ctbz razao`, explicados para leigos (ativo = o que a empresa
tem; passivo = o que deve; patrimônio líquido = capital + lucros acumulados), com o razão para
explicar um saldo estranho.

### 7. Registrar

Lucro do exercício, distribuição feita, aceites e decisões vão para o dossiê, com data.

## O que pode alterar

| Ação | Risco | Consequência |
|---|---|---|
| `lucros distribuir` | **alto** | lucros declarados no IRPF dos sócios; refaz enquanto o exercício permitir |
| `lucros informe aceitar carta-responsabilidade` / `termo-debitos` | **alto** | sem desfazer |
| `lucros informe decidir …` | **alto** | sem desfazer; não regularizar a pendência zera a distribuição |
| `balanco reabrir --ano A` | **alto** | informe indisponível até a análise; sem desfazer |
| `pendencias aceitar CHAVE` | **alto** | declaração legal sem revogação |

## Arquivos

| Arquivo | Conteúdo |
|---|---|
| [`SKILL.md`](SKILL.md) | instruções que o agente segue |
