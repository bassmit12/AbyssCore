CREATE SCHEMA IF NOT EXISTS inventory;
CREATE TABLE IF NOT EXISTS inventory.items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id UUID NOT NULL,
    name TEXT NOT NULL,
    item_type TEXT NOT NULL,
    rarity TEXT NOT NULL DEFAULT 'common',
    effect_type TEXT NOT NULL,
    effect_value INT NOT NULL DEFAULT 0,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
