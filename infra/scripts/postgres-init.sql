-- AbyssCore database init
-- Creates separate schemas per service (logical separation, single Postgres instance for local dev)

CREATE SCHEMA IF NOT EXISTS game;
CREATE SCHEMA IF NOT EXISTS dungeon;
CREATE SCHEMA IF NOT EXISTS combat;
CREATE SCHEMA IF NOT EXISTS inventory;
CREATE SCHEMA IF NOT EXISTS leaderboard;
