# Concurrency

- The server accepts multiple TCP clients concurrently.
- Each connection is handled by its own goroutine.
- The in-memory store uses a RWMutex for safe concurrent access.
