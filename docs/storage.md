# Storage

vulnkv uses an in-memory map with optional TTL per key. Persistence is done via a write-ahead log (WAL).

## WAL

Each mutating operation appends a record:

- SET: key, value, optional expiration
- DEL: key
- EXPIRE: key, new expiration

Records are appended in the same order as in-memory changes. On restart, the WAL is replayed to rebuild state.

## Snapshots

Snapshotting is stubbed for now. See `internal/persistence/snapshot.go`.
