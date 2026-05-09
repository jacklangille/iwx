package requestctx

import (
	"context"
	"net/http"
	"strings"

	"github.com/nats-io/nuid"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
	userIDKey    contextKey = "user_id"
)

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
	HeaderTraceID       = "X-Trace-ID"
)

func WithRequestID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, requestIDKey, strings.TrimSpace(value))
}

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return strings.TrimSpace(value)
}

func WithTraceID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, traceIDKey, strings.TrimSpace(value))
}

func TraceID(ctx context.Context) string {
	value, _ := ctx.Value(traceIDKey).(string)
	return strings.TrimSpace(value)
}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserID(ctx context.Context) (int64, bool) {
	value, ok := ctx.Value(userIDKey).(int64)
	return value, ok
}

func ExtractOrGenerate(r *http.Request) (requestID string, traceID string) {
	if r == nil {
		next := nuid.Next()
		return next, next
	}

	requestID = firstNonBlank(
		r.Header.Get(HeaderRequestID),
		r.Header.Get(HeaderCorrelationID),
		r.Header.Get(HeaderTraceID),
	)
	if requestID == "" {
		requestID = nuid.Next()
	}

	traceID = firstNonBlank(
		r.Header.Get(HeaderTraceID),
		r.Header.Get(HeaderCorrelationID),
		r.Header.Get(HeaderRequestID),
	)
	if traceID == "" {
		traceID = requestID
	}

	return requestID, traceID
}

func ApplyHeaders(w http.ResponseWriter, requestID string, traceID string) {
	if w == nil {
		return
	}

	if strings.TrimSpace(requestID) != "" {
		w.Header().Set(HeaderRequestID, strings.TrimSpace(requestID))
	}
	if strings.TrimSpace(traceID) != "" {
		w.Header().Set(HeaderTraceID, strings.TrimSpace(traceID))
		w.Header().Set(HeaderCorrelationID, strings.TrimSpace(traceID))
	}
}

func WithHTTPContext(ctx context.Context, requestID string, traceID string) context.Context {
	ctx = WithRequestID(ctx, requestID)
	ctx = WithTraceID(ctx, traceID)
	return ctx
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
