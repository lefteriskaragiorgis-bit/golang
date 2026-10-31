package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/lefteriskaragiorgis/pokedex/internal/handler"
)

func main() {
	templates := template.Must(
		template.ParseGlob("templates/*.html"),
	)

	mux := http.NewServeMux()

	handler := &handler.PokemonHandler{
		Templates: templates,
	}

	mux.HandleFunc("GET /pokemon", handler.List)
	mux.HandleFunc("GET /pokemon/{id}", handler.GetByID)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
