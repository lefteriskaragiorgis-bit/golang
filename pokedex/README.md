# Pokédex

A Pokédex web application built with Go, using [PokéAPI](https://pokeapi.co/) as its Pokémon data source.

## Requirements

- Go
- Docker and Docker Compose

## Run the application

Start the PostgreSQL container:

```bash
docker compose up -d
```

Start the web server:

```bash
go run ./cmd/server
```

The application will be available at:

```text
http://localhost:8080
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

## Useful commands

Run a specific package's tests:

```bash
go test ./internal/pokemon
```

Run tests with verbose output:

```bash
go test -v ./...
```

Stop the containers:

```bash
docker compose down
```

Stop containers and remove the PostgreSQL volume:

```bash
docker compose down -v
```

**Note:** The last command permanently deletes the local PostgreSQL data.