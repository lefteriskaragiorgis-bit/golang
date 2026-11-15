package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/url"

	_ "github.com/lib/pq"

	"github.com/lefteriskaragiorgis-bit/golang/internal/config"
	"github.com/lefteriskaragiorgis-bit/golang/internal/seeder"
)

func main() {
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

	pokemonSeeder := seeder.NewPokemonSeeder(db)

	if err := pokemonSeeder.SeedPokemon(context.Background(), "data/pokemon.json"); err != nil {
		log.Fatal(err)
	}

	log.Println("Pokémon import completed")
}
