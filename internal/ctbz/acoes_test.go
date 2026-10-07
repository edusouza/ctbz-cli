package ctbz

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestAcoes(t *testing.T) {
	s := &Store{Dir: filepath.Join(t.TempDir(), "novo")}
	acoes, invalidas, err := s.LoadAcoes()
	if err != nil || len(acoes) != 0 || invalidas != 0 {
		t.Fatalf("sem arquivo: %v %d %v", acoes, invalidas, err)
	}
	t0 := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)
	for i, a := range []Acao{
		{Data: t0, CNPJ: "22222222000122", Comando: "ctbz caixa remover", Metodo: "DELETE", Caminho: "/api/plataforma/x/1", Status: 200, Resultado: ResultadoEnviada, ID: "1"},
		{Data: t0.Add(time.Minute), Comando: "ctbz api", Metodo: "POST", Caminho: "/api/plataforma/y", Resultado: ResultadoSemResposta},
	} {
		if err := s.AppendAcao(a); err != nil {
			t.Fatalf("ação %d: %v", i, err)
		}
	}
	f, err := os.OpenFile(s.acoesPath(), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("linha corrompida\n\n")
	f.Close()

	acoes, invalidas, err = s.LoadAcoes()
	if err != nil || len(acoes) != 2 || invalidas != 1 {
		t.Fatalf("leitura: %+v %d %v", acoes, invalidas, err)
	}
	if acoes[0].ID != "1" || acoes[0].Status != 200 || !acoes[0].Data.Equal(t0) || acoes[1].Resultado != ResultadoSemResposta {
		t.Errorf("ações = %+v", acoes)
	}
	if runtime.GOOS != "windows" {
		if fi, _ := os.Stat(s.acoesPath()); fi.Mode().Perm() != 0o600 {
			t.Errorf("permissão %v, quero 0600", fi.Mode().Perm())
		}
	}
}

func TestLoadAcoesLongLineAndEmptyObjects(t *testing.T) {
	s := &Store{Dir: t.TempDir()}
	good := Acao{
		Data:      time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
		Comando:   "ctbz api",
		Metodo:    "POST",
		Caminho:   "/api/plataforma/z",
		Resultado: ResultadoEnviada,
	}
	if err := s.AppendAcao(good); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(s.acoesPath(), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	// Linha > 1 MiB sem quebrar o load inteiro (issue #319).
	long := make([]byte, (1<<20)+64)
	for i := range long {
		long[i] = 'x'
	}
	f.Write(long)
	f.Write([]byte("\n"))
	f.WriteString("null\n{}\n{\"metodo\":\"GET\"}\n{\"caminho\":\"/x\"}\nnot-json\n")
	f.Close()

	acoes, invalidas, err := s.LoadAcoes()
	if err != nil {
		t.Fatalf("load com linha longa: %v", err)
	}
	if len(acoes) != 1 {
		t.Fatalf("ações = %+v, quero só a válida", acoes)
	}
	if acoes[0].Metodo != "POST" || acoes[0].Caminho != "/api/plataforma/z" {
		t.Errorf("ação inesperada: %+v", acoes[0])
	}
	// long + null + {} + só metodo + só caminho + not-json
	if invalidas != 6 {
		t.Errorf("invalidas = %d, quero 6", invalidas)
	}
}
