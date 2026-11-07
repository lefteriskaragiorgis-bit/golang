package main

import (
	"context"
	"log"

	"github.com/lefteriskaragiorgis-bit/golang/internal/importer"
	"github.com/lefteriskaragiorgis-bit/golang/internal/pokemon"
)

func main() {
	api := pokemon.NewAPIClient()
	pokemonImporter := importer.NewPokemonImporter(api)

	if err := pokemonImporter.ImportToJSON(context.Background(), "data/pokemon.json"); err != nil {
		log.Fatal(err)
	}

	log.Println("Pokémon import completed")
}
