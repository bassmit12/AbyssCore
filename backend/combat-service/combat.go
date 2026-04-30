package combat

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("combat", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// Monster templates by dungeon floor level
var monsterTemplates = []struct {
	name   string
	hp     int
	damage int
}{
	{"Skeleton", 20, 5},
	{"Goblin", 15, 7},
	{"Orc", 35, 10},
	{"Dark Knight", 60, 15},
	{"Dragon (Mini)", 80, 20},
}

// ─── Types ───────────────────────────────────────────────────────────────────

type Monster struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Damage int    `json:"damage"`
	Status string `json:"status"`
}

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

type SpawnRequest struct {
	FloorID string `json:"floor_id"`
	RoomX   int    `json:"room_x"`
	RoomY   int    `json:"room_y"`
	Level   int    `json:"level"`
}

type GetMonsterResponse struct {
	Monster *Monster `json:"monster"`
}

// ─── API ─────────────────────────────────────────────────────────────────────

// Attack resolves a hero attack on a monster.
// Publishes: combat.attack.initiated, combat.result, combat.monster.killed, game.player.died
//
//encore:api auth method=POST path=/combat/attack
func Attack(ctx context.Context, req *AttackRequest) (*CombatResult, error) {
	// Load monster
	monster, err := getMonster(ctx, req.MonsterID)
	if err != nil {
		return nil, err
	}
	if monster.Status == "dead" {
		return nil, errors.New("monster is already dead")
	}

	// Load hero HP (we need it to check death)
	heroHP, heroMaxHP, heroClass, err := getHeroStats(ctx, req.HeroID)
	if err != nil {
		return nil, err
	}

	metrics.CombatEventsTotal.WithLabelValues("hit").Inc()
	publishEvent(ctx, "combat.attack.initiated", map[string]any{
		"hero_id": req.HeroID, "monster_id": req.MonsterID,
	})

	// Resolve damage
	heroDmg := heroDamage(heroClass)
	monsterDmg := monsterDamage(monster, heroClass)

	// Apply damage
	monster.HP -= heroDmg
	if monster.HP < 0 {
		monster.HP = 0
	}
	heroHP -= monsterDmg
	if heroHP < 0 {
		heroHP = 0
	}

	monsterDied := monster.HP == 0
	heroDied := heroHP == 0

	result := &CombatResult{
		HeroDamageDealt:   heroDmg,
		MonsterDamageBack: monsterDmg,
		MonsterDied:       monsterDied,
		HeroDied:          heroDied,
	}

	// Update monster HP
	status := "alive"
	if monsterDied {
		status = "dead"
	}
	_, err = db.Exec(ctx, `
		UPDATE combat.monsters SET hp = $1, status = $2 WHERE id = $3::uuid
	`, monster.HP, status, req.MonsterID)
	if err != nil {
		return nil, fmt.Errorf("update monster: %w", err)
	}

	// Log combat
	_, _ = db.Exec(ctx, `
		INSERT INTO combat.combat_log
		  (hero_id, monster_id, hero_damage_dealt, monster_damage_back, monster_died, hero_died)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
	`, req.HeroID, req.MonsterID, heroDmg, monsterDmg, monsterDied, heroDied)

	// Build message
	result.Message = buildMessage(monster.Name, heroDmg, monsterDmg, monsterDied, heroDied)

	// Publish outcome events
	publishEvent(ctx, "combat.result", map[string]any{
		"hero_id": req.HeroID, "monster_id": req.MonsterID,
		"hero_hp": heroHP, "hero_max_hp": heroMaxHP,
		"result": result,
	})

	if monsterDied {
		metrics.MonstersKilledTotal.Inc()
		metrics.CombatEventsTotal.WithLabelValues("kill").Inc()
		xpReward := monster.MaxHP / 2
		publishEvent(ctx, "combat.monster.killed", map[string]any{
			"hero_id":    req.HeroID,
			"monster_id": req.MonsterID,
			"xp_reward":  xpReward,
		})
	}

	if heroDied {
		metrics.HeroDeathsTotal.Inc()
		metrics.CombatEventsTotal.WithLabelValues("death").Inc()
		publishEvent(ctx, "game.player.died", map[string]any{
			"hero_id": req.HeroID,
		})
	}

	return result, nil
}

