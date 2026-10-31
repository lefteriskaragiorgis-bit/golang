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

const pokeAPIURL = "https://pokeapi.co/api/v2/"
const pokeAPIList = pokeAPIURL + "pokemon/"

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
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Type    []string `json:"type"`
	Height  int      `json:"height"`
	Weight  int      `json:"weight"`
	HP      int      `json:"hp"`
	Attack  int      `json:"attack"`
	Defense int      `json:"defense"`
	NextID  int      `json:"-"`
	PrevID  int      `json:"-"`

	Sprites struct {
		Other struct {
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
			} `json:"official-artwork"`
		} `json:"other"`
	} `json:"sprites"`
}

func FetchPokemons(ctx context.Context, limit, offset int) ([]PokemonResponse, error) {
	endpoint, err := url.Parse(pokeAPIList)
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

	client := &http.Client{}

	resp, err := client.Do(req)
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

		pokemons = append(pokemons, PokemonResponse{
			ID:   id,
			Name: entry.Name,
		})
	}

	return pokemons, nil
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

func FetchPokemonByID(ctx context.Context, id int) (PokemonResponse, error) {
	endpoint, err := url.Parse(
		fmt.Sprintf("%s%d/", pokeAPIList, id),
	)
	if err != nil {
		return PokemonResponse{}, err
	}

	fmt.Printf("Fetching Pokémon with ID %d from endpoint: %s\n", id, endpoint.String())

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, endpoint.String(), nil,
	)
	if err != nil {
		return PokemonResponse{}, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return PokemonResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PokemonResponse{}, fmt.Errorf("pokeAPI returned status %d", resp.StatusCode)
	}

	var apiResponse PokemonResponse

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return PokemonResponse{}, err
	}

	apiResponse.NextID = id + 1
	if id > 1 {
		apiResponse.PrevID = id - 1
	} else {
		apiResponse.PrevID = 0 // No previous Pokémon for ID 1
	}

	return apiResponse, nil
}
