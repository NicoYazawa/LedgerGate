package log

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestNewWritesJSONWithService(t *testing.T) {
	originalStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stderr = originalStderr
		_ = reader.Close()
		_ = writer.Close()
	})

	New("gateway", "info").Info("gateway started")

	if err := writer.Close(); err != nil {
		t.Fatalf("close log writer: %v", err)
	}
	os.Stderr = originalStderr

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read log output: %v", err)
	}
	var record map[string]any
	if err := json.Unmarshal(output, &record); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", output, err)
	}
	if record["service"] != "gateway" {
		t.Errorf("service = %v, want gateway", record["service"])
	}
	if record["msg"] != "gateway started" {
		t.Errorf("msg = %v, want gateway started", record["msg"])
	}
}
