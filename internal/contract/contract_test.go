package contract

import (
	"encoding/json"
	"strings"
	"testing"
)

type item struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
}

type resposta struct {
	CNPJ     string   `json:"cnpj"`
	Ativo    bool     `json:"ativo"`
	Valor    float64  `json:"valor"`
	Tags     []string `json:"tags"`
	Itens    []item   `json:"itens"`
	Opcional *item    `json:"opcional" contract:"optional"`
	Ignorado string   `json:"-"`
	privado  string
}

func TestCheck(t *testing.T) {
	for _, tc := range []struct {
		name, json string
		want       []string
	}{
		{"conforme", `{"cnpj":"1","ativo":true,"valor":1.5,"tags":["a"],"itens":[{"id":1,"label":"x"}]}`, nil},
		{"nulos são aceitos", `{"cnpj":null,"ativo":null,"valor":null,"tags":null,"itens":null,"opcional":null}`, nil},
		{"campo removido", `{"ativo":true,"valor":1,"tags":[],"itens":[]}`, []string{"cnpj: campo removido (esperado texto)"}},
		{"tipo mudou", `{"cnpj":1,"ativo":true,"valor":"1","tags":[],"itens":[{"id":"x","label":"y"}]}`, []string{
			"cnpj: tipo mudou (esperado texto, veio número)",
			"itens[].id: tipo mudou (esperado número, veio texto)",
			"valor: tipo mudou (esperado número, veio texto)",
		}},
		{"campo novo", `{"cnpj":"1","ativo":true,"valor":1,"tags":[],"itens":[],"extra":{"a":1}}`, []string{"extra: campo novo (objeto)"}},
		{"raiz errada", `[]`, []string{"(raiz): tipo mudou (esperado objeto, veio lista)"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Check([]byte(tc.json), resposta{})
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, f := range r.Findings {
				got = append(got, f.String())
			}
			if strings.Join(got, "\n") != strings.Join(tc.want, "\n") {
				t.Errorf("obtido:\n%s\nesperado:\n%s", strings.Join(got, "\n"), strings.Join(tc.want, "\n"))
			}
		})
	}
	if _, err := Check([]byte(`{`), resposta{}); err == nil {
		t.Error("esperava erro para JSON inválido")
	}
}

func TestCheckDedupesListItems(t *testing.T) {
	r, _ := Check([]byte(`[{"id":1,"label":"a","x":1},{"id":2,"label":"b","x":2}]`), []item{})
	if r.String() != "[].x: campo novo (número)" {
		t.Errorf("relatório: %q", r.String())
	}
}

func TestReportBrokenAndAdded(t *testing.T) {
	r, _ := Check([]byte(`{"ativo":true,"valor":1,"tags":[],"itens":[],"novo":1}`), resposta{})
	if len(r.Broken()) != 1 || len(r.Added()) != 1 {
		t.Errorf("broken=%v added=%v", r.Broken(), r.Added())
	}
}

func TestCheckRootList(t *testing.T) {
	r, err := Check([]byte(`[{"id":1,"label":"a"},{"id":2}]`), []item{})
	if err != nil {
		t.Fatal(err)
	}
	if r.String() != "[].label: campo removido (esperado texto)" {
		t.Errorf("relatório: %q", r.String())
	}
}

