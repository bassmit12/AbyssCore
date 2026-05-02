package combat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"encore.app/game-service"
	"encore.app/inventory-service"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("combat", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// ─── Types ────────────────────────────────────────────────────────────────────

type Intent struct {
	Type  string `json:"type"`  // attack | defend | buff | debuff
	Value int    `json:"value"` // damage / block amount
}

type EncounterMonster struct {
	ID          string            `json:"id"`
	EncounterID string            `json:"encounter_id"`
	Name        string            `json:"name"`
	HP          int               `json:"hp"`
	MaxHP       int               `json:"max_hp"`
	Block       int               `json:"block"`
	Damage      int               `json:"damage"`
	Intents     []Intent          `json:"intents"`
	IntentIndex int               `json:"intent_index"`
	Status      string            `json:"status"`
	Statuses    map[string]int    `json:"statuses"` // vulnerable, weak, etc.
}

type HeroCombatState struct {
	EncounterID string         `json:"encounter_id"`
	HeroID      string         `json:"hero_id"`
	HP          int            `json:"hp"`
	MaxHP       int            `json:"max_hp"`
	Block       int            `json:"block"`
	Energy      int            `json:"energy"`
	MaxEnergy   int            `json:"max_energy"`
	DrawPile    []string       `json:"draw_pile"`
	Hand        []string       `json:"hand"`
	DiscardPile []string       `json:"discard_pile"`
	Statuses    map[string]int `json:"statuses"`
	Monsters    []EncounterMonster `json:"monsters"`
}

// CardResponse is the enriched card shape sent to the gateway.
type CardResponse struct {
	ID     string `json:"id"`
	DefID  string `json:"def_id"`
	Name   string `json:"name"`
	Cost   int    `json:"cost"`
	Type   string `json:"type"`
	Effect string `json:"effect"`
}

// HeroCombatStateResponse is the API-facing hero state with full card objects in hand.
type HeroCombatStateResponse struct {
	EncounterID      string            `json:"encounter_id"`
	HeroID           string            `json:"hero_id"`
	HP               int               `json:"hp"`
	MaxHP            int               `json:"max_hp"`
	Block            int               `json:"block"`
	Energy           int               `json:"energy"`
	MaxEnergy        int               `json:"max_energy"`
	Hand             []CardResponse    `json:"hand"`
	DrawPileCount    int               `json:"draw_pile_count"`
	DiscardPileCount int               `json:"discard_pile_count"`
	Statuses         map[string]int    `json:"statuses"`
}

type FullEncounterState struct {
	EncounterID string                  `json:"encounter_id"`
	HeroState   HeroCombatStateResponse `json:"hero_state"`
	Monsters    []EncounterMonster      `json:"monsters"`
	TurnNumber  int                     `json:"turn_number"`
	Status      string                  `json:"status"`
	Message     string                  `json:"message"`
}

type StartEncounterRequest struct {
	HeroID   string `json:"hero_id"`
	NodeID   string `json:"node_id"`
	NodeType string `json:"node_type"`
}

type PlayCardRequest struct {
	EncounterID string `json:"encounter_id"`
	HeroID      string `json:"hero_id"`
	CardID      string `json:"card_id"`   // hero_decks.id (the card in hand)
	TargetID    string `json:"target_id"` // encounter_monster.id, empty for non-targeted
}

type EndTurnRequest struct {
	EncounterID string `json:"encounter_id"`
	HeroID      string `json:"hero_id"`
}

// ─── Monster templates (by node type + floor scaling) ───────────────────────

type monsterTemplate struct {
	name    string
	hp      int
	damage  int
	intents []Intent
}

var combatTemplates = []monsterTemplate{
	{"Skeleton",    20, 5,  []Intent{{"attack", 5}, {"attack", 5}}},
	{"Goblin",      18, 7,  []Intent{{"attack", 7}, {"defend", 6}, {"attack", 7}}},
	{"Orc",         35, 10, []Intent{{"attack", 10}, {"buff", 2}, {"attack", 12}}},
}
var eliteTemplates = []monsterTemplate{
	{"Dark Knight",   60, 15, []Intent{{"attack", 15}, {"attack", 15}, {"defend", 10}, {"attack", 18}}},
	{"Shadow Wraith", 50, 12, []Intent{{"debuff", 2}, {"attack", 12}, {"attack", 12}}},
}
var bossTemplates = []monsterTemplate{
	{"Dragon (Mini)",  80, 20, []Intent{{"attack", 20}, {"buff", 3}, {"attack", 25}, {"defend", 15}, {"attack", 20}}},
	{"The Collector",  70, 16, []Intent{{"attack", 16}, {"summon", 0}, {"attack", 20}, {"attack", 16}}},
}

// ─── API ──────────────────────────────────────────────────────────────────────

// StartEncounter creates a new encounter for a hero at a map node.
// Spawns monsters based on node type, initialises hero combat state from their deck.
//
//encore:api auth method=POST path=/combat/encounter/start
func StartEncounter(ctx context.Context, req *StartEncounterRequest) (*FullEncounterState, error) {
	// Load hero HP from game schema
	heroHP, heroMaxHP, heroClass, _, _, err := getHeroStats(ctx, req.HeroID)
	if err != nil {
		return nil, fmt.Errorf("load hero: %w", err)
	}

	// Create encounter
	var encID string
	err = db.QueryRow(ctx, `
		INSERT INTO combat.encounters (hero_id, node_id, node_type)
		VALUES ($1::uuid, $2::uuid, $3) RETURNING id
	`, req.HeroID, req.NodeID, req.NodeType).Scan(&encID)
	if err != nil {
		return nil, fmt.Errorf("create encounter: %w", err)
	}

	// Spawn monsters
	templates := pickTemplates(req.NodeType)
	count := monsterCount(req.NodeType)
	monsters := make([]EncounterMonster, 0, count)
	for i := 0; i < count; i++ {
		t := templates[rand.Intn(len(templates))]
		intentsJSON, _ := json.Marshal(t.intents)
		m := EncounterMonster{}
		err := db.QueryRow(ctx, `
			INSERT INTO combat.encounter_monsters
			  (encounter_id, name, hp, max_hp, damage, intents, status)
			VALUES ($1::uuid, $2, $3, $3, $4, $5, 'alive')
			RETURNING id, encounter_id, name, hp, max_hp, block, damage, intents::text, intent_index, status, statuses::text
		`, encID, t.name, t.hp, t.damage, string(intentsJSON)).Scan(
			&m.ID, &m.EncounterID, &m.Name, &m.HP, &m.MaxHP,
			&m.Block, &m.Damage,
			new(string), // intents raw — we'll re-unmarshal below
			&m.IntentIndex, &m.Status, new(string),
		)
		if err != nil {
			return nil, fmt.Errorf("spawn monster: %w", err)
		}
		m.Intents = t.intents
		m.Statuses = map[string]int{}
		monsters = append(monsters, m)
	}

	// Load hero deck for this run → build draw pile
	// Seed starter deck first if the hero has no cards yet
	_ = callService(ctx, "POST", "http://127.0.0.1:4000/deck/seed-starter",
		map[string]string{"hero_id": req.HeroID, "class": heroClass}, &struct{}{})
	deck, err := loadHeroDeck(ctx, req.HeroID)
	if err != nil {
		return nil, fmt.Errorf("load deck: %w", err)
	}
	shuffled := shuffleDeck(deck)

	// Draw opening hand (5 cards)
	hand, draw := drawCards(shuffled, []string{}, 5)

	energyBonus := relicBonus(ctx, req.HeroID, "extra_energy")
	maxEnergy := 3 + energyBonus

	drawJSON, _ := json.Marshal(draw)
	handJSON, _ := json.Marshal(hand)
	discardJSON, _ := json.Marshal([]string{})
	statusesJSON, _ := json.Marshal(map[string]int{})

	// Apply relic bonuses at start of combat
	startBlock := relicBonus(ctx, req.HeroID, "block_start_combat")
	startStrength := relicBonus(ctx, req.HeroID, "strength")
	startDex := relicBonus(ctx, req.HeroID, "dexterity_start")
	startStatuses := map[string]int{}
	if startStrength > 0 {
		startStatuses["strength"] = startStrength
	}
	if startDex > 0 {
		startStatuses["dexterity"] = startDex
	}
	// extra_draw_start relic
	extraDraw := relicBonus(ctx, req.HeroID, "extra_draw_start")
	if extraDraw > 0 {
		hand, draw = drawCards(draw, hand, extraDraw)
		handJSON, _ = json.Marshal(hand)
		drawJSON, _ = json.Marshal(draw)
	}
	if len(startStatuses) > 0 {
		statusesJSON, _ = json.Marshal(startStatuses)
	}

	_, err = db.Exec(ctx, `
		INSERT INTO combat.hero_combat_state
		  (encounter_id, hero_id, hp, max_hp, block, energy, max_energy,
		   draw_pile, hand, discard_pile, statuses)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $6, $7, $8, $9, $10)
	`, encID, req.HeroID, heroHP, heroMaxHP, startBlock, maxEnergy,
		string(drawJSON), string(handJSON), string(discardJSON), string(statusesJSON))
	if err != nil {
		return nil, fmt.Errorf("init hero state: %w", err)
	}

	_ = heroClass // used for display only

	return buildFullState(ctx, encID, "Combat begins!")
}

// PlayCard resolves a card played from the hero's hand.
//
//encore:api auth method=POST path=/combat/encounter/play-card
func PlayCard(ctx context.Context, req *PlayCardRequest) (*FullEncounterState, error) {
	enc, heroState, monsters, err := loadEncounterState(ctx, req.EncounterID)
	if err != nil {
		return nil, err
	}
	if enc.Status != "active" {
		return nil, errors.New("encounter is not active")
	}

	// Find card in hand
	cardIdx := -1
	for i, id := range heroState.Hand {
		if id == req.CardID {
			cardIdx = i
			break
		}
	}
	if cardIdx == -1 {
		return nil, fmt.Errorf("card %s not in hand", req.CardID)
	}

	// Load card definition
	card, err := loadCard(ctx, req.CardID)
	if err != nil {
		return nil, fmt.Errorf("load card: %w", err)
	}

	// Check energy
	cost := card.Cost
	if heroState.Energy < cost {
		return nil, fmt.Errorf("not enough energy (have %d, need %d)", heroState.Energy, cost)
	}

	// Remove card from hand → discard (unless exhaust)
	heroState.Hand = append(heroState.Hand[:cardIdx], heroState.Hand[cardIdx+1:]...)
	heroState.DiscardPile = append(heroState.DiscardPile, req.CardID)
	heroState.Energy -= cost

	msg := fmt.Sprintf("You play %s.", card.Name)

	// ─── Apply card effects ───────────────────────────────────────────────
	var effect map[string]interface{}
	_ = json.Unmarshal([]byte(card.Effect), &effect)

	// Damage
	if dmgRaw, ok := effect["damage"]; ok {
		dmg := int(dmgRaw.(float64))
		// Strength bonus
		dmg += heroState.Statuses["strength"]
		// Weak penalty
		if heroState.Statuses["weak"] > 0 {
			dmg = int(float64(dmg) * 0.75)
		}

		allEnemies, _ := effect["all_enemies"].(bool)
		if allEnemies {
			for i := range monsters {
				if monsters[i].Status == "alive" {
					monsters[i] = dealDamageToMonster(monsters[i], dmg)
				}
			}
			msg += fmt.Sprintf(" Deals %d to all enemies.", dmg)
		} else {
			// Find target
			tIdx := findMonster(monsters, req.TargetID)
			if tIdx == -1 {
				// Auto-target first alive
				for i, m := range monsters {
					if m.Status == "alive" {
						tIdx = i
						break
					}
				}
			}
			if tIdx >= 0 {
				monsters[tIdx] = dealDamageToMonster(monsters[tIdx], dmg)
				msg += fmt.Sprintf(" Deals %d damage to %s.", dmg, monsters[tIdx].Name)
				if monsters[tIdx].Status == "dead" {
					msg += fmt.Sprintf(" %s is slain!", monsters[tIdx].Name)
				}
			}
		}

		// Multi-hit (blade dance style)
		if hitsRaw, ok := effect["hits"]; ok {
			hits := int(hitsRaw.(float64)) - 1
			tIdx := findMonster(monsters, req.TargetID)
			if tIdx == -1 {
				for i, m := range monsters {
					if m.Status == "alive" {
						tIdx = i
						break
					}
				}
			}
			for h := 0; h < hits && tIdx >= 0 && monsters[tIdx].Status == "alive"; h++ {
				monsters[tIdx] = dealDamageToMonster(monsters[tIdx], dmg)
			}
		}
	}

	// Block
	if blockRaw, ok := effect["block"]; ok {
		blk := int(blockRaw.(float64))
		blk += heroState.Statuses["dexterity"]
		heroState.Block += blk
		msg += fmt.Sprintf(" Gain %d Block.", blk)
	}

	// Damage equals block (Bodyslam)
	if _, ok := effect["damage_equals_block"]; ok {
		dmg := heroState.Block + heroState.Statuses["strength"]
		tIdx := findMonster(monsters, req.TargetID)
		if tIdx == -1 {
			for i, m := range monsters {
				if m.Status == "alive" {
					tIdx = i
					break
				}
			}
		}
		if tIdx >= 0 {
			monsters[tIdx] = dealDamageToMonster(monsters[tIdx], dmg)
			msg += fmt.Sprintf(" Deals %d damage (= Block).", dmg)
		}
	}

	// Draw cards
	if drawRaw, ok := effect["draw"]; ok {
		n := int(drawRaw.(float64))
		heroState.Hand, heroState.DrawPile = drawCards(heroState.DrawPile, heroState.Hand, n)
		msg += fmt.Sprintf(" Draw %d.", n)
	}

	// Discard
	if discardRaw, ok := effect["discard"]; ok {
		n := int(discardRaw.(float64))
		if len(heroState.Hand) > 0 {
			for i := 0; i < n && len(heroState.Hand) > 0; i++ {
				last := heroState.Hand[len(heroState.Hand)-1]
				heroState.Hand = heroState.Hand[:len(heroState.Hand)-1]
				heroState.DiscardPile = append(heroState.DiscardPile, last)
			}
		}
	}

	// Apply Vulnerable to target
	if vulnRaw, ok := effect["apply_vulnerable"]; ok {
		stacks := int(vulnRaw.(float64))
		tIdx := findMonster(monsters, req.TargetID)
		if tIdx == -1 {
			for i, m := range monsters {
				if m.Status == "alive" {
					tIdx = i
					break
				}
			}
		}
		if tIdx >= 0 {
			monsters[tIdx].Statuses["vulnerable"] += stacks
			msg += fmt.Sprintf(" Apply %d Vulnerable.", stacks)
		}
	}

	// Apply Weak to target
	if weakRaw, ok := effect["apply_weak"]; ok {
		stacks := int(weakRaw.(float64))
		tIdx := findMonster(monsters, req.TargetID)
		if tIdx >= 0 {
			monsters[tIdx].Statuses["weak"] += stacks
		}
	}

	// Strength this turn (Flex)
	if strRaw, ok := effect["strength_this_turn"]; ok {
		heroState.Statuses["strength"] += int(strRaw.(float64))
		heroState.Statuses["_strength_temp"] += int(strRaw.(float64)) // will remove end of turn
	}

	// Permanent strength (Inflame)
	if strRaw, ok := effect["strength"]; ok {
		heroState.Statuses["strength"] += int(strRaw.(float64))
	}

	// Permanent dexterity (Footwork)
	if dexRaw, ok := effect["dexterity"]; ok {
		heroState.Statuses["dexterity"] += int(dexRaw.(float64))
	}

	// Enemy loses strength this turn (Dark Shackles)
	if strRaw, ok := effect["enemy_strength_this_turn"]; ok {
		delta := int(strRaw.(float64))
		for i := range monsters {
			if monsters[i].Status == "alive" {
				monsters[i].Statuses["strength"] += delta
			}
		}
	}

	// Persist updated state
	if err := saveEncounterState(ctx, req.EncounterID, heroState, monsters); err != nil {
		return nil, err
	}

	// Check win condition
	allDead := true
	for _, m := range monsters {
		if m.Status == "alive" {
			allDead = false
			break
		}
	}
	if allDead {
		_, _ = db.Exec(ctx, `UPDATE combat.encounters SET status='won', updated_at=now() WHERE id=$1::uuid`, req.EncounterID)
		goldReward := 10 + enc.TurnNumber*2
		_ = game.AddGold(ctx, req.HeroID, &game.AddGoldRequest{Gold: goldReward})
		// Roll loot drop
		_, _ = inventory.RollLootDrop(ctx, &inventory.LootDropRequest{HeroID: req.HeroID})
		// Heal from burning blood relic
		if healAmt := relicBonus(ctx, req.HeroID, "heal_after_combat"); healAmt > 0 {
			_ = game.HealHero(ctx, req.HeroID, &game.HealRequest{Amount: healAmt})
		}
		publishEvent(ctx, "combat.encounter.won", map[string]any{
			"hero_id": req.HeroID, "encounter_id": req.EncounterID, "gold_reward": goldReward,
		})
		msg += fmt.Sprintf(" All enemies defeated! +%d gold.", goldReward)
		return buildFullStateWithMsg(ctx, req.EncounterID, msg)
	}

	return buildFullStateWithMsg(ctx, req.EncounterID, msg)
}

// EndTurn discards hand, draws new hand, resets energy, then runs monster turns.
//
//encore:api auth method=POST path=/combat/encounter/end-turn
func EndTurn(ctx context.Context, req *EndTurnRequest) (*FullEncounterState, error) {
	enc, heroState, monsters, err := loadEncounterState(ctx, req.EncounterID)
	if err != nil {
		return nil, err
	}
	if enc.Status != "active" {
		return nil, errors.New("encounter is not active")
	}

	// Remove temp strength (Flex)
	if tmp := heroState.Statuses["_strength_temp"]; tmp > 0 {
		heroState.Statuses["strength"] -= tmp
		delete(heroState.Statuses, "_strength_temp")
	}

	// Discard hand
	heroState.DiscardPile = append(heroState.DiscardPile, heroState.Hand...)
	heroState.Hand = []string{}

	// Reset block
	heroState.Block = 0

	// Tick down hero statuses
	heroState.Statuses = tickStatuses(heroState.Statuses)

	// Restore energy (+ relic bonus)
	heroState.Energy = heroState.MaxEnergy + relicBonus(ctx, req.HeroID, "energy_per_turn")

	// Draw new hand (5 cards); if draw pile runs dry, shuffle discard in mid-draw
	drawCount := 5
	// First pass: draw what's available
	heroState.Hand, heroState.DrawPile = drawCards(heroState.DrawPile, heroState.Hand, drawCount)
	// If we still need more and discard has cards, shuffle discard into draw and continue
	if len(heroState.Hand) < drawCount && len(heroState.DiscardPile) > 0 {
		heroState.DrawPile = shuffleDeck(heroState.DiscardPile)
		heroState.DiscardPile = []string{}
		needed := drawCount - len(heroState.Hand)
		heroState.Hand, heroState.DrawPile = drawCards(heroState.DrawPile, heroState.Hand, needed)
	}

	msg := "Turn ended. "

	// ─── Monster turns ────────────────────────────────────────────────────
	for i := range monsters {
		m := &monsters[i]
		if m.Status != "alive" {
			continue
		}
		if len(m.Intents) == 0 {
			continue
		}
		intent := m.Intents[m.IntentIndex%len(m.Intents)]
		m.IntentIndex = (m.IntentIndex + 1) % len(m.Intents)

		switch intent.Type {
		case "attack":
			dmg := intent.Value + m.Statuses["strength"]
			if m.Statuses["weak"] > 0 {
				dmg = int(float64(dmg) * 0.75)
			}
			// Hero vulnerable
			if heroState.Statuses["vulnerable"] > 0 {
				dmg = int(float64(dmg) * 1.5)
			}
			// Hero dex reduces damage like block (simplified: just block)
			actual := dmg - heroState.Block
			if actual < 0 {
				actual = 0
			}
			heroState.Block -= dmg
			if heroState.Block < 0 {
				heroState.Block = 0
			}
			heroState.HP -= actual
			if heroState.HP < 0 {
				heroState.HP = 0
			}
			msg += fmt.Sprintf("%s attacks for %d. ", m.Name, dmg)

		case "defend":
			m.Block += intent.Value
			msg += fmt.Sprintf("%s gains %d Block. ", m.Name, intent.Value)

		case "buff":
			m.Statuses["strength"] += intent.Value
			msg += fmt.Sprintf("%s gains %d Strength. ", m.Name, intent.Value)

		case "debuff":
			heroState.Statuses["vulnerable"] += intent.Value
			msg += fmt.Sprintf("%s applies %d Vulnerable. ", m.Name, intent.Value)
		}

		// Tick monster statuses
		m.Statuses = tickStatuses(m.Statuses)
		m.Block = 0 // monsters lose block between turns (simplified)
	}

	// Update turn number
	_, _ = db.Exec(ctx, `
		UPDATE combat.encounters SET turn_number=turn_number+1, updated_at=now() WHERE id=$1::uuid
	`, req.EncounterID)

	// Check hero death
	if heroState.HP == 0 {
		_, _ = db.Exec(ctx, `UPDATE combat.encounters SET status='lost', updated_at=now() WHERE id=$1::uuid`, req.EncounterID)
		_, _ = callServiceNoResp(ctx, "POST", fmt.Sprintf("http://127.0.0.1:4000/hero/%s/kill", req.HeroID), nil)
		// Fetch player_id and name for leaderboard event
		_, _, _, heroPlayerID, heroName, _ := getHeroStats(ctx, req.HeroID)
		publishEvent(ctx, "game.player.died", map[string]any{
			"hero_id":   req.HeroID,
			"hero_name": heroName,
			"player_id": heroPlayerID,
		})
		msg += "You have died."
	}

	if err := saveEncounterState(ctx, req.EncounterID, heroState, monsters); err != nil {
		return nil, err
	}

	return buildFullStateWithMsg(ctx, req.EncounterID, msg)
}

// GetEncounter returns the current state of an encounter.
//
//encore:api auth method=GET path=/combat/encounter/:id
func GetEncounter(ctx context.Context, id string) (*FullEncounterState, error) {
	return buildFullState(ctx, id, "")
}

// ─── Legacy endpoints (kept for backward compat during transition) ───────────

type SpawnRequest struct {
	FloorID string `json:"floor_id"`
	RoomX   int    `json:"room_x"`
	RoomY   int    `json:"room_y"`
	Level   int    `json:"level"`
}

type Monster struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Damage int    `json:"damage"`
	Status string `json:"status"`
}

