package requestctx

import (
	"net/http/httptest"
	"testing"
)

func TestExtractOrGeneratePrefersHeaders(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set(HeaderRequestID, "req-1")
	request.Header.Set(HeaderTraceID, "trace-1")

	requestID, traceID := ExtractOrGenerate(request)
	if requestID != "req-1" {
		t.Fatalf("expected request id req-1, got %q", requestID)
	}
	if traceID != "trace-1" {
		t.Fatalf("expected trace id trace-1, got %q", traceID)
	}
}

func TestApplyHeadersSetsRequestTraceAndCorrelationIDs(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	ApplyHeaders(recorder, "req-9", "trace-9")

	if got := recorder.Header().Get(HeaderRequestID); got != "req-9" {
		t.Fatalf("expected request id header req-9, got %q", got)
	}
	if got := recorder.Header().Get(HeaderTraceID); got != "trace-9" {
		t.Fatalf("expected trace id header trace-9, got %q", got)
	}
	if got := recorder.Header().Get(HeaderCorrelationID); got != "trace-9" {
		t.Fatalf("expected correlation id header trace-9, got %q", got)
	}
}
