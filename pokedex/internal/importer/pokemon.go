package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lefteriskaragiorgis/pokedex/internal/pokemon"
)

const (
	batchSize  = 20
	batchDelay = 1 * time.Second
)

type PokemonImporter struct {
	API *pokemon.APIClient
}

func NewPokemonImporter(api *pokemon.APIClient) *PokemonImporter {
	return &PokemonImporter{
		API: api,
	}
}

func (i *PokemonImporter) Import(ctx context.Context) ([]pokemon.PokemonResponse, error) {
	var allPokemon []pokemon.PokemonResponse

	for offset := 0; ; offset += batchSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		results, err := i.API.FetchPokemons(ctx, batchSize, offset, nil)
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
