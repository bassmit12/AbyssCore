-- Rename user_id -> player_id
ALTER TABLE game.heroes RENAME COLUMN user_id TO player_id;

-- Replace status TEXT with alive BOOL
ALTER TABLE game.heroes ADD COLUMN alive BOOLEAN NOT NULL DEFAULT TRUE;
UPDATE game.heroes SET alive = (status = 'alive');
ALTER TABLE game.heroes DROP COLUMN status;

-- Add updated_at
ALTER TABLE game.heroes ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
