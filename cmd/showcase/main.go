// Command showcase serves a page per component group of github.com/go-htemel/ui, with each live
// component next to the Go source that produced it.
//
//	go run ./cmd/showcase -addr :8080
package main

import (
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("showcase listening", "addr", *addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
