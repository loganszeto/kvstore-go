package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/loganszeto/vulnkv/internal/protocol"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:7379", "server address")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	if flag.NArg() > 0 {
		if err := sendCommand(writer, flag.Args()); err != nil {
			fmt.Fprintf(os.Stderr, "send: %v\n", err)
			os.Exit(1)
		}
		resp, err := protocol.ReadResponse(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read: %v\n", err)
			os.Exit(1)
		}
		printResponse(resp)
		return
	}

	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := in.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}
		if strings.EqualFold(args[0], "QUIT") || strings.EqualFold(args[0], "EXIT") {
			return
		}
		if err := sendCommand(writer, args); err != nil {
			fmt.Fprintf(os.Stderr, "send: %v\n", err)
			continue
		}
		resp, err := protocol.ReadResponse(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read: %v\n", err)
			return
		}
		printResponse(resp)
	}
}

func sendCommand(w *bufio.Writer, args []string) error {
	cmd := strings.ToUpper(args[0])
	switch cmd {
	case "SET":
		if len(args) < 3 {
			return fmt.Errorf("usage: SET key value")
		}
		key := args[1]
		value := strings.Join(args[2:], " ")
		if _, err := w.WriteString(fmt.Sprintf("SET %s %d\n", key, len(value))); err != nil {
			return err
		}
		if _, err := w.WriteString(value + "\n"); err != nil {
			return err
		}
	case "SETEX":
		if len(args) < 4 {
			return fmt.Errorf("usage: SETEX key ttl value")
		}
		key := args[1]
		ttl := args[2]
		value := strings.Join(args[3:], " ")
		if _, err := w.WriteString(fmt.Sprintf("SETEX %s %s %d\n", key, ttl, len(value))); err != nil {
			return err
		}
		if _, err := w.WriteString(value + "\n"); err != nil {
			return err
		}
	default:
		if _, err := w.WriteString(strings.Join(args, " ") + "\n"); err != nil {
			return err
		}
	}
	return w.Flush()
}

func printResponse(resp protocol.Response) {
	switch resp.Kind {
	case "OK":
		fmt.Println("OK")
	case "ERR":
		fmt.Println("ERR", resp.Err)
	case "NOT_FOUND":
		fmt.Println("NOT_FOUND")
	case "VALUE":
		fmt.Println(string(resp.Value))
	case "INT":
		fmt.Println(resp.Int)
	case "ARRAY":
		for _, item := range resp.Array {
			fmt.Println(item)
		}
	default:
		fmt.Println("ERR unknown response")
	}
}
