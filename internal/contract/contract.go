// Package contract verifica se respostas JSON da API ainda têm os campos que a CLI usa.
//
// O contrato de um endpoint é o próprio struct Go usado para decodificar a resposta:
// cada campo com tag json é obrigatório, a menos que tenha a tag `contract:"optional"`.
// Assim não existe um esquema paralelo para manter: mudou o struct, mudou o contrato.
package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Kind classifica uma divergência entre a resposta e o contrato.
type Kind string

const (
	Missing     Kind = "campo removido" // campo obrigatório ausente: quebra a CLI
	TypeChanged Kind = "tipo mudou"     // tipo JSON incompatível: quebra a CLI
	Added       Kind = "campo novo"     // campo que a CLI não usa: só informativo
)

// Finding é uma divergência num caminho da resposta (ex.: "empresaAtual.cnpj").
type Finding struct {
	Path     string
	Kind     Kind
	Expected string
	Got      string
}

func (f Finding) String() string {
	switch f.Kind {
	case TypeChanged:
		return fmt.Sprintf("%s: %s (esperado %s, veio %s)", f.Path, f.Kind, f.Expected, f.Got)
	case Missing:
		return fmt.Sprintf("%s: %s (esperado %s)", f.Path, f.Kind, f.Expected)
	default:
		return fmt.Sprintf("%s: %s (%s)", f.Path, f.Kind, f.Got)
	}
}

// Report é o resultado de uma verificação.
type Report struct{ Findings []Finding }

// Broken lista as divergências que quebram a CLI (removidos e tipos mudados).
func (r Report) Broken() []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Kind != Added {
			out = append(out, f)
		}
	}
	return out
}

// Added lista os campos novos (informativo).
func (r Report) Added() []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Kind == Added {
			out = append(out, f)
		}
	}
	return out
}

// String formata o relatório, uma divergência por linha.
func (r Report) String() string {
	lines := make([]string, len(r.Findings))
	for i, f := range r.Findings {
		lines[i] = f.String()
	}
	return strings.Join(lines, "\n")
}

// Check compara a resposta JSON com o tipo que a decodifica (ex.: DadosEmpresa{}).
func Check(data []byte, v any) (Report, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return Report{}, fmt.Errorf("resposta não é JSON: %w", err)
	}
	var r Report
	check(&r, "", doc, reflect.TypeOf(v))
	r.Findings = dedupe(r.Findings)
	sort.Slice(r.Findings, func(i, j int) bool { return r.Findings[i].Path < r.Findings[j].Path })
	return r, nil
}

// rawJSON (json.RawMessage) marca um campo cujo formato ainda não foi verificado: é aceito
// como vier e mantido inteiro na poda.
var rawJSON = reflect.TypeOf(json.RawMessage(nil))

var unmarshaler = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

// opaco diz se o tipo aceita qualquer JSON: json.RawMessage ou um tipo com UnmarshalJSON
// próprio (ex.: uma data que vem como número ou texto), que valida o formato sozinho.
func opaco(t reflect.Type) bool {
	return t == rawJSON || reflect.PointerTo(t).Implements(unmarshaler)
}

// maxElements limita quantos itens de cada lista são verificados.
const maxElements = 5

func check(r *Report, path string, val any, t reflect.Type) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if val == nil || t.Kind() == reflect.Interface || opaco(t) {
		return // null é aceito em qualquer campo; any e tipos opacos aceitam qualquer coisa
	}
	if want, got := expected(t), jsonKind(val); want != "qualquer" && want != got {
		r.Findings = append(r.Findings, Finding{Path: label(path), Kind: TypeChanged, Expected: want, Got: got})
		return
	}
	switch t.Kind() {
	case reflect.Struct:
		obj := val.(map[string]any)
		known := map[string]bool{}
		for _, f := range fields(t) {
			known[f.name] = true
			child, ok := obj[f.name]
			if !ok {
				if !f.optional {
					r.Findings = append(r.Findings, Finding{Path: join(path, f.name), Kind: Missing, Expected: expected(f.typ)})
				}
				continue
			}
			check(r, join(path, f.name), child, f.typ)
		}
		for k, v := range obj {
			if !known[k] {
				r.Findings = append(r.Findings, Finding{Path: join(path, k), Kind: Added, Got: jsonKind(v)})
			}
		}
	case reflect.Slice, reflect.Array:
		for i, item := range val.([]any) {
			if i == maxElements {
				break
			}
			check(r, path+"[]", item, t.Elem())
		}
	case reflect.Map:
		for k, item := range val.(map[string]any) {
			check(r, join(path, k), item, t.Elem())
		}
	}
}

type field struct {
	name     string
	typ      reflect.Type
	optional bool
}

// fields lê os campos com tag json, incluindo os de structs embutidos.
func fields(t reflect.Type) []field {
	var out []field
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag := sf.Tag.Get("json")
		if sf.Anonymous && tag == "" {
			out = append(out, fields(sf.Type)...)
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if !sf.IsExported() || name == "-" {
			continue
		}
		if name == "" {
			name = sf.Name
		}
		out = append(out, field{name: name, typ: sf.Type, optional: sf.Tag.Get("contract") == "optional"})
	}
	return out
}

// expected descreve o tipo JSON que um tipo Go aceita.
func expected(t reflect.Type) string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(json.Number("")) {
		return "número"
	}
	switch t.Kind() {
	case reflect.String:
		return "texto"
	case reflect.Bool:
		return "booleano"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "número"
	case reflect.Slice, reflect.Array:
		return "lista"
	case reflect.Struct, reflect.Map:
		return "objeto"
	}
	return "qualquer"
}

func jsonKind(v any) string {
	switch v.(type) {
	case string:
		return "texto"
	case bool:
		return "booleano"
	case json.Number:
		return "número"
	case []any:
		return "lista"
	case map[string]any:
		return "objeto"
	case nil:
		return "nulo"
	}
	return fmt.Sprintf("%T", v)
}

func join(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

func label(path string) string {
	if path == "" {
		return "(raiz)"
	}
	return path
}

// dedupe remove repetições (o mesmo campo em vários itens de uma lista).
func dedupe(fs []Finding) []Finding {
	seen := map[Finding]bool{}
	out := fs[:0]
	for _, f := range fs {
		if !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	return out
}

// Prune remove da resposta os campos que o tipo não declara, mantendo a estrutura.
// Fixtures podadas guardam só o que a CLI usa: menos dados reais para revisar e
// testes de contrato offline que verificam exatamente o contrato.
func Prune(data []byte, v any) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("resposta não é JSON: %w", err)
	}
	doc = prune(doc, reflect.TypeOf(v))
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func prune(val any, t reflect.Type) any {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if opaco(t) {
		return val
	}
	switch t.Kind() {
	case reflect.Struct:
		obj, ok := val.(map[string]any)
		if !ok {
			return val
		}
		out := map[string]any{}
		for _, f := range fields(t) {
			if child, ok := obj[f.name]; ok {
				out[f.name] = prune(child, f.typ)
			}
		}
		return out
	case reflect.Slice, reflect.Array:
		items, ok := val.([]any)
		if !ok {
			return val
		}
		for i := range items {
			items[i] = prune(items[i], t.Elem())
		}
		return items
	case reflect.Map:
		obj, ok := val.(map[string]any)
		if !ok {
			return val
		}
		for k, item := range obj {
			obj[k] = prune(item, t.Elem())
		}
		return obj
	}
	return val
}
