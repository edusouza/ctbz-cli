#!/usr/bin/env python3
"""Gera o catálogo de endpoints da plataforma Contabilizei a partir do front.

Baixa o painel (/painel-de-controle/) e o front de notas de entrada (/nota-entrada/),
com todos os seus arquivos JavaScript, usando a sessão salva pelo `ctbz login`;
encontra as chamadas HTTP e grava um Markdown agrupado por base de API e por domínio.

Formas de chamada reconhecidas (código minificado):

- axios por instância: r["c"].post("caminho", …)
- sufixo concatenado: r["c"].put("guia/".concat(id, "/confirmar-pagamento"))
- caminho em variável: var e = "caminho"; … r["c"].get(e)   (inclui ternários)
- objeto de configuração: Object(r["c"])({method: "post", url: "caminho", …})
- vue-resource (notas de entrada): this.$http.post("/api/emissor/…")

Uso:
    ctbz login
    scripts/extrair-endpoints.py > docs/api/catalogo.md

Variáveis: CTBZ_HOME (diretório da sessão, padrão ~/.config/ctbz).
Só usa a biblioteca padrão do Python. Testes: python3 -m unittest discover -s scripts
"""
import collections
import concurrent.futures
import datetime
import gzip
import json
import os
import re
import sys
import urllib.request

BASE = "https://app.contabilizei.com.br"
PAINEL = BASE + "/painel-de-controle/"
NOTA_ENTRADA = BASE + "/nota-entrada/"
UA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0 Safari/537.36"

# Exportações do módulo que cria as instâncias axios (ver docs/frontend).
INSTANCIAS = {
    "c": "/api/plataforma/",
    "d": "/api/legado/",
    "b": "/api/legado/",
    "e": "/api/multiusuario/",
    "f": "/api/fintech/",
    "a": "/api/leads/hubspot/",
}
# Base das chamadas relativas do front de notas de entrada (vue-resource).
BASE_NOTA_ENTRADA = "/api/emissor/"

METODOS = "get|post|put|delete|patch"
LITERAL = r'["`]([^"`]*)["`]'
# Cadeia de .concat(x, "sufixo") depois de um literal.
CONCAT = r'((?:\.concat\([^()]*(?:\([^()]*\)[^()]*)*\))*)'

RE_CHAMADA = re.compile(r'\w+\["([a-f])"\]\.(' + METODOS + r')\(\s*' + LITERAL + CONCAT)
RE_CHAMADA_ID = re.compile(r'\w+\["([a-f])"\]\.(' + METODOS + r')\(\s*([A-Za-z_$][\w$]*)\s*[,)]')
RE_CONFIG = re.compile(r'\w+\["([a-f])"\](?:\.request)?\)?\(\s*\{')
RE_HTTP = re.compile(r'\$http\.(' + METODOS + r')\(\s*' + LITERAL + CONCAT)
RE_HTTP_ID = re.compile(r'\$http\.(' + METODOS + r')\(\s*([A-Za-z_$][\w$]*)\s*[,)]')
RE_BASE_DECLARADA = re.compile(r'baseURL:"(/api/[^"]+)"')
RE_CONCAT_ARGS = re.compile(r'\.concat\(([^()]*(?:\([^()]*\)[^()]*)*)\)')
RE_PATH = re.compile(r'^/?[a-z][\w\-./?=&{}]*$')
# Distância máxima, para trás, entre a atribuição de uma variável e a chamada que a usa.
JANELA_VARIAVEL = 400


def com_sufixo(prefixo, cadeia):
    """Junta prefixo e .concat(...): só mantém o molde completo ({} no lugar das variáveis)
    quando um sufixo literal traz texto; senão fica o prefixo, como no catálogo antigo."""
    if not cadeia:
        return prefixo
    molde, tem_texto = prefixo, False
    for args in RE_CONCAT_ARGS.findall(cadeia):
        for arg in dividir_args(args):
            m = re.fullmatch(r'\s*' + LITERAL + r'\s*', arg)
            if m:
                molde += m.group(1)
                tem_texto = tem_texto or bool(re.search(r'[A-Za-z]', m.group(1)))
            else:
                molde += "{}"
    return molde if tem_texto else prefixo


def dividir_args(args):
    """Divide os argumentos de uma chamada pelas vírgulas de primeiro nível."""
    partes, nivel, atual, aspas = [], 0, "", None
    for ch in args:
        if aspas:
            atual += ch
            if ch == aspas:
                aspas = None
            continue
        if ch in "\"'`":
            aspas = ch
        elif ch in "([{":
            nivel += 1
        elif ch in ")]}":
            nivel -= 1
        elif ch == "," and nivel == 0:
            partes.append(atual)
            atual = ""
            continue
        atual += ch
    if atual.strip():
        partes.append(atual)
    return partes


def objeto(js, inicio):
    """Texto do objeto literal que começa em js[inicio] == "{", com chaves balanceadas."""
    nivel, aspas = 0, None
    for i in range(inicio, min(len(js), inicio + 2000)):
        ch = js[i]
        if aspas:
            if ch == aspas and js[i - 1] != "\\":
                aspas = None
        elif ch in "\"'`":
            aspas = ch
        elif ch == "{":
            nivel += 1
        elif ch == "}":
            nivel -= 1
            if nivel == 0:
                return js[inicio:i + 1]
    return js[inicio:inicio + 2000]


def caminhos_da_variavel(js, nome, fim):
    """Literais de caminho atribuídos a `nome` pouco antes da posição `fim`."""
    trecho = js[max(0, fim - JANELA_VARIAVEL):fim]
    atrib = list(re.finditer(r'(?<![\w$.])' + re.escape(nome) + r'\s*=(?!=)([^;]*)', trecho))
    if not atrib:
        return []
    expr = atrib[-1].group(1)
    return [c for c in re.findall(LITERAL, expr) if RE_PATH.match(c)]


