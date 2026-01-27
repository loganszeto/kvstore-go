package stats

import "sync/atomic"

type Stats struct {
	gets   atomic.Int64
	sets   atomic.Int64
	dels   atomic.Int64
	hits   atomic.Int64
	misses atomic.Int64
	errors atomic.Int64
}

func New() *Stats {
	return &Stats{}
}

func (s *Stats) RecordGet(hit bool) {
	s.gets.Add(1)
	if hit {
		s.hits.Add(1)
	} else {
		s.misses.Add(1)
	}
}

func (s *Stats) RecordSet() {
	s.sets.Add(1)
}

func (s *Stats) RecordDel() {
	s.dels.Add(1)
}

func (s *Stats) RecordError() {
	s.errors.Add(1)
}

func (s *Stats) Snapshot() map[string]int64 {
	return map[string]int64{
		"gets":   s.gets.Load(),
		"sets":   s.sets.Load(),
		"dels":   s.dels.Load(),
		"hits":   s.hits.Load(),
		"misses": s.misses.Load(),
		"errors": s.errors.Load(),
	}
}
