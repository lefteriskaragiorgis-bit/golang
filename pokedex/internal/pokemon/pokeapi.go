package pokemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const pokeAPIURL = "https://pokeapi.co/api/v2"

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type PokemonsResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []PokeAPIEntry `json:"results"`
}

type PokeAPIEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokemonResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	NextID int    `json:"-"`
	PrevID int    `json:"-"`

	Types   []PokemonType  `json:"types"`
	Stats   []PokemonStat  `json:"stats"`
	Sprites PokemonSprites `json:"sprites"`
}

type PokemonType struct {
	Slot int              `json:"slot"`
	Type NamedAPIResource `json:"type"`
}

type PokemonStat struct {
	BaseStat int              `json:"base_stat"`
	Effort   int              `json:"effort"`
	Stat     NamedAPIResource `json:"stat"`
}

type NamedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokemonSprites struct {
	Other struct {
		OfficialArtwork struct {
			FrontDefault string `json:"front_default"`
		} `json:"official-artwork"`
	} `json:"other"`
}

func NewAPIClient() *APIClient {
	return &APIClient{
		BaseURL:    pokeAPIURL,
		HTTPClient: http.DefaultClient,
	}
}

func (c *APIClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}

	return http.DefaultClient
}

func (c *APIClient) pokemonURL(path string) string {
	return strings.TrimRight(c.BaseURL, "/") + "/pokemon/" + path
}

func (c *APIClient) FetchPokemons(
	ctx context.Context,
	limit, offset int,
	cache *Cache,
) ([]PokemonResponse, error) {
	endpoint, err := url.Parse(c.pokemonURL(""))
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, endpoint.String(), nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pokeAPI returned status %d", resp.StatusCode)
	}

	var apiResponse PokemonsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	pokemons := make([]PokemonResponse, 0, len(apiResponse.Results))

	for _, entry := range apiResponse.Results {
		id, err := pokemonIDFromURL(entry.URL)
		if err != nil {
			return nil, err
		}

		p, err := c.FetchPokemonByID(ctx, id, cache)
		if err != nil {
			return nil, err
		}

		pokemons = append(pokemons, p)
	}

	return pokemons, nil
}

func (c *APIClient) FetchPokemonByID(
	ctx context.Context,
	id int,
	cache *Cache,
) (PokemonResponse, error) {
	if cache != nil {
		if cachedPokemon, exists := cache.Get(id); exists {
			fmt.Printf("Cache HIT: Pokémon %d\n", id)
			return cachedPokemon, nil
		}

		fmt.Printf("Cache MISS: Pokémon %d\n", id)
	}

	endpoint := c.pokemonURL(strconv.Itoa(id) + "/")

	p, err := c.fetchPokemon(ctx, endpoint)
	if err != nil {
		return PokemonResponse{}, err
	}

	p.NextID = id + 1
	if id > 1 {
		p.PrevID = id - 1
	}
	if cache != nil {
		cache.Set(id, p)
	}

	return p, nil
}

func (c *APIClient) FetchPokemonByName(
	ctx context.Context,
	name string,
	cache *Cache,
) (PokemonResponse, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	if name == "" {
		return PokemonResponse{}, fmt.Errorf("pokemon name cannot be empty")
	}

	endpoint := c.pokemonURL(url.PathEscape(name) + "/")

	p, err := c.fetchPokemon(ctx, endpoint)
	if err != nil {
		return PokemonResponse{}, err
	}

	p.NextID = p.ID + 1
	if p.ID > 1 {
		p.PrevID = p.ID - 1
	}

	cache.Set(p.ID, p)

	return p, nil
}

func (c *APIClient) fetchPokemon(
	ctx context.Context,
	endpoint string,
) (PokemonResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, endpoint, nil,
	)
	if err != nil {
		return PokemonResponse{}, err
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return PokemonResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PokemonResponse{}, fmt.Errorf(
			"pokeAPI returned status %d", resp.StatusCode,
		)
	}

	var p PokemonResponse
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return PokemonResponse{}, err
	}

	return p, nil
}

func pokemonIDFromURL(rawURL string) (int, error) {
	trimmed := strings.TrimRight(rawURL, "/")
	idString := trimmed[strings.LastIndex(trimmed, "/")+1:]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return 0, fmt.Errorf("invalid pokemon URL %q: %w", rawURL, err)
	}

	return id, nil
}
