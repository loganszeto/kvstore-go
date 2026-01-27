package server

import (
	"bufio"
	"errors"
	"io"
	"net"

	"github.com/loganszeto/vulnkv/internal/protocol"
)

func (s *Server) handleConn(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)
	writer := bufio.NewWriter(c)

	for {
		req, err := protocol.ReadRequest(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			s.stats.RecordError()
			_ = protocol.WriteResponse(writer, protocol.Response{Kind: "ERR", Err: err.Error()})
			_ = writer.Flush()
			continue
		}
		resp := s.dispatch(req)
		if err := protocol.WriteResponse(writer, resp); err != nil {
			return
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}
