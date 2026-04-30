CREATE TABLE IF NOT EXISTS leaderboard.runs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id          UUID NOT NULL,
    hero_name        TEXT NOT NULL,
    player_id        TEXT NOT NULL,
    floors_cleared   INT  NOT NULL DEFAULT 0,
    monsters_killed  INT  NOT NULL DEFAULT 0,
    items_found      INT  NOT NULL DEFAULT 0,
    score            INT  NOT NULL DEFAULT 0,
    completed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_runs_score ON leaderboard.runs(score DESC);
