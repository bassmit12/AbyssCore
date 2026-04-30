CREATE TABLE IF NOT EXISTS inventory.item_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    type        TEXT NOT NULL CHECK (type IN ('weapon', 'armor', 'potion')),
    value       INT  NOT NULL,
    rarity      TEXT NOT NULL CHECK (rarity IN ('common', 'uncommon', 'rare')),
    drop_weight INT  NOT NULL DEFAULT 10
);

-- Seed item definitions
INSERT INTO inventory.item_definitions (name, type, value, rarity, drop_weight) VALUES
    ('Rusty Sword',      'weapon', 5,  'common',   30),
    ('Iron Sword',       'weapon', 10, 'uncommon', 15),
    ('Shadow Blade',     'weapon', 18, 'rare',      5),
    ('Leather Armor',    'armor',  4,  'common',   30),
    ('Chain Mail',       'armor',  8,  'uncommon', 15),
    ('Dragon Scale',     'armor',  15, 'rare',      5),
    ('Health Potion',    'potion', 20, 'common',   40),
    ('Mega Potion',      'potion', 50, 'uncommon', 15),
    ('Elixir',           'potion', 99, 'rare',      5)
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS inventory.hero_inventory (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id     UUID NOT NULL,
    item_def_id UUID NOT NULL REFERENCES inventory.item_definitions(id),
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_hero_inventory_hero_id ON inventory.hero_inventory(hero_id);
