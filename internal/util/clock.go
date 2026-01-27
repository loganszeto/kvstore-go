package util

import "time"

type Clock interface {
	NowMs() int64
}

type RealClock struct{}

func (RealClock) NowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
