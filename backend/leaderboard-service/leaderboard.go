package leaderboard

// leaderboard-service: tracks completed runs and rankings.
// Phase 4E in PLAN.md

import "context"

type Run struct {
	HeroID          string `json:"hero_id"`
	HeroName        string `json:"hero_name"`
	FloorsCleared   int    `json:"floors_cleared"`
	MonstersKilled  int    `json:"monsters_killed"`
	ItemsFound      int    `json:"items_found"`
	Score           int    `json:"score"`
}

type LeaderboardResponse struct {
	Runs []Run `json:"runs"`
}

// GetLeaderboard returns the top 10 runs.
//
//encore:api public method=GET path=/leaderboard
func GetLeaderboard(ctx context.Context) (*LeaderboardResponse, error) {
	// TODO: Phase 4E
	panic("not implemented")
}
