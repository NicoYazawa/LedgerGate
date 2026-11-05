package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NicoYazawa/LedgerGate/internal/platform/config"
)

func TestHandlerServesHealthReadinessAndMetrics(t *testing.T) {
	handler := NewHandler(testConfig(), Dependencies{}, http.NotFoundHandler(), discardLogger())

	for _, path := range []string{"/healthz", "/readyz"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, response.Code)
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200", response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
		t.Errorf("Content-Type = %q, want Prometheus text content type", contentType)
	}
	body := response.Body.String()
	for _, metric := range []string{"ledgergate_http_requests_total", "ledgergate_http_in_flight"} {
		if !strings.Contains(body, metric) {
			t.Errorf("metrics output does not contain %q:\n%s", metric, body)
		}
	}
}

func TestReadinessReturnsServiceUnavailableWhenDependencyFails(t *testing.T) {
	deps := Dependencies{Ready: []ReadyCheck{{
		Name: "database",
		Check: func(context.Context) error {
			return errors.New("connection refused")
		},
	}}}
	handler := NewHandler(testConfig(), deps, http.NotFoundHandler(), discardLogger())
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	if !strings.Contains(response.Body.String(), "database") {
		t.Errorf("body = %q, want failed dependency name", response.Body.String())
	}
}

func TestServeDrainsInflightRequestBeforeShutdown(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	routes := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "complete")
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	cfg := testConfig()
	cfg.ShutdownTimeout = 2 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, cfg, listener, Dependencies{}, routes, discardLogger())
	}()

	responseDone := make(chan *http.Response, 1)
	requestErr := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + listener.Addr().String() + "/slow")
		if err != nil {
			requestErr <- err
			return
		}
		responseDone <- response
	}()

	select {
	case <-started:
	case err := <-requestErr:
		t.Fatalf("request error = %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("request did not reach handler")
	}

	cancel()
	select {
	case err := <-done:
		t.Fatalf("Serve() returned before draining request: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	select {
	case response := <-responseDone:
		defer func() {
			_ = response.Body.Close()
		}()
		if response.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", response.StatusCode)
		}
	case err := <-requestErr:
		t.Fatalf("request error = %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("draining request did not complete")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve() did not return after drain")
	}
}

func TestRunReportsListenError(t *testing.T) {
	err := Run(context.Background(), config.Config{
		Service: "gateway", HTTPAddr: "not-an-address", LogLevel: "info",
	}, Dependencies{}, nil)
	if err == nil || !strings.Contains(err.Error(), "listen on") {
		t.Errorf("Run() error = %v, want listen error", err)
	}
}

func TestRunReturnsWhenContextIsAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(ctx, testConfig(), Dependencies{}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestServeClosesInflightRequestAtShutdownTimeout(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	routes := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	cfg := testConfig()
	cfg.ShutdownTimeout = 20 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, cfg, listener, Dependencies{}, routes, discardLogger())
	}()

	requestDone := make(chan struct{})
	go func() {
		response, _ := http.Get("http://" + listener.Addr().String() + "/slow")
		if response != nil {
			_ = response.Body.Close()
		}
		close(requestDone)
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("request did not reach handler")
	}
	cancel()

	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "shutdown HTTP server") {
			t.Errorf("Serve() error = %v, want shutdown timeout", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve() did not return after shutdown timeout")
	}
	close(release)
	select {
	case <-requestDone:
	case <-time.After(2 * time.Second):
		t.Fatal("request did not finish")
	}
}

func TestNewHandlerFallsBackWhenDependenciesAreNil(t *testing.T) {
	handler := NewHandler(testConfig(), Dependencies{}, nil, nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if response.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", response.Code)
	}
}

func testConfig() config.Config {
	return config.Config{
		Env: "test", Service: "gateway", HTTPAddr: "127.0.0.1:0",
		ShutdownTimeout: time.Second, LogLevel: "debug",
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
