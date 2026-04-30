CREATE TABLE IF NOT EXISTS combat.monsters (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    floor_id   UUID NOT NULL,
    room_x     INT  NOT NULL,
    room_y     INT  NOT NULL,
    name       TEXT NOT NULL,
    hp         INT  NOT NULL,
    max_hp     INT  NOT NULL,
    damage     INT  NOT NULL,
    status     TEXT NOT NULL DEFAULT 'alive' CHECK (status IN ('alive', 'dead')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_monsters_floor_id ON combat.monsters(floor_id);
CREATE INDEX IF NOT EXISTS idx_monsters_status   ON combat.monsters(status);

CREATE TABLE IF NOT EXISTS combat.combat_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id             UUID NOT NULL,
    monster_id          UUID NOT NULL,
    hero_damage_dealt   INT  NOT NULL,
    monster_damage_back INT  NOT NULL,
    monster_died        BOOLEAN NOT NULL,
    hero_died           BOOLEAN NOT NULL,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_combat_log_hero_id ON combat.combat_log(hero_id);
