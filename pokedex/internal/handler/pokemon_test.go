package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lefteriskaragiorgis-bit/golang/internal/pokemon"
)

func TestSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/pokemon/pikachu/" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")

			_, _ = w.Write([]byte(`{
				"id": 25,
				"name": "pikachu",
				"height": 4,
				"weight": 60
			}`))
		},
	))
	defer server.Close()

	cache := pokemon.NewCache()

	apiClient := &pokemon.APIClient{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	handler := &PokemonHandler{
		Cache: cache,
		API:   apiClient,
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/pokemon/search?name=pikachu",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d",
			http.StatusOK, rec.Code)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected application/json, got %s", contentType)
	}

	var result pokemon.PokemonResponse

	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result.Name != "pikachu" {
		t.Errorf("expected pikachu, got %s", result.Name)
	}

	if result.ID != 25 {
		t.Errorf("expected ID 25, got %d", result.ID)
	}
}
