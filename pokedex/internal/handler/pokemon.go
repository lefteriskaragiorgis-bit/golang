package handler

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/lefteriskaragiorgis/pokedex/internal/pokemon"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type PokemonHandler struct {
	Templates *template.Template
}

func (h *PokemonHandler) List(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.String())

	// Parse query parameters
	limit, offset := h.parseListQueryParams(w, r)

	// Fetch Pokémon from PokéAPI
	pokemons, err := pokemon.FetchPokemons(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Could not retrieve Pokémon", http.StatusInternalServerError)
		return
	}

	// Render HTML template
	err = h.Templates.ExecuteTemplate(w, "index.html", pokemons)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}

func (h *PokemonHandler) parseListQueryParams(w http.ResponseWriter, r *http.Request) (int, int) {
	limit, offset := defaultLimit, 0

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > maxLimit {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return limit, offset
		}
		limit = parsed
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return limit, offset
		}
		offset = parsed
	}

	return limit, offset
}

func (h *PokemonHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.String())

	id := r.PathValue("id")

	pokemonID, err := strconv.Atoi(id)
	if err != nil || pokemonID < 1 {
		http.Error(w, "Invalid Pokémon ID", http.StatusBadRequest)
		return
	}

	// Fetch Pokémon from PokéAPI
	pokemons, err := pokemon.FetchPokemonByID(r.Context(), pokemonID)
	if err != nil {
		http.Error(w, "Could not retrieve Pokémon", http.StatusInternalServerError)
		return
	}

	// Render HTML template
	err = h.Templates.ExecuteTemplate(w, "pokemon.html", pokemons)
	if err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}
