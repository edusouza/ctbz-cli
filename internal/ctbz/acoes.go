package ctbz

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Resultados possíveis de uma escrita registrada.
const (
	ResultadoEnviada     = "enviada"      // HTTP 2xx
	ResultadoRecusada    = "recusada"     // resposta HTTP de erro
	ResultadoSemResposta = "sem resposta" // falha de conexão: resultado incerto
)

// Acao é uma escrita enviada à Contabilizei, como registrada em acoes.jsonl. Não guarda
// corpo de requisição nem de resposta, que podem ter dados pessoais ou senhas.
type Acao struct {
	Data      time.Time `json:"data"`
	CNPJ      string    `json:"cnpj,omitempty"`
	Comando   string    `json:"comando"`
	Metodo    string    `json:"metodo"`
	Caminho   string    `json:"caminho"`
	Status    int       `json:"status,omitempty"` // 0 sem resposta
	Resultado string    `json:"resultado"`
	ID        string    `json:"id,omitempty"`
	// Arquivos são o nome e o tamanho dos arquivos de um envio multipart, nunca o conteúdo
	// ("documento: extrato.csv (42 KB)").
	Arquivos string `json:"arquivos,omitempty"`
}

func (s *Store) acoesPath() string { return filepath.Join(s.Dir, "acoes.jsonl") }

// AppendAcao acrescenta uma linha a acoes.jsonl (criado com permissão 0600).
func (s *Store) AppendAcao(a Acao) error {
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return err
	}
	line, err := json.Marshal(a)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.acoesPath(), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return errors.Join(err, f.Close())
}

// LoadAcoes lê o registro em ordem cronológica. Sem arquivo, devolve lista vazia; linhas
// ilegíveis são puladas e contadas em invalidas.
func (s *Store) LoadAcoes() (acoes []Acao, invalidas int, err error) {
	data, err := os.ReadFile(s.acoesPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("lendo o registro de ações: %w", err)
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var a Acao
		if json.Unmarshal(line, &a) != nil {
			invalidas++
			continue
		}
		acoes = append(acoes, a)
	}
	return acoes, invalidas, sc.Err()
}
