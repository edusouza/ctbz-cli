// Comando notices regenera o THIRD_PARTY_NOTICES com a licença de cada módulo compilado no
// binário (em todos os sistemas que o GoReleaser gera) e da biblioteca padrão do Go:
//
//	go run ./tools/notices
//
// Rode depois de adicionar ou atualizar uma dependência; TestNoticesCobreGoMod falha se o
// arquivo estiver desatualizado em relação ao go.mod.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const arquivo = "THIRD_PARTY_NOTICES"

// sistemas são os GOOS do .goreleaser.yaml: dependências exclusivas de um sistema
// (mousetrap só entra no Windows) também vão no arquivo.
var sistemas = []string{"linux", "darwin", "windows"}

type modulo struct {
	Path, Version, Dir string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	mods := map[string]modulo{}
	for _, goos := range sistemas {
		lista, err := dependencias(goos)
		if err != nil {
			return err
		}
		for _, m := range lista {
			mods[m.Path] = m
		}
	}
	goroot, err := goEnv("GOROOT")
	if err != nil {
		return err
	}
	ordenados := make([]modulo, 0, len(mods)+1)
	ordenados = append(ordenados, modulo{Path: "Go (biblioteca padrão)", Dir: goroot})
	caminhos := make([]string, 0, len(mods))
	for p := range mods {
		caminhos = append(caminhos, p)
	}
	sort.Strings(caminhos)
	for _, p := range caminhos {
		ordenados = append(ordenados, mods[p])
	}
	texto, err := montar(ordenados)
	if err != nil {
		return err
	}
	return os.WriteFile(arquivo, []byte(texto), 0o644)
}

// dependencias lista os módulos (fora o principal) compilados em ./cmd/ctbz para o GOOS.
func dependencias(goos string) ([]modulo, error) {
	cmd := exec.Command("go", "list", "-deps",
		"-f", "{{with .Module}}{{if not .Main}}{{.Path}}\t{{.Version}}\t{{.Dir}}{{end}}{{end}}",
		"./cmd/ctbz")
	cmd.Env = append(os.Environ(), "GOOS="+goos)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listando dependências para %s: %w", goos, err)
	}
	var mods []modulo
	for _, l := range strings.Split(string(out), "\n") {
		campos := strings.Split(l, "\t")
		if len(campos) != 3 {
			continue
		}
		mods = append(mods, modulo{Path: campos[0], Version: campos[1], Dir: campos[2]})
	}
	return mods, nil
}

func goEnv(nome string) (string, error) {
	out, err := exec.Command("go", "env", nome).Output()
	if err != nil {
		return "", fmt.Errorf("lendo go env %s: %w", nome, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// montar escreve o arquivo: um cabeçalho e, por módulo, o nome, a versão e os arquivos de
// licença e de aviso encontrados na raiz do módulo.
func montar(mods []modulo) (string, error) {
	var b bytes.Buffer
	b.WriteString("Avisos de terceiros\n\n")
	b.WriteString("O ctbz inclui o software abaixo, distribuído sob as licenças reproduzidas a seguir.\n")
	b.WriteString("Arquivo gerado por `go run ./tools/notices`; não edite à mão.\n")
	for _, m := range mods {
		arquivos, err := arquivosDeLicenca(m.Dir)
		if err != nil {
			return "", err
		}
		if len(arquivos) == 0 {
			return "", fmt.Errorf("nenhum arquivo de licença em %s (%s)", m.Path, m.Dir)
		}
		b.WriteString("\n" + strings.Repeat("=", 78) + "\n")
		b.WriteString(strings.TrimSpace(m.Path + " " + m.Version))
		b.WriteString("\n" + strings.Repeat("=", 78) + "\n")
		for _, a := range arquivos {
			conteudo, err := os.ReadFile(filepath.Join(m.Dir, a))
			if err != nil {
				return "", fmt.Errorf("lendo licença de %s: %w", m.Path, err)
			}
			b.WriteString("\n")
			b.Write(bytes.TrimSpace(conteudo))
			b.WriteString("\n")
		}
	}
	return b.String(), nil
}

// arquivosDeLicenca devolve, em ordem alfabética, os arquivos LICENSE*, LICENCE*, COPYING* e
// NOTICE* da raiz de dir.
func arquivosDeLicenca(dir string) ([]string, error) {
	entradas, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("lendo %s: %w", dir, err)
	}
	var nomes []string
	for _, e := range entradas {
		if e.IsDir() {
			continue
		}
		n := strings.ToUpper(e.Name())
		for _, prefixo := range []string{"LICENSE", "LICENCE", "COPYING", "NOTICE"} {
			if strings.HasPrefix(n, prefixo) {
				nomes = append(nomes, e.Name())
				break
			}
		}
	}
	sort.Strings(nomes)
	return nomes, nil
}
