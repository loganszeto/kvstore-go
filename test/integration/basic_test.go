package integration

import (
	"bufio"
	"context"
	"net"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/loganszeto/kvstore-go/internal/protocol"
)

func startServer(t *testing.T, dataDir string) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr = ln.Addr().String()
	ln.Close()

	cmd := exec.CommandContext(context.Background(), "go", "run", "./cmd/kv-server", "--addr", addr, "--data_dir", dataDir)
	cmd.Dir = filepath.Clean("../..")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	waitForReady(t, addr, 3*time.Second)

	return addr, func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}
}

func waitForReady(t *testing.T, addr string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server not ready on %s", addr)
}

func send(t *testing.T, addr string, line string, body string) protocol.Response {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	if _, err := w.WriteString(line + "\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if body != "" {
		if _, err := w.WriteString(body + "\n"); err != nil {
			t.Fatalf("write body: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	resp, err := protocol.ReadResponse(r)
	if err != nil {
		t.Fatalf("read resp: %v", err)
	}
	return resp
}

func TestBasicSetGet(t *testing.T) {
	dir := t.TempDir()
	addr, stop := startServer(t, dir)
	defer stop()

	resp := send(t, addr, "SET hello 5", "world")
	if resp.Kind != "OK" {
		t.Fatalf("expected OK, got %v", resp.Kind)
	}
	resp = send(t, addr, "GET hello", "")
	if resp.Kind != "VALUE" || string(resp.Value) != "world" {
		t.Fatalf("expected world, got %v %q", resp.Kind, string(resp.Value))
	}
	resp = send(t, addr, "DEL hello", "")
	if resp.Kind != "INT" || resp.Int != 1 {
		t.Fatalf("expected INT 1, got %v %d", resp.Kind, resp.Int)
	}
	resp = send(t, addr, "GET hello", "")
	if resp.Kind != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %v", resp.Kind)
	}
}
