# Dossiê da empresa

O dossiê é a memória do assessor entre sessões: o que não muda (ou muda pouco) e o que já
foi decidido. Com ele, o agente não redescobre a empresa a cada conversa.

## Onde

```sh
D="${CTBZ_HOME:-$HOME/.config/ctbz}/dossie"
CNPJ=$(ctbz status -o json | jq -r .cnpj)     # sem jq: leia o campo "cnpj" da saída
mkdir -p "$D" && chmod 700 "$D"
F="$D/$CNPJ.md"; touch "$F" && chmod 600 "$F"
```

Um arquivo por CNPJ. Se o usuário trocar de empresa (`ctbz empresa usar`), troque de arquivo.

## O que nunca gravar

Senhas, códigos OTP, código de acesso do Simples, usuário/senha de prefeitura ou Dataprev,
tokens, dados de cartão, CPF completo dos sócios. Use nomes e os IDs da Contabilizei.

## Como montar (primeira sessão)

Rode em paralelo e resuma:

```sh
ctbz empresa -o json
ctbz empresa atividades -o json
ctbz prolabore -o json            # sócios e seus ids (--socio)
ctbz prolabore fator-r -o json    # se alguma atividade depende do Fator R
ctbz contas-bancarias -o json     # ids (--conta-bancaria) e se é integrada
ctbz plano -o json
```

## Modelo

```markdown
# Dossiê — RAZÃO SOCIAL (CNPJ 00000000000100)

Atualizado em: 2026-10-20

## Perfil
- Regime: Simples Nacional · Ramo: serviço · Início na Contabilizei: 03/2024
- Plano: … (categoria …)
- Atividades: 6204-0/00 Consultoria em TI (Anexo III ou V, depende do Fator R)
- Sujeita ao Fator R: sim · Motor do Fator R / gestão inteligente: sim
- Município/UF: … · Emissor de NFS-e: habilitado
- Certificado digital: válido até 2027-07-22

## Sócios
| Nome | ID (--socio) | Pró-labore | Gestão |
|---|---|---|---|
| FULANO | 1000000000000001 | R$ 1.621,00 | INTELIGENTE |

## Contas bancárias
| Banco | ID (--conta-bancaria) | Integrada | Formato do extrato |
|---|---|---|---|
| Contabilizei Conta PJ | 1000000000000001 | sim | — |

## Clientes recorrentes (tomadores)
- ACME LTDA (CNPJ …): ~R$ 8.000/mês

## Decisões e preferências
- 2026-09-10: pró-labore mantido na gestão inteligente (Fator R 28,7%).
- Guias em ~/Documentos/guias.

## Histórico de checkups
- 2026-10-20: 1 guia do mês (DAS R$ 1.000, vence 20/10), extrato de setembro pendente.

## Escritas feitas pelo agente
- 2026-10-20 `ctbz impostos confirmar 123` → confirmado (usuário pagou pelo app do banco).
```

## Quando atualizar

- Perfil: na criação e quando algo mudar (CNAE, sócio, conta, plano, certificado).
- Decisões: toda vez que o usuário escolher algo que vale para o futuro.
- Checkup: uma linha por checkup, com o essencial.
- Escritas: toda escrita enviada (data, comando sem segredos, `situacao`, motivo).
  `ctbz acoes` é o registro técnico; o dossiê guarda o porquê.

Mantenha o arquivo curto: resuma ou apague entradas de checkup com mais de 12 meses.
