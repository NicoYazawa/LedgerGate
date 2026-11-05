package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoggingIncludesCorrelationFields(t *testing.T) {
	const requestID = "req_test-123"
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})).
		With("service", "gateway")
	handler := RequestID(Logging(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})))
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Request-ID", requestID)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
	}
	if got := response.Header().Get("X-Request-ID"); got != requestID {
		t.Errorf("X-Request-ID = %q, want %q", got, requestID)
	}
	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &record); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", output.Bytes(), err)
	}
	for key, want := range map[string]string{"service": "gateway", "request_id": requestID} {
		if got := record[key]; got != want {
			t.Errorf("%s = %v, want %s", key, got, want)
		}
	}
	traceID, _ := record["trace_id"].(string)
	if len(traceID) != 32 {
		t.Errorf("trace_id = %q, want 32 hex characters", traceID)
	}
}

func TestRequestIDAddsCorrelationIDsWhenHeadersAreAbsent(t *testing.T) {
	var requestID, traceID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID = RequestIDFromContext(r.Context())
		traceID = TraceIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if requestID == "" {
		t.Error("request ID is empty")
	}
	if traceID == "" {
		t.Error("trace ID is empty")
	}
	if len(traceID) != 32 {
		t.Errorf("trace ID length = %d, want 32", len(traceID))
	}
	parts := strings.Split(response.Header().Get("traceparent"), "-")
	if len(parts) != 4 || parts[1] != traceID || len(parts[2]) != 16 || len(parts[3]) != 2 {
		t.Errorf("traceparent = %q, want a valid W3C header for trace %q", response.Header().Get("traceparent"), traceID)
	}
}

func TestLoggingPreservesResponseWriterFlushing(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	flushed := make(chan struct{}, 1)
	handler := RequestID(Logging(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
		flushed <- struct{}{}
	})))
	server := httptest.NewServer(handler)
	defer server.Close()

	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	_ = response.Body.Close()
	select {
	case <-flushed:
	case <-time.After(time.Second):
		t.Fatal("handler did not flush")
	}
}

func TestContextFallbacksReturnEmptyStrings(t *testing.T) {
	if got := RequestIDFromContext(context.Background()); got != "" {
		t.Errorf("RequestIDFromContext() = %q, want empty", got)
	}
	if got := TraceIDFromContext(context.Background()); got != "" {
		t.Errorf("TraceIDFromContext() = %q, want empty", got)
	}
}
