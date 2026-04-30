package inventory

// inventory-service: manages hero inventory and loot drops.
// Phase 4D in PLAN.md

import "context"

type Item struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`  // weapon | armor | potion
	Value  int    `json:"value"` // damage / defense / heal amount
	Rarity string `json:"rarity"` // common | uncommon | rare
}

type Inventory struct {
	HeroID string `json:"hero_id"`
	Items  []Item `json:"items"`
}

// GetInventory returns the hero's current inventory.
//
//encore:api auth method=GET path=/inventory/:heroID
func GetInventory(ctx context.Context, heroID string) (*Inventory, error) {
	// TODO: Phase 4D
	panic("not implemented")
}

type UseItemRequest struct {
	HeroID string `json:"hero_id"`
	ItemID string `json:"item_id"`
}

// UseItem uses a consumable item (e.g. potion).
//
//encore:api auth method=POST path=/inventory/use
func UseItem(ctx context.Context, req *UseItemRequest) error {
	// TODO: Phase 4D
	panic("not implemented")
}
