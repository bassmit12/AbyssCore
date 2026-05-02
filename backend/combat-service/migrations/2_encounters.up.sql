-- Keep the old monsters table (it's referenced in old code still running)
-- Add new card-based combat tables alongside

CREATE TABLE IF NOT EXISTS combat.encounters (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id      UUID NOT NULL,
    node_id      UUID NOT NULL,           -- map.nodes.id
    node_type    TEXT NOT NULL DEFAULT 'combat',
    status       TEXT NOT NULL DEFAULT 'active',  -- active | won | lost
    turn_number  INT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Monsters spawned for this encounter
CREATE TABLE IF NOT EXISTS combat.encounter_monsters (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    encounter_id UUID NOT NULL REFERENCES combat.encounters(id),
    name         TEXT NOT NULL,
    hp           INT NOT NULL,
    max_hp       INT NOT NULL,
    block        INT NOT NULL DEFAULT 0,
    damage       INT NOT NULL,
    intents      JSONB NOT NULL DEFAULT '[]',  -- [{"type":"attack","value":8}, ...]
    intent_index INT NOT NULL DEFAULT 0,
    status       TEXT NOT NULL DEFAULT 'alive',
    statuses     JSONB NOT NULL DEFAULT '{}'   -- {"vulnerable":2, "weak":1, ...}
);

-- Hero state per encounter (hand, draw/discard piles, block, energy, statuses)
CREATE TABLE IF NOT EXISTS combat.hero_combat_state (
    encounter_id UUID PRIMARY KEY REFERENCES combat.encounters(id),
    hero_id      UUID NOT NULL,
    hp           INT NOT NULL,
    max_hp       INT NOT NULL,
    block        INT NOT NULL DEFAULT 0,
    energy       INT NOT NULL DEFAULT 3,
    max_energy   INT NOT NULL DEFAULT 3,
    draw_pile    JSONB NOT NULL DEFAULT '[]',    -- [card_deck_id, ...]
    hand         JSONB NOT NULL DEFAULT '[]',
    discard_pile JSONB NOT NULL DEFAULT '[]',
    statuses     JSONB NOT NULL DEFAULT '{}',    -- {"strength":2, "vulnerable":0, ...}
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
