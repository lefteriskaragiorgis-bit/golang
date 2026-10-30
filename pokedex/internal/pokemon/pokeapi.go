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

const pokeAPIURL = "https://pokeapi.co/api/v2/pokemon/"

type PokeAPIResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []PokeAPIEntry `json:"results"`
}

type PokeAPIEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func FetchPokemons(ctx context.Context, limit, offset int) ([]Pokemon, error) {
	endpoint, err := url.Parse(pokeAPIURL)
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

	var apiResponse PokeAPIResponse

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	pokemons := make([]Pokemon, 0, len(apiResponse.Results))

	for _, entry := range apiResponse.Results {
		id, err := pokemonIDFromURL(entry.URL)
		if err != nil {
			return nil, err
		}

		pokemons = append(pokemons, Pokemon{
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
