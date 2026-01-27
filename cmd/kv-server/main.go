package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/loganszeto/vulnkv/internal/persistence"
	"github.com/loganszeto/vulnkv/internal/server"
	"github.com/loganszeto/vulnkv/internal/stats"
	"github.com/loganszeto/vulnkv/internal/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:7379", "listen address")
	dataDir := flag.String("data_dir", "./data", "data directory")
	fsync := flag.Bool("fsync", false, "fsync on each write")
	flag.Parse()

	wal, err := persistence.OpenWAL(*dataDir, persistence.Options{Fsync: *fsync})
	if err != nil {
		log.Fatalf("open wal: %v", err)
	}
	defer wal.Close()

	st := store.NewStore(store.Options{})

	if err := persistence.Replay(persistence.WALPath(*dataDir), st); err != nil {
		log.Fatalf("replay wal: %v", err)
	}

	stats := stats.New()
	srv := server.New(*addr, st, wal, stats)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("vulnkv listening on %s", *addr)
	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
