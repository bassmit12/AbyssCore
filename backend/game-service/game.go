package game

// game-service: manages hero state and player actions.
// Phase 4B in PLAN.md

import "context"

// Hero represents a player's character in a run.
type Hero struct {
	ID       string `json:"id"`
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
	Class    string `json:"class"`
	HP       int    `json:"hp"`
	MaxHP    int    `json:"max_hp"`
	Level    int    `json:"level"`
	XP       int    `json:"xp"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

type CreateHeroRequest struct {
	Name  string `json:"name"`
	Class string `json:"class"` // warrior | rogue | mage
}

// CreateHero creates a new hero for the authenticated player.
//
//encore:api auth method=POST path=/hero
func CreateHero(ctx context.Context, req *CreateHeroRequest) (*Hero, error) {
	// TODO: Phase 4B
	panic("not implemented")
}

// GetHero returns the hero state.
//
//encore:api auth method=GET path=/hero/:id
func GetHero(ctx context.Context, id string) (*Hero, error) {
	// TODO: Phase 4B
	panic("not implemented")
}

type MoveRequest struct {
	Direction string `json:"direction"` // up | down | left | right
}

// Move moves the hero in the dungeon and publishes dungeon.player.moved.
//
//encore:api auth method=POST path=/hero/:id/move
func Move(ctx context.Context, id string, req *MoveRequest) (*Hero, error) {
	// TODO: Phase 4B
	panic("not implemented")
}
