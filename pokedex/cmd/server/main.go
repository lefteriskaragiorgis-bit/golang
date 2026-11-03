package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/lefteriskaragiorgis/pokedex/internal/handler"
	"github.com/lefteriskaragiorgis/pokedex/internal/pokemon"
)

func main() {
	templates := template.Must(template.ParseGlob("templates/*.html"))

	pokemonHandler := &handler.PokemonHandler{
		Templates: templates,
		Cache:     pokemon.NewCache(),
		API:       pokemon.NewAPIClient(),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", pokemonHandler.Index)
	mux.HandleFunc("GET /pokemon", pokemonHandler.List)
	mux.HandleFunc("GET /pokemon/{id}", pokemonHandler.GetByID)
	mux.HandleFunc("GET /api/pokemon", pokemonHandler.ListJSON)
	mux.HandleFunc("GET /api/pokemon/search", pokemonHandler.Search)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
