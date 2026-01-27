package integration

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/loganszeto/vulnkv/internal/protocol"
)

func TestConcurrency(t *testing.T) {
	dir := t.TempDir()
	addr, stop := startServer(t, dir)
	defer stop()

	const goroutines = 50
	const loops = 50

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close()
			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)
			key := fmt.Sprintf("k:%d", id)
			for j := 0; j < loops; j++ {
				if err := writeCommand(writer, fmt.Sprintf("SET %s 1", key), "x"); err != nil {
					errCh <- err
					return
				}
				if _, err := protocol.ReadResponse(reader); err != nil {
					errCh <- err
					return
				}
				if err := writeCommand(writer, fmt.Sprintf("GET %s", key), ""); err != nil {
					errCh <- err
					return
				}
				resp, err := protocol.ReadResponse(reader)
				if err != nil {
					errCh <- err
					return
				}
				if resp.Kind != "VALUE" {
					errCh <- fmt.Errorf("get err %v", resp.Kind)
					return
				}
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func writeCommand(w *bufio.Writer, line string, body string) error {
	if _, err := w.WriteString(line + "\n"); err != nil {
		return err
	}
	if body != "" {
		if _, err := w.WriteString(body + "\n"); err != nil {
			return err
		}
	}
	return w.Flush()
}
