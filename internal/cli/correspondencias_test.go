package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	pathCorrespondencias = "/api/plataforma/escritorio-virtual/recuperar-mensagens?cursor="
	pathEnderecoEntrega  = "/api/plataforma/escritorio-virtual/endereco-entrega"
	pathAutorizacao      = "/api/plataforma/escritorio-virtual/verifica-autorizacao-recebimento"
)

func correspondenciasFake(t *testing.T) *escritaFake {
	t.Helper()
	f := newEscritaFake(t, map[string]string{
		pathCorrespondencias: fixture(t, "correspondencias"),
		pathEnderecoEntrega:  fixture(t, "endereco_entrega"),
		pathAutorizacao:      `{"autoriza":false}`,
		"/api/plataforma/escritorio-virtual/pesquisarEndereco/01310100": `{"cep":"01310100","logradouro":"Avenida Paulista","bairro":"Bela Vista","cidade":"São Paulo","uf":"SP"}`,
	})
	f.onWrite = func(f *escritaFake) {
		f.gets[pathEnderecoEntrega] = `{"cep":"01310100","logradouro":"Avenida Paulista","numero":"1000","cidade":"São Paulo","uf":"SP"}`
		f.gets[pathAutorizacao] = `{"autoriza":true}`
	}
	return f
}

func TestCorrespondencias(t *testing.T) {
	correspondenciasFake(t)
	out, stderr, code := execCLI(t, "", "correspondencias", "-o", "csv")
	want := "id,data,remetente,descricao,situacao,valor_envio\n501,2026-09-15,RECEITA FEDERAL,Intimação,DISPONIVEL,12.50\n"
	if code != ExitOK || out != want {
		t.Errorf("código %d, %s\n%s", code, stderr, out)
	}
}

func TestCorrespondenciasSemServico(t *testing.T) {
	writeServer(t, http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(560)
		io.WriteString(w, `{"detalhes":[{"detalhe":"Serviço não contratado"}]}`)
	})
	out, stderr, code := execCLI(t, "", "correspondencias", "-o", "json")
	if code != ExitOK || out != "[]\n" || !strings.Contains(stderr, "não contratou o escritório virtual") {
		t.Errorf("listagem: código %d, %q, %s", code, out, stderr)
	}
	_, stderr, code = execCLI(t, "", "correspondencias", "endereco")
	if code != ExitError || !strings.Contains(stderr, "não contratou o escritório virtual") {
		t.Errorf("endereço: código %d, %s", code, stderr)
	}
}

func TestCorrespondenciasEndereco(t *testing.T) {
	f := correspondenciasFake(t)
	out, _, code := execCLI(t, "", "correspondencias", "endereco", "-o", "csv")
	if code != ExitOK || out != "cep,logradouro,numero,complemento,bairro,cidade,uf\n01001000,Praça da Sé,100,sala 1,Sé,São Paulo,SP\n" {
		t.Errorf("mostrar: código %d\n%s", code, out)
	}
	out, stderr, code := execCLI(t, "", "correspondencias", "endereco", "--cep", "01310-100", "--numero", "1000", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != `POST /api/plataforma/escritorio-virtual/salvar-endereco-correspondencia {"cep":"01310100","numero":"1000","complemento":""}` {
		t.Fatalf("trocar: código %d, %q\n%s", code, f.writes, stderr)
	}
	if !strings.Contains(stderr, "de Praça da Sé, 100 - sala 1 - São Paulo/SP - CEP 01001000 para Avenida Paulista, 1000 - São Paulo/SP - CEP 01310100 (risco médio)") {
		t.Errorf("stderr:\n%s", stderr)
	}
	if !strings.HasPrefix(out, "acao,situacao,id,cep,") || !strings.Contains(out, "endereco,alterado,,01310100,Avenida Paulista,1000") {
		t.Errorf("saída:\n%s", out)
	}
	if _, stderr, code := execCLI(t, "", "correspondencias", "endereco", "--cep", "123", "--numero", "1", "--yes"); code != ExitUsage || !strings.Contains(stderr, "8 dígitos") {
		t.Errorf("CEP inválido: código %d, %s", code, stderr)
	}
}

func TestCorrespondenciasAutorizar(t *testing.T) {
	f := correspondenciasFake(t)
	out, stderr, code := execCLI(t, "", "correspondencias", "autorizar", "on", "--yes", "-o", "csv")
	if code != ExitOK || len(f.writes) != 1 || f.writes[0] != "PUT /api/plataforma/escritorio-virtual/autorizar-recebimento/true " ||
		!strings.Contains(stderr, "o envio pode ser cobrado") || out != "acao,situacao,id,autoriza\nautorizar,autorizado,,true\n" {
		t.Fatalf("código %d, %q\n%s\n%s", code, f.writes, stderr, out)
	}
	_, stderr, code = execCLI(t, "", "correspondencias", "autorizar", "on", "--yes")
	if code != ExitOK || len(f.writes) != 1 || !strings.Contains(stderr, "já está autorizado") {
		t.Errorf("sem mudança: código %d, %d escritas, %s", code, len(f.writes), stderr)
	}
}

func TestCorrespondenciasBaixar(t *testing.T) {
	pdf := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, "%PDF-correspondencia") }))
	t.Cleanup(pdf.Close)
	arquivo := pdf.URL + "/correspondencia.pdf"
	newEscritaFake(t, map[string]string{"/api/plataforma/escritorio-virtual/gerar-link-download-correspondencia/501": `{"url":"` + arquivo + `"}`})
	dir := t.TempDir()
	out, stderr, code := execCLI(t, "", "correspondencias", "baixar", "501", "--dir", dir)
	path := filepath.Join(dir, "correspondencia-501.pdf")
	if b, _ := os.ReadFile(path); code != ExitOK || string(b) != "%PDF-correspondencia" || out != path+"\n" {
		t.Errorf("código %d, %s\n%s", code, stderr, out)
	}
	if _, stderr, code := execCLI(t, "", "correspondencias", "baixar", "501", "--dir", dir); code != ExitError || !strings.Contains(stderr, "já existe") {
		t.Errorf("já existe: código %d, %s", code, stderr)
	}
}
