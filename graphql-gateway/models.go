package main

import "time"

// Models mirror the Encore service response shapes.

// ─── Hero ─────────────────────────────────────────────────────────────────────

type HeroModel struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Class  string `json:"class"`
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Level  int    `json:"level"`
	XP     int    `json:"xp"`
	Gold   int    `json:"gold"`
	Alive  bool   `json:"alive"`
	RunID  string `json:"run_id,omitempty"`
}

// ─── Map / Run ────────────────────────────────────────────────────────────────

type MapNodeModel struct {
	ID       string `json:"id"`
	FloorID  string `json:"floor_id"`
	Floor    int    `json:"col"`    // map col → floor field for resolver
	Position int    `json:"row"`    // map row → position field for resolver
	Type     string `json:"type"`
	Visited  bool   `json:"cleared"` // cleared ≈ visited
	Available bool  `json:"available,omitempty"`
}

type MapEdgeModel struct {
	FromNodeID string `json:"from_node_id"`
	ToNodeID   string `json:"to_node_id"`
}

type FloorGraphModel struct {
	RunID         string         `json:"run_id"`
	HeroID        string         `json:"hero_id"`
	CurrentNodeID string         `json:"current_node_id,omitempty"`
	Nodes         []MapNodeModel `json:"nodes"`
	Edges         []MapEdgeModel `json:"edges"`
}

type RunModel struct {
	ID             string    `json:"id"`
	HeroID         string    `json:"hero_id"`
	HeroName       string    `json:"hero_name"`
	FloorsCleared  int       `json:"floors_cleared"`
	MonstersKilled int       `json:"monsters_killed"`
	Score          int       `json:"score"`
	CreatedAt      time.Time `json:"created_at"`
}

// ─── Cards ────────────────────────────────────────────────────────────────────

type CardModel struct {
	ID     string `json:"id"`
	DefID  string `json:"def_id"`
	Name   string `json:"name"`
	Cost   int    `json:"cost"`
	Type   string `json:"type"`
	Effect string `json:"effect"`
}

type CardDefinitionModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Type        string `json:"type"`
	Cost        int    `json:"cost"`
	Effect      string `json:"effect"`
	Rarity      string `json:"rarity"`
	Description string `json:"description"`
}

type HeroDeckModel struct {
	HeroID string      `json:"hero_id"`
	Cards  []CardModel `json:"cards"`
}

// ─── Relics ───────────────────────────────────────────────────────────────────

type RelicModel struct {
	ID          string `json:"id"`
	DefID       string `json:"def_id"`
	Name        string `json:"name"`
	Rarity      string `json:"rarity"`
	Description string `json:"description"`
}

// ─── Encounter ────────────────────────────────────────────────────────────────

type MonsterIntentModel struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type EncounterMonsterModel struct {
	ID      string               `json:"id"`
	Name    string               `json:"name"`
	HP      int                  `json:"hp"`
	MaxHP   int                  `json:"max_hp"`
	Block   int                  `json:"block"`
	Status  string               `json:"status"`
	Intents []MonsterIntentModel `json:"intents"`
}

type StatusEffectModel struct {
	Name   string `json:"name"`
	Stacks int    `json:"stacks"`
}

type HeroCombatStateModel struct {
	HP               int            `json:"hp"`
	MaxHP            int            `json:"max_hp"`
	Block            int            `json:"block"`
	Energy           int            `json:"energy"`
	MaxEnergy        int            `json:"max_energy"`
	Hand             []CardModel    `json:"hand"`
	DrawPileCount    int            `json:"draw_pile_count"`
	DiscardPileCount int            `json:"discard_pile_count"`
	// Backend sends {"strength":2,"weak":1} — a map, not a slice.
	Statuses         map[string]int `json:"statuses"`
}

type EncounterStateModel struct {
	EncounterID string                 `json:"encounter_id"`
	HeroState   HeroCombatStateModel   `json:"hero_state"`
	Monsters    []EncounterMonsterModel `json:"monsters"`
	TurnNumber  int                    `json:"turn_number"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
}

type CardRewardsModel struct {
	EncounterID string               `json:"encounter_id"`
	Cards       []CardDefinitionModel `json:"cards"`
}

// ─── Shop ─────────────────────────────────────────────────────────────────────

type ShopItemModel struct {
	CardDef CardDefinitionModel `json:"card_def"`
	Price   int                 `json:"price"`
}

type ShopInventoryModel struct {
	Items []ShopItemModel `json:"items"`
}

// ─── Inventory ─────────────────────────────────────────────────────────────────

type InventoryItemModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       int    `json:"value"`
	Rarity      string `json:"rarity"`
	Description string `json:"description"`
}

type HeroInventoryModel struct {
	HeroID string               `json:"hero_id"`
	Items  []InventoryItemModel `json:"items"`
}

// ─── Events ───────────────────────────────────────────────────────────────────

type EventChoiceModel struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type RandomEventModel struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Choices     []EventChoiceModel `json:"choices"`
}

type EventOutcomeModel struct {
	ChoiceID    string `json:"choice_id"`
	Description string `json:"description"`
	GoldDelta   int    `json:"gold_delta"`
	HPDelta     int    `json:"hp_delta"`
}
