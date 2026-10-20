#!/usr/bin/env python3
"""Compara cenários de pró-labore para uma atividade sujeita ao Fator R (Anexo III x V).

Estimativa de apoio: o cálculo oficial é o da Contabilizei (`ctbz impostos calculo`) e o
valor exato mês a mês é o da gestão inteligente. Só usa a biblioteca padrão.

Exemplos:
    ctbz impostos faturamento -o json > fat.json
    ctbz impostos tabela-irrf -o json > irrf.json
    python3 simular.py --faturamento-json fat.json --irrf-json irrf.json --socios 1

    python3 simular.py --rbt12 360000 --folha12 36000 --prolabore 3000 --socios 2

Com menos de 12 meses de atividade, a receita e a folha somadas são anualizadas (média dos
meses × 12, LC 123, art. 18, § 2º). Os meses de atividade vêm de --meses-atividade ou, com
--faturamento-json, contam a partir do primeiro mês com faturamento ou pró-labore.
"""

import argparse
import json
import math
import re
import sys

# LC 123/2006, anexos III e V (redação da LC 155/2016): (limite da faixa, alíquota, dedução).
ANEXOS = {
    "III": [
        (180_000, 0.060, 0),
        (360_000, 0.112, 9_360),
        (720_000, 0.135, 17_640),
        (1_800_000, 0.160, 35_640),
        (3_600_000, 0.210, 125_640),
        (4_800_000, 0.330, 648_000),
    ],
    "V": [
        (180_000, 0.155, 0),
        (360_000, 0.180, 4_500),
        (720_000, 0.195, 9_900),
        (1_800_000, 0.205, 17_100),
        (3_600_000, 0.230, 62_100),
        (4_800_000, 0.305, 540_000),
    ],
}
FATOR_R_MINIMO = 0.28
SUBLIMITE = 3_600_000
TETO = 4_800_000

# IRRF mensal de 2026 (Lei 15.270/2025): redutor = REDUTOR_FIXO − REDUTOR_FATOR × rendimento,
# para rendimentos entre a isenção e --irrf-redutor-ate.
REDUTOR_FIXO = 978.62
REDUTOR_FATOR = 0.133145


def aliquota_efetiva(anexo, rbt12):
    if rbt12 <= 0:
        return ANEXOS[anexo][0][1]
    for limite, nominal, deducao in ANEXOS[anexo]:
        if rbt12 <= limite:
            break
    # Acima de R$ 4,8 milhões vale a última faixa (limite, nominal e dedução do laço).
    return (rbt12 * nominal - deducao) / rbt12


def irrf(bruto, inss, a):
    """IRRF mensal sobre o pró-labore de um sócio, pela tabela de `ctbz impostos tabela-irrf`."""
    if not a.tabela_irrf or bruto <= a.irrf_isento_ate:
        return 0.0
    deducoes = max(inss + a.dependentes * a.deducao_dependente, a.desconto_simplificado)
    base = bruto - deducoes
    devido = 0.0
    for faixa in a.tabela_irrf:
        if base >= faixa["base"]:
            devido = base * faixa["aliquota"] - faixa["deducao"]
    if bruto <= a.irrf_redutor_ate:
        devido -= max(REDUTOR_FIXO - REDUTOR_FATOR * bruto, 0.0)
    return max(devido, 0.0)


def ler_tabela_irrf(caminho):
    """Lê a saída de `ctbz impostos tabela-irrf -o json`."""
    if not caminho:
        return []
    with open(caminho, encoding="utf-8") as f:
        linhas = json.load(f)
    tabela = []
    for linha in linhas:
        tabela.append(
            {
                "base": inicio_da_faixa(linha.get("base_calculo")),
                "aliquota": numero(linha.get("aliquota")) / 100,
                "deducao": numero(linha.get("deducao")),
            }
        )
    return sorted(tabela, key=lambda f: f["base"])


def inicio_da_faixa(texto):
    """Início da faixa em textos como "até R$ 2.428,80" (0) ou "de R$ 2.428,81 até R$ 2.826,65"."""
    texto = str(texto or "").strip().lower()
    if texto.startswith("até"):
        return 0.0
    valores = re.findall(r"\d[\d.,]*", texto)
    return numero(valores[0]) if valores else 0.0