func TestAnonymize(t *testing.T) {
	in := `{
	  "razaoSocial": "EMPRESA REAL LTDA",
	  "cnpj": "12.345.678/0001-90",
	  "cpf": "12345678901",
	  "email": "pessoa@gmail.com",
	  "endereco": {"logradouro": "RUA DE VERDADE", "numero": "1234", "cep": "80000123"},
	  "id": 4801207998644224,
	  "outroId": 4801207998644224,
	  "dataAbertura": 1784581200000,
	  "valor": 139,
	  "saldo": 9876.54,
	  "guia": {"id": 7, "valor": {"label": 4321.98, "descricao": null}, "competencia": {"mes": 7, "ano": 2026}},
	  "impostoBruto": 777.77,
	  "status": "ATIVO",
	  "nomeMesCompetencia": "Setembro",
	  "totalPaginas": 3,
	  "chamados": [{"subject": "assunto real 1"}, {"subject": "assunto real 2"}, {"subject": "3"}, {"subject": "4"}],
	  "detalhe": "pagamento do CNPJ 12.345.678/0001-90 por pessoa@gmail.com",
	  "url": "https://storage.exemplo.com/assinado?token=abc",
	  "ref": "5e3050a571f2d301",
	  "prolabore": "R$ 100,00",
	  "descontos": "R$ 11,00",
	  "historico": "pago R$ 2.345,67 em julho",
	  "contrato": "` + strings.Repeat("x", 400) + `"
	}`
	out, err := Anonymize([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, leaked := range []string{"EMPRESA REAL", "12.345.678", "12345678901", "pessoa@gmail", "RUA DE VERDADE", `"1234"`, "80000123", "4801207998644224", "9876.54", "4321.98", "777.77", "storage.exemplo", "5e3050a5", "assunto real", "100,00", "11,00", "2.345,67"} {
		if strings.Contains(s, leaked) {
			t.Errorf("vazou %q:\n%s", leaked, s)
		}
	}
	for _, kept := range []string{`"totalPaginas": 3`, `"nomeMesCompetencia": "Setembro"`, `"mes": 7`, `"ano": 2026`, `"id": 7`, `"status": "ATIVO"`, `"dataAbertura": 1784581200000`, `"cnpj": "00.000.000/0000-00"`, `"<texto omitido>"`} {
		if !strings.Contains(s, kept) {
			t.Errorf("esperava %q:\n%s", kept, s)
		}
	}
	if strings.Count(s, `"subject"`) != maxItems {
		t.Errorf("listas deveriam ser cortadas em %d itens:\n%s", maxItems, s)
	}
	// O mesmo ID real vira o mesmo ID fictício.
	if !strings.Contains(s, `"id": 1000000000000001`) || !strings.Contains(s, `"outroId": 1000000000000001`) {
		t.Errorf("IDs fictícios inconsistentes:\n%s", s)
	}
	again, _ := Anonymize([]byte(in))
	if string(again) != s {
		t.Error("anonimização não é determinística")
	}
}

func TestRawMessageNaoVerificado(t *testing.T) {
	type pagina struct {
		Itens []json.RawMessage `json:"itens"`
	}
	in := `{"itens":[{"qualquer":1,"coisa":["a"]},"texto"]}`
	r, err := Check([]byte(in), pagina{})
	if err != nil || len(r.Findings) != 0 {
		t.Errorf("json.RawMessage deveria aceitar qualquer item: %v %v", err, r.Findings)
	}
	out, _ := Prune([]byte(in), pagina{})
	if !strings.Contains(string(out), `"coisa"`) {
		t.Errorf("a poda deveria manter o item inteiro:\n%s", out)
	}
}

// dataFlex simula um tipo com UnmarshalJSON próprio (ex.: data em número ou texto).
type dataFlex struct{ v string }

func (d *dataFlex) UnmarshalJSON(b []byte) error { d.v = string(b); return nil }

func TestUnmarshalerProprioNaoVerificado(t *testing.T) {
	type item struct {
		Data dataFlex `json:"data"`
	}
	in := `[{"data":1789732800000},{"data":"2026-09-18"}]`
	r, err := Check([]byte(in), []item{})
	if err != nil || len(r.Findings) != 0 {
		t.Errorf("tipo com UnmarshalJSON deveria aceitar número e texto: %v %v", err, r.Findings)
	}
	r, _ = Check([]byte(`[{}]`), []item{})
	if len(r.Findings) != 1 || r.Findings[0].Kind != Missing {
		t.Errorf("campo ausente continua obrigatório: %v", r.Findings)
	}
	out, _ := Prune([]byte(in), []item{})
	if !strings.Contains(string(out), "1789732800000") {
		t.Errorf("a poda deveria manter o valor:\n%s", out)
	}
}

func TestAnonymizeNomesConhecidos(t *testing.T) {
	sessao := `{"nome":"Maria Clara da Silva","empresa":{"razaoSocial":"ACME SOLUCOES LTDA","status":"ATIVO"},"nomeMesX":"Julho"}`
	nomes := Nomes([]byte(sessao))
	in := `[{"descricao":"Retirada para sócio - Maria Clara da Silva","conta":"Caixa"},{"historico":"pagamento acme solucoes ltda","status":"ATIVO"}]`
	out, err := Anonymize([]byte(in), nomes...)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "Maria Clara") || strings.Contains(strings.ToLower(s), "acme solucoes") {
		t.Errorf("nome conhecido vazou:\n%s", s)
	}
	if !strings.Contains(s, "Retirada para sócio - FULANO DE TAL") || !strings.Contains(s, `"conta": "Caixa"`) {
		t.Errorf("o resto do texto deveria ficar:\n%s", s)
	}
}

func TestPrune(t *testing.T) {
	in := `{"cnpj":"1","segredo":"x","itens":[{"id":1,"label":"a","cpf":"123"}],"opcional":null,"ativo":true}`
	out, err := Prune([]byte(in), resposta{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, gone := range []string{"segredo", `"cpf"`} {
		if strings.Contains(s, gone) {
			t.Errorf("campo não declarado ficou: %s\n%s", gone, s)
		}
	}
	for _, kept := range []string{`"cnpj": "1"`, `"label": "a"`, `"opcional": null`, `"ativo": true`} {
		if !strings.Contains(s, kept) {
			t.Errorf("faltou %s:\n%s", kept, s)
		}
	}
	r, _ := Check(out, resposta{})
	if len(r.Added()) != 0 {
		t.Errorf("fixture podada não deveria ter campos novos: %v", r.Added())
	}
}
