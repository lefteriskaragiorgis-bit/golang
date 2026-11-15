
-- Main Pokémon table
CREATE TABLE pokemon (
    id INTEGER PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    height INTEGER NOT NULL CHECK (height >= 0),
    weight INTEGER NOT NULL CHECK (weight >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pokémon types
CREATE TABLE types (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

-- Pokémon <-> types relationship
CREATE TABLE pokemon_types (
    pokemon_id INTEGER NOT NULL REFERENCES pokemon(id) ON DELETE CASCADE,
    type_id INTEGER NOT NULL REFERENCES types(id),
    slot SMALLINT NOT NULL CHECK (slot > 0),

    PRIMARY KEY (pokemon_id, type_id),
    UNIQUE (pokemon_id, slot)
);

-- Stat definitions
CREATE TABLE stats (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

-- Pokémon <-> stats relationship and values
CREATE TABLE pokemon_stats (
    pokemon_id INTEGER NOT NULL REFERENCES pokemon(id) ON DELETE CASCADE,
    stat_id INTEGER NOT NULL REFERENCES stats(id),
    base_stat INTEGER NOT NULL CHECK (base_stat >= 0),
    effort INTEGER NOT NULL DEFAULT 0 CHECK (effort >= 0),

    PRIMARY KEY (pokemon_id, stat_id)
);

-- Pokémon artwork and sprite URLs
CREATE TABLE pokemon_sprites (
    pokemon_id INTEGER PRIMARY KEY REFERENCES pokemon(id) ON DELETE CASCADE,
    official_artwork TEXT
);

-- Useful indexes for filtering and relationships
CREATE INDEX idx_pokemon_name ON pokemon(name);
CREATE INDEX idx_pokemon_types_type_id ON pokemon_types(type_id);
CREATE INDEX idx_pokemon_stats_stat_id ON pokemon_stats(stat_id);
