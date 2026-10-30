package handler

import (
	"html/template"
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
	limit := defaultLimit
	offset := 0

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > maxLimit {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}

		limit = parsed
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsed
	}

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
