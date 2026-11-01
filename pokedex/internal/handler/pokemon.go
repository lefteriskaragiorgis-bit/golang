package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/lefteriskaragiorgis/pokedex/internal/pokemon"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type PokemonHandler struct {
	Templates *template.Template
	Cache     *pokemon.Cache
}

func (h *PokemonHandler) Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/pokemon", http.StatusSeeOther)
}

func (h *PokemonHandler) List(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.String())

	limit, offset := h.parseListQueryParams(w, r)

	pokemons, err := pokemon.FetchPokemons(r.Context(), limit, offset, h.Cache)
	if err != nil {
		http.Error(w, "Could not retrieve Pokémon", http.StatusInternalServerError)
		return
	}

	if err := h.Templates.ExecuteTemplate(w, "index.html", pokemons); err != nil {
		log.Printf("Could not render page: %v", err)
	}
}

func (h *PokemonHandler) ListJSON(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.String())

	limit, offset := h.parseListQueryParams(w, r)

	pokemons, err := pokemon.FetchPokemons(r.Context(), limit, offset, h.Cache)
	if err != nil {
		http.Error(w, "Could not retrieve Pokémon", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(pokemons); err != nil {
		log.Printf("Could not encode Pokémon: %v", err)
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
	pokemons, err := pokemon.FetchPokemonByID(r.Context(), pokemonID, h.Cache)
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

func (h *PokemonHandler) Search(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	if strings.TrimSpace(name) == "" {
		http.Error(w, "Pokemon name is required", http.StatusBadRequest)
		return
	}

	pokemon, err := pokemon.FetchPokemonByName(r.Context(), name, h.Cache)
	if err != nil {
		http.Error(w, "Pokemon not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(pokemon); err != nil {
		log.Printf("Could not encode Pokémon: %v", err)
	}
}
