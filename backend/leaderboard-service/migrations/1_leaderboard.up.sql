CREATE SCHEMA IF NOT EXISTS leaderboard;
CREATE TABLE IF NOT EXISTS leaderboard.runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id UUID NOT NULL,
    hero_name TEXT NOT NULL,
    hero_class TEXT NOT NULL,
    floors_reached INT NOT NULL DEFAULT 0,
    monsters_killed INT NOT NULL DEFAULT 0,
    items_collected INT NOT NULL DEFAULT 0,
    score INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
