package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/lefteriskaragiorgis-bit/golang/internal/pokemon"
	"github.com/lefteriskaragiorgis-bit/golang/internal/repository"
)

type PokemonService struct {
	Repository  *repository.PokemonRepository
	DetailCache *pokemon.Cache[pokemon.Pokemon]
	ListCache   *pokemon.Cache[[]pokemon.PokemonListItem]
}

func (s *PokemonService) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]pokemon.PokemonListItem, error) {
	if s.ListCache != nil {
		if cachedList, exists := s.ListCache.Get(fmt.Sprintf("list_%d_%d", limit, offset)); exists {
			fmt.Printf("Cache HIT: Pokémon list %d_%d\n", limit, offset)
			return cachedList, nil
		}
		fmt.Printf("Cache MISS: Pokémon list %d_%d\n", limit, offset)
	}

	pokemons, err := s.Repository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	if s.ListCache != nil {
		s.ListCache.Set(fmt.Sprintf("list_%d_%d", limit, offset), pokemons)
	}
	return pokemons, nil
}

func (s *PokemonService) GetByID(
	ctx context.Context,
	id int,
) (pokemon.Pokemon, error) {
	if s.DetailCache != nil {
		if cachedPokemon, exists := s.DetailCache.Get(fmt.Sprintf("%d", id)); exists {
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
	if s.DetailCache != nil {
		s.DetailCache.Set(fmt.Sprintf("%d", id), *p)
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

	if s.DetailCache != nil {
		s.DetailCache.Set(fmt.Sprintf("%d", p.ID), *p)
	}

	return *p, nil
}

func (s *PokemonService) Update(
	ctx context.Context,
	pokemon pokemon.PokemonUpdateRequest,
) (*pokemon.Pokemon, error) {
	existingPokemon, err := s.Repository.FindByID(ctx, fmt.Sprintf("%d", pokemon.ID))
	if err != nil {
		return nil, err
	}

	existingPokemon.Description = pokemon.Description

	err = s.Repository.UpdatePokemon(ctx, existingPokemon)
	if err != nil {
		return nil, err
	}

	// update pokemon in cache
	if s.DetailCache != nil {
		s.DetailCache.Set(fmt.Sprintf("%d", existingPokemon.ID), *existingPokemon)
	}

	fmt.Printf("Updated Pokémon %s in database and cache\n", existingPokemon.Name)
	return existingPokemon, nil
}
