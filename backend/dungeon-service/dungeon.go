package dungeon

import (
	"context"
	"fmt"
	"math/rand"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("dungeon", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// ─── Types ───────────────────────────────────────────────────────────────────

type Room struct {
	X           int      `json:"x"`
	Y           int      `json:"y"`
	HasChest    bool     `json:"has_chest"`
	ChestOpened bool     `json:"chest_opened"`
	Exits       []string `json:"exits"`
	Monsters    []string `json:"monsters"` // monster IDs, populated by combat-service
}

type Floor struct {
	DungeonID string    `json:"dungeon_id"`
	Level     int       `json:"level"`
	Rooms     [][]Room  `json:"rooms"`
}

type StartDungeonRequest struct {
	HeroID string `json:"hero_id"`
}

// ─── API ─────────────────────────────────────────────────────────────────────

// StartDungeon creates a dungeon run and generates floor 1.
//
//encore:api auth method=POST path=/runs/new
func StartDungeon(ctx context.Context, req *StartDungeonRequest) (*Floor, error) {
	// Create dungeon record
	var dungeonID string
	err := db.QueryRow(ctx, `
		INSERT INTO dungeon.dungeons (hero_id) VALUES ($1::uuid) RETURNING id
	`, req.HeroID).Scan(&dungeonID)
	if err != nil {
		return nil, fmt.Errorf("create dungeon: %w", err)
	}

	floor, err := generateFloor(ctx, dungeonID, 1)
	if err != nil {
		return nil, err
	}

	publishEvent(ctx, "dungeon.floor.entered", map[string]any{
		"hero_id":    req.HeroID,
		"dungeon_id": dungeonID,
		"level":      1,
	})

	return floor, nil
}

// GetFloor returns a dungeon floor with all rooms.
//
//encore:api auth method=GET path=/dungeon/:dungeonID/floor/:level
func GetFloor(ctx context.Context, dungeonID string, level int) (*Floor, error) {
	var floorID string
	err := db.QueryRow(ctx, `
		SELECT id FROM dungeon.floors WHERE dungeon_id = $1::uuid AND level = $2
	`, dungeonID, level).Scan(&floorID)
	if err != nil {
		return nil, fmt.Errorf("floor not found: %w", err)
	}

	rooms, err := loadRooms(ctx, floorID)
	if err != nil {
		return nil, err
	}

	return &Floor{
		DungeonID: dungeonID,
		Level:     level,
		Rooms:     rooms,
	}, nil
}

type OpenChestRequest struct {
	HeroID    string `json:"hero_id"`
	DungeonID string `json:"dungeon_id"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
}

type OpenChestResponse struct {
	ItemFound bool   `json:"item_found"`
	Message   string `json:"message"`
}

// OpenChest opens a chest in the current room, granting loot to the hero.
//
//encore:api auth method=POST path=/dungeon/:dungeonID/chest
func OpenChest(ctx context.Context, dungeonID string, req *OpenChestRequest) (*OpenChestResponse, error) {
	// Find the current floor
	var floorID string
	err := db.QueryRow(ctx, `
		SELECT id FROM dungeon.floors WHERE dungeon_id = $1::uuid ORDER BY level DESC LIMIT 1
	`, dungeonID).Scan(&floorID)
	if err != nil {
		return nil, fmt.Errorf("floor not found: %w", err)
	}

	// Check room has an unopened chest
	var hasChest, chestOpened bool
	err = db.QueryRow(ctx, `
		SELECT has_chest, chest_opened FROM dungeon.rooms WHERE floor_id = $1::uuid AND x = $2 AND y = $3
	`, floorID, req.X, req.Y).Scan(&hasChest, &chestOpened)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}
	if !hasChest {
		return &OpenChestResponse{ItemFound: false, Message: "There is no chest here."}, nil
	}
	if chestOpened {
		return &OpenChestResponse{ItemFound: false, Message: "This chest has already been opened."}, nil
	}

	// Mark chest as opened
	_, err = db.Exec(ctx, `
		UPDATE dungeon.rooms SET chest_opened = TRUE WHERE floor_id = $1::uuid AND x = $2 AND y = $3
	`, floorID, req.X, req.Y)
	if err != nil {
		return nil, fmt.Errorf("mark chest opened: %w", err)
	}

	// Publish event — inventory-service will roll loot and add to hero's bag
	publishEvent(ctx, "dungeon.chest.opened", map[string]any{
		"hero_id":    req.HeroID,
		"dungeon_id": dungeonID,
		"x":          req.X,
		"y":          req.Y,
	})

	return &OpenChestResponse{ItemFound: true, Message: "You open the chest and find something inside!"}, nil
}

// DescendFloor generates the next floor when the hero reaches the stairs.
//
//encore:api auth method=POST path=/dungeon/:dungeonID/descend
func DescendFloor(ctx context.Context, dungeonID string) (*Floor, error) {
	var currentLevel int
	err := db.QueryRow(ctx, `
		SELECT COALESCE(MAX(level), 0) FROM dungeon.floors WHERE dungeon_id = $1::uuid
	`, dungeonID).Scan(&currentLevel)
	if err != nil {
		return nil, err
	}

	return generateFloor(ctx, dungeonID, currentLevel+1)
}

// ─── Procedural Generation ───────────────────────────────────────────────────

const (
	gridWidth  = 8
	gridHeight = 8
)

// generateFloor creates a new floor using a random walk algorithm.
// - Starts at (0,0), walks randomly to fill ~60% of the grid
// - Each room gets random exits, chest chance, monster count scales with level
func generateFloor(ctx context.Context, dungeonID string, level int) (*Floor, error) {
	var floorID string
	err := db.QueryRow(ctx, `
		INSERT INTO dungeon.floors (dungeon_id, level, width, height)
		VALUES ($1::uuid, $2, $3, $4) RETURNING id
	`, dungeonID, level, gridWidth, gridHeight).Scan(&floorID)
	if err != nil {
		return nil, fmt.Errorf("create floor: %w", err)
	}

	// Random walk to determine which cells are rooms
	visited := randomWalk(gridWidth, gridHeight)

	// Build rooms and compute exits
	roomGrid := make([][]Room, gridHeight)
	for y := 0; y < gridHeight; y++ {
		roomGrid[y] = make([]Room, gridWidth)
		for x := 0; x < gridWidth; x++ {
			if !visited[y][x] {
				continue
			}

			exits := computeExits(visited, x, y)
			hasChest := rand.Float32() < chestChance(level)

			roomGrid[y][x] = Room{
				X:        x,
				Y:        y,
				HasChest: hasChest,
				Exits:    exits,
			}

			_, err = db.Exec(ctx, `
				INSERT INTO dungeon.rooms (floor_id, x, y, has_chest, exits)
				VALUES ($1::uuid, $2, $3, $4, $5)
			`, floorID, x, y, hasChest, exits)
			if err != nil {
				return nil, fmt.Errorf("insert room: %w", err)
			}
		}
	}

	// Spawn monsters via combat-service (Phase 4C)
	// For now, mark monster positions in room grid
	spawnMonsters(ctx, floorID, roomGrid, level)

	return &Floor{
		DungeonID: dungeonID,
		Level:     level,
		Rooms:     roomGrid,
	}, nil
}

// randomWalk fills a grid using a drunk-walk algorithm. Returns which cells are active rooms.
func randomWalk(w, h int) [][]bool {
	visited := make([][]bool, h)
	for i := range visited {
		visited[i] = make([]bool, w)
	}

	x, y := 0, 0
	visited[y][x] = true
	target := (w * h * 6) / 10 // ~60% fill

	dirs := [][2]int{{0, -1}, {0, 1}, {1, 0}, {-1, 0}}
	count := 1

	for count < target {
		d := dirs[rand.Intn(4)]
		nx, ny := x+d[0], y+d[1]
		if nx >= 0 && nx < w && ny >= 0 && ny < h {
			x, y = nx, ny
			if !visited[y][x] {
				visited[y][x] = true
				count++
			}
		}
	}
	return visited
}

func computeExits(visited [][]bool, x, y int) []string {
	exits := []string{}
	if y > 0 && visited[y-1][x] {
		exits = append(exits, "north")
	}
	if y < gridHeight-1 && visited[y+1][x] {
		exits = append(exits, "south")
	}
	if x < gridWidth-1 && visited[y][x+1] {
		exits = append(exits, "east")
	}
	if x > 0 && visited[y][x-1] {
		exits = append(exits, "west")
	}
	return exits
}

func chestChance(level int) float32 {
	base := float32(0.15)
	if level > 5 {
		base = 0.20
	}
	return base
}

// spawnMonsters calls combat-service to place monsters in rooms.
// Stub until Phase 4C and Phase 5 are wired.
func spawnMonsters(ctx context.Context, floorID string, rooms [][]Room, level int) {
	// TODO: Phase 4C - call combat-service to seed monsters for this floor
}

func loadRooms(ctx context.Context, floorID string) ([][]Room, error) {
	rows, err := db.Query(ctx, `
		SELECT x, y, has_chest, chest_opened, exits FROM dungeon.rooms WHERE floor_id = $1::uuid ORDER BY y, x
	`, floorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grid := make([][]Room, gridHeight)
	for i := range grid {
		grid[i] = make([]Room, gridWidth)
	}

	for rows.Next() {
		var r Room
		var exits []string
		if err := rows.Scan(&r.X, &r.Y, &r.HasChest, &r.ChestOpened, &exits); err != nil {
			return nil, err
		}
		r.Exits = exits
		grid[r.Y][r.X] = r
	}
	return grid, rows.Err()
}

