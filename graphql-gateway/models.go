package main

import "time"

// Models mirror the Encore service response shapes.
// These are the Go structs that GraphQL resolvers return.

type HeroModel struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Class     string      `json:"class"`
	HP        int         `json:"hp"`
	MaxHP     int         `json:"max_hp"`
	Level     int         `json:"level"`
	XP        int         `json:"xp"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	Inventory []ItemModel `json:"inventory,omitempty"`
}

type RoomModel struct {
	X        int            `json:"x"`
	Y        int            `json:"y"`
	HasChest bool           `json:"has_chest"`
	Exits    []string       `json:"exits"`
	Monsters []MonsterModel `json:"monsters,omitempty"`
}

type FloorModel struct {
	DungeonID string        `json:"dungeon_id"`
	Level     int           `json:"level"`
	Rooms     [][]RoomModel `json:"rooms"`
}

type MonsterModel struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Damage int    `json:"damage"`
	Status string `json:"status"`
}

type ItemModel struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Value  int    `json:"value"`
	Rarity string `json:"rarity"`
}

type CombatResultModel struct {
	HeroDamageDealt   int    `json:"hero_damage_dealt"`
	MonsterDamageBack int    `json:"monster_damage_back"`
	MonsterDied       bool   `json:"monster_died"`
	HeroDied          bool   `json:"hero_died"`
	Message           string `json:"message"`
}

type CombatEventModel struct {
	HeroID    string             `json:"hero_id"`
	MonsterID string             `json:"monster_id"`
	Result    *CombatResultModel `json:"result"`
	Timestamp time.Time          `json:"timestamp"`
}

type RunModel struct {
	HeroID         string `json:"hero_id"`
	HeroName       string `json:"hero_name"`
	FloorsCleared  int    `json:"floors_cleared"`
	MonstersKilled int    `json:"monsters_killed"`
	ItemsFound     int    `json:"items_found"`
	Score          int    `json:"score"`
}
