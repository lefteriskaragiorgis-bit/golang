package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lefteriskaragiorgis-bit/golang/internal/pokemon"
)

type PokemonRepository struct {
	DB *sql.DB
}

func (r *PokemonRepository) FindByID(
	ctx context.Context,
	id string,
) (*pokemon.Pokemon, error) {

	var p pokemon.Pokemon
	var typesJSON []byte
	var statsJSON []byte
	var spritesJSON []byte

	err := r.DB.QueryRowContext(ctx, `
        SELECT
            p.id,
            p.name,
            p.height,
            p.weight,
            p.description,

            COALESCE(
                (
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'slot', pt.slot,
                            'type', jsonb_build_object(
                          	'name', t.name,
                                'url', '/type/' || t.id
                            )
                        )
                        ORDER BY pt.slot
                    )
                    FROM pokemon_types pt
                    JOIN types t ON t.id = pt.type_id
                    WHERE pt.pokemon_id = p.id
                ),
                '[]'::jsonb
            ),

            COALESCE(
                (
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'base_stat', ps.base_stat,
                            'effort', ps.effort,
                            'stat', jsonb_build_object(
                                'name', s.name,
                                'url', '/stat/' || s.id
                            )
                        )
                        ORDER BY s.id
                    )
                    FROM pokemon_stats ps
                    JOIN stats s ON s.id = ps.stat_id
                    WHERE ps.pokemon_id = p.id
                ),
                '[]'::jsonb
            ),

            jsonb_build_object(
                'other', jsonb_build_object(
                    'official-artwork', jsonb_build_object(
                        'front_default',
                        COALESCE(spr.official_artwork, '')
                    )
                )
            )

        FROM pokemon p
        LEFT JOIN pokemon_sprites spr
            ON spr.pokemon_id = p.id
        WHERE p.id = $1
    `, id).Scan(
		&p.ID,
		&p.Name,
		&p.Height,
		&p.Weight,
		&p.Description,
		&typesJSON,
		&statsJSON,
		&spritesJSON,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(typesJSON, &p.Types); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(statsJSON, &p.Stats); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(spritesJSON, &p.Sprites); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PokemonRepository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]pokemon.PokemonListItem, error) {
	const query = `
        SELECT
            p.id,
            p.name,
            COALESCE(ps.official_artwork, '') AS image_url
        FROM pokemon p
        LEFT JOIN pokemon_sprites ps
            ON ps.pokemon_id = p.id
        ORDER BY p.id
        LIMIT $1 OFFSET $2
    `

	rows, err := r.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pokemons := make([]pokemon.PokemonListItem, 0, limit)

	for rows.Next() {
		var p pokemon.PokemonListItem

		if err := rows.Scan(&p.ID, &p.Name, &p.ImageURL); err != nil {
			return nil, err
		}

		pokemons = append(pokemons, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pokemons, nil
}

func (r *PokemonRepository) SearchPokemonByName(ctx context.Context, name string) (*pokemon.Pokemon, error) {
	var pokemon pokemon.Pokemon

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, name, height, weight
         FROM pokemon
         WHERE name = $1`,
		name,
	).Scan(
		&pokemon.ID,
		&pokemon.Name,
		&pokemon.Height,
		&pokemon.Weight,
	)

	if err != nil {
		return nil, err
	}

	return &pokemon, nil
}
