package combat

// combat-service: resolves attacks, manages monster HP, publishes combat events.
// Phase 4C in PLAN.md

import "context"

type AttackRequest struct {
	HeroID    string `json:"hero_id"`
	MonsterID string `json:"monster_id"`
}

type CombatResult struct {
	HeroDamageDealt   int    `json:"hero_damage_dealt"`
	MonsterDamageBack int    `json:"monster_damage_back"`
	MonsterDied       bool   `json:"monster_died"`
	HeroDied          bool   `json:"hero_died"`
	Message           string `json:"message"`
}

// Attack resolves a hero attack on a monster.
// Publishes: combat.attack.initiated, combat.result, combat.monster.killed (if applicable)
// Also publishes game.player.died if hero HP reaches 0.
//
//encore:api auth method=POST path=/combat/attack
func Attack(ctx context.Context, req *AttackRequest) (*CombatResult, error) {
	// TODO: Phase 4C
	panic("not implemented")
}
