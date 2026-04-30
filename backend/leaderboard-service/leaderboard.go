package leaderboard

import (
	"context"
	"fmt"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("leaderboard", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// ─── Types ───────────────────────────────────────────────────────────────────

type Run struct {
	ID             string `json:"id"`
	HeroID         string `json:"hero_id"`
	HeroName       string `json:"hero_name"`
	PlayerID       string `json:"player_id"`
	FloorsCleared  int    `json:"floors_cleared"`
	MonstersKilled int    `json:"monsters_killed"`
	ItemsFound     int    `json:"items_found"`
	Score          int    `json:"score"`
}

type FinalizeRunRequest struct {
	HeroID         string `json:"hero_id"`
	HeroName       string `json:"hero_name"`
	PlayerID       string `json:"player_id"`
	FloorsCleared  int    `json:"floors_cleared"`
	MonstersKilled int    `json:"monsters_killed"`
	ItemsFound     int    `json:"items_found"`
}

type LeaderboardResponse struct {
	Runs []Run `json:"runs"`
}

// ─── API ─────────────────────────────────────────────────────────────────────

// GetLeaderboard returns the top 10 runs by score.
//
//encore:api public method=GET path=/leaderboard
func GetLeaderboard(ctx context.Context) (*LeaderboardResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, hero_id, hero_name, player_id, floors_cleared, monsters_killed, items_found, score
		FROM leaderboard.runs
		ORDER BY score DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []Run{}
	for rows.Next() {
		var r Run
		if err := rows.Scan(
			&r.ID, &r.HeroID, &r.HeroName, &r.PlayerID,
			&r.FloorsCleared, &r.MonstersKilled, &r.ItemsFound, &r.Score,
		); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return &LeaderboardResponse{Runs: runs}, rows.Err()
}

// FinalizeRun saves a completed run. Called by RabbitMQ consumer on game.player.died.
// Also exposed as API for testing.
//
//encore:api auth method=POST path=/leaderboard/finalize
func FinalizeRun(ctx context.Context, req *FinalizeRunRequest) (*Run, error) {
	score := calculateScore(req.FloorsCleared, req.MonstersKilled, req.ItemsFound)

	run := &Run{}
	err := db.QueryRow(ctx, `
		INSERT INTO leaderboard.runs
		  (hero_id, hero_name, player_id, floors_cleared, monsters_killed, items_found, score)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
		RETURNING id, hero_id, hero_name, player_id, floors_cleared, monsters_killed, items_found, score
	`, req.HeroID, req.HeroName, req.PlayerID,
		req.FloorsCleared, req.MonstersKilled, req.ItemsFound, score,
	).Scan(
		&run.ID, &run.HeroID, &run.HeroName, &run.PlayerID,
		&run.FloorsCleared, &run.MonstersKilled, &run.ItemsFound, &run.Score,
	)
	if err != nil {
		return nil, fmt.Errorf("finalize run: %w", err)
	}
	return run, nil
}

// calculateScore: floors are worth most, monsters and items add bonus
func calculateScore(floors, monsters, items int) int {
	return floors*500 + monsters*50 + items*25
}
