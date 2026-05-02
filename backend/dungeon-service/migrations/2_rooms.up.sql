CREATE TABLE IF NOT EXISTS dungeon.rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    floor_id UUID NOT NULL REFERENCES dungeon.floors(id) ON DELETE CASCADE,
    x INT NOT NULL,
    y INT NOT NULL,
    has_chest BOOLEAN NOT NULL DEFAULT FALSE,
    exits TEXT[] NOT NULL DEFAULT '{}'
);
