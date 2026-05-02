CREATE SCHEMA IF NOT EXISTS dungeon;
CREATE TABLE IF NOT EXISTS dungeon.dungeons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    current_floor INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS dungeon.floors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dungeon_id UUID NOT NULL REFERENCES dungeon.dungeons(id),
    level INT NOT NULL,
    width INT NOT NULL DEFAULT 8,
    height INT NOT NULL DEFAULT 8,
    layout JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
