package pokemon

type Pokemons struct {
	Count    int               `json:"count"`
	Next     string            `json:"next"`
	Previous string            `json:"previous"`
	Results  []PokemonListItem `json:"results"`
}

type PokemonEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokemonListItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type Pokemon struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Height      int    `json:"height"`
	Weight      int    `json:"weight"`
	Description string `json:"description,omitempty"`
	NextID      int    `json:"-"`
	PrevID      int    `json:"-"`

	Types   []PokemonType  `json:"types"`
	Stats   []PokemonStat  `json:"stats"`
	Sprites PokemonSprites `json:"sprites"`
}

type PokemonType struct {
	Slot int           `json:"slot"`
	Type NamedResource `json:"type"`
}

type PokemonStat struct {
	BaseStat int           `json:"base_stat"`
	Effort   int           `json:"effort"`
	Stat     NamedResource `json:"stat"`
}

type NamedResource struct {
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

type PokemonUpdateRequest struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}
