package store

type Options struct{}

func NewStore(_ Options) Store {
	return NewMemTable()
}
