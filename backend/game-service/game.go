package game

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("game", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// Hero stats by class
var classStats = map[string]struct{ hp, damage int }{
	"warrior": {hp: 120, damage: 12},
	"rogue":   {hp: 80, damage: 18},
	"mage":    {hp: 70, damage: 22},
}

// XP required to reach next level
func xpForLevel(level int) int {
	return level * 100
}

// ─── Types ───────────────────────────────────────────────────────────────────

type Hero struct {
	ID        string    `json:"id"`
	PlayerID  string    `json:"player_id"`
	Name      string    `json:"name"`
	Class     string    `json:"class"`
	HP        int       `json:"hp"`
	MaxHP     int       `json:"max_hp"`
	Level     int       `json:"level"`
	XP        int       `json:"xp"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	DungeonID string    `json:"dungeon_id,omitempty"`
	Alive     bool      `json:"alive"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateHeroRequest struct {
	Name  string `json:"name"`
	Class string `json:"class"`
}

type MoveRequest struct {
	Direction string `json:"direction"`
}

// ─── API ─────────────────────────────────────────────────────────────────────

// CreateHero creates a new hero for the authenticated player.
//
//encore:api auth method=POST path=/hero
func CreateHero(ctx context.Context, req *CreateHeroRequest) (*Hero, error) {
	uid := string(auth.UserID())

	stats, ok := classStats[req.Class]
	if !ok {
		return nil, errors.New("invalid class: must be warrior, rogue, or mage")
	}
	if req.Name == "" {
		return nil, errors.New("hero name is required")
	}

	hero := &Hero{}
	err := db.QueryRow(ctx, `
		INSERT INTO game.heroes (player_id, name, class, hp, max_hp, level, xp, x, y, alive)
		VALUES ($1, $2, $3, $4, $4, 1, 0, 0, 0, TRUE)
		RETURNING id, player_id, name, class, hp, max_hp, level, xp, x, y, alive, updated_at
	`, uid, req.Name, req.Class, stats.hp).Scan(
		&hero.ID, &hero.PlayerID, &hero.Name, &hero.Class,
		&hero.HP, &hero.MaxHP, &hero.Level, &hero.XP,
		&hero.X, &hero.Y, &hero.Alive, &hero.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create hero: %w", err)
	}
	return hero, nil
}

// GetHero returns a hero by ID.
//
//encore:api auth method=GET path=/hero/:id
func GetHero(ctx context.Context, id string) (*Hero, error) {
	hero := &Hero{}
	err := db.QueryRow(ctx, `
		SELECT id, player_id, name, class, hp, max_hp, level, xp, x, y,
		       COALESCE(dungeon_id::text, '') AS dungeon_id, alive, updated_at
		FROM game.heroes WHERE id = $1
	`, id).Scan(
		&hero.ID, &hero.PlayerID, &hero.Name, &hero.Class,
		&hero.HP, &hero.MaxHP, &hero.Level, &hero.XP,
		&hero.X, &hero.Y, &hero.DungeonID, &hero.Alive, &hero.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("hero not found: %w", err)
	}
	return hero, nil
}

// Move moves the hero on the current floor.
//
//encore:api auth method=POST path=/hero/:id/move
func Move(ctx context.Context, id string, req *MoveRequest) (*Hero, error) {
	hero, err := GetHero(ctx, id)
	if err != nil {
		return nil, err
	}
	if !hero.Alive {
		return nil, errors.New("hero is dead")
	}

	dx, dy := directionDelta(req.Direction)
	if dx == 0 && dy == 0 {
		return nil, fmt.Errorf("invalid direction: %s", req.Direction)
	}

	newX := hero.X + dx
	newY := hero.Y + dy

	_, err = db.Exec(ctx, `
		UPDATE game.heroes SET x = $1, y = $2, updated_at = now() WHERE id = $3
	`, newX, newY, id)
	if err != nil {
		return nil, fmt.Errorf("move hero: %w", err)
	}

	hero.X = newX
	hero.Y = newY

	// Publish event (Phase 5 - RabbitMQ)
	publishEvent(ctx, "dungeon.player.moved", map[string]any{
		"hero_id":   id,
		"direction": req.Direction,
		"x":         newX,
		"y":         newY,
	})

	return hero, nil
}

// AwardXP adds XP to a hero and handles level ups. Called by combat-service via RabbitMQ consumer.
//
//encore:api auth method=POST path=/hero/:id/xp
func AwardXP(ctx context.Context, id string, req *AwardXPRequest) (*Hero, error) {
	hero, err := GetHero(ctx, id)
	if err != nil {
		return nil, err
	}

	hero.XP += req.Amount
	leveledUp := false

	for hero.XP >= xpForLevel(hero.Level) {
		hero.XP -= xpForLevel(hero.Level)
		hero.Level++
		hero.MaxHP += 10
		hero.HP = min(hero.HP+10, hero.MaxHP)
		leveledUp = true
	}

	_, err = db.Exec(ctx, `
		UPDATE game.heroes
		SET xp = $1, level = $2, hp = $3, max_hp = $4, updated_at = now()
		WHERE id = $5
	`, hero.XP, hero.Level, hero.HP, hero.MaxHP, id)
	if err != nil {
		return nil, err
	}

	if leveledUp {
		publishEvent(ctx, "game.hero.levelup", map[string]any{
			"hero_id": id,
			"level":   hero.Level,
		})
	}

	return hero, nil
}

type AwardXPRequest struct {
	Amount int `json:"amount"`
}

// MarkDead marks a hero as dead. Called by combat-service.
//
//encore:api auth method=POST path=/hero/:id/die
func MarkDead(ctx context.Context, id string) error {
	_, err := db.Exec(ctx, `
		UPDATE game.heroes SET alive = FALSE, updated_at = now() WHERE id = $1
	`, id)
	return err
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func directionDelta(dir string) (dx, dy int) {
	switch dir {
	case "north", "up":
		return 0, -1
	case "south", "down":
		return 0, 1
	case "east", "right":
		return 1, 0
	case "west", "left":
		return -1, 0
	}
	return 0, 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// publishEvent is a stub until Phase 5 wires RabbitMQ
func publishEvent(ctx context.Context, routingKey string, payload map[string]any) {
	// TODO: Phase 5 - publish to abysscore.events exchange
}
