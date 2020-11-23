// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	_ "embed"

	"context"
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"net/http"
	"os/signal"

	"github.com/dimfeld/httptreemux/v5"
)

var (
	//go:embed README.md
	defaultIndex []byte

	bind          string
	index         string
	retention     time.Duration
	uploadLimitMB int
)

func init() {
	flag.StringVar(&bind, "bind", "[::1]:8080", "addr to bind")
	flag.StringVar(&index, "index", "./index.html",
		"homepage to show (default: ./index.html)")
	flag.DurationVar(&retention, "retention", 5*time.Minute,
		"retention policy for the current paste (default: 5min)")
	flag.IntVar(&uploadLimitMB, "size-limit", 16,
		"maximum size for uploads in MB (default: 16)")
	flag.Parse()
}

func main() {
	log.Println("starting...")

	r := httptreemux.New()
	r.GET("/health", func(w http.ResponseWriter, r *http.Request, p map[string]string) {})

	r.GET("/", func(w http.ResponseWriter, r *http.Request, p map[string]string) {
		if _, err := os.Stat(index); os.IsNotExist(err) {
			w.Write(defaultIndex)
			return
		}
		http.ServeFile(w, r, index)
	})

	state := NewState(retention, int64(uploadLimitMB*1024*1024))
	state.AddPUTStyleUploading(r)
	state.AddNullPointerStyleUploading(r)
	state.AddPasteViewing(r)
	state.AddPasteDeletion(r)

	srv := &http.Server{
		Addr:              bind,
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server terminated abnormally: %s\n", err)
		}
	}()

	log.Println("ready")

	quit := make(chan os.Signal)

	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down…")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %s\n", err)
	}

	log.Println("goodbye")
}
