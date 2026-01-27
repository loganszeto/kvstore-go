package protocol

type CmdType int

const (
	CmdPing CmdType = iota
	CmdGet
	CmdSet
	CmdDel
	CmdExists
	CmdExpire
	CmdSetEx
	CmdKeys
	CmdStats
)

type Request struct {
	Type       CmdType
	Key        string
	TTLSeconds int64
	Value      []byte
	Prefix     string
}

type Response struct {
	Kind  string
	Value []byte
	Int   int64
	Array []string
	Err   string
}