def separar_base(caminho, base_padrao):
    """Caminhos absolutos /api/<x>/… vão para a base /api/<x>/; os demais, para base_padrao."""
    m = re.match(r'^(/api/[\w-]+/)(.*)$', caminho)
    if m:
        return m.group(1), m.group(2)
    return base_padrao, caminho


def extrair(fontes, chamadas=None, base_vue=BASE_NOTA_ENTRADA):
    """Encontra as chamadas nas fontes JavaScript. Devolve {base: {(MÉTODO, caminho)}} e as
    bases declaradas em axios.create."""
    chamadas = chamadas if chamadas is not None else collections.defaultdict(set)
    bases = set()
    for js in fontes:
        bases.update(RE_BASE_DECLARADA.findall(js))
        for inst, metodo, caminho, cadeia in RE_CHAMADA.findall(js):
            chamadas[INSTANCIAS[inst]].add((metodo.upper(), com_sufixo(caminho, cadeia)))
        for m in RE_CHAMADA_ID.finditer(js):
            for caminho in caminhos_da_variavel(js, m.group(3), m.start()):
                chamadas[INSTANCIAS[m.group(1)]].add((m.group(2).upper(), caminho))
        for m in RE_CONFIG.finditer(js):
            inst, corpo = m.group(1), objeto(js, m.end() - 1)
            url = re.search(r'(?<![\w$])url:\s*' + LITERAL + CONCAT, corpo)
            if not url:
                continue
            metodo = re.search(r'(?<![\w$])method:\s*["\'](\w+)["\']', corpo)
            chamadas[INSTANCIAS[inst]].add(
                ((metodo.group(1) if metodo else "get").upper(), com_sufixo(url.group(1), url.group(2))))
        for metodo, caminho, cadeia in RE_HTTP.findall(js):
            base, rel = separar_base(com_sufixo(caminho, cadeia), base_vue)
            chamadas[base].add((metodo.upper(), rel))
        for m in RE_HTTP_ID.finditer(js):
            for caminho in caminhos_da_variavel(js, m.group(2), m.start()):
                base, rel = separar_base(caminho, base_vue)
                chamadas[base].add((m.group(1).upper(), rel))
    return chamadas, bases


def escrever(out, chamadas, bases, n_arquivos, hoje):
    total = sum(len(v) for v in chamadas.values())
    out.write("# Catálogo de endpoints (gerado)\n\n")
    out.write(f"> Gerado por `scripts/extrair-endpoints.py` em {hoje} "
              f"a partir de {n_arquivos} arquivos JavaScript do painel e das notas de entrada. "
              f"{total} chamadas encontradas.\n>\n"
              "> A base de cada chamada é inferida pela instância axios usada no código; "
              "caminhos terminados em `/` ou `=` recebem parâmetros concatenados pelo front, "
              "e `{}` marca um valor concatenado no meio do caminho.\n"
              "> Endpoints já testados estão em [endpoints-verificados.md](endpoints-verificados.md).\n\n")
    out.write("Bases declaradas no front: " + ", ".join(f"`{b}`" for b in sorted(bases)) + "\n\n")
    for base in sorted(chamadas, key=lambda b: (-len(chamadas[b]), b)):
        itens = chamadas[base]
        out.write(f"## `{base}` ({len(itens)})\n\n")
        por_dominio = collections.defaultdict(list)
        for metodo, caminho in itens:
            dominio = caminho.lstrip("/").split("/")[0].split("?")[0] or "(raiz)"
            por_dominio[dominio].append((metodo, caminho))
        for dominio in sorted(por_dominio):
            out.write(f"### {dominio}\n\n| Método | Caminho |\n|---|---|\n")
            for metodo, caminho in sorted(por_dominio[dominio], key=lambda t: (t[1], t[0])):
                out.write(f"| {metodo} | `{caminho}` |\n")
            out.write("\n")


def sessao_cookie():
    home = os.environ.get("CTBZ_HOME") or os.path.join(
        os.environ.get("XDG_CONFIG_HOME") or os.path.expanduser("~/.config"), "ctbz")
    with open(os.path.join(home, "session.json")) as f:
        sess = json.load(f)
    return "; ".join(f"{k}={v['Value']}" for k, v in sess["cookies"].items())


def baixar(url, cookie):
    req = urllib.request.Request(url, headers={"User-Agent": UA, "Cookie": cookie, "Accept-Encoding": "gzip"})
    with urllib.request.urlopen(req, timeout=60) as r:
        data = r.read()
        if r.headers.get("Content-Encoding") == "gzip":
            data = gzip.decompress(data)
    return data.decode("utf-8", errors="ignore")


def baixar_front(raiz, cookie):
    """Baixa o HTML de um front e os scripts que ele referencia."""
    html = baixar(raiz, cookie)
    if "form-login" in html or 'location.replace("/login' in html:
        sys.exit("sessão expirada: rode `ctbz login`")
    scripts = sorted(set(re.findall(r'(?:src|href)="(?:\./|/nota-entrada/)?((?:static/)?js/[^"]+\.js)"', html)))
    with concurrent.futures.ThreadPoolExecutor(8) as ex:
        return list(ex.map(lambda s: baixar(raiz + s, cookie), scripts))


def main():
    cookie = sessao_cookie()
    painel = baixar_front(PAINEL, cookie)
    notas = baixar_front(NOTA_ENTRADA, cookie)
    chamadas, bases = extrair(painel)
    extrair(notas, chamadas)
    escrever(sys.stdout, chamadas, bases, len(painel) + len(notas), datetime.date.today().isoformat())


if __name__ == "__main__":
    main()
