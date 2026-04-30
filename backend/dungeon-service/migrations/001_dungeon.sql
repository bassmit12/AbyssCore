CREATE TABLE IF NOT EXISTS dungeon.dungeons (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id    UUID NOT NULL,
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dungeon.floors (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dungeon_id UUID NOT NULL REFERENCES dungeon.dungeons(id),
    level      INT  NOT NULL,
    width      INT  NOT NULL DEFAULT 8,
    height     INT  NOT NULL DEFAULT 8,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(dungeon_id, level)
);

CREATE TABLE IF NOT EXISTS dungeon.rooms (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    floor_id   UUID NOT NULL REFERENCES dungeon.floors(id),
    x          INT  NOT NULL,
    y          INT  NOT NULL,
    has_chest  BOOLEAN NOT NULL DEFAULT FALSE,
    exits      TEXT[] NOT NULL DEFAULT '{}',
    UNIQUE(floor_id, x, y)
);

CREATE INDEX IF NOT EXISTS idx_rooms_floor_id ON dungeon.rooms(floor_id);
