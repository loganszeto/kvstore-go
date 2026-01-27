package integration

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRecovery(t *testing.T) {
	dir := t.TempDir()
	addr, stop := startServer(t, dir)

	resp := send(t, addr, "SET a 1", "1")
	if resp.Kind != "OK" {
		t.Fatalf("expected OK, got %v", resp.Kind)
	}
	resp = send(t, addr, "SET b 1", "2")
	if resp.Kind != "OK" {
		t.Fatalf("expected OK, got %v", resp.Kind)
	}
	stop()

	time.Sleep(200 * time.Millisecond)
	addr, stop = startServer(t, filepath.Clean(dir))
	defer stop()

	resp = send(t, addr, "GET a", "")
	if resp.Kind != "VALUE" || string(resp.Value) != "1" {
		t.Fatalf("expected 1, got %v %q", resp.Kind, string(resp.Value))
	}
	resp = send(t, addr, "GET b", "")
	if resp.Kind != "VALUE" || string(resp.Value) != "2" {
		t.Fatalf("expected 2, got %v %q", resp.Kind, string(resp.Value))
	}
}
