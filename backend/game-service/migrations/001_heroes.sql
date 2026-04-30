CREATE TABLE IF NOT EXISTS game.heroes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id   TEXT NOT NULL,
    name        TEXT NOT NULL,
    class       TEXT NOT NULL CHECK (class IN ('warrior', 'rogue', 'mage')),
    hp          INT  NOT NULL,
    max_hp      INT  NOT NULL,
    level       INT  NOT NULL DEFAULT 1,
    xp          INT  NOT NULL DEFAULT 0,
    x           INT  NOT NULL DEFAULT 0,
    y           INT  NOT NULL DEFAULT 0,
    dungeon_id  UUID,
    alive       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_heroes_player_id ON game.heroes(player_id);
