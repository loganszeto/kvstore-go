package integration

import (
	"testing"
	"time"
)

func TestTTL(t *testing.T) {
	dir := t.TempDir()
	addr, stop := startServer(t, dir)
	defer stop()

	resp := send(t, addr, "SETEX temp 1 1", "x")
	if resp.Kind != "OK" {
		t.Fatalf("expected OK, got %v", resp.Kind)
	}
	time.Sleep(1200 * time.Millisecond)
	resp = send(t, addr, "GET temp", "")
	if resp.Kind != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %v", resp.Kind)
	}
}