type MonsterListResponse struct{ Monsters []Monster }

// ─── Helpers ──────────────────────────────────────────────────────────────────

func pickTemplates(nodeType string) []monsterTemplate {
	switch nodeType {
	case "elite":
		return eliteTemplates
	case "boss":
		return bossTemplates
	default:
		return combatTemplates
	}
}

func monsterCount(nodeType string) int {
	if nodeType == "boss" {
		return 1
	}
	if nodeType == "elite" {
		return 1
	}
	return 1 + rand.Intn(2) // 1 or 2
}

func dealDamageToMonster(m EncounterMonster, dmg int) EncounterMonster {
	if m.Statuses["vulnerable"] > 0 {
		dmg = int(float64(dmg) * 1.5)
	}
	actual := dmg - m.Block
	if actual < 0 {
		actual = 0
	}
	m.Block -= dmg
	if m.Block < 0 {
		m.Block = 0
	}
	m.HP -= actual
	if m.HP < 0 {
		m.HP = 0
	}
	if m.HP == 0 {
		m.Status = "dead"
	}
	return m
}

func findMonster(monsters []EncounterMonster, targetID string) int {
	for i, m := range monsters {
		if m.ID == targetID {
			return i
		}
	}
	return -1
}

func tickStatuses(s map[string]int) map[string]int {
	out := map[string]int{}
	for k, v := range s {
		if k == "strength" || k == "dexterity" {
			out[k] = v // permanent
		} else if v > 1 {
			out[k] = v - 1
		}
		// 0 or 1 → expired
	}
	return out
}

