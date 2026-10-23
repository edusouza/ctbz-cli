package ctbz

import (
	"context"
	"net/http"
	"time"
)

// Exchange é uma requisição feita pelo Client e a resposta (ou o erro de conexão), como o
// modo verbose a recebe. Traz os valores reais dos cabeçalhos e dos corpos: quem mostra é
// responsável por mascarar os segredos.
type Exchange struct {
	Method     string
	URL        string
	Header     http.Header // cabeçalhos enviados, inclusive Cookie
	Body       []byte
	Status     int // 0 sem resposta
	RespHeader http.Header
	RespBody   []byte
	Elapsed    time.Duration
	Err        error
}

// Tracer recebe cada troca feita pelo Client (modo verbose).
type Tracer func(Exchange)

type tracerKey struct{}

// WithTracer devolve um contexto em que todas as chamadas do Client são passadas ao tracer.
func WithTracer(ctx context.Context, t Tracer) context.Context {
	return context.WithValue(ctx, tracerKey{}, t)
}

// TracerFrom devolve o tracer do contexto, ou nil.
func TracerFrom(ctx context.Context) Tracer {
	t, _ := ctx.Value(tracerKey{}).(Tracer)
	return t
}
