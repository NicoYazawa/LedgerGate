package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/NicoYazawa/LedgerGate/internal/platform/config"
	"github.com/NicoYazawa/LedgerGate/internal/platform/httpx"
	platformlog "github.com/NicoYazawa/LedgerGate/internal/platform/log"
)

type ReadyCheck struct {
	Name  string
	Check func(context.Context) error
}
type Dependencies struct{ Ready []ReadyCheck }
type metrics struct {
	requests atomic.Uint64
	inFlight atomic.Int64
}

func Run(ctx context.Context, cfg config.Config, deps Dependencies, routes http.Handler) error {
	logger := platformlog.New(cfg.Service, cfg.LogLevel)
	listener, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.HTTPAddr, err)
	}
	logger.Info("server.started", "addr", listener.Addr().String())
	return Serve(ctx, cfg, listener, deps, routes, logger)
}
func Serve(
	ctx context.Context,
	cfg config.Config,
	listener net.Listener,
	deps Dependencies,
	routes http.Handler,
	logger *slog.Logger,
) error {
	handler := NewHandler(cfg, deps, routes, logger)
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			<-serverErrors
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		if err := <-serverErrors; !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		logger.Info("server.stopped")
		return nil
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	}
}
func NewHandler(cfg config.Config, deps Dependencies, routes http.Handler, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = platformlog.New(cfg.Service, cfg.LogLevel)
	}
	if routes == nil {
		routes = http.NotFoundHandler()
	}
	serviceMetrics := &metrics{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("GET /readyz", readiness(deps))
	mux.HandleFunc("GET /metrics", serviceMetrics.serve)
	mux.Handle("/", routes)
	var handler http.Handler = mux
	handler = serviceMetrics.instrument(handler)
	handler = httpx.Logging(logger, handler)
	return httpx.RequestID(handler)
}
func health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}
func readiness(deps Dependencies) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		checks := make(map[string]string, len(deps.Ready))
		status := http.StatusOK
		statusText := "ok"
		for _, dependency := range deps.Ready {
			if err := dependency.Check(request.Context()); err != nil {
				checks[dependency.Name] = "unavailable"
				status = http.StatusServiceUnavailable
				statusText = "unavailable"
				continue
			}
			checks[dependency.Name] = "ok"
		}
		writeJSON(writer, status, map[string]any{"status": statusText, "checks": checks})
	}
}
func (metrics *metrics) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		metrics.inFlight.Add(1)
		defer metrics.inFlight.Add(-1)
		defer metrics.requests.Add(1)
		next.ServeHTTP(writer, request)
	})
}

func (metrics *metrics) serve(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = fmt.Fprintf(writer, "# TYPE ledgergate_http_requests_total counter\nledgergate_http_requests_total %d\n# TYPE ledgergate_http_in_flight gauge\nledgergate_http_in_flight %d\n", metrics.requests.Load(), metrics.inFlight.Load())
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
