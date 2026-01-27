package store

import (
	"sort"
	"sync"
)

type entry struct {
	v           []byte
	expiresAtMs int64
}

type MemTable struct {
	mu sync.RWMutex
	m  map[string]entry
}

func NewMemTable() *MemTable {
	return &MemTable{
		m: make(map[string]entry),
	}
}

func (t *MemTable) Get(key string) ([]byte, bool) {
	now := NowMs()
	t.mu.RLock()
	ent, ok := t.m[key]
	t.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if IsExpired(ent.expiresAtMs, now) {
		t.mu.Lock()
		delete(t.m, key)
		t.mu.Unlock()
		return nil, false
	}
	out := make([]byte, len(ent.v))
	copy(out, ent.v)
	return out, true
}

func (t *MemTable) Set(key string, value []byte, expiresAtMs int64) {
	buf := make([]byte, len(value))
	copy(buf, value)
	t.mu.Lock()
	t.m[key] = entry{v: buf, expiresAtMs: expiresAtMs}
	t.mu.Unlock()
}

func (t *MemTable) Del(key string) bool {
	t.mu.Lock()
	_, ok := t.m[key]
	if ok {
		delete(t.m, key)
	}
	t.mu.Unlock()
	return ok
}

func (t *MemTable) Exists(key string) bool {
	now := NowMs()
	t.mu.RLock()
	ent, ok := t.m[key]
	t.mu.RUnlock()
	if !ok {
		return false
	}
	if IsExpired(ent.expiresAtMs, now) {
		t.mu.Lock()
		delete(t.m, key)
		t.mu.Unlock()
		return false
	}
	return true
}

func (t *MemTable) Expire(key string, expiresAtMs int64) bool {
	now := NowMs()
	t.mu.Lock()
	ent, ok := t.m[key]
	if !ok {
		t.mu.Unlock()
		return false
	}
	if IsExpired(ent.expiresAtMs, now) {
		delete(t.m, key)
		t.mu.Unlock()
		return false
	}
	ent.expiresAtMs = expiresAtMs
	t.m[key] = ent
	t.mu.Unlock()
	return true
}

func (t *MemTable) Keys(prefix string) []string {
	now := NowMs()
	out := make([]string, 0)
	t.mu.Lock()
	for k, ent := range t.m {
		if IsExpired(ent.expiresAtMs, now) {
			delete(t.m, k)
			continue
		}
		if prefix == "" || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, k)
		}
	}
	t.mu.Unlock()
	sort.Strings(out)
	return out
}
