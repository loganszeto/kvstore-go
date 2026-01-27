package persistence

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

var (
	magic      = [4]byte{'V', 'K', 'V', '1'}
	ErrCorrupt = errors.New("wal record corrupt")
)

type Op byte

const (
	OpSet    Op = 1
	OpDel    Op = 2
	OpExpire Op = 3
)

type Record struct {
	Op          Op
	Key         string
	Value       []byte
	ExpiresAtMs int64
}

func Encode(rec Record) ([]byte, error) {
	keyBytes := []byte(rec.Key)
	buf := bytes.NewBuffer(nil)
	if _, err := buf.Write(magic[:]); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(byte(rec.Op)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(keyBytes))); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(rec.Value))); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, rec.ExpiresAtMs); err != nil {
		return nil, err
	}
	if _, err := buf.Write(keyBytes); err != nil {
		return nil, err
	}
	if _, err := buf.Write(rec.Value); err != nil {
		return nil, err
	}
	crc := crc32.ChecksumIEEE(buf.Bytes())
	if err := binary.Write(buf, binary.LittleEndian, crc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeFrom(r io.Reader) (Record, error) {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return Record{}, err
	}
	if header != magic {
		return Record{}, ErrCorrupt
	}
	opByte := make([]byte, 1)
	if _, err := io.ReadFull(r, opByte); err != nil {
		return Record{}, err
	}
	var keyLen uint32
	if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
		return Record{}, err
	}
	var valLen uint32
	if err := binary.Read(r, binary.LittleEndian, &valLen); err != nil {
		return Record{}, err
	}
	var expiresAt int64
	if err := binary.Read(r, binary.LittleEndian, &expiresAt); err != nil {
		return Record{}, err
	}
	keyBytes := make([]byte, keyLen)
	if _, err := io.ReadFull(r, keyBytes); err != nil {
		return Record{}, err
	}
	valBytes := make([]byte, valLen)
	if _, err := io.ReadFull(r, valBytes); err != nil {
		return Record{}, err
	}
	var crc uint32
	if err := binary.Read(r, binary.LittleEndian, &crc); err != nil {
		return Record{}, err
	}

	body := bytes.NewBuffer(nil)
	body.Write(header[:])
	body.WriteByte(opByte[0])
	binary.Write(body, binary.LittleEndian, keyLen)
	binary.Write(body, binary.LittleEndian, valLen)
	binary.Write(body, binary.LittleEndian, expiresAt)
	body.Write(keyBytes)
	body.Write(valBytes)
	if crc32.ChecksumIEEE(body.Bytes()) != crc {
		return Record{}, ErrCorrupt
	}

	return Record{
		Op:          Op(opByte[0]),
		Key:         string(keyBytes),
		Value:       valBytes,
		ExpiresAtMs: expiresAt,
	}, nil
}
