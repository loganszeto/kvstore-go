package persistence

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
)

const walFileName = "wal.log"

type Options struct {
	FlushEveryMs int
	Fsync        bool
}

type WAL struct {
	mu   sync.Mutex
	f    *os.File
	buf  *bufio.Writer
	opts Options
}

func WALPath(dir string) string {
	return filepath.Join(dir, walFileName)
}

func OpenWAL(dir string, opts Options) (*WAL, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := WALPath(dir)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &WAL{
		f:    f,
		buf:  bufio.NewWriter(f),
		opts: opts,
	}, nil
}

func (w *WAL) Append(rec Record) error {
	data, err := Encode(rec)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := w.buf.Write(data); err != nil {
		return err
	}
	if err := w.buf.Flush(); err != nil {
		return err
	}
	if w.opts.Fsync {
		return w.f.Sync()
	}
	return nil
}

func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.buf.Flush(); err != nil {
		return err
	}
	return w.f.Close()
}
