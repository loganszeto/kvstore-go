# vulnkv

vulnkv is a Redis-lite key-value store built in Go. It supports multiple concurrent clients over TCP and persists writes with a WAL.

## Features

- Concurrent TCP server
- WAL persistence + crash recovery
- TTL support
- CLI client + benchmark tool

## Quickstart

```
go run ./cmd/kv-server --data_dir ./data
go run ./cmd/kv-cli SET hello world
```

## Protocol

See `docs/protocol.md` for the full spec. Examples:

```
SET hello 5
world
GET hello
```

## Persistence model

Mutations append a WAL record before updating in-memory state. On restart, the WAL is replayed in order. Durability can be increased by enabling `--fsync`.

## Benchmarks

```
go run ./cmd/kv-bench --addr 127.0.0.1:7379 --ops 10000 --threads 8 --clients 8
```

## CVE/CPE caching demo

```
./scripts/load_cve_sample.sh
```

## Gotcha to avoid

Don’t start with NVD, Postgres, or CVE matching logic. Start with protocol, SET/GET, WAL, and recovery. The CVE/CPE demo is just sample data.
