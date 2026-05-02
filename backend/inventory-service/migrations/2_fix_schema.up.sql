-- Create item_definitions if migration 1 didn't (schema may vary)
CREATE TABLE IF NOT EXISTS inventory.item_definitions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    type         TEXT NOT NULL,
    value        INT  NOT NULL DEFAULT 0,
    rarity       TEXT NOT NULL DEFAULT 'common',
    drop_weight  INT  NOT NULL DEFAULT 10
);

-- Create hero_inventory if missing
CREATE TABLE IF NOT EXISTS inventory.hero_inventory (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id     UUID NOT NULL,
    item_def_id UUID NOT NULL REFERENCES inventory.item_definitions(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add description column (idempotent)
ALTER TABLE inventory.item_definitions ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

-- Seed item definitions if empty
INSERT INTO inventory.item_definitions (name, type, value, rarity, description, drop_weight)
SELECT * FROM (VALUES
  ('Health Potion',    'potion', 20, 'common',   'Restore 20 HP.',                           30),
  ('Greater Potion',   'potion', 40, 'uncommon', 'Restore 40 HP.',                           15),
  ('Max Potion',       'potion', 60, 'rare',     'Restore 60 HP.',                            5),
  ('Gold Coins',       'gold',   25, 'common',   'A pouch of 25 gold.',                      25),
  ('Gold Pouch',       'gold',   50, 'uncommon', 'A pouch of 50 gold.',                      10),
  ('Strange Artifact', 'relic',   0, 'rare',     'An artifact of unknown power. +5 max HP.',  8),
  ('Enchanted Gem',    'relic',   0, 'epic',     'Glows faintly. +1 energy per turn.',        3),
  ('Sharp Dagger',     'weapon', 10, 'common',   'A crude blade. +10 attack.',               12),
  ('Iron Shield',      'armor',   8, 'common',   'Sturdy iron. +8 block.',                   10),
  ('Magic Tome',       'relic',   0, 'uncommon', 'Arcane knowledge. +1 card draw per turn.',  7)
) AS v(name, type, value, rarity, description, drop_weight)
WHERE NOT EXISTS (SELECT 1 FROM inventory.item_definitions LIMIT 1);
