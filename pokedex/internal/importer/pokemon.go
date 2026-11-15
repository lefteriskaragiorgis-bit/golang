package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	batchSize  = 20
	batchDelay = 1 * time.Second
)

const pokeAPIURL = "https://pokeapi.co/api/v2"

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type PokemonImporter struct {
	API *APIClient
}

type PokeAPIPokemonsResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []PokeAPIEntry `json:"results"`
}

type PokeAPIEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokeAPIPokemonResponse struct {
	ID      int                   `json:"id"`
	Name    string                `json:"name"`
	Height  int                   `json:"height"`
	Weight  int                   `json:"weight"`
	Types   []PokeAPIPokemonType  `json:"types"`
	Stats   []PokeAPIPokemonStat  `json:"stats"`
	Sprites PokeAPIPokemonSprites `json:"sprites"`
}

type PokeAPIPokemonType struct {
	Slot int                     `json:"slot"`
	Type PokeAPINamedAPIResource `json:"type"`
}

type PokeAPIPokemonStat struct {
	BaseStat int                     `json:"base_stat"`
	Effort   int                     `json:"effort"`
	Stat     PokeAPINamedAPIResource `json:"stat"`
}

type PokeAPINamedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokeAPIPokemonSprites struct {
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

func NewPokemonImporter(api *APIClient) *PokemonImporter {
	return &PokemonImporter{
		API: api,
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
) ([]PokeAPIPokemonResponse, error) {
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

	var apiResponse PokeAPIPokemonsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	pokemons := make([]PokeAPIPokemonResponse, 0, len(apiResponse.Results))

	for _, entry := range apiResponse.Results {
		id, err := pokemonIDFromURL(entry.URL)
		if err != nil {
			return nil, err
		}

		p, err := c.FetchPokemonByID(ctx, id)
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
) (PokeAPIPokemonResponse, error) {
	endpoint := c.pokemonURL(strconv.Itoa(id) + "/")

	p, err := c.fetchPokemon(ctx, endpoint)
	if err != nil {
		return PokeAPIPokemonResponse{}, err
	}

	return p, nil
}

func (c *APIClient) fetchPokemon(
	ctx context.Context,
	endpoint string,
) (PokeAPIPokemonResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, endpoint, nil,
	)
	if err != nil {
		return PokeAPIPokemonResponse{}, err
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return PokeAPIPokemonResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PokeAPIPokemonResponse{}, fmt.Errorf(
			"pokeAPI returned status %d", resp.StatusCode,
		)
	}

	var p PokeAPIPokemonResponse
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return PokeAPIPokemonResponse{}, err
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

func (i *PokemonImporter) Import(ctx context.Context) ([]PokeAPIPokemonResponse, error) {
	var allPokemon []PokeAPIPokemonResponse

	for offset := 0; ; offset += batchSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		results, err := i.API.FetchPokemons(ctx, batchSize, offset)
		if err != nil {
			return nil, fmt.Errorf("fetch Pokémon at offset %d: %w", offset, err)
		}

		allPokemon = append(allPokemon, results...)

		fmt.Printf("Imported %d Pokémon so far\n", len(allPokemon))

		if len(results) < batchSize {
			break
		}

		timer := time.NewTimer(batchDelay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		}
	}

	return allPokemon, nil
}

func (i *PokemonImporter) ImportToJSON(ctx context.Context, outputPath string) error {
	allPokemon, err := i.Import(ctx)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(allPokemon); err != nil {
		return fmt.Errorf("encode Pokémon JSON: %w", err)
	}

	fmt.Printf(
		"Successfully imported %d Pokémon to %s\n",
		len(allPokemon),
		outputPath,
	)

	return nil
}
