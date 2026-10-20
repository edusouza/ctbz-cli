#!/usr/bin/env python3
"""Compara cenários de pró-labore para uma atividade sujeita ao Fator R (Anexo III x V).

Estimativa de apoio: o cálculo oficial é o da Contabilizei (`ctbz impostos calculo`) e o
valor exato mês a mês é o da gestão inteligente. Só usa a biblioteca padrão.

Exemplos:
    ctbz impostos faturamento -o json > fat.json
    ctbz impostos tabela-irrf -o json > irrf.json
    python3 simular.py --faturamento-json fat.json --irrf-json irrf.json --socios 1

    python3 simular.py --rbt12 360000 --folha12 36000 --prolabore 3000 --socios 2
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


def aliquota_efetiva(anexo, rbt12):
    if rbt12 <= 0:
        return ANEXOS[anexo][0][1]
    for limite, nominal, deducao in ANEXOS[anexo]:
        if rbt12 <= limite:
            return (rbt12 * nominal - deducao) / rbt12
    raise SystemExit("RBT12 acima de R$ 4,8 milhões: fora do Simples Nacional")


def irrf(base, tabela, isento_ate, bruto):
    if bruto <= isento_ate:
        return 0.0
    devido = 0.0
    for faixa in tabela:
        if base >= faixa["base"]:
            devido = base * faixa["aliquota"] - faixa["deducao"]
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
    valores = re.findall(r"\d[\d.]*,\d+|\d+(?:\.\d+)?", texto)
    return numero(valores[0]) if valores else 0.0


def numero(v):
    """Aceita 1234.56, "1234.56", "1.234,56", "R$ 1.234,56" e "27,5%"; vazio vale 0."""
    if v is None:
        return 0.0
    if isinstance(v, (int, float)):
        return float(v)
    s = "".join(c for c in str(v) if c.isdigit() or c in ",.")
    if not s:
        return 0.0
    if "," in s:
        s = s.replace(".", "").replace(",", ".")
    return float(s)


def cenario(nome, prolabore, folha12, a):
    socios = a.socios
    fator_r = folha12 / a.rbt12 if a.rbt12 else 0.0
    anexo = "III" if fator_r >= FATOR_R_MINIMO else "V"
    das = a.faturamento_mes * aliquota_efetiva(anexo, a.rbt12)
    por_socio = prolabore / socios
    inss_socio = a.inss / 100 * min(por_socio, a.teto_inss)
    irrf_socio = irrf(por_socio - inss_socio, a.tabela_irrf, a.irrf_isento_ate, por_socio)
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


def argumentos():
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--faturamento-json", help="saída de `ctbz impostos faturamento -o json`")
    p.add_argument("--rbt12", type=float, help="faturamento dos últimos 12 meses")
    p.add_argument("--folha12", type=float, help="pró-labore (folha) dos últimos 12 meses")
    p.add_argument("--faturamento-mes", type=float, help="faturamento do mês simulado (padrão: RBT12/12)")
    p.add_argument("--prolabore", type=float, help="pró-labore mensal atual, somando os sócios")
    p.add_argument("--socios", type=int, default=1, help="sócios com pró-labore (divide em partes iguais)")
    p.add_argument("--salario-minimo", type=float, default=1621.0, help="de `ctbz prolabore parametros`")
    p.add_argument("--inss", type=float, default=11.0, help="alíquota do INSS do sócio, em %%")
    p.add_argument("--teto-inss", type=float, default=8475.55, help="pró-labore no teto do INSS, por sócio")
    p.add_argument("--irrf-json", help="saída de `ctbz impostos tabela-irrf -o json` (sem ela, IRRF = 0)")
    p.add_argument("--irrf-isento-ate", type=float, default=5000.0, help="pró-labore isento de IRRF, por sócio")
    p.add_argument("--json", action="store_true", help="saída em JSON")
    a = p.parse_args()

    if a.faturamento_json:
        with open(a.faturamento_json, encoding="utf-8") as f:
            meses = json.load(f)
        if not meses:
            p.error("faturamento vazio")
        a.rbt12 = a.rbt12 if a.rbt12 is not None else sum(numero(m.get("faturamento")) for m in meses)
        a.folha12 = a.folha12 if a.folha12 is not None else sum(numero(m.get("prolabore")) for m in meses)
        if a.prolabore is None:
            a.prolabore = numero(meses[-1].get("prolabore"))
    if a.rbt12 is None or a.folha12 is None or a.prolabore is None:
        p.error("informe --faturamento-json ou --rbt12, --folha12 e --prolabore")
    if a.socios < 1:
        p.error("--socios precisa ser pelo menos 1")
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
    ]
    if not a.tabela_irrf:
        avisos.append("IRRF não estimado: passe --irrf-json com `ctbz impostos tabela-irrf -o json`.")
    else:
        avisos.append(
            "IRRF sem dependentes e sem o redutor entre R$ 5.000 e R$ 7.350 (Lei 15.270/2025): o valor"
            " real pode ser menor."
        )
    if a.rbt12 > 3_600_000:
        avisos.append("RBT12 acima do sublimite de R$ 3,6 milhões: ISS fora do DAS, não incluído aqui.")

    if a.json:
        json.dump({"rbt12": a.rbt12, "folha12": a.folha12, "faturamento_mes": a.faturamento_mes,
                   "cenarios": cenarios, "avisos": avisos}, sys.stdout, ensure_ascii=False, indent=2)
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
