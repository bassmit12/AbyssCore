CREATE SCHEMA IF NOT EXISTS combat;
CREATE TABLE IF NOT EXISTS combat.monsters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    floor_id UUID NOT NULL,
    name TEXT NOT NULL,
    hp INT NOT NULL,
    max_hp INT NOT NULL,
    damage INT NOT NULL,
    room_x INT NOT NULL,
    room_y INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'alive',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
