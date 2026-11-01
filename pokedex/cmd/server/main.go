package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/lefteriskaragiorgis/pokedex/internal/handler"
	"github.com/lefteriskaragiorgis/pokedex/internal/pokemon"
)

func main() {
	templates := template.Must(
		template.ParseGlob("templates/*.html"),
	)

	cache := pokemon.NewCache()

	pokemonHandler := &handler.PokemonHandler{
		Templates: templates,
		Cache:     cache,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /pokemon", pokemonHandler.List)
	mux.HandleFunc("GET /pokemon/{id}", pokemonHandler.GetByID)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