func shuffleDeck(deck []string) []string {
	out := make([]string, len(deck))
	copy(out, deck)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func drawCards(drawPile, hand []string, n int) (newHand, newDraw []string) {
	newHand = make([]string, len(hand))
	copy(newHand, hand)
	newDraw = make([]string, len(drawPile))
	copy(newDraw, drawPile)
	for i := 0; i < n && len(newDraw) > 0; i++ {
		newHand = append(newHand, newDraw[0])
		newDraw = newDraw[1:]
	}
	return newHand, newDraw
}

type cardInfo struct {
	ID     string
	DefID  string
	Name   string
	Cost   int
	Effect string
	Type   string
}

func loadCard(ctx context.Context, cardID string) (*cardInfo, error) {
	// Call deck-service for card info
	var result struct {
		ID          string `json:"id"`
		DefID       string `json:"def_id"`
		Name        string `json:"name"`
		Cost        int    `json:"cost"`
		Effect      string `json:"effect"`
		Type        string `json:"type"`
	}
	err := callService(ctx, "GET", fmt.Sprintf("http://127.0.0.1:4000/deck/card/%s", cardID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("load card from deck-service: %w", err)
	}
	return &cardInfo{
		ID: result.ID, DefID: result.DefID, Name: result.Name,
		Cost: result.Cost, Effect: result.Effect, Type: result.Type,
	}, nil
}

func loadHeroDeck(ctx context.Context, heroID string) ([]string, error) {
	// Call deck-service to get hero deck card IDs
	var result struct {
		HeroID string `json:"hero_id"`
		Cards  []struct {
			ID string `json:"id"`
		} `json:"cards"`
	}
	err := callService(ctx, "GET", fmt.Sprintf("http://127.0.0.1:4000/deck/hero/%s", heroID), nil, &result)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(result.Cards))
	for _, c := range result.Cards {
		ids = append(ids, c.ID)
	}
	// If deck is empty (new hero), give starter deck
	if len(ids) == 0 {
		return []string{}, nil
	}
	return ids, nil
}

func getHeroStats(ctx context.Context, heroID string) (hp, maxHP int, class, playerID, name string, err error) {
	// Call game-service API — no cross-schema queries
	var result struct {
		HP       int    `json:"hp"`
		MaxHP    int    `json:"max_hp"`
		Class    string `json:"class"`
		PlayerID string `json:"player_id"`
		Name     string `json:"name"`
	}
	err = callService(ctx, "GET", fmt.Sprintf("http://127.0.0.1:4000/hero/%s/stats", heroID), nil, &result)
	if err != nil {
		return 0, 0, "", "", "", err
	}
	return result.HP, result.MaxHP, result.Class, result.PlayerID, result.Name, nil
}

func relicBonus(ctx context.Context, heroID, effectKey string) int {
	// Call deck-service to get relic bonus value
	var result struct {
		Value int `json:"value"`
	}
	url := fmt.Sprintf("http://127.0.0.1:4000/deck/relic-bonus/%s/%s", heroID, effectKey)
	if err := callService(ctx, "GET", url, nil, &result); err != nil {
		return 0
	}
	return result.Value
}

type encounterRow struct {
	ID         string
	HeroID     string
	NodeID     string
	NodeType   string
	Status     string
	TurnNumber int
}

func loadEncounterState(ctx context.Context, encID string) (*encounterRow, *HeroCombatState, []EncounterMonster, error) {
	// Load encounter
	enc := &encounterRow{}
	err := db.QueryRow(ctx, `
		SELECT id, hero_id, node_id, node_type, status, turn_number
		FROM combat.encounters WHERE id = $1::uuid
	`, encID).Scan(&enc.ID, &enc.HeroID, &enc.NodeID, &enc.NodeType, &enc.Status, &enc.TurnNumber)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("encounter not found: %w", err)
	}

	// Load hero state
	hs := &HeroCombatState{}
	var drawRaw, handRaw, discardRaw, statusRaw string
	err = db.QueryRow(ctx, `
		SELECT encounter_id, hero_id, hp, max_hp, block, energy, max_energy,
		       draw_pile::text, hand::text, discard_pile::text, statuses::text
		FROM combat.hero_combat_state WHERE encounter_id = $1::uuid
	`, encID).Scan(&hs.EncounterID, &hs.HeroID, &hs.HP, &hs.MaxHP, &hs.Block, &hs.Energy, &hs.MaxEnergy,
		&drawRaw, &handRaw, &discardRaw, &statusRaw)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("hero state not found: %w", err)
	}
	_ = json.Unmarshal([]byte(drawRaw), &hs.DrawPile)
	_ = json.Unmarshal([]byte(handRaw), &hs.Hand)
	_ = json.Unmarshal([]byte(discardRaw), &hs.DiscardPile)
	_ = json.Unmarshal([]byte(statusRaw), &hs.Statuses)
	if hs.DrawPile == nil {
		hs.DrawPile = []string{}
	}
	if hs.Hand == nil {
		hs.Hand = []string{}
	}
	if hs.DiscardPile == nil {
		hs.DiscardPile = []string{}
	}
	if hs.Statuses == nil {
		hs.Statuses = map[string]int{}
	}

	// Load monsters
	rows, err := db.Query(ctx, `
		SELECT id, encounter_id, name, hp, max_hp, block, damage,
		       intents::text, intent_index, status, statuses::text
		FROM combat.encounter_monsters WHERE encounter_id = $1::uuid ORDER BY id
	`, encID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	var monsters []EncounterMonster
	for rows.Next() {
		m := EncounterMonster{}
		var intentsRaw, statusesRaw string
		if err := rows.Scan(&m.ID, &m.EncounterID, &m.Name, &m.HP, &m.MaxHP,
			&m.Block, &m.Damage, &intentsRaw, &m.IntentIndex, &m.Status, &statusesRaw); err != nil {
			return nil, nil, nil, err
		}
		_ = json.Unmarshal([]byte(intentsRaw), &m.Intents)
		_ = json.Unmarshal([]byte(statusesRaw), &m.Statuses)
		if m.Statuses == nil {
			m.Statuses = map[string]int{}
		}
		monsters = append(monsters, m)
	}
	return enc, hs, monsters, rows.Err()
}

