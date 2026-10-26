package main

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestArquivosDeLicenca(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"LICENSE.txt", "NOTICE", "COPYING", "license-extra.md", "README.md", "main.go"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "LICENSES"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := arquivosDeLicenca(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"COPYING", "LICENSE.txt", "NOTICE", "license-extra.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("arquivosDeLicenca = %v, quer %v", got, want)
	}
}

func TestMontar(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("\nTexto da licença\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := montar([]modulo{{Path: "exemplo.com/mod", Version: "v1.2.3", Dir: dir}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "exemplo.com/mod v1.2.3\n") || !strings.Contains(got, "\nTexto da licença\n") {
		t.Errorf("montar sem o módulo ou a licença:\n%s", got)
	}
}

func TestMontarSemLicenca(t *testing.T) {
	if _, err := montar([]modulo{{Path: "exemplo.com/mod", Version: "v1.0.0", Dir: t.TempDir()}}); err == nil {
		t.Error("montar aceitou módulo sem arquivo de licença")
	}
}

// TestNoticesCobreGoMod falha quando o go.mod tem um módulo, ou uma versão, que não está
// no THIRD_PARTY_NOTICES: rode go run ./tools/notices.
func TestNoticesCobreGoMod(t *testing.T) {
	avisos, err := os.ReadFile(filepath.Join("..", "..", arquivo))
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dentro := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		switch {
		case l == "require (":
			dentro = true
			continue
		case l == ")":
			dentro = false
			continue
		case strings.HasPrefix(l, "require "):
			l = strings.TrimPrefix(l, "require ")
		case !dentro:
			continue
		}
		campos := strings.Fields(l)
		if len(campos) < 2 {
			continue
		}
		if !strings.Contains(string(avisos), "\n"+campos[0]+" "+campos[1]+"\n") {
			t.Errorf("%s %s ausente do %s: rode go run ./tools/notices", campos[0], campos[1], arquivo)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}
