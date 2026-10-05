"""Testes do extrator de endpoints com trechos de JavaScript minificado.

Rode com: python3 -m unittest discover -s scripts
"""
import importlib.util
import io
import os
import unittest

_spec = importlib.util.spec_from_file_location(
    "extrair_endpoints", os.path.join(os.path.dirname(__file__), "extrair-endpoints.py"))
ex = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(ex)


def extrair(*fontes):
    chamadas, _ = ex.extrair(list(fontes))
    return {base: sorted(itens) for base, itens in chamadas.items()}


class ExtrairTest(unittest.TestCase):
    def test_chamada_literal(self):
        self.assertEqual(extrair('n["c"].get("appbar/get").then(x)'),
                         {"/api/plataforma/": [("GET", "appbar/get")]})

    def test_concat_com_sufixo_literal(self):
        js = ('return r["c"].put("impostos/v5/impostos-a-pagar/guia/".concat(t.id,"/confirmar-pagamento"),{tipo:e})'
              ';r["c"].get("relatorios-ms/gerar-balancete/".concat(a,"/").concat(o))'
              ';r["c"].put("impostos/v2/impostos-a-pagar/guia/".concat(Object(i.a)(t),"/v2/recalcular"))')
        self.assertEqual(extrair(js), {"/api/plataforma/": [
            ("GET", "relatorios-ms/gerar-balancete/"),
            ("PUT", "impostos/v2/impostos-a-pagar/guia/{}/v2/recalcular"),
            ("PUT", "impostos/v5/impostos-a-pagar/guia/{}/confirmar-pagamento"),
        ]})

    def test_delete_com_concat(self):
        js = 'u["c"].delete("/movimentacao-financeira/desmembrar/desfazer/".concat(e))'
        self.assertEqual(extrair(js), {"/api/plataforma/": [
            ("DELETE", "/movimentacao-financeira/desmembrar/desfazer/")]})

    def test_caminho_em_variavel(self):
        js = ('salvar:function(t){var e="/novo-emissor/clientes/salvar-cliente-nacional";return u["c"].post(e,t)},'
              'exterior:function(t){var n=t.exterior?"novo-emissor/clientes/salvar-cliente-exterior":"novo-emissor/clientes/salvar-cliente-nacional";'
              'return u["c"].post(n,t)},emitir:function(t){const a="novo-emissor/v2/emissao/emitir";return u["c"].post(a,{nota:t})},'
              'nao:function(t){var x=t.url;return u["c"].get(x)}')
        self.assertEqual(extrair(js), {"/api/plataforma/": [
            ("POST", "/novo-emissor/clientes/salvar-cliente-nacional"),
            ("POST", "novo-emissor/clientes/salvar-cliente-exterior"),
            ("POST", "novo-emissor/clientes/salvar-cliente-nacional"),
            ("POST", "novo-emissor/v2/emissao/emitir"),
        ]})

    def test_axios_com_objeto(self):
        js = ('Object(r["c"])({method:"post",url:"upload-documentos/extrato/enviar",data:t,headers:{}})'
              ';Object(r["c"])({url:"documentos/envio-documento/enviar/consolidado",method:"POST",data:n})'
              ';r["c"].request({method:"delete",url:"/movimentacao-financeira/extrato",params:{id:e}})'
              ';Object(r["d"])({url:"documentos/".concat(e,"/anexo")})')
        self.assertEqual(extrair(js), {
            "/api/plataforma/": [
                ("DELETE", "/movimentacao-financeira/extrato"),
                ("POST", "documentos/envio-documento/enviar/consolidado"),
                ("POST", "upload-documentos/extrato/enviar"),
            ],
            "/api/legado/": [("GET", "documentos/{}/anexo")],
        })

    def test_vue_resource_notas_de_entrada(self):
        js = ('this.$http.post("/api/emissor/notasentrada/manifestar/",{notas:t})'
              ';t.$http.post("/api/emissor/classificacaonotas/salvarloteclassificacao/".concat(e))'
              ';this.$http.get("notasentrada/manifestacao/0?mes="+t)'
              ';var u="/api/emissor/parametrosEmpresa/create";this.$http.post(u,t)')
        self.assertEqual(extrair(js), {"/api/emissor/": [
            ("GET", "notasentrada/manifestacao/0?mes="),
            ("POST", "classificacaonotas/salvarloteclassificacao/"),
            ("POST", "notasentrada/manifestar/"),
            ("POST", "parametrosEmpresa/create"),
        ]})

    def test_bases_declaradas(self):
        _, bases = ex.extrair(['c=axios.create({baseURL:"/api/plataforma/",withCredentials:!0})'])
        self.assertEqual(bases, {"/api/plataforma/"})

    def test_escrever(self):
        chamadas, bases = ex.extrair(['n["c"].get("appbar/get");n["e"].post("invites/enviar",t)'])
        out = io.StringIO()
        ex.escrever(out, chamadas, bases, 2, "2026-10-04")
        texto = out.getvalue()
        self.assertIn("em 2026-10-04 a partir de 2 arquivos", texto)
        self.assertIn("2 chamadas encontradas", texto)
        self.assertIn("## `/api/multiusuario/` (1)\n\n### invites\n\n| Método | Caminho |\n|---|---|\n| POST | `invites/enviar` |\n", texto)


if __name__ == "__main__":
    unittest.main()
