package persistence

import (
	"errors"
	"io"
	"os"

	"github.com/loganszeto/vulnkv/internal/store"
)

func Replay(walPath string, st store.Store) error {
	f, err := os.Open(walPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()

	for {
		rec, err := DecodeFrom(f)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, ErrCorrupt) {
				return nil
			}
			return err
		}
		switch rec.Op {
		case OpSet:
			st.Set(rec.Key, rec.Value, rec.ExpiresAtMs)
		case OpDel:
			st.Del(rec.Key)
		case OpExpire:
			st.Expire(rec.Key, rec.ExpiresAtMs)
		}
	}
}
