package seeder

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lefteriskaragiorgis-bit/golang/internal/pokemon"
)

type PokemonSeeder struct {
	DB        *sql.DB
	InputPath string
}

func NewPokemonSeeder(db *sql.DB) *PokemonSeeder {
	return &PokemonSeeder{
		DB: db,
	}
}

func (p *PokemonSeeder) SeedPokemon(ctx context.Context, inputPath string) error {
	p.InputPath = inputPath
	pokemons, err := p.loadPokemon()
	if err != nil {
		return err
	}

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() // no-op after a successful Commit

	// description is intentionally left out: it is managed by admins via the CRUD endpoints.
	pokemonStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO pokemon (id, name, height, weight)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    height = EXCLUDED.height,
		    weight = EXCLUDED.weight,
		    updated_at = NOW()`)
	if err != nil {
		return fmt.Errorf("prepare pokemon stmt: %w", err)
	}
	defer pokemonStmt.Close()

	typeStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO pokemon_types (pokemon_id, type_id, slot)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`)
	if err != nil {
		return fmt.Errorf("prepare pokemon_types stmt: %w", err)
	}
	defer typeStmt.Close()

	statStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO pokemon_stats (pokemon_id, stat_id, base_stat, effort)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (pokemon_id, stat_id) DO UPDATE
		SET base_stat = EXCLUDED.base_stat,
		    effort = EXCLUDED.effort`)
	if err != nil {
		return fmt.Errorf("prepare pokemon_stats stmt: %w", err)
	}
	defer statStmt.Close()

	spriteStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO pokemon_sprites (pokemon_id, official_artwork)
		VALUES ($1, $2)
		ON CONFLICT (pokemon_id) DO UPDATE
		SET official_artwork = EXCLUDED.official_artwork`)
	if err != nil {
		return fmt.Errorf("prepare pokemon_sprites stmt: %w", err)
	}
	defer spriteStmt.Close()

	// Caches so each type/stat name is only resolved once.
	typeIDs := make(map[string]int)
	statIDs := make(map[string]int)

	for _, pk := range pokemons {
		if _, err := pokemonStmt.ExecContext(ctx, pk.ID, pk.Name, pk.Height, pk.Weight); err != nil {
			return fmt.Errorf("insert pokemon %q: %w", pk.Name, err)
		}

		for _, t := range pk.Types {
			typeID, err := getOrCreateID(ctx, tx, typeIDs, "types", t.Type.Name)
			if err != nil {
				return fmt.Errorf("resolve type %q for %q: %w", t.Type.Name, pk.Name, err)
			}
			if _, err := typeStmt.ExecContext(ctx, pk.ID, typeID, t.Slot); err != nil {
				return fmt.Errorf("insert pokemon_type %q for %q: %w", t.Type.Name, pk.Name, err)
			}
		}

		for _, s := range pk.Stats {
			statID, err := getOrCreateID(ctx, tx, statIDs, "stats", s.Stat.Name)
			if err != nil {
				return fmt.Errorf("resolve stat %q for %q: %w", s.Stat.Name, pk.Name, err)
			}
			if _, err := statStmt.ExecContext(ctx, pk.ID, statID, s.BaseStat, s.Effort); err != nil {
				return fmt.Errorf("insert pokemon_stat %q for %q: %w", s.Stat.Name, pk.Name, err)
			}
		}

		if _, err := spriteStmt.ExecContext(ctx,
			pk.ID,
			nullIfEmpty(pk.Sprites.Other.OfficialArtwork.FrontDefault),
		); err != nil {
			return fmt.Errorf("insert sprites for %q: %w", pk.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (p *PokemonSeeder) loadPokemon() ([]pokemon.PokemonResponse, error) {
	data, err := os.ReadFile(p.InputPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", p.InputPath, err)
	}

	var pokemons []pokemon.PokemonResponse
	if err := json.Unmarshal(data, &pokemons); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", p.InputPath, err)
	}

	return pokemons, nil
}

// getOrCreateID returns the id for a name in a lookup table ("types" or "stats"),
// inserting it when missing. The table argument is only ever a hardcoded constant.
func getOrCreateID(ctx context.Context, tx *sql.Tx, cache map[string]int, table, name string) (int, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}

	var id int
	query := fmt.Sprintf(`
		INSERT INTO %s (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, table)

	if err := tx.QueryRowContext(ctx, query, name).Scan(&id); err != nil {
		return 0, err
	}

	cache[name] = id
	return id, nil
}

// nullIfEmpty stores NULL instead of an empty string for nullable columns.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
