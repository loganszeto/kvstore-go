package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/websocket"
	"google.golang.org/api/googleapi"

	"github.com/loganszeto/kvstore-go/internal/persistence"
	"github.com/loganszeto/kvstore-go/internal/protocol"
	"github.com/loganszeto/kvstore-go/internal/server"
	"github.com/loganszeto/kvstore-go/internal/stats"
	"github.com/loganszeto/kvstore-go/internal/store"
)

const (
	defaultDataDir = "/tmp/kv-demo"
	defaultObject  = "wal.log"
)

func main() {
	port := getenv("PORT", "8080")
	dataDir := getenv("KV_DEMO_DATA_DIR", defaultDataDir)
	bucket := os.Getenv("KV_DEMO_BUCKET")
	object := getenv("KV_DEMO_OBJECT", defaultObject)

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("data dir: %v", err)
	}

	ctx := context.Background()
	var uploader walUploader = noopUploader{}
	if bucket == "" {
		log.Printf("KV_DEMO_BUCKET not set; running without GCS persistence")
	} else {
		gcs, err := newGCSWAL(ctx, bucket, object)
		if err != nil {
			log.Fatalf("gcs client: %v", err)
		}
		uploader = gcs
	}

	walPath := persistence.WALPath(dataDir)
	if err := uploader.Download(ctx, walPath); err != nil {
		log.Fatalf("download wal: %v", err)
	}

	wal, err := persistence.OpenWAL(dataDir, persistence.Options{Fsync: true})
	if err != nil {
		log.Fatalf("open wal: %v", err)
	}
	defer wal.Close()

	st := store.NewStore(store.Options{})
	if err := persistence.Replay(walPath, st); err != nil {
		log.Fatalf("replay wal: %v", err)
	}

	stats := stats.New()

	handler := &wsHandler{
		st:       st,
		wal:      wal,
		stats:    stats,
		uploader: uploader,
		walPath:  walPath,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler.handleWS)
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("kv-demo ok\n"))
	})

	logged := withLogging(mux)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           logged,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("kv demo listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

type wsHandler struct {
	st       store.Store
	wal      *persistence.WAL
	stats    *stats.Stats
	uploader walUploader
	walPath  string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func (h *wsHandler) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
			continue
		}

		req, err := parseRequest(payload)
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("ERR "+err.Error()+"\n"))
			continue
		}

		resp := server.Dispatch(h.st, h.wal, h.stats, req)
		if h.isMutating(req) {
			if err := h.uploader.Upload(r.Context(), h.walPath); err != nil {
				resp = protocol.Response{Kind: "ERR", Err: "persistence upload failed: " + err.Error()}
			}
		}

		data, err := encodeResponse(resp)
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("ERR "+err.Error()+"\n"))
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

func (h *wsHandler) isMutating(req protocol.Request) bool {
	switch req.Type {
	case protocol.CmdSet, protocol.CmdSetEx, protocol.CmdDel, protocol.CmdExpire:
		return true
	default:
		return false
	}
}

func parseRequest(payload []byte) (protocol.Request, error) {
	reader := bufio.NewReader(bytes.NewReader(payload))
	return protocol.ReadRequest(reader)
}

func encodeResponse(resp protocol.Response) ([]byte, error) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	if err := protocol.WriteResponse(writer, resp); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type gcsWAL struct {
	client      *storage.Client
	bucket      string
	object      string
	mu          sync.Mutex
	lastUpload  time.Time
	minInterval time.Duration
}

type walUploader interface {
	Download(ctx context.Context, path string) error
	Upload(ctx context.Context, path string) error
}

type noopUploader struct{}

func (noopUploader) Download(_ context.Context, _ string) error { return nil }
func (noopUploader) Upload(_ context.Context, _ string) error   { return nil }

func newGCSWAL(ctx context.Context, bucket, object string) (*gcsWAL, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &gcsWAL{
		client:      client,
		bucket:      bucket,
		object:      object,
		minInterval: 1 * time.Second,
	}, nil
}

func (g *gcsWAL) Download(ctx context.Context, path string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	obj := g.client.Bucket(g.bucket).Object(g.object)
	rc, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return err
	}
	defer rc.Close()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, rc); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (g *gcsWAL) Upload(ctx context.Context, path string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.lastUpload.IsZero() && time.Since(g.lastUpload) < g.minInterval {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	obj := g.client.Bucket(g.bucket).Object(g.object)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, f); err != nil {
		_ = w.Close()
		if isRateLimitErr(err) {
			log.Printf("gcs wal upload rate limited: %v", err)
			return nil
		}
		return err
	}
	if err := w.Close(); err != nil {
		if isRateLimitErr(err) {
			log.Printf("gcs wal upload rate limited: %v", err)
			return nil
		}
		return err
	}
	g.lastUpload = time.Now()
	return nil
}

func getenv(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}

func isRateLimitErr(err error) bool {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		return gerr.Code == 429
	}
	return strings.Contains(err.Error(), "rateLimitExceeded") || strings.Contains(err.Error(), "Error 429")
}
