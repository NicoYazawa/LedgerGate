package httpx

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get("X-Request-ID")
		if !validID(requestID, 128) {
			requestID = "req_" + rand.Text()
		}
		traceID := randomID(32)
		parentSpanID := randomID(16)

		ctx := context.WithValue(context.WithValue(request.Context(), requestIDKey, requestID), traceIDKey, traceID)
		request = request.WithContext(ctx)
		request.Header.Set("X-Request-ID", requestID)
		writer.Header().Set("X-Request-ID", requestID)
		writer.Header().Set("traceparent", "00-"+traceID+"-"+parentSpanID+"-01")

		next.ServeHTTP(writer, request)
	})
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		tracked := &statusWriter{ResponseWriter: writer, status: http.StatusOK}

		next.ServeHTTP(tracked, request)

		logger.InfoContext(request.Context(), "http.request",
			"request_id", RequestIDFromContext(request.Context()), "trace_id", TraceIDFromContext(request.Context()),
			"method", request.Method, "path", request.URL.Path,
			"status", tracked.status, "duration_ms", time.Since(started).Milliseconds())
	})
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}

func TraceIDFromContext(ctx context.Context) string {
	traceID, _ := ctx.Value(traceIDKey).(string)
	return traceID
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (writer *statusWriter) WriteHeader(status int) {
	if writer.wroteHeader {
		return
	}
	writer.status = status
	writer.wroteHeader = true
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *statusWriter) Write(data []byte) (int, error) {
	if !writer.wroteHeader {
		writer.WriteHeader(http.StatusOK)
	}
	return writer.ResponseWriter.Write(data)
}

func (writer *statusWriter) Unwrap() http.ResponseWriter { return writer.ResponseWriter }

func randomID(length int) string {
	sum := sha256.Sum256([]byte(rand.Text()))
	return hex.EncodeToString(sum[:])[:length]
}

func validID(value string, maximumLength int) bool {
	return value != "" && len(value) <= maximumLength && strings.IndexFunc(value, unicode.IsControl) == -1
}