func saveEncounterState(ctx context.Context, encID string, hs *HeroCombatState, monsters []EncounterMonster) error {
	drawJSON, _ := json.Marshal(hs.DrawPile)
	handJSON, _ := json.Marshal(hs.Hand)
	discardJSON, _ := json.Marshal(hs.DiscardPile)
	statusJSON, _ := json.Marshal(hs.Statuses)

	_, err := db.Exec(ctx, `
		UPDATE combat.hero_combat_state
		SET hp=$1, block=$2, energy=$3, draw_pile=$4, hand=$5, discard_pile=$6, statuses=$7, updated_at=now()
		WHERE encounter_id=$8::uuid
	`, hs.HP, hs.Block, hs.Energy,
		string(drawJSON), string(handJSON), string(discardJSON), string(statusJSON),
		encID)
	if err != nil {
		return fmt.Errorf("save hero state: %w", err)
	}

	for _, m := range monsters {
		intentsJSON, _ := json.Marshal(m.Intents)
		statusesJSON, _ := json.Marshal(m.Statuses)
		_, err = db.Exec(ctx, `
			UPDATE combat.encounter_monsters
			SET hp=$1, block=$2, status=$3, intents=$4, intent_index=$5, statuses=$6
			WHERE id=$7::uuid
		`, m.HP, m.Block, m.Status, string(intentsJSON), m.IntentIndex, string(statusesJSON), m.ID)
		if err != nil {
			return fmt.Errorf("save monster: %w", err)
		}
	}
	return nil
}

