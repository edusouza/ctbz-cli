#!/usr/bin/env python3
"""Gera skills/assessor/references/comandos.md a partir de docs/referencia.

O mapa dá ao agente o uso e as flags de todos os comandos sem que ele precise rodar
`ctbz --help`. Rode depois de `go run ./tools/gendocs`:

    python3 scripts/gerar-comandos-skill.py
"""

import pathlib
import re
import sys

RAIZ = pathlib.Path(__file__).resolve().parent.parent
REFERENCIA = RAIZ / "docs" / "referencia"
DESTINO = RAIZ / "skills" / "assessor" / "references" / "comandos.md"

# Flags que todo comando de escrita tem; o protocolo de escrita da skill já as explica.
FLAGS_OMITIDAS = ("--help", "--dry-run", "--yes")

CABECALHO = """# Mapa de comandos do ctbz

Gerado de `docs/referencia` por `scripts/gerar-comandos-skill.py`; não edite à mão.

Todos os comandos aceitam `-o table|json|csv`. Os de escrita aceitam também `--dry-run`
(mostra a requisição sem enviar) e `-y/--yes` (envia sem perguntar; obrigatório sem terminal).
Campos do JSON de cada comando: [campos-json.md](campos-json.md).
"""


def secao(texto, titulo):
    m = re.search(r"## " + titulo + r"\n\n```\n(.*?)\n```", texto, re.S)
    return m.group(1) if m else ""


def comando(caminho):
    texto = caminho.read_text(encoding="utf-8")
    linhas = texto.splitlines()
    resumo = linhas[2].strip() if len(linhas) > 2 else ""
    uso = secao(texto, "Uso").strip() or caminho.stem.replace("_", " ")
    flags = [
        linha.strip()
        for linha in secao(texto, "Flags").splitlines()
        if linha.strip() and not any(f in linha for f in FLAGS_OMITIDAS)
    ]
    return uso, resumo, flags


def gerar():
    grupos = {}
    for caminho in sorted(REFERENCIA.glob("ctbz_*.md")):
        topo = caminho.stem.split("_")[1]
        grupos.setdefault(topo, []).append(comando(caminho))
    partes = [CABECALHO]
    for topo, comandos in grupos.items():
        partes.append(f"\n## {topo}\n")
        for uso, resumo, flags in comandos:
            partes.append(f"\n`{uso}`: {resumo}\n")
            if flags:
                partes.append("\n```text\n" + "\n".join(flags) + "\n```\n")
    return "".join(partes)


def main():
    DESTINO.parent.mkdir(parents=True, exist_ok=True)
    DESTINO.write_text(gerar(), encoding="utf-8")
    print(f"gravado {DESTINO.relative_to(RAIZ)}", file=sys.stderr)


if __name__ == "__main__":
    main()
