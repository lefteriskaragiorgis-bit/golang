# Pokédex

A Pokédex web application built with Go, using [PokéAPI](https://pokeapi.co/) as its Pokémon data source.

## Requirements

- Go
- Docker and Docker Compose

## Run the application

Start PostgreSQL:

```bash
docker compose up -d postgres
```

Start the web server:

```bash
go run ./cmd/server
```

The application will be available at:

http://localhost:8080

## Database migrations

Migrations are managed by a small Go migration runner. SQL files are stored in the `migrations/` directory, with separate `.up.sql` and `.down.sql` files.

Apply all pending migrations:

```bash
go run ./cmd/migrate up
```

Roll back the latest migration:

```bash
go run ./cmd/migrate down
```

Check applied migrations:

```bash
docker compose exec postgres \
  psql -U pokedex -d pokedex \
  -c "SELECT * FROM schema_migrations ORDER BY version;"
```

Create a new migration manually by adding two numbered SQL files:

```text
migrations/
├── 000001_create_pokemon_tables.up.sql
├── 000001_create_pokemon_tables.down.sql
├── 000002_create_types_tables.up.sql
└── 000002_create_types_tables.down.sql
```

## Run the importer

Import Pokémon data into the JSON file:

```bash
go run ./cmd/import-pokemon
```

## Run tests

Run all tests:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

## Useful Docker commands

Stop containers:

```bash
docker compose down
```

Stop containers and remove PostgreSQL data:

```bash
docker compose down -v
```

**Warning:** The last command permanently deletes your local PostgreSQL data.


TODO: make sure to add Auth for the admin endpoints to update pokemons