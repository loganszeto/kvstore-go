package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var ErrInvalidRequest = errors.New("invalid request")

func ReadRequest(r *bufio.Reader) (Request, error) {
	line, err := readLine(r)
	if err != nil {
		return Request{}, err
	}
	if line == "" {
		return Request{}, ErrInvalidRequest
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Request{}, ErrInvalidRequest
	}
	cmd := strings.ToUpper(fields[0])
	switch cmd {
	case "PING":
		return Request{Type: CmdPing}, nil
	case "GET":
		if len(fields) != 2 {
			return Request{}, ErrInvalidRequest
		}
		return Request{Type: CmdGet, Key: fields[1]}, nil
	case "DEL":
		if len(fields) != 2 {
			return Request{}, ErrInvalidRequest
		}
		return Request{Type: CmdDel, Key: fields[1]}, nil
	case "EXISTS":
		if len(fields) != 2 {
			return Request{}, ErrInvalidRequest
		}
		return Request{Type: CmdExists, Key: fields[1]}, nil
	case "EXPIRE":
		if len(fields) != 3 {
			return Request{}, ErrInvalidRequest
		}
		ttl, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return Request{}, ErrInvalidRequest
		}
		return Request{Type: CmdExpire, Key: fields[1], TTLSeconds: ttl}, nil
	case "KEYS":
		if len(fields) != 2 {
			return Request{}, ErrInvalidRequest
		}
		prefix := fields[1]
		if strings.HasSuffix(prefix, "*") {
			prefix = strings.TrimSuffix(prefix, "*")
		}
		return Request{Type: CmdKeys, Prefix: prefix}, nil
	case "SET":
		if len(fields) != 3 {
			return Request{}, ErrInvalidRequest
		}
		key := fields[1]
		n, err := strconv.Atoi(fields[2])
		if err != nil || n < 0 {
			return Request{}, ErrInvalidRequest
		}
		val, err := readBytes(r, n)
		if err != nil {
			return Request{}, err
		}
		return Request{Type: CmdSet, Key: key, Value: val}, nil
	case "SETEX":
		if len(fields) != 4 {
			return Request{}, ErrInvalidRequest
		}
		key := fields[1]
		ttl, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return Request{}, ErrInvalidRequest
		}
		n, err := strconv.Atoi(fields[3])
		if err != nil || n < 0 {
			return Request{}, ErrInvalidRequest
		}
		val, err := readBytes(r, n)
		if err != nil {
			return Request{}, err
		}
		return Request{Type: CmdSetEx, Key: key, TTLSeconds: ttl, Value: val}, nil
	case "STATS":
		return Request{Type: CmdStats}, nil
	default:
		return Request{}, fmt.Errorf("%w: unknown command %q", ErrInvalidRequest, cmd)
	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line, nil
}

func readBytes(r *bufio.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b == '\r' {
		next, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if next != '\n' {
			return nil, ErrInvalidRequest
		}
		return buf, nil
	}
	if b != '\n' {
		return nil, ErrInvalidRequest
	}
	return buf, nil
}
