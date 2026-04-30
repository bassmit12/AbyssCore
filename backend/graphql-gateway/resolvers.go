package main

import (
	"context"
	"fmt"
	"net/http"
)

// Resolver holds all GraphQL resolvers.
// Each resolver calls the appropriate Encore service via HTTP.
type Resolver struct{}

// ─── Query Resolvers ─────────────────────────────────────────────────────────

func (r *Resolver) Hero(ctx context.Context, id string) (*HeroModel, error) {
	return encoreCall[HeroModel](ctx, http.MethodGet, fmt.Sprintf("/hero/%s", id), nil)
}

func (r *Resolver) Floor(ctx context.Context, dungeonID, level string) (*FloorModel, error) {
	return encoreCall[FloorModel](ctx, http.MethodGet, fmt.Sprintf("/dungeon/%s/floor/%s", dungeonID, level), nil)
}

func (r *Resolver) Inventory(ctx context.Context, heroID string) ([]ItemModel, error) {
	type resp struct {
		Items []ItemModel `json:"items"`
	}
	result, err := encoreCall[resp](ctx, http.MethodGet, fmt.Sprintf("/inventory/%s", heroID), nil)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (r *Resolver) Leaderboard(ctx context.Context) ([]RunModel, error) {
	type resp struct {
		Runs []RunModel `json:"runs"`
	}
	result, err := encoreCall[resp](ctx, http.MethodGet, "/leaderboard", nil)
	if err != nil {
		return nil, err
	}
	return result.Runs, nil
}

// ─── Mutation Resolvers ───────────────────────────────────────────────────────

func (r *Resolver) CreateHero(ctx context.Context, name, class string) (*HeroModel, error) {
	return encoreCall[HeroModel](ctx, http.MethodPost, "/hero", map[string]string{
		"name":  name,
		"class": class,
	})
}

func (r *Resolver) StartDungeon(ctx context.Context, heroID string) (*FloorModel, error) {
	return encoreCall[FloorModel](ctx, http.MethodPost, "/dungeon/start", map[string]string{
		"hero_id": heroID,
	})
}

func (r *Resolver) MoveHero(ctx context.Context, heroID, direction string) (*HeroModel, error) {
	return encoreCall[HeroModel](ctx, http.MethodPost, fmt.Sprintf("/hero/%s/move", heroID), map[string]string{
		"direction": direction,
	})
}

func (r *Resolver) Attack(ctx context.Context, heroID, monsterID string) (*CombatResultModel, error) {
	return encoreCall[CombatResultModel](ctx, http.MethodPost, "/combat/attack", map[string]string{
		"hero_id":    heroID,
		"monster_id": monsterID,
	})
}

func (r *Resolver) UseItem(ctx context.Context, heroID, itemID string) (*HeroModel, error) {
	return encoreCall[HeroModel](ctx, http.MethodPost, "/inventory/use", map[string]string{
		"hero_id": heroID,
		"item_id": itemID,
	})
}

// ─── Subscription Resolvers ───────────────────────────────────────────────────
// Subscriptions receive events forwarded from RabbitMQ via an internal SSE/channel bridge.
// Implemented in Phase 5 when RabbitMQ consumers are wired.

func (r *Resolver) CombatEvents(ctx context.Context, heroID string) (<-chan *CombatEventModel, error) {
	// TODO: Phase 5 - subscribe to RabbitMQ combat.result events for this heroID
	ch := make(chan *CombatEventModel)
	return ch, nil
}

func (r *Resolver) HeroUpdated(ctx context.Context, heroID string) (<-chan *HeroModel, error) {
	// TODO: Phase 5 - subscribe to game.hero.updated events
	ch := make(chan *HeroModel)
	return ch, nil
}
