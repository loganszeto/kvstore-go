package persistence

import (
	"errors"

	"github.com/loganszeto/kvstore-go/internal/store"
)

var ErrSnapshotUnsupported = errors.New("snapshot not implemented")

func SaveSnapshot(_ string, _ store.Store) error {
	return ErrSnapshotUnsupported
}

func LoadSnapshot(_ string, _ store.Store) error {
	return ErrSnapshotUnsupported
}
