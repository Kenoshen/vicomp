package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mwingfield/vicomp/internal/db"
	"github.com/mwingfield/vicomp/internal/pdf"
	"github.com/mwingfield/vicomp/internal/store"
	"github.com/mwingfield/vicomp/internal/web"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbURL := env("DATABASE_URL", "postgres://vicomp:vicomp@localhost:5432/vicomp?sslmode=disable")
	gotenbergURL := env("GOTENBERG_URL", "http://localhost:3000")
	listenAddr := env("LISTEN_ADDR", ":8080")
	migrationsDir := env("MIGRATIONS_DIR", "migrations")
	templatesDir := env("TEMPLATES_DIR", "templates")
	staticDir := env("STATIC_DIR", "static")

	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool, migrationsDir); err != nil {
		return err
	}
	log.Println("migrations applied")

	srv, err := web.New(store.New(pool), pdf.NewClient(gotenbergURL), templatesDir, staticDir)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:              listenAddr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", listenAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
