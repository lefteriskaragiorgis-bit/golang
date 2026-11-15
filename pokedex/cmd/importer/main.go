package main

import (
	"context"
	"log"

	"github.com/lefteriskaragiorgis-bit/golang/internal/importer"
)

func main() {
	api := importer.NewAPIClient()
	pokemonImporter := importer.NewPokemonImporter(api)

	if err := pokemonImporter.ImportToJSON(context.Background(), "data/pokemon.json"); err != nil {
		log.Fatal(err)
	}

	log.Println("Pokémon import completed")
}
