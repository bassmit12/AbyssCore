package game

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
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
	Gold      int       `json:"gold"`
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
	rawUID, _ := auth.UserID()
	uid := string(rawUID)

	stats, ok := classStats[strings.ToLower(req.Class)]
	if !ok {
		return nil, errors.New("invalid class: must be warrior, rogue, or mage")
	}
	if req.Name == "" {
		return nil, errors.New("hero name is required")
	}

	hero := &Hero{}
	err := db.QueryRow(ctx, `
		INSERT INTO game.heroes (player_id, name, class, hp, max_hp, damage, level, xp, x, y, alive)
		VALUES ($1, $2, $3, $4, $4, $5, 1, 0, 0, 0, TRUE)
		RETURNING id, player_id, name, class, hp, max_hp, level, xp, x, y, alive, updated_at
	`, uid, req.Name, req.Class, stats.hp, stats.damage).Scan(
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
		       COALESCE(dungeon_id::text, '') AS dungeon_id, alive, COALESCE(gold,0) AS gold, updated_at
		FROM game.heroes WHERE id = $1
	`, id).Scan(
		&hero.ID, &hero.PlayerID, &hero.Name, &hero.Class,
		&hero.HP, &hero.MaxHP, &hero.Level, &hero.XP,
		&hero.X, &hero.Y, &hero.DungeonID, &hero.Alive, &hero.Gold, &hero.UpdatedAt,
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

// ─── Endpoints for inter-service calls ────────────────────────────────────────

type HeroStatsResponse struct {
	HP       int    `json:"hp"`
	MaxHP    int    `json:"max_hp"`
	Class    string `json:"class"`
	Alive    bool   `json:"alive"`
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
}

// GetHeroStats returns basic stats for a hero (called by combat-service).
//
//encore:api public method=GET path=/hero/:id/stats
func GetHeroStats(ctx context.Context, id string) (*HeroStatsResponse, error) {
	r := &HeroStatsResponse{}
	err := db.QueryRow(ctx, `SELECT hp, max_hp, class, alive, player_id, name FROM game.heroes WHERE id=$1::uuid`, id).
		Scan(&r.HP, &r.MaxHP, &r.Class, &r.Alive, &r.PlayerID, &r.Name)
	if err != nil {
		return nil, fmt.Errorf("hero not found: %w", err)
	}
	return r, nil
}

type AddGoldRequest struct {
	Gold int `json:"gold"`
}

// AddGold adds gold to a hero (called by combat-service on victory).
//
//encore:api public method=POST path=/hero/:id/add-gold
func AddGold(ctx context.Context, id string, req *AddGoldRequest) error {
	_, err := db.Exec(ctx, `UPDATE game.heroes SET gold = gold + $1 WHERE id=$2::uuid`, req.Gold, id)
	return err
}

type SetAliveRequest struct {
	Alive bool `json:"alive"`
}

// SetAlive marks a hero alive or dead (called by combat-service on death).
//
//encore:api public method=POST path=/hero/:id/set-alive
func SetAlive(ctx context.Context, id string, req *SetAliveRequest) error {
	_, err := db.Exec(ctx, `UPDATE game.heroes SET alive=$1 WHERE id=$2::uuid`, req.Alive, id)
	return err
}

type HealRequest struct {
	Amount int `json:"amount"`
}

// HealHero adds HP to a hero (capped at max_hp).
//
//encore:api public method=POST path=/hero/:id/heal
func HealHero(ctx context.Context, id string, req *HealRequest) error {
	_, err := db.Exec(ctx, `
		UPDATE game.heroes SET hp = LEAST(hp + $1, max_hp) WHERE id=$2::uuid
	`, req.Amount, id)
	return err
}

type DamageRequest struct {
	Amount int `json:"amount"`
}

// DamageHero subtracts HP from a hero.
//
//encore:api public method=POST path=/hero/:id/damage
func DamageHero(ctx context.Context, id string, req *DamageRequest) error {
	_, err := db.Exec(ctx, `
		UPDATE game.heroes SET hp = GREATEST(hp - $1, 0) WHERE id=$2::uuid
	`, req.Amount, id)
	return err
}

type DeductGoldRequest struct {
	Gold int `json:"gold"`
}

type DeductGoldResponse struct {
	Gold int `json:"gold"` // remaining gold
}

// DeductGold removes gold from a hero (e.g. shop purchase). Returns error if insufficient funds.
//
//encore:api public method=POST path=/hero/:id/deduct-gold
func DeductGold(ctx context.Context, id string, req *DeductGoldRequest) (*DeductGoldResponse, error) {
	var remaining int
	err := db.QueryRow(ctx, `
		UPDATE game.heroes SET gold = gold - $1 WHERE id=$2::uuid AND gold >= $1 RETURNING gold
	`, req.Gold, id).Scan(&remaining)
	if err != nil {
		return nil, fmt.Errorf("insufficient gold or hero not found")
	}
	return &DeductGoldResponse{Gold: remaining}, nil
}

// ─── Random Events ────────────────────────────────────────────────────────────

type EventChoice struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type RandomEvent struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Choices     []EventChoice `json:"choices"`
}

type EventOutcome struct {
	ChoiceID    string `json:"choice_id"`
	Description string `json:"description"`
	GoldDelta   int    `json:"gold_delta"`
	HPDelta     int    `json:"hp_delta"`
}

type ResolveEventRequest struct {
	HeroID   string `json:"hero_id"`
	EventID  string `json:"event_id"`
	ChoiceID string `json:"choice_id"`
}

var randomEvents = []RandomEvent{
	{
		ID:          "mysterious_merchant",
		Title:       "Mysterious Merchant",
		Description: "A hooded figure offers you a deal from the shadows...",
		Choices: []EventChoice{
			{ID: "buy", Label: "Buy a secret potion (30 gold)", Description: "Spend 30 gold for a powerful healing potion."},
			{ID: "steal", Label: "Attempt to steal", Description: "Risk taking without paying — might end badly."},
			{ID: "ignore", Label: "Walk away", Description: "You have bigger problems than this."},
		},
	},
	{
		ID:          "ancient_shrine",
		Title:       "Ancient Shrine",
		Description: "A crumbling altar radiates faint warmth. An offering bowl sits at its base.",
		Choices: []EventChoice{
			{ID: "offer_gold", Label: "Offer 20 gold", Description: "The gods may favor you."},
			{ID: "offer_blood", Label: "Offer blood (lose 10 HP)", Description: "A blood sacrifice. Perhaps power awaits."},
			{ID: "take", Label: "Loot the altar", Description: "Take whatever is here and move on."},
		},
	},
	{
		ID:          "wounded_adventurer",
		Title:       "Wounded Adventurer",
		Description: "A battered hero slumps against the wall. 'Please... help me...'",
		Choices: []EventChoice{
			{ID: "heal", Label: "Use a potion to heal them (lose 20 HP)", Description: "Give up your own life force to save them."},
			{ID: "loot", Label: "Take their belongings (+30 gold)", Description: "They won't need it much longer."},
			{ID: "help_fight", Label: "Help them fight off their pursuers", Description: "Could be rewarding... or fatal."},
		},
	},
	{
		ID:          "cursed_fountain",
		Title:       "Cursed Fountain",
		Description: "A dark fountain bubbles ominously. The liquid smells of iron.",
		Choices: []EventChoice{
			{ID: "drink", Label: "Drink from the fountain", Description: "Unknown effects. Could be anything."},
			{ID: "fill_vial", Label: "Fill a vial (+15 gold)", Description: "Someone might pay for this cursed water."},
			{ID: "avoid", Label: "Avoid it entirely", Description: "Some things are best left alone."},
		},
	},
	{
		ID:          "trapped_chest",
		Title:       "Trapped Chest",
		Description: "A glittering chest sits in the middle of the room. You spot a tripwire.",
		Choices: []EventChoice{
			{ID: "disarm", Label: "Carefully disarm and open (+40 gold)", Description: "High risk, high reward."},
			{ID: "smash", Label: "Smash it open (take 15 damage)", Description: "Take the hit. Get the loot."},
			{ID: "leave", Label: "Leave it", Description: "Greed is a slow death."},
		},
	},
}

// GetRandomEvent returns a random event for a floor encounter.
//
//encore:api public method=GET path=/game/event
func GetRandomEvent(ctx context.Context) (*RandomEvent, error) {
	idx := rand.Intn(len(randomEvents))
	e := randomEvents[idx]
	return &e, nil
}

// ResolveEvent applies the outcome of a player's event choice.
//
//encore:api public method=POST path=/game/event/resolve
func ResolveEvent(ctx context.Context, req *ResolveEventRequest) (*EventOutcome, error) {
	outcome := applyEventChoice(req.EventID, req.ChoiceID)

	if outcome.GoldDelta > 0 {
		_, _ = db.Exec(ctx, `UPDATE game.heroes SET gold = gold + $1 WHERE id=$2::uuid`, outcome.GoldDelta, req.HeroID)
	} else if outcome.GoldDelta < 0 {
		_, _ = db.Exec(ctx, `UPDATE game.heroes SET gold = GREATEST(gold + $1, 0) WHERE id=$2::uuid`, outcome.GoldDelta, req.HeroID)
	}

	if outcome.HPDelta > 0 {
		_, _ = db.Exec(ctx, `UPDATE game.heroes SET hp = LEAST(hp + $1, max_hp) WHERE id=$2::uuid`, outcome.HPDelta, req.HeroID)
	} else if outcome.HPDelta < 0 {
		_, _ = db.Exec(ctx, `UPDATE game.heroes SET hp = GREATEST(hp + $1, 1) WHERE id=$2::uuid`, outcome.HPDelta, req.HeroID)
	}

	return outcome, nil
}

func applyEventChoice(eventID, choiceID string) *EventOutcome {
	r := rand.Float32()
	switch eventID {
	case "mysterious_merchant":
		switch choiceID {
		case "buy":
			return &EventOutcome{ChoiceID: choiceID, Description: "You buy a glowing vial. Your wounds close slightly.", GoldDelta: -30, HPDelta: 25}
		case "steal":
			if r < 0.4 {
				return &EventOutcome{ChoiceID: choiceID, Description: "You snatched the potion and ran!", GoldDelta: 0, HPDelta: 20}
			}
			return &EventOutcome{ChoiceID: choiceID, Description: "The merchant's blade finds your ribs.", GoldDelta: -10, HPDelta: -20}
		default:
			return &EventOutcome{ChoiceID: choiceID, Description: "You walk away. Perhaps wise.", GoldDelta: 0, HPDelta: 0}
		}
	case "ancient_shrine":
		switch choiceID {
		case "offer_gold":
			if r < 0.6 {
				return &EventOutcome{ChoiceID: choiceID, Description: "The shrine glows. You feel restored!", GoldDelta: -20, HPDelta: 30}
			}
			return &EventOutcome{ChoiceID: choiceID, Description: "Nothing happens. Gold wasted.", GoldDelta: -20, HPDelta: 0}
		case "offer_blood":
			if r < 0.7 {
				return &EventOutcome{ChoiceID: choiceID, Description: "Power surges through you. +20 max HP restored.", GoldDelta: 0, HPDelta: 20}
			}
			return &EventOutcome{ChoiceID: choiceID, Description: "The shrine rejects your offering.", GoldDelta: 0, HPDelta: -10}
		default:
			return &EventOutcome{ChoiceID: choiceID, Description: "You find a few coins in the offering bowl.", GoldDelta: 15, HPDelta: 0}
		}
	case "wounded_adventurer":
		switch choiceID {
		case "heal":
			return &EventOutcome{ChoiceID: choiceID, Description: "They recover and press their last potion into your hands.", GoldDelta: 0, HPDelta: -10}
		case "loot":
			return &EventOutcome{ChoiceID: choiceID, Description: "You take their coins. You tell yourself they'd have done the same.", GoldDelta: 30, HPDelta: 0}
		default:
			if r < 0.5 {
				return &EventOutcome{ChoiceID: choiceID, Description: "You drive off the attackers! The adventurer is grateful and shares their gold.", GoldDelta: 25, HPDelta: 0}
			}
			return &EventOutcome{ChoiceID: choiceID, Description: "You fight bravely but take a hit in the chaos.", GoldDelta: 10, HPDelta: -15}
		}
	case "cursed_fountain":
		switch choiceID {
		case "drink":
			roll := rand.Intn(3)
			if roll == 0 {
				return &EventOutcome{ChoiceID: choiceID, Description: "You feel strangely invigorated!", GoldDelta: 0, HPDelta: 25}
			} else if roll == 1 {
				return &EventOutcome{ChoiceID: choiceID, Description: "The liquid burns. But you feel tougher.", GoldDelta: 0, HPDelta: -10}
			}
			return &EventOutcome{ChoiceID: choiceID, Description: "Nothing happens. The water tastes foul.", GoldDelta: 0, HPDelta: -5}
		case "fill_vial":
			return &EventOutcome{ChoiceID: choiceID, Description: "An alchemist would pay for this. +15 gold.", GoldDelta: 15, HPDelta: 0}
		default:
			return &EventOutcome{ChoiceID: choiceID, Description: "You move on. No harm done.", GoldDelta: 0, HPDelta: 0}
		}
	case "trapped_chest":
		switch choiceID {
		case "disarm":
			if r < 0.6 {
				return &EventOutcome{ChoiceID: choiceID, Description: "Careful fingers save you. The chest is yours!", GoldDelta: 40, HPDelta: 0}
			}
			return &EventOutcome{ChoiceID: choiceID, Description: "A hidden spring fires a dart into your neck.", GoldDelta: 20, HPDelta: -20}
		case "smash":
			return &EventOutcome{ChoiceID: choiceID, Description: "The trap fires. Blood and gold. Worth it?", GoldDelta: 35, HPDelta: -15}
		default:
			return &EventOutcome{ChoiceID: choiceID, Description: "You leave it. The trap would have been nasty.", GoldDelta: 0, HPDelta: 0}
		}
	}
	return &EventOutcome{ChoiceID: choiceID, Description: "Nothing happens.", GoldDelta: 0, HPDelta: 0}
}

