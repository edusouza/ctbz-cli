package ctbz

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// maxAcaoLine is the per-line budget for acoes.jsonl. Longer lines are counted
// as invalid instead of aborting the whole load (bufio.Scanner's ErrTooLong
// used to do that, so one corrupted line hid every action).
const maxAcaoLine = 1 << 20

// LoadAcoes lê o registro em ordem cronológica. Sem arquivo, devolve lista vazia; linhas
// ilegíveis — JSON inválido, linha longa demais, ou objeto sem Metodo/Caminho — são
// puladas e contadas em invalidas.
func (s *Store) LoadAcoes() (acoes []Acao, invalidas int, err error) {
	f, err := os.Open(s.acoesPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("lendo o registro de ações: %w", err)
	}
	defer func() { _ = f.Close() }()

	br := bufio.NewReader(f)
	for {
		line, readErr := br.ReadSlice('\n')
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > maxAcaoLine {
			invalidas++
		} else if len(trimmed) > 0 {
			var a Acao
			if json.Unmarshal(trimmed, &a) != nil || a.Metodo == "" || a.Caminho == "" {
				invalidas++
			} else {
				acoes = append(acoes, a)
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			if errors.Is(readErr, bufio.ErrBufferFull) {
				// Consome só o resto desta linha longa; o \n encerra o drain
				// e o loop externo processa as linhas seguintes.
				for {
					_, drainErr := br.ReadSlice('\n')
					if drainErr == nil {
						break
					}
					if errors.Is(drainErr, io.EOF) {
						return acoes, invalidas, nil
					}
					if !errors.Is(drainErr, bufio.ErrBufferFull) {
						return acoes, invalidas, fmt.Errorf("lendo o registro de ações: %w", drainErr)
					}
				}
				continue
			}
			return acoes, invalidas, fmt.Errorf("lendo o registro de ações: %w", readErr)
		}
	}
	return acoes, invalidas, nil
}
