package inventory

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("inventory", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// ─── Types ───────────────────────────────────────────────────────────────────

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       int    `json:"value"`
	Rarity      string `json:"rarity"`
	Description string `json:"description"`
}

type Inventory struct {
	HeroID string `json:"hero_id"`
	Items  []Item `json:"items"`
}

type UseItemRequest struct {
	HeroID string `json:"hero_id"`
	ItemID string `json:"item_id"`
}

type LootDropRequest struct {
	HeroID    string `json:"hero_id"`
	MonsterID string `json:"monster_id"`
}

// ─── API ─────────────────────────────────────────────────────────────────────

// GetInventory returns all items in a hero's inventory.
//
//encore:api public method=GET path=/inventory/:heroID
func GetInventory(ctx context.Context, heroID string) (*Inventory, error) {
	rows, err := db.Query(ctx, `
		SELECT hi.id, d.name, d.type, d.value, d.rarity, d.description
		FROM inventory.hero_inventory hi
		JOIN inventory.item_definitions d ON d.id = hi.item_def_id
		WHERE hi.hero_id = $1::uuid
		ORDER BY hi.created_at DESC
	`, heroID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inv := &Inventory{HeroID: heroID, Items: []Item{}}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Value, &item.Rarity, &item.Description); err != nil {
			return nil, err
		}
		inv.Items = append(inv.Items, item)
	}
	return inv, rows.Err()
}

// UseItem uses a consumable from the hero's inventory.
// Potions are consumed and HP is restored. Gold items add gold.
//
//encore:api public method=POST path=/inventory/use
func UseItem(ctx context.Context, req *UseItemRequest) error {
	// Load item
	var itemType string
	var itemValue int
	err := db.QueryRow(ctx, `
		SELECT d.type, d.value
		FROM inventory.hero_inventory hi
		JOIN inventory.item_definitions d ON d.id = hi.item_def_id
		WHERE hi.id = $1::uuid AND hi.hero_id = $2::uuid
	`, req.ItemID, req.HeroID).Scan(&itemType, &itemValue)
	if err != nil {
		return fmt.Errorf("item not found in inventory: %w", err)
	}

	if itemType != "potion" && itemType != "gold" {
		return errors.New("only potions and gold items can be used directly")
	}

	// Remove from inventory
	_, err = db.Exec(ctx, `
		DELETE FROM inventory.hero_inventory WHERE id = $1::uuid
	`, req.ItemID)
	if err != nil {
		return err
	}

	// Publish heal event so game-service updates HP
	publishEvent(ctx, "inventory.item.used", map[string]any{
		"hero_id":   req.HeroID,
		"item_type": itemType,
		"heal":      itemValue,
	})

	return nil
}

// RollLootDrop rolls for a random loot drop and adds it to hero inventory.
// Called by combat-service on combat win. Returns nil if no drop.
//
//encore:api public method=POST path=/inventory/loot
func RollLootDrop(ctx context.Context, req *LootDropRequest) (*Item, error) {
	// 60% chance of any drop
	if rand.Float32() > 0.6 {
		return nil, nil // no drop
	}

	// Weighted random pick from item_definitions
	type itemDef struct {
		id          string
		name        string
		typ         string
		value       int
		rarity      string
		description string
		weight      int
	}

	rows, err := db.Query(ctx, `
		SELECT id, name, type, value, rarity, description, drop_weight FROM inventory.item_definitions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []itemDef
	totalWeight := 0
	for rows.Next() {
		var d itemDef
		if err := rows.Scan(&d.id, &d.name, &d.typ, &d.value, &d.rarity, &d.description, &d.weight); err != nil {
			return nil, err
		}
		defs = append(defs, d)
		totalWeight += d.weight
	}

	if len(defs) == 0 {
		return nil, nil
	}

	// Pick weighted random
	roll := rand.Intn(totalWeight)
	cumulative := 0
	var chosen *itemDef
	for i := range defs {
		cumulative += defs[i].weight
		if roll < cumulative {
			chosen = &defs[i]
			break
		}
	}
	if chosen == nil {
		return nil, nil
	}

	// Add to hero inventory
	var invID string
	err = db.QueryRow(ctx, `
		INSERT INTO inventory.hero_inventory (hero_id, item_def_id)
		VALUES ($1::uuid, $2::uuid) RETURNING id
	`, req.HeroID, chosen.id).Scan(&invID)
	if err != nil {
		return nil, err
	}

	item := &Item{
		ID:          invID,
		Name:        chosen.name,
		Type:        chosen.typ,
		Value:       chosen.value,
		Rarity:      chosen.rarity,
		Description: chosen.description,
	}

	publishEvent(ctx, "inventory.item.dropped", map[string]any{
		"hero_id": req.HeroID,
		"item":    item,
	})

	return item, nil
}

