// Command server runs the portfolio site: one binary, assets embedded.
package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"portfolio"
	"portfolio/internal/handler"
	"portfolio/internal/store"
	"portfolio/internal/terminal"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dbPath := flag.String("db", "inquiries.db", "path to the SQLite inquiry database")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Error("open store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	srv, err := handler.New(portfolio.Templates, portfolio.Static, st, terminal.New(), log)
	if err != nil {
		log.Error("init server", "err", err)
		os.Exit(1)
	}

	httpSrv := &http.Server{
		Addr:              *addr,
		Handler:           srv,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Info("listening", "addr", *addr)
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Error("server", "err", err)
		os.Exit(1)
	}
}
