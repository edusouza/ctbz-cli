package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/edusouza/ctbz-cli/internal/output"
)

// now é trocado nos testes para fixar a data atual.
var now = time.Now

// dateFromMillis converte epoch em milissegundos (formato das APIs); zero vira data vazia.
func dateFromMillis(ms int64) output.Date {
	if ms == 0 {
		return output.Date{}
	}
	return output.DateFromMillis(ms)
}

// dateFromMillisUTC é para campos que representam só uma data e vêm como epoch à meia-noite UTC
// (ex.: lançamentos do razão): o dia é o do instante em UTC, sem converter para o fuso local.
func dateFromMillisUTC(ms int64) output.Date {
	if ms == 0 {
		return output.Date{}
	}
	t := time.UnixMilli(ms).UTC()
	return output.Date{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)}
}

// diasEntre conta dias de calendário de a até b (negativo se b já passou).
func diasEntre(a, b time.Time) int {
	da := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
	db := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
	return int(db.Sub(da).Hours() / 24)
}

// competencia formata mês e ano como "07/2026"; zero vira vazio.
func competencia(mes, ano int) any {
	if mes == 0 || ano == 0 {
		return nil
	}
	return fmt.Sprintf("%02d/%d", mes, ano)
}

var mesesAbreviados = map[string]int{
	"jan": 1, "fev": 2, "mar": 3, "abr": 4, "mai": 5, "jun": 6,
	"jul": 7, "ago": 8, "set": 9, "out": 10, "nov": 11, "dez": 12,
}

// competenciaDeTexto converte competências escritas pela API ("Jul de 2026",
// "Julho / 2026", "set./26") para "07/2026". Texto desconhecido volta como veio.
func competenciaDeTexto(s string) any {
	if s == "" {
		return nil
	}
	campos := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return r == ' ' || r == '/' || r == '.'
	})
	var mes, ano int
	for _, c := range campos {
		if c == "de" {
			continue
		}
		if len(c) >= 3 {
			if m, ok := mesesAbreviados[c[:3]]; ok && mes == 0 {
				mes = m
				continue
			}
		}
		if n, err := strconv.Atoi(c); err == nil {
			if n < 100 {
				n += 2000
			}
			ano = n
		}
	}
	if mes == 0 || ano == 0 {
		return s
	}
	return competencia(mes, ano)
}
