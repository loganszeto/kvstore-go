package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/loganszeto/kvstore-go/internal/protocol"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:7379", "server address")
	clients := flag.Int("clients", 10, "number of connections")
	threads := flag.Int("threads", 10, "goroutines")
	ops := flag.Int("ops", 10000, "total operations")
	ratioGet := flag.Float64("ratio_get", 0.8, "get ratio")
	valueSize := flag.Int("value_size", 128, "value size bytes")
	flag.Parse()

	if *threads <= 0 || *clients <= 0 {
		fmt.Fprintln(os.Stderr, "threads and clients must be > 0")
		os.Exit(1)
	}

	value := strings.Repeat("x", *valueSize)
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key:%d", i)
	}

	var opsDone atomic.Int64
	latCh := make(chan time.Duration, *ops)

	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", *addr)
			if err != nil {
				return
			}
			defer conn.Close()
			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
			for {
				idx := int(opsDone.Add(1)) - 1
				if idx >= *ops {
					return
				}
				key := keys[rng.Intn(len(keys))]
				doGet := rng.Float64() < *ratioGet
				startOp := time.Now()
				if doGet {
					if _, err := writer.WriteString("GET " + key + "\n"); err != nil {
						return
					}
				} else {
					if _, err := writer.WriteString(fmt.Sprintf("SET %s %d\n", key, len(value))); err != nil {
						return
					}
					if _, err := writer.WriteString(value + "\n"); err != nil {
						return
					}
				}
				if err := writer.Flush(); err != nil {
					return
				}
				if _, err := protocol.ReadResponse(reader); err != nil {
					return
				}
				latCh <- time.Since(startOp)
			}
		}(i)
		if (i+1)%*clients == 0 {
			time.Sleep(5 * time.Millisecond)
		}
	}

	wg.Wait()
	close(latCh)

	elapsed := time.Since(start)
	totalOps := opsDone.Load()
	if totalOps > int64(*ops) {
		totalOps = int64(*ops)
	}
	fmt.Printf("Total ops: %d\n", totalOps)
	fmt.Printf("Elapsed: %s\n", elapsed)
	fmt.Printf("Ops/sec: %.2f\n", float64(totalOps)/elapsed.Seconds())

	var lats []time.Duration
	for d := range latCh {
		lats = append(lats, d)
	}
	printLatencyStats(lats)
}

func printLatencyStats(lats []time.Duration) {
	if len(lats) == 0 {
		fmt.Println("No latency samples")
		return
	}
	sortDurations(lats)
	p50 := lats[len(lats)*50/100]
	p95 := lats[len(lats)*95/100]
	p99 := lats[len(lats)*99/100]
	fmt.Printf("p50: %s\n", p50)
	fmt.Printf("p95: %s\n", p95)
	fmt.Printf("p99: %s\n", p99)
}

func sortDurations(vals []time.Duration) {
	if len(vals) < 2 {
		return
	}
	quickSort(vals, 0, len(vals)-1)
}

func quickSort(a []time.Duration, lo, hi int) {
	if lo >= hi {
		return
	}
	p := partition(a, lo, hi)
	quickSort(a, lo, p-1)
	quickSort(a, p+1, hi)
}

func partition(a []time.Duration, lo, hi int) int {
	pivot := a[hi]
	i := lo
	for j := lo; j < hi; j++ {
		if a[j] < pivot {
			a[i], a[j] = a[j], a[i]
			i++
		}
	}
	a[i], a[hi] = a[hi], a[i]
	return i
}
