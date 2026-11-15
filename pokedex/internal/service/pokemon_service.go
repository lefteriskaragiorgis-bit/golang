package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/lefteriskaragiorgis-bit/golang/internal/pokemon"
	"github.com/lefteriskaragiorgis-bit/golang/internal/repository"
)

type PokemonService struct {
	Repository *repository.PokemonRepository
	Cache      *pokemon.Cache
}

func (s *PokemonService) GetPokemons(
	ctx context.Context,
	limit, offset int,
) ([]pokemon.Pokemons, error) {
	results, err := s.Repository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	pokemons := make([]pokemon.Pokemons, 0, len(results))

	return pokemons, nil
}

func (s *PokemonService) GetByID(
	ctx context.Context,
	id int,
) (pokemon.Pokemon, error) {
	if s.Cache != nil {
		if cachedPokemon, exists := s.Cache.Get(id); exists {
			fmt.Printf("Cache HIT: Pokémon %d\n", id)
			return cachedPokemon, nil
		}

		fmt.Printf("Cache MISS: Pokémon %d\n", id)
	}

	p, err := s.Repository.FindByID(ctx, fmt.Sprintf("%d", id))
	if err != nil {
		return pokemon.Pokemon{}, err
	}

	p.NextID = id + 1
	if id > 1 {
		p.PrevID = id - 1
	}
	if s.Cache != nil {
		s.Cache.Set(id, *p)
	}

	return *p, nil
}

func (s *PokemonService) GetPokemonByName(
	ctx context.Context,
	name string,
) (pokemon.Pokemon, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	if name == "" {
		return pokemon.Pokemon{}, fmt.Errorf("pokemon name cannot be empty")
	}

	p, err := s.Repository.SearchPokemonByName(ctx, name)
	if err != nil {
		return pokemon.Pokemon{}, err
	}

	p.NextID = p.ID + 1
	if p.ID > 1 {
		p.PrevID = p.ID - 1
	}

	if s.Cache != nil {
		s.Cache.Set(p.ID, *p)
	}

	return *p, nil
}
