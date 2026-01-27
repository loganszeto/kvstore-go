package server

import (
	"context"
	"net"

	"github.com/loganszeto/vulnkv/internal/persistence"
	"github.com/loganszeto/vulnkv/internal/stats"
	"github.com/loganszeto/vulnkv/internal/store"
)

type Server struct {
	addr  string
	st    store.Store
	wal   *persistence.WAL
	stats *stats.Stats
}

func New(addr string, st store.Store, wal *persistence.WAL, stats *stats.Stats) *Server {
	return &Server{
		addr:  addr,
		st:    st,
		wal:   wal,
		stats: stats,
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			continue
		}
		go s.handleConn(conn)
	}
}
