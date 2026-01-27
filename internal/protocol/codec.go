package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var ErrInvalidResponse = errors.New("invalid response")

func WriteResponse(w *bufio.Writer, resp Response) error {
	switch resp.Kind {
	case "OK":
		_, err := w.WriteString("OK\n")
		return err
	case "ERR":
		_, err := w.WriteString("ERR " + resp.Err + "\n")
		return err
	case "NOT_FOUND":
		_, err := w.WriteString("NOT_FOUND\n")
		return err
	case "VALUE":
		if _, err := w.WriteString(fmt.Sprintf("VALUE %d\n", len(resp.Value))); err != nil {
			return err
		}
		if _, err := w.Write(resp.Value); err != nil {
			return err
		}
		_, err := w.WriteString("\n")
		return err
	case "INT":
		_, err := w.WriteString(fmt.Sprintf("INT %d\n", resp.Int))
		return err
	case "ARRAY":
		if _, err := w.WriteString(fmt.Sprintf("ARRAY %d\n", len(resp.Array))); err != nil {
			return err
		}
		for _, item := range resp.Array {
			if _, err := w.WriteString(item + "\n"); err != nil {
				return err
			}
		}
		return nil
	default:
		return ErrInvalidResponse
	}
}

func ReadResponse(r *bufio.Reader) (Response, error) {
	line, err := readLine(r)
	if err != nil {
		return Response{}, err
	}
	if line == "" {
		return Response{}, ErrInvalidResponse
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Response{}, ErrInvalidResponse
	}
	switch fields[0] {
	case "OK":
		return Response{Kind: "OK"}, nil
	case "ERR":
		msg := strings.TrimPrefix(line, "ERR ")
		if msg == "ERR" {
			msg = ""
		}
		return Response{Kind: "ERR", Err: msg}, nil
	case "NOT_FOUND":
		return Response{Kind: "NOT_FOUND"}, nil
	case "VALUE":
		if len(fields) != 2 {
			return Response{}, ErrInvalidResponse
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil || n < 0 {
			return Response{}, ErrInvalidResponse
		}
		val, err := readBytes(r, n)
		if err != nil {
			return Response{}, err
		}
		return Response{Kind: "VALUE", Value: val}, nil
	case "INT":
		if len(fields) != 2 {
			return Response{}, ErrInvalidResponse
		}
		n, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return Response{}, ErrInvalidResponse
		}
		return Response{Kind: "INT", Int: n}, nil
	case "ARRAY":
		if len(fields) != 2 {
			return Response{}, ErrInvalidResponse
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil || n < 0 {
			return Response{}, ErrInvalidResponse
		}
		items := make([]string, 0, n)
		for i := 0; i < n; i++ {
			item, err := readLine(r)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return Response{}, ErrInvalidResponse
				}
				return Response{}, err
			}
			items = append(items, item)
		}
		return Response{Kind: "ARRAY", Array: items}, nil
	default:
		return Response{}, ErrInvalidResponse
	}
}