def numero(v):
    """Lê números da CLI (1234.56) e textos em reais ("R$ 1.234,56", "1.234", "27,5%", "-10,5").

    Com vírgula e ponto, o último separador é o decimal; só com vírgula, ela é o decimal; só
    com pontos, eles são de milhar quando há mais de um ou quando o único deixa 3 dígitos
    depois dele. Vazio vale 0.
    """
    if v is None:
        return 0.0
    if isinstance(v, (int, float)):
        return float(v)
    texto = str(v).strip()
    negativo = texto.startswith("-") or texto.startswith("−")
    s = "".join(c for c in texto if c.isdigit() or c in ",.")
    if not s:
        return 0.0
    if "," in s and "." in s:
        decimal = "," if s.rfind(",") > s.rfind(".") else "."
        milhar = "." if decimal == "," else ","
        s = s.replace(milhar, "").replace(decimal, ".")
    elif "," in s:
        s = s.replace(",", ".")
    elif s.count(".") > 1 or len(s) - s.rfind(".") - 1 == 3:
        s = s.replace(".", "")
    valor = float(s)
    return -valor if negativo else valor


def cenario(nome, prolabore, folha12, a):
    socios = a.socios
    fator_r = folha12 / a.rbt12 if a.rbt12 else 0.0
    anexo = "III" if fator_r >= FATOR_R_MINIMO else "V"
    das = a.faturamento_mes * aliquota_efetiva(anexo, a.rbt12)
    por_socio = prolabore / socios
    inss_socio = a.inss / 100 * min(por_socio, a.teto_inss)
    irrf_socio = irrf(por_socio, inss_socio, a)
    return {
        "cenario": nome,
        "prolabore_mensal": round(prolabore, 2),
        "fator_r": round(fator_r * 100, 2),
        "anexo": anexo,
        "das": round(das, 2),
        "inss": round(inss_socio * socios, 2),
        "irrf": round(irrf_socio * socios, 2),
        "total": round(das + (inss_socio + irrf_socio) * socios, 2),
    }


def meses_de_atividade(meses):
    """Meses a partir do primeiro com faturamento ou pró-labore (0 se nenhum)."""
    for i, m in enumerate(meses):
        if numero(m.get("faturamento")) > 0 or numero(m.get("prolabore")) > 0:
            return len(meses) - i
    return 0


def argumentos():
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--faturamento-json", help="saída de `ctbz impostos faturamento -o json`")
    p.add_argument("--rbt12", type=float, help="faturamento somado do período (até 12 meses)")
    p.add_argument("--folha12", type=float, help="pró-labore (folha) somado do mesmo período")
    p.add_argument("--meses-atividade", type=int, help="meses de atividade no período, se menos de 12")
    p.add_argument("--faturamento-mes", type=float, help="faturamento do mês simulado (padrão: RBT12/12)")
    p.add_argument("--prolabore", type=float, help="pró-labore mensal atual, somando os sócios")
    p.add_argument("--socios", type=int, default=1, help="sócios com pró-labore (divide em partes iguais)")
    p.add_argument("--salario-minimo", type=float, default=1621.0, help="de `ctbz prolabore parametros`")
    p.add_argument("--inss", type=float, default=11.0, help="alíquota do INSS do sócio, em %%")
    p.add_argument("--teto-inss", type=float, default=8475.55, help="pró-labore no teto do INSS, por sócio")
    p.add_argument("--irrf-json", help="saída de `ctbz impostos tabela-irrf -o json` (sem ela, IRRF = 0)")
    p.add_argument("--irrf-isento-ate", type=float, default=5000.0, help="pró-labore isento de IRRF, por sócio")
    p.add_argument("--irrf-redutor-ate", type=float, default=7350.0, help="limite do redutor do IRRF de 2026")
    p.add_argument("--desconto-simplificado", type=float, default=607.20, help="desconto simplificado mensal do IRRF")
    p.add_argument("--dependentes", type=int, default=0, help="dependentes por sócio, para o IRRF")
    p.add_argument("--deducao-dependente", type=float, default=189.59, help="dedução mensal por dependente")
    p.add_argument("--json", action="store_true", help="saída em JSON")
    a = p.parse_args()

    meses_json = None
    if a.faturamento_json:
        with open(a.faturamento_json, encoding="utf-8") as f:
            meses = json.load(f)
        if not meses:
            p.error("faturamento vazio")
        if a.rbt12 is None:
            a.rbt12 = sum(numero(m.get("faturamento")) for m in meses)
        if a.folha12 is None:
            a.folha12 = sum(numero(m.get("prolabore")) for m in meses)
        if a.prolabore is None:
            # O mês mais recente pode vir sem pró-labore enquanto a Contabilizei calcula.
            valores = [m.get("prolabore") for m in meses if m.get("prolabore") is not None]
            a.prolabore = numero(valores[-1]) if valores else 0.0
        meses_json = min(meses_de_atividade(meses), 12)
    if a.rbt12 is None or a.folha12 is None or a.prolabore is None:
        p.error("informe --faturamento-json ou --rbt12, --folha12 e --prolabore")
    for nome in ("rbt12", "folha12", "prolabore", "faturamento_mes", "salario_minimo", "inss", "teto_inss",
                 "irrf_isento_ate", "irrf_redutor_ate", "desconto_simplificado", "dependentes", "deducao_dependente"):
        valor = getattr(a, nome)
        if valor is not None and valor < 0:
            p.error(f"--{nome.replace('_', '-')} não pode ser negativo")
    if a.socios < 1:
        p.error("--socios precisa ser pelo menos 1")

    a.meses = a.meses_atividade if a.meses_atividade is not None else (meses_json or 12)
    if not 1 <= a.meses <= 12:
        p.error("--meses-atividade precisa estar entre 1 e 12")
    if a.meses < 12:
        a.rbt12 = a.rbt12 * 12 / a.meses
        a.folha12 = a.folha12 * 12 / a.meses
    if a.faturamento_mes is None:
        a.faturamento_mes = a.rbt12 / 12
    a.tabela_irrf = ler_tabela_irrf(a.irrf_json)
    return a


