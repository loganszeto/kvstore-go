package server

import (
	"sort"

	"github.com/loganszeto/kvstore-go/internal/persistence"
	"github.com/loganszeto/kvstore-go/internal/protocol"
	"github.com/loganszeto/kvstore-go/internal/store"
)

func (s *Server) dispatch(req protocol.Request) protocol.Response {
	switch req.Type {
	case protocol.CmdPing:
		return protocol.Response{Kind: "OK"}
	case protocol.CmdGet:
		val, ok := s.st.Get(req.Key)
		s.stats.RecordGet(ok)
		if !ok {
			return protocol.Response{Kind: "NOT_FOUND"}
		}
		return protocol.Response{Kind: "VALUE", Value: val}
	case protocol.CmdSet:
		if err := s.wal.Append(persistence.Record{
			Op:          persistence.OpSet,
			Key:         req.Key,
			Value:       req.Value,
			ExpiresAtMs: 0,
		}); err != nil {
			s.stats.RecordError()
			return protocol.Response{Kind: "ERR", Err: err.Error()}
		}
		s.st.Set(req.Key, req.Value, 0)
		s.stats.RecordSet()
		return protocol.Response{Kind: "OK"}
	case protocol.CmdSetEx:
		expiresAt := store.NowMs() + req.TTLSeconds*1000
		if err := s.wal.Append(persistence.Record{
			Op:          persistence.OpSet,
			Key:         req.Key,
			Value:       req.Value,
			ExpiresAtMs: expiresAt,
		}); err != nil {
			s.stats.RecordError()
			return protocol.Response{Kind: "ERR", Err: err.Error()}
		}
		s.st.Set(req.Key, req.Value, expiresAt)
		s.stats.RecordSet()
		return protocol.Response{Kind: "OK"}
	case protocol.CmdDel:
		if err := s.wal.Append(persistence.Record{
			Op:  persistence.OpDel,
			Key: req.Key,
		}); err != nil {
			s.stats.RecordError()
			return protocol.Response{Kind: "ERR", Err: err.Error()}
		}
		ok := s.st.Del(req.Key)
		s.stats.RecordDel()
		if ok {
			return protocol.Response{Kind: "INT", Int: 1}
		}
		return protocol.Response{Kind: "INT", Int: 0}
	case protocol.CmdExists:
		if s.st.Exists(req.Key) {
			return protocol.Response{Kind: "INT", Int: 1}
		}
		return protocol.Response{Kind: "INT", Int: 0}
	case protocol.CmdExpire:
		expiresAt := store.NowMs() + req.TTLSeconds*1000
		if err := s.wal.Append(persistence.Record{
			Op:          persistence.OpExpire,
			Key:         req.Key,
			ExpiresAtMs: expiresAt,
		}); err != nil {
			s.stats.RecordError()
			return protocol.Response{Kind: "ERR", Err: err.Error()}
		}
		if s.st.Expire(req.Key, expiresAt) {
			return protocol.Response{Kind: "INT", Int: 1}
		}
		return protocol.Response{Kind: "INT", Int: 0}
	case protocol.CmdKeys:
		keys := s.st.Keys(req.Prefix)
		return protocol.Response{Kind: "ARRAY", Array: keys}
	case protocol.CmdStats:
		snap := s.stats.Snapshot()
		keys := make([]string, 0, len(snap))
		for k := range snap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make([]string, 0, len(keys))
		for _, k := range keys {
			out = append(out, k+" "+itoa(snap[k]))
		}
		return protocol.Response{Kind: "ARRAY", Array: out}
	default:
		s.stats.RecordError()
		return protocol.Response{Kind: "ERR", Err: "unknown command"}
	}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		d := n % 10
		buf = append(buf, byte('0'+d))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
