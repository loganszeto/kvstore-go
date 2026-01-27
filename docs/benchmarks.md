# Benchmarks

Run a quick load test:

```
go run ./cmd/kv-bench --addr 127.0.0.1:7379 --ops 10000 --threads 8 --clients 8
```

This prints ops/sec and p50/p95/p99 latencies.