def brl(v, largura=0):
    """1234.5 → "1.234,50", alinhado à direita em `largura`."""
    s = f"{v:,.2f}".replace(",", "_").replace(".", ",").replace("_", ".")
    return s.rjust(largura)


def main():
    a = argumentos()
    if a.rbt12 <= 0:
        print("Sem faturamento nos últimos 12 meses: o Fator R ainda não se aplica, não há o que simular.")
        return
    # Pró-labore que, mantido por 12 meses, deixa o Fator R em 28% (arredondado para cima).
    necessario = max(math.ceil(round(FATOR_R_MINIMO * a.rbt12 / 12 * 100, 6)) / 100, a.salario_minimo * a.socios)
    minimo = a.salario_minimo * a.socios

    cenarios = [cenario("atual", a.prolabore, a.folha12, a)]
    cenarios.append(cenario("salario_minimo", minimo, minimo * 12, a))
    cenarios.append(cenario("fator_r_28", necessario, necessario * 12, a))
    base = cenarios[0]["total"]
    for c in cenarios:
        c["diferenca_vs_atual"] = round(c["total"] - base, 2)

    avisos = [
        "Estimativa mensal em regime: os cenários supõem o mesmo pró-labore nos 12 meses; o Fator R"
        " é uma média móvel e muda aos poucos. O valor exato mês a mês é o da gestão inteligente.",
        "INSS do sócio (11%) e IRRF são retidos do sócio; nos anexos III e V a contribuição"
        " patronal já está no DAS.",
        "A folha considerada é só o pró-labore. Com funcionários, some salários e FGTS em --folha12.",
    ]
    if a.meses < 12:
        avisos.append(
            f"Empresa com {a.meses} mês(es) de atividade: receita e folha anualizadas (média × 12,"
            " LC 123, art. 18, § 2º). Confira com o início da atividade (`ctbz empresa`, abertura)."
        )
    if not a.tabela_irrf:
        avisos.append("IRRF não estimado: passe --irrf-json com `ctbz impostos tabela-irrf -o json`.")
    else:
        avisos.append(
            "IRRF com o desconto simplificado (ou INSS e dependentes, se maior) e o redutor de 2026"
            " (Lei 15.270/2025)."
        )
    if a.rbt12 > TETO:
        avisos.append(
            "RBT12 acima de R$ 4,8 milhões: alíquota da 6ª faixa. A exclusão do Simples depende da"
            " receita do ano-calendário, não da RBT12: confira com o contador."
        )
    elif a.rbt12 > SUBLIMITE:
        avisos.append(
            "RBT12 acima de R$ 3,6 milhões: se a receita do ano-calendário passar do sublimite, ISS e"
            " ICMS saem do DAS (não incluído aqui)."
        )

    if a.json:
        json.dump({"rbt12": a.rbt12, "folha12": a.folha12, "faturamento_mes": a.faturamento_mes,
                   "meses_atividade": a.meses, "cenarios": cenarios, "avisos": avisos},
                  sys.stdout, ensure_ascii=False, indent=2)
        print()
        return

    print(f"RBT12: R$ {brl(a.rbt12)} · Folha 12 meses: R$ {brl(a.folha12)} · "
          f"Faturamento do mês: R$ {brl(a.faturamento_mes)}")
    cab = f"{'Cenário':<16}{'Pró-labore':>14}{'Fator R':>9}{'Anexo':>7}{'DAS':>13}{'INSS':>11}{'IRRF':>11}{'Total':>13}{'Dif.':>12}"
    print(cab)
    for c in cenarios:
        print(f"{c['cenario']:<16}{brl(c['prolabore_mensal'], 14)}{brl(c['fator_r'], 8)}%{c['anexo']:>7}"
              f"{brl(c['das'], 13)}{brl(c['inss'], 11)}{brl(c['irrf'], 11)}{brl(c['total'], 13)}"
              f"{brl(c['diferenca_vs_atual'], 12)}")
    for aviso in avisos:
        print("- " + aviso)


if __name__ == "__main__":
    main()
