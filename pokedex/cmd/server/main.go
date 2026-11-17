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

	detailCache := pokemon.NewCache[pokemon.Pokemon]()
	listCache := pokemon.NewCache[[]pokemon.PokemonListItem]()
	defer detailCache.Close()
	defer listCache.Close()

	templates := template.Must(template.ParseGlob("templates/*.html"))
	repository := &repository.PokemonRepository{DB: db}
	service := &service.PokemonService{Repository: repository, DetailCache: detailCache, ListCache: listCache}

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

	mux.HandleFunc("PUT /admin/pokemon", pokemonHandler.Update)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
