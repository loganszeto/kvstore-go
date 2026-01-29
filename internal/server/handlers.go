package server

import (
	"sort"

	"github.com/loganszeto/kvstore-go/internal/persistence"
	"github.com/loganszeto/kvstore-go/internal/protocol"
	"github.com/loganszeto/kvstore-go/internal/stats"
	"github.com/loganszeto/kvstore-go/internal/store"
)

func (s *Server) dispatch(req protocol.Request) protocol.Response {
	return Dispatch(s.st, s.wal, s.stats, req)
}

func Dispatch(st store.Store, wal *persistence.WAL, stats *stats.Stats, req protocol.Request) protocol.Response {
	switch req.Type {
	case protocol.CmdPing:
		return protocol.Response{Kind: "OK"}
	case protocol.CmdGet:
		val, ok := st.Get(req.Key)
		stats.RecordGet(ok)
		if !ok {
			return protocol.Response{Kind: "NOT_FOUND"}
		}
		return protocol.Response{Kind: "VALUE", Value: val}
	case protocol.CmdSet:
		if err := wal.Append(persistence.Record{
			Op:          persistence.OpSet,
			Key:         req.Key,
			Value:       req.Value,
			ExpiresAtMs: 0,
		}); err != nil {
			stats.RecordError()
			return protocol.Response{Kind: "ERR", Err: err.Error()}
		}
		st.Set(req.Key, req.Value, 0)
		stats.RecordSet()
		return protocol.Response{Kind: "OK"}
	case protocol.CmdSetEx:
		expiresAt := store.NowMs() + req.TTLSeconds*1000
		if err := wal.Append(persistence.Record{
			Op:          persistence.OpSet,
			Key:         req.Key,
			Value:       req.Value,
			ExpiresAtMs: expiresAt,
		}); err != nil {
			stats.RecordError()
			return protocol.Response{Kind: "ERR", Err: err.Error()}
		}
		st.Set(req.Key, req.Value, expiresAt)
		stats.RecordSet()
		return protocol.Response{Kind: "OK"}
	case protocol.CmdDel:
		if err := wal.Append(persistence.Record{
			Op:  persistence.OpDel,
			Key: req.Key,
		}); err != nil {
			stats.RecordError()
			return protocol.Response{Kind: "ERR", Err: err.Error()}
		}
		ok := st.Del(req.Key)
		stats.RecordDel()
		if ok {
			return protocol.Response{Kind: "INT", Int: 1}
		}
		return protocol.Response{Kind: "INT", Int: 0}
	case protocol.CmdExists:
		if st.Exists(req.Key) {
			return protocol.Response{Kind: "INT", Int: 1}
		}
		return protocol.Response{Kind: "INT", Int: 0}
	case protocol.CmdExpire:
		expiresAt := store.NowMs() + req.TTLSeconds*1000
		if err := wal.Append(persistence.Record{
			Op:          persistence.OpExpire,
			Key:         req.Key,
			ExpiresAtMs: expiresAt,
		}); err != nil {
			stats.RecordError()
			return protocol.Response{Kind: "ERR", Err: err.Error()}
		}
		if st.Expire(req.Key, expiresAt) {
			return protocol.Response{Kind: "INT", Int: 1}
		}
		return protocol.Response{Kind: "INT", Int: 0}
	case protocol.CmdKeys:
		keys := st.Keys(req.Prefix)
		return protocol.Response{Kind: "ARRAY", Array: keys}
	case protocol.CmdStats:
		snap := stats.Snapshot()
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
		stats.RecordError()
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
