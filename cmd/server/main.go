package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
    "time"

    "lighthouse/internal/version"
)

func envOrDefault(key, def string) string {
    if v := strings.TrimSpace(os.Getenv(key)); v != "" {
        return v
    }
    return def
}

func main() {
    // Config
    var (
        portFlag    = flag.Int("port", 2001, "Port to listen on (overrides PORT env)")
        addrFlag    = flag.String("addr", "", "Bind address (default 0.0.0.0)")
        readTOFlag  = flag.Duration("read-timeout", 5*time.Second, "HTTP read timeout")
        writeTOFlag = flag.Duration("write-timeout", 10*time.Second, "HTTP write timeout")
        idleTOFlag  = flag.Duration("idle-timeout", 60*time.Second, "HTTP idle timeout")
    )
    flag.Parse()

    // Environment overrides
    if p := envOrDefault("PORT", ""); p != "" {
        if n, err := strconv.Atoi(p); err == nil {
            *portFlag = n
        }
    }

    bindAddr := *addrFlag
    if bindAddr == "" {
        bindAddr = "0.0.0.0"
    }

    mux := http.NewServeMux()

    startedAt := time.Now()
    var reqCount uint64

    // Handlers
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        _, _ = fmt.Fprintln(w, version.Get())
    })

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        reqCount++
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        _, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>LighthouseTest</title>
  <style>
    body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Ubuntu, Cantarell, Noto Sans, Helvetica, Arial, "Apple Color Emoji", "Segoe UI Emoji"; margin: 2rem; }
    code, pre { background: #f5f5f5; padding: 2px 4px; border-radius: 4px; }
    .badge { display: inline-block; padding: 0.2rem 0.5rem; border-radius: 9999px; background: #eef; color: #335; margin-left: .5rem; font-size: .85rem; }
    .grid { display: grid; gap: 1rem; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); }
    .card { border: 1px solid #e5e5e5; border-radius: 8px; padding: 1rem; }
  </style>
  <meta http-equiv="refresh" content="30"/>
  </head>
<body>
  <h1>LighthouseTest <span class="badge">v%s</span></h1>
  <p>Small Go web service listening on <code>%s:%d</code>.</p>
  <div class="grid">
    <div class="card">
      <h3>Uptime</h3>
      <p>%s</p>
    </div>
    <div class="card">
      <h3>Requests Served</h3>
      <p>%d</p>
    </div>
    <div class="card">
      <h3>Endpoints</h3>
      <ul>
        <li><code>/</code> — this page</li>
        <li><code>/healthz</code> — health check</li>
        <li><code>/version</code> — plain text version</li>
      </ul>
    </div>
  </div>
</body>
</html>`, version.Get(), bindAddr, *portFlag, time.Since(startedAt).Round(time.Second), reqCount)
    })

    srv := &http.Server{
        Addr:              fmt.Sprintf("%s:%d", bindAddr, *portFlag),
        Handler:           logRequests(mux),
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       *readTOFlag,
        WriteTimeout:      *writeTOFlag,
        IdleTimeout:       *idleTOFlag,
    }

    // Run server
    go func() {
        log.Printf("listening on http://%s (version=%s)", srv.Addr, version.Get())
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    log.Println("shutting down...")
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("graceful shutdown failed: %v", err)
    }
}

func logRequests(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        ww := &respWriter{ResponseWriter: w, status: 200}
        next.ServeHTTP(ww, r)
        log.Printf("%s %s %d %s", r.Method, r.URL.Path, ww.status, time.Since(start))
    })
}

type respWriter struct {
    http.ResponseWriter
    status int
}

func (w *respWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

