package store

type Store interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, expiresAtMs int64)
	Del(key string) bool
	Exists(key string) bool
	Expire(key string, expiresAtMs int64) bool
	Keys(prefix string) []string
}
