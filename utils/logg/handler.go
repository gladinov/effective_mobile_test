package logg

import (
	"context"
	"log/slog"

	"github.com/gladinov/contracts/trace"
)

type traceHandler struct {
	next slog.Handler
}

func New(next slog.Handler) *traceHandler {
	return &traceHandler{next: next}
}

func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := trace.TraceIDFromContext(ctx); ok {
		r.AddAttrs(slog.String("trace_id", traceID))
	}
	return h.next.Handle(ctx, r)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{next: h.next.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{next: h.next.WithGroup(name)}
}
