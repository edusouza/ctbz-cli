# Pistas para classificar lançamentos

A classificação final é **sempre** uma das opções de `ctbz extratos contas --competencia M`
(ou `ctbz caixa contas`). Use as pistas para escolher e o trecho entre aspas para buscar na
lista (`ctbz contas --busca TRECHO` também ajuda; acentos contam).

| No extrato aparece… | Provável natureza | Buscar | Observação |
|---|---|---|---|
| Pix/TED de cliente, com nota emitida | receita de serviço/venda | "receita", "venda", "serviço" | vira conciliação com a nota |
| Pix/TED de cliente, **sem** nota | receita sem nota | — | falta emitir a nota: avise o usuário |
| TED/Pix do próprio sócio | aporte, empréstimo do sócio ou devolução | "sócio", "capital", "empréstimo" | pergunte; exige `--socio` |
| Transferência para o sócio | pró-labore ou distribuição de lucros | "pró-labore", "lucros" | antecipação de lucros exige `--socio` |
| DAS, DARF, guia | imposto | "impostos" | no caixa exige `--guia ID` ou `--sem-guia` |
| Tarifa, pacote de serviços, IOF | despesa bancária | "tarifa", "banc" | |
| Mensalidade Contabilizei | contabilidade | "contabilidade" | costuma vir automática |
| Aluguel, condomínio | ocupação | "aluguel", "condom" | boleto misto: desmembrar |
| Software, SaaS, hospedagem, domínio | serviços/tecnologia | "software", "internet", "serviços" | |
| Energia, água, telefone, internet | utilidades | "energia", "telefone", "internet" | |
| Combustível, Uber, passagem | transporte/viagem | "viagem", "transporte", "combust" | só se for da empresa |
| Fornecedor com NF-e de compra | compra | "fornecedor", "mercadoria" | a nota aparece em `notas entrada` |
| Aplicação/resgate (CDB, fundo) | investimento | "aplicação", "investimento" | gera pendência de extrato de aplicação |
| Rendimento de aplicação | receita financeira | "rendimento", "financeira" | |
| Empréstimo, financiamento recebido | passivo | "empréstimo", "financiamento" | a rotina pode pedir o contrato |
| Estorno, devolução | reverte o lançamento original | — | use a classificação do original |
| Gasto pessoal do sócio na conta PJ | não é despesa da empresa | "sócio", "lucros" | trate como retirada do sócio e oriente a separar as contas |

Na dúvida, pergunte ao usuário em vez de chutar. Uma classificação errada muda o resultado
contábil e, numa conciliação, a natureza fiscal.
