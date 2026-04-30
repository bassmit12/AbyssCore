package dungeon

// dungeon-service: procedural floor generation and dungeon state.
// Phase 4A in PLAN.md

import "context"

type Room struct {
	X        int      `json:"x"`
	Y        int      `json:"y"`
	HasChest bool     `json:"has_chest"`
	Monsters []string `json:"monsters"` // monster IDs
	Exits    []string `json:"exits"`    // "north" | "south" | "east" | "west"
}

type Floor struct {
	DungeonID string  `json:"dungeon_id"`
	Level     int     `json:"level"`
	Rooms     [][]Room `json:"rooms"`
}

type StartDungeonRequest struct {
	HeroID string `json:"hero_id"`
}

// StartDungeon creates a new dungeon run and generates the first floor.
//
//encore:api auth method=POST path=/dungeon/start
func StartDungeon(ctx context.Context, req *StartDungeonRequest) (*Floor, error) {
	// TODO: Phase 4A
	panic("not implemented")
}

// GetFloor returns the layout of a specific dungeon floor.
//
//encore:api auth method=GET path=/dungeon/:dungeonID/floor/:level
func GetFloor(ctx context.Context, dungeonID string, level int) (*Floor, error) {
	// TODO: Phase 4A
	panic("not implemented")
}
