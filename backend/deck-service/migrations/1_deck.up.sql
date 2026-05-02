CREATE SCHEMA IF NOT EXISTS deck;

-- Master list of all cards in the game
CREATE TABLE IF NOT EXISTS deck.card_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    class       TEXT NOT NULL DEFAULT 'any',   -- warrior | rogue | mage | any
    type        TEXT NOT NULL,                  -- attack | skill | power
    cost        INT NOT NULL DEFAULT 1,         -- energy cost (0-3)
    effect      JSONB NOT NULL DEFAULT '{}',    -- { "damage": 6, "block": 0, "draw": 0, ... }
    rarity      TEXT NOT NULL DEFAULT 'common', -- common | uncommon | rare
    description TEXT NOT NULL DEFAULT ''
);

-- A hero's current deck (cards they've collected this run)
CREATE TABLE IF NOT EXISTS deck.hero_decks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id     UUID NOT NULL,
    card_def_id UUID NOT NULL REFERENCES deck.card_definitions(id),
    upgraded    BOOLEAN NOT NULL DEFAULT FALSE,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Master list of relics
CREATE TABLE IF NOT EXISTS deck.relic_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    rarity      TEXT NOT NULL DEFAULT 'common',
    effect      JSONB NOT NULL DEFAULT '{}',
    description TEXT NOT NULL DEFAULT ''
);

-- Relics a hero currently holds
CREATE TABLE IF NOT EXISTS deck.hero_relics (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id     UUID NOT NULL,
    relic_def_id UUID NOT NULL REFERENCES deck.relic_definitions(id),
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (hero_id, relic_def_id)
);
