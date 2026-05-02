CREATE SCHEMA IF NOT EXISTS map;

-- A run is one full playthrough attempt by a hero
CREATE TABLE IF NOT EXISTS map.runs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hero_id     UUID NOT NULL,
    status      TEXT NOT NULL DEFAULT 'active', -- active | won | lost
    current_floor INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Each floor is a set of nodes arranged in columns
CREATE TABLE IF NOT EXISTS map.floors (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id      UUID NOT NULL REFERENCES map.runs(id),
    level       INT NOT NULL,
    cols        INT NOT NULL DEFAULT 7,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A node is a stop on the map: combat, elite, event, shop, rest, boss
CREATE TABLE IF NOT EXISTS map.nodes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    floor_id    UUID NOT NULL REFERENCES map.floors(id),
    col         INT NOT NULL,
    row         INT NOT NULL,
    type        TEXT NOT NULL, -- combat | elite | event | shop | rest | boss
    cleared     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- DAG edges between nodes (player can travel from_node -> to_node)
CREATE TABLE IF NOT EXISTS map.paths (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    floor_id    UUID NOT NULL REFERENCES map.floors(id),
    from_node_id UUID NOT NULL REFERENCES map.nodes(id),
    to_node_id   UUID NOT NULL REFERENCES map.nodes(id),
    UNIQUE (from_node_id, to_node_id)
);

-- Track which node the hero is currently at
CREATE TABLE IF NOT EXISTS map.hero_positions (
    hero_id     UUID PRIMARY KEY,
    run_id      UUID NOT NULL REFERENCES map.runs(id),
    node_id     UUID REFERENCES map.nodes(id), -- NULL = not yet placed (just started floor)
    floor_id    UUID NOT NULL REFERENCES map.floors(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
