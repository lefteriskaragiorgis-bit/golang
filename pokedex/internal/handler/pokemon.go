package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/lefteriskaragiorgis-bit/golang/internal/pokemon"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type PokemonHandler struct {
	Templates *template.Template
	Cache     *pokemon.Cache
	API       *pokemon.APIClient
}

func (h *PokemonHandler) Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/pokemon", http.StatusSeeOther)
}

func (h *PokemonHandler) List(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.String())

	limit, offset, err := parseListQueryParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pokemons, err := h.API.FetchPokemons(r.Context(), limit, offset, h.Cache)
	if err != nil {
		http.Error(w, "Could not retrieve Pokémon", http.StatusInternalServerError)
		return
	}

	if err := h.Templates.ExecuteTemplate(w, "index.html", pokemons); err != nil {
		http.Error(w, "Could not render page", http.StatusInternalServerError)
	}
}

func (h *PokemonHandler) ListJSON(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.String())

	limit, offset, err := parseListQueryParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pokemons, err := h.API.FetchPokemons(r.Context(), limit, offset, h.Cache)
	if err != nil {
		http.Error(w, "Could not retrieve Pokémon", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(pokemons); err != nil {
		log.Printf("Could not encode Pokémon: %v", err)
	}
}

func parseListQueryParams(r *http.Request) (int, int, error) {
	limit, offset := defaultLimit, 0

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > maxLimit {
			return 0, 0, fmt.Errorf("invalid limit")
		}
		limit = parsed
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return 0, 0, fmt.Errorf("invalid offset")
		}
		offset = parsed
	}

	return limit, offset, nil
}

func (h *PokemonHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.String())

	id := r.PathValue("id")

	pokemonID, err := strconv.Atoi(id)
	if err != nil || pokemonID < 1 {
		http.Error(w, "Invalid Pokémon ID", http.StatusBadRequest)
		return
	}

	pokemons, err := h.API.FetchPokemonByID(r.Context(), pokemonID, h.Cache)
	if err != nil {
		http.Error(w, "Could not retrieve Pokémon", http.StatusInternalServerError)
		return
	}

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

	pokemon, err := h.API.FetchPokemonByName(r.Context(), name, h.Cache)
	if err != nil {
		http.Error(w, "Pokemon not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(pokemon); err != nil {
		log.Printf("Could not encode Pokémon: %v", err)
	}
}
