package store

import "time"

func NowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func IsExpired(expiresAtMs, nowMs int64) bool {
	return expiresAtMs > 0 && nowMs >= expiresAtMs
}