func buildFullState(ctx context.Context, encID string, msg string) (*FullEncounterState, error) {
	return buildFullStateWithMsg(ctx, encID, msg)
}

// ─── HTTP helpers (inter-service calls within Encore) ────────────────────────

func callService(ctx context.Context, method, url string, body any, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("service returned %d: %s", resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func callServiceNoResp(ctx context.Context, method, url string, body any) (int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

func buildFullStateWithMsg(ctx context.Context, encID string, msg string) (*FullEncounterState, error) {
	enc, hs, monsters, err := loadEncounterState(ctx, encID)
	if err != nil {
		return nil, err
	}
	hs.Monsters = monsters

	// Hydrate hand: resolve each card ID to a full CardResponse
	hand := make([]CardResponse, 0, len(hs.Hand))
	for _, cardID := range hs.Hand {
		c, err := loadCard(ctx, cardID)
		if err != nil {
			// If we can't load a card just skip it rather than failing the whole request
			continue
		}
		hand = append(hand, CardResponse{
			ID: c.ID, DefID: c.DefID, Name: c.Name,
			Cost: c.Cost, Type: c.Type, Effect: c.Effect,
		})
	}

	heroResp := HeroCombatStateResponse{
		EncounterID:      hs.EncounterID,
		HeroID:           hs.HeroID,
		HP:               hs.HP,
		MaxHP:            hs.MaxHP,
		Block:            hs.Block,
		Energy:           hs.Energy,
		MaxEnergy:        hs.MaxEnergy,
		Hand:             hand,
		DrawPileCount:    len(hs.DrawPile),
		DiscardPileCount: len(hs.DiscardPile),
		Statuses:         hs.Statuses,
	}
	return &FullEncounterState{
		EncounterID: encID,
		HeroState:   heroResp,
		Monsters:    monsters,
		TurnNumber:  enc.TurnNumber,
		Status:      enc.Status,
		Message:     msg,
	}, nil
}
