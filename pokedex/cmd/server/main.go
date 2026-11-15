package main

import (
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/lefteriskaragiorgis-bit/golang/internal/config"
	"github.com/lefteriskaragiorgis-bit/golang/internal/handler"
	"github.com/lefteriskaragiorgis-bit/golang/internal/pokemon"
	"github.com/lefteriskaragiorgis-bit/golang/internal/repository"
	"github.com/lefteriskaragiorgis-bit/golang/internal/service"
)

func main() {
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	templates := template.Must(template.ParseGlob("templates/*.html"))
	cache := pokemon.NewCache()
	repository := &repository.PokemonRepository{DB: db}
	service := &service.PokemonService{Repository: repository, Cache: cache}

	pokemonHandler := &handler.PokemonHandler{
		Templates: templates,
		Service:   service,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", pokemonHandler.Index)
	mux.HandleFunc("GET /pokemon", pokemonHandler.List)
	mux.HandleFunc("GET /pokemon/{id}", pokemonHandler.GetByID)
	mux.HandleFunc("GET /api/pokemon", pokemonHandler.ListJSON)
	mux.HandleFunc("GET /api/pokemon/search", pokemonHandler.Search)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