// SpawnMonster seeds a monster in a room. Called by dungeon-service.
//
//encore:api auth method=POST path=/combat/spawn
func SpawnMonster(ctx context.Context, req *SpawnRequest) (*Monster, error) {
	template := monsterTemplates[req.Level%len(monsterTemplates)]
	// Scale HP and damage with floor level
	hp := template.hp + (req.Level-1)*8
	dmg := template.damage + (req.Level-1)*3

	m := &Monster{}
	err := db.QueryRow(ctx, `
		INSERT INTO combat.monsters (floor_id, room_x, room_y, name, hp, max_hp, damage, status)
		VALUES ($1::uuid, $2, $3, $4, $5, $5, $6, 'alive')
		RETURNING id, name, hp, max_hp, damage, status
	`, req.FloorID, req.RoomX, req.RoomY, template.name, hp, dmg).Scan(
		&m.ID, &m.Name, &m.HP, &m.MaxHP, &m.Damage, &m.Status,
	)
	return m, err
}

// GetMonstersByFloor returns all monsters on a floor.
//
//encore:api auth method=GET path=/combat/floor/:floorID/monsters
func GetMonstersByFloor(ctx context.Context, floorID string) (*struct{ Monsters []Monster }, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, hp, max_hp, damage, status
		FROM combat.monsters WHERE floor_id = $1::uuid ORDER BY room_y, room_x
	`, floorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monsters []Monster
	for rows.Next() {
		var m Monster
		if err := rows.Scan(&m.ID, &m.Name, &m.HP, &m.MaxHP, &m.Damage, &m.Status); err != nil {
			return nil, err
		}
		monsters = append(monsters, m)
	}
	return &struct{ Monsters []Monster }{Monsters: monsters}, rows.Err()
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func getMonster(ctx context.Context, id string) (*Monster, error) {
	m := &Monster{}
	err := db.QueryRow(ctx, `
		SELECT id, name, hp, max_hp, damage, status FROM combat.monsters WHERE id = $1::uuid
	`, id).Scan(&m.ID, &m.Name, &m.HP, &m.MaxHP, &m.Damage, &m.Status)
	if err != nil {
		return nil, fmt.Errorf("monster not found: %w", err)
	}
	return m, nil
}

func getHeroStats(ctx context.Context, heroID string) (hp, maxHP int, class string, err error) {
	// In a real impl, call game-service API. Here we do a cross-schema query
	// since we're in single-postgres local dev. In k8s each service has its own DB.
	err = db.QueryRow(ctx, `
		SELECT hp, max_hp, class FROM game.heroes WHERE id = $1::uuid
	`, heroID).Scan(&hp, &maxHP, &class)
	return
}

func heroDamage(class string) int {
	base := map[string]int{"warrior": 12, "rogue": 18, "mage": 22}
	dmg := base[class]
	if dmg == 0 {
		dmg = 10
	}
	// ±20% variance
	variance := dmg / 5
	return dmg - variance + rand.Intn(variance*2+1)
}

func monsterDamage(m *Monster, heroClass string) int {
	// Rogues dodge 20% of the time
	if heroClass == "rogue" && rand.Float32() < 0.2 {
		return 0
	}
	variance := m.Damage / 5
	if variance == 0 {
		variance = 1
	}
	return m.Damage - variance + rand.Intn(variance*2+1)
}

func buildMessage(monsterName string, heroDmg, monsterDmg int, monsterDied, heroDied bool) string {
	msg := fmt.Sprintf("You deal %d damage to %s.", heroDmg, monsterName)
	if monsterDied {
		msg += fmt.Sprintf(" %s is slain!", monsterName)
	} else {
		msg += fmt.Sprintf(" %s strikes back for %d damage.", monsterName, monsterDmg)
	}
	if heroDied {
		msg += " You have died."
	}
	return msg
}

func publishEvent(ctx context.Context, routingKey string, payload map[string]any) {
	// TODO: Phase 5 - RabbitMQ
}
