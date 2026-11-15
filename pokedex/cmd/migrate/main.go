package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	_ "github.com/lib/pq"

	"github.com/lefteriskaragiorgis-bit/golang/internal/config"
	"github.com/lefteriskaragiorgis-bit/golang/internal/migrate"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run ./cmd/migrate [up|down]")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	dbURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:   net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		Path:   cfg.DBName,
	}

	query := dbURL.Query()
	query.Set("sslmode", cfg.DBSSLMode)
	dbURL.RawQuery = query.Encode()

	db, err := sql.Open("postgres", dbURL.String())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal(err)
	}

	runner := migrate.NewMigrationRunner(db, "migrations")

	switch os.Args[1] {
	case "up":
		err = runner.Up(context.Background())
	case "down":
		err = runner.Down(context.Background())
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		log.Fatal(err)
	}
}
