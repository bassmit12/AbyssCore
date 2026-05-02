-- Fix column names to match leaderboard.go expectations
DO $$
BEGIN
    -- floors_reached -> floors_cleared
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'leaderboard' AND table_name = 'runs' AND column_name = 'floors_reached'
    ) THEN
        ALTER TABLE leaderboard.runs RENAME COLUMN floors_reached TO floors_cleared;
    END IF;

    -- items_collected -> items_found
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'leaderboard' AND table_name = 'runs' AND column_name = 'items_collected'
    ) THEN
        ALTER TABLE leaderboard.runs RENAME COLUMN items_collected TO items_found;
    END IF;
END $$;

-- Add missing columns
ALTER TABLE leaderboard.runs
    ADD COLUMN IF NOT EXISTS player_id UUID,
    ADD COLUMN IF NOT EXISTS monsters_killed INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS items_found INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS floors_cleared INT NOT NULL DEFAULT 0;

-- hero_class is in the original schema but not used by Go code — keep it, it's fine
