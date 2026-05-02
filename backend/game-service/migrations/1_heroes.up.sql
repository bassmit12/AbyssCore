CREATE SCHEMA IF NOT EXISTS game;
CREATE TABLE IF NOT EXISTS game.heroes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    class TEXT NOT NULL,
    level INT NOT NULL DEFAULT 1,
    xp INT NOT NULL DEFAULT 0,
    hp INT NOT NULL,
    max_hp INT NOT NULL,
    damage INT NOT NULL,
    dodge_chance FLOAT NOT NULL DEFAULT 0,
    x INT NOT NULL DEFAULT 0,
    y INT NOT NULL DEFAULT 0,
    floor_id UUID,
    dungeon_id UUID,
    status TEXT NOT NULL DEFAULT 'alive',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
