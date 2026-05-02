package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/graph-gophers/graphql-go"
)

// Resolver holds all GraphQL resolvers.
type Resolver struct{}

// ─── Query: me ───────────────────────────────────────────────────────────────

type UserResolver struct {
	id    string
	email string
}

func (u *UserResolver) ID() graphql.ID  { return graphql.ID(u.id) }
func (u *UserResolver) Email() string   { return u.email }
func (u *UserResolver) Roles() []string { return []string{} }

func (r *Resolver) Me(ctx context.Context) (*UserResolver, error) {
	return nil, nil
}

// ─── Query: hero ─────────────────────────────────────────────────────────────

func (r *Resolver) Hero(ctx context.Context, args struct{ ID graphql.ID }) (*HeroResolver, error) {
	m, err := encoreCall[HeroModel](ctx, http.MethodGet, fmt.Sprintf("/hero/%s", args.ID), nil)
	if err != nil {
		return nil, err
	}
	return &HeroResolver{m}, nil
}

// ─── Query: floorGraph ───────────────────────────────────────────────────────

func (r *Resolver) FloorGraph(ctx context.Context, args struct{ HeroId graphql.ID }) (*FloorGraphResolver, error) {
	m, err := encoreCall[FloorGraphModel](ctx, http.MethodGet, fmt.Sprintf("/map/hero/%s/graph", args.HeroId), nil)
	if err != nil {
		return nil, err
	}
	return &FloorGraphResolver{m}, nil
}

// ─── Query: heroDeck ─────────────────────────────────────────────────────────

func (r *Resolver) HeroDeck(ctx context.Context, args struct{ HeroId graphql.ID }) (*HeroDeckResolver, error) {
	m, err := encoreCall[HeroDeckModel](ctx, http.MethodGet, fmt.Sprintf("/deck/hero/%s", args.HeroId), nil)
	if err != nil {
		return nil, err
	}
	return &HeroDeckResolver{m}, nil
}

// ─── Query: heroRelics ───────────────────────────────────────────────────────

func (r *Resolver) HeroRelics(ctx context.Context, args struct{ HeroId graphql.ID }) ([]*RelicResolver, error) {
	type resp struct {
		Relics []RelicModel `json:"relics"`
	}
	result, err := encoreCall[resp](ctx, http.MethodGet, fmt.Sprintf("/deck/hero/%s/relics", args.HeroId), nil)
	if err != nil {
		return nil, err
	}
	out := make([]*RelicResolver, len(result.Relics))
	for i := range result.Relics {
		out[i] = &RelicResolver{&result.Relics[i]}
	}
	return out, nil
}

// ─── Query: encounterState ───────────────────────────────────────────────────

func (r *Resolver) EncounterState(ctx context.Context, args struct{ EncounterId graphql.ID }) (*EncounterStateResolver, error) {
	m, err := encoreCall[EncounterStateModel](ctx, http.MethodGet, fmt.Sprintf("/combat/encounter/%s", args.EncounterId), nil)
	if err != nil {
		return nil, err
	}
	return &EncounterStateResolver{m}, nil
}

// ─── Query: cardRewards ──────────────────────────────────────────────────────

func (r *Resolver) CardRewards(ctx context.Context, args struct{ HeroId graphql.ID }) (*CardRewardsResolver, error) {
	// Look up hero class so we can filter cards
	type heroResp struct {
		Class string `json:"class"`
	}
	hero, err := encoreCall[heroResp](ctx, http.MethodGet, fmt.Sprintf("/hero/%s", args.HeroId), nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch hero: %w", err)
	}
	m, err := encoreCall[CardRewardsModel](ctx, http.MethodGet, fmt.Sprintf("/deck/rewards/%s", hero.Class), nil)
	if err != nil {
		return nil, err
	}
	return &CardRewardsResolver{m}, nil
}

// ─── Query: leaderboard ──────────────────────────────────────────────────────

func (r *Resolver) Leaderboard(ctx context.Context) ([]*RunResolver, error) {
	type resp struct {
		Runs []RunModel `json:"runs"`
	}
	result, err := encoreCall[resp](ctx, http.MethodGet, "/leaderboard", nil)
	if err != nil {
		return nil, err
	}
	out := make([]*RunResolver, len(result.Runs))
	for i := range result.Runs {
		out[i] = &RunResolver{&result.Runs[i]}
	}
	return out, nil
}

// ─── Mutation: createHero ────────────────────────────────────────────────────

func (r *Resolver) CreateHero(ctx context.Context, args struct {
	Name  string
	Class string
}) (*HeroResolver, error) {
	m, err := encoreCall[HeroModel](ctx, http.MethodPost, "/hero", map[string]string{
		"name":  args.Name,
		"class": strings.ToLower(args.Class),
	})
	if err != nil {
		return nil, err
	}
	return &HeroResolver{m}, nil
}

// ─── Mutation: startRun ──────────────────────────────────────────────────────

func (r *Resolver) StartRun(ctx context.Context, args struct{ HeroId graphql.ID }) (*FloorGraphResolver, error) {
	m, err := encoreCall[FloorGraphModel](ctx, http.MethodPost, "/map/runs", map[string]string{
		"hero_id": string(args.HeroId),
	})
	if err != nil {
		return nil, err
	}
	return &FloorGraphResolver{m}, nil
}

// ─── Mutation: travelToNode ──────────────────────────────────────────────────

func (r *Resolver) TravelToNode(ctx context.Context, args struct {
	HeroId graphql.ID
	NodeId graphql.ID
}) (*FloorGraphResolver, error) {
	m, err := encoreCall[FloorGraphModel](ctx, http.MethodPost, "/map/travel", map[string]string{
		"hero_id": string(args.HeroId),
		"node_id": string(args.NodeId),
	})
	if err != nil {
		return nil, err
	}
	return &FloorGraphResolver{m}, nil
}

// ─── Mutation: startEncounter ────────────────────────────────────────────────

func (r *Resolver) StartEncounter(ctx context.Context, args struct {
	HeroId graphql.ID
	NodeId graphql.ID
}) (*EncounterStateResolver, error) {
	m, err := encoreCall[EncounterStateModel](ctx, http.MethodPost, "/combat/encounter/start", map[string]string{
		"hero_id": string(args.HeroId),
		"node_id": string(args.NodeId),
	})
	if err != nil {
		return nil, err
	}
	return &EncounterStateResolver{m}, nil
}

// ─── Mutation: playCard ──────────────────────────────────────────────────────

func (r *Resolver) PlayCard(ctx context.Context, args struct {
	EncounterId graphql.ID
	HeroId      graphql.ID
	CardId      graphql.ID
	TargetId    *graphql.ID
}) (*EncounterStateResolver, error) {
	body := map[string]string{
		"encounter_id": string(args.EncounterId),
		"hero_id":      string(args.HeroId),
		"card_id":      string(args.CardId),
	}
	if args.TargetId != nil {
		body["target_id"] = string(*args.TargetId)
	}
	m, err := encoreCall[EncounterStateModel](ctx, http.MethodPost, "/combat/encounter/play-card", body)
	if err != nil {
		return nil, err
	}
	return &EncounterStateResolver{m}, nil
}

// ─── Mutation: endTurn ───────────────────────────────────────────────────────

func (r *Resolver) EndTurn(ctx context.Context, args struct {
	EncounterId graphql.ID
	HeroId      graphql.ID
}) (*EncounterStateResolver, error) {
	m, err := encoreCall[EncounterStateModel](ctx, http.MethodPost, "/combat/encounter/end-turn", map[string]string{
		"encounter_id": string(args.EncounterId),
		"hero_id":      string(args.HeroId),
	})
	if err != nil {
		return nil, err
	}
	return &EncounterStateResolver{m}, nil
}

// ─── Mutation: pickCardReward ────────────────────────────────────────────────

func (r *Resolver) PickCardReward(ctx context.Context, args struct {
	EncounterId graphql.ID
	HeroId      graphql.ID
	CardDefId   graphql.ID
}) (*HeroDeckResolver, error) {
	// Add the chosen card to the hero's deck
	_, err := encoreCall[map[string]interface{}](ctx, http.MethodPost, "/deck/add", map[string]string{
		"hero_id":     string(args.HeroId),
		"card_def_id": string(args.CardDefId),
	})
	if err != nil {
		return nil, err
	}
	// Return updated deck
	m, err := encoreCall[HeroDeckModel](ctx, http.MethodGet, fmt.Sprintf("/deck/hero/%s", args.HeroId), nil)
	if err != nil {
		return nil, err
	}
	return &HeroDeckResolver{m}, nil
}

// ─── Mutation: skipCardReward ────────────────────────────────────────────────

func (r *Resolver) SkipCardReward(ctx context.Context, args struct {
	EncounterId graphql.ID
	HeroId      graphql.ID
}) (bool, error) {
	// Skip is a no-op — just return true
	return true, nil
}

// ─── Mutation: rest ──────────────────────────────────────────────────────────

func (r *Resolver) Rest(ctx context.Context, args struct {
	HeroId graphql.ID
	NodeId graphql.ID
	Action string
}) (*HeroResolver, error) {
	// Heal hero for 30% of max HP at rest sites
	_, _ = encoreCall[map[string]interface{}](ctx, http.MethodPost, fmt.Sprintf("/hero/%s/heal", args.HeroId), map[string]interface{}{
		"amount": 30,
	})
	// Mark node cleared
	_, _ = encoreCall[map[string]interface{}](ctx, http.MethodPost, "/map/clear", map[string]string{
		"hero_id": string(args.HeroId),
		"node_id": string(args.NodeId),
	})
	// Return updated hero
	m, err := encoreCall[HeroModel](ctx, http.MethodGet, fmt.Sprintf("/hero/%s", args.HeroId), nil)
	if err != nil {
		return nil, err
	}
	return &HeroResolver{m}, nil
}

// ─── Mutation: shopCards ─────────────────────────────────────────────────────

func (r *Resolver) ShopCards(ctx context.Context, args struct{ HeroId graphql.ID }) (*ShopInventoryResolver, error) {
	type heroResp struct {
		Class string `json:"class"`
	}
	hero, err := encoreCall[heroResp](ctx, http.MethodGet, fmt.Sprintf("/hero/%s", args.HeroId), nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch hero: %w", err)
	}
	m, err := encoreCall[ShopInventoryModel](ctx, http.MethodGet, fmt.Sprintf("/deck/shop/%s", hero.Class), nil)
	if err != nil {
		return nil, err
	}
	return &ShopInventoryResolver{m}, nil
}

// ─── Mutation: buyCard ───────────────────────────────────────────────────────

func (r *Resolver) BuyCard(ctx context.Context, args struct {
	HeroId    graphql.ID
	CardDefId graphql.ID
	Price     int32
}) (*HeroResolver, error) {
	// Deduct gold first
	_, err := encoreCall[map[string]interface{}](ctx, http.MethodPost, fmt.Sprintf("/hero/%s/deduct-gold", args.HeroId), map[string]interface{}{
		"gold": int(args.Price),
	})
	if err != nil {
		return nil, fmt.Errorf("not enough gold: %w", err)
	}
	// Add card to deck
	_, err = encoreCall[map[string]interface{}](ctx, http.MethodPost, "/deck/shop/buy", map[string]string{
		"hero_id":     string(args.HeroId),
		"card_def_id": string(args.CardDefId),
	})
	if err != nil {
		return nil, fmt.Errorf("could not add card: %w", err)
	}
	// Return updated hero
	m, err := encoreCall[HeroModel](ctx, http.MethodGet, fmt.Sprintf("/hero/%s", args.HeroId), nil)
	if err != nil {
		return nil, err
	}
	return &HeroResolver{m}, nil
}

// ─── Mutation: submitScore ───────────────────────────────────────────────────

func (r *Resolver) SubmitScore(ctx context.Context, args struct{ HeroId graphql.ID }) (*RunResolver, error) {
	m, err := encoreCall[RunModel](ctx, http.MethodPost, "/leaderboard/submit", map[string]string{
		"hero_id": string(args.HeroId),
	})
	if err != nil {
		return nil, err
	}
	return &RunResolver{m}, nil
}

// ─── Subscriptions (stub) ────────────────────────────────────────────────────

func (r *Resolver) EncounterUpdated(ctx context.Context, args struct{ EncounterId graphql.ID }) (<-chan *EncounterStateResolver, error) {
	ch := make(chan *EncounterStateResolver)
	go func() { <-ctx.Done(); close(ch) }()
	return ch, nil
}

func (r *Resolver) HeroUpdated(ctx context.Context, args struct{ HeroId graphql.ID }) (<-chan *HeroResolver, error) {
	ch := make(chan *HeroResolver)
	go func() { <-ctx.Done(); close(ch) }()
	return ch, nil
}

// ─── Type Resolvers ───────────────────────────────────────────────────────────

// Hero
type HeroResolver struct{ m *HeroModel }

func (h *HeroResolver) ID() graphql.ID { return graphql.ID(h.m.ID) }
func (h *HeroResolver) Name() string   { return h.m.Name }
func (h *HeroResolver) Class() string  { return strings.ToUpper(h.m.Class) }
func (h *HeroResolver) Hp() int32      { return int32(h.m.HP) }
func (h *HeroResolver) MaxHp() int32   { return int32(h.m.MaxHP) }
func (h *HeroResolver) Level() int32   { return int32(h.m.Level) }
func (h *HeroResolver) Xp() int32      { return int32(h.m.XP) }
func (h *HeroResolver) Gold() int32    { return int32(h.m.Gold) }
func (h *HeroResolver) Alive() bool    { return h.m.Alive }
func (h *HeroResolver) RunId() *graphql.ID {
	if h.m.RunID == "" {
		return nil
	}
	id := graphql.ID(h.m.RunID)
	return &id
}

// FloorGraph
type FloorGraphResolver struct{ m *FloorGraphModel }

func (f *FloorGraphResolver) RunId() graphql.ID { return graphql.ID(f.m.RunID) }
func (f *FloorGraphResolver) HeroId() graphql.ID { return graphql.ID(f.m.HeroID) }
func (f *FloorGraphResolver) CurrentNodeId() *graphql.ID {
	if f.m.CurrentNodeID == "" {
		return nil
	}
	id := graphql.ID(f.m.CurrentNodeID)
	return &id
}
func (f *FloorGraphResolver) Nodes() []*MapNodeResolver {
	out := make([]*MapNodeResolver, len(f.m.Nodes))
	for i := range f.m.Nodes {
		out[i] = &MapNodeResolver{&f.m.Nodes[i]}
	}
	return out
}
func (f *FloorGraphResolver) Edges() []*MapEdgeResolver {
	out := make([]*MapEdgeResolver, len(f.m.Edges))
	for i := range f.m.Edges {
		out[i] = &MapEdgeResolver{&f.m.Edges[i]}
	}
	return out
}

// MapNode
type MapNodeResolver struct{ m *MapNodeModel }

func (n *MapNodeResolver) ID() graphql.ID  { return graphql.ID(n.m.ID) }
func (n *MapNodeResolver) Floor() int32    { return int32(n.m.Floor) }
func (n *MapNodeResolver) Position() int32 { return int32(n.m.Position) }
func (n *MapNodeResolver) Type() string    { return strings.ToUpper(n.m.Type) }
func (n *MapNodeResolver) Visited() bool   { return n.m.Visited }
func (n *MapNodeResolver) Available() bool { return n.m.Available }

// MapEdge
type MapEdgeResolver struct{ m *MapEdgeModel }

func (e *MapEdgeResolver) FromNodeId() graphql.ID { return graphql.ID(e.m.FromNodeID) }
func (e *MapEdgeResolver) ToNodeId() graphql.ID   { return graphql.ID(e.m.ToNodeID) }

// Run / leaderboard
type RunResolver struct{ m *RunModel }

func (r *RunResolver) ID() graphql.ID       { return graphql.ID(r.m.ID) }
func (r *RunResolver) HeroId() graphql.ID   { return graphql.ID(r.m.HeroID) }
func (r *RunResolver) HeroName() string     { return r.m.HeroName }
func (r *RunResolver) PlayerName() string   { return r.m.HeroName }
func (r *RunResolver) FloorsCleared() int32 { return int32(r.m.FloorsCleared) }
func (r *RunResolver) MonstersKilled() int32 { return int32(r.m.MonstersKilled) }
func (r *RunResolver) Score() int32         { return int32(r.m.Score) }

// Card
type CardResolver struct{ m *CardModel }

func (c *CardResolver) ID() graphql.ID    { return graphql.ID(c.m.ID) }
func (c *CardResolver) DefId() graphql.ID { return graphql.ID(c.m.DefID) }
func (c *CardResolver) Name() string      { return c.m.Name }
func (c *CardResolver) Cost() int32       { return int32(c.m.Cost) }
func (c *CardResolver) Type() string      { return strings.ToUpper(c.m.Type) }
func (c *CardResolver) Effect() string    { return c.m.Effect }

// CardDefinition
type CardDefinitionResolver struct{ m *CardDefinitionModel }

func (c *CardDefinitionResolver) ID() graphql.ID    { return graphql.ID(c.m.ID) }
func (c *CardDefinitionResolver) Name() string      { return c.m.Name }
func (c *CardDefinitionResolver) Class() string     { return c.m.Class }
func (c *CardDefinitionResolver) Type() string      { return strings.ToUpper(c.m.Type) }
func (c *CardDefinitionResolver) Cost() int32       { return int32(c.m.Cost) }
func (c *CardDefinitionResolver) Effect() string    { return c.m.Effect }
func (c *CardDefinitionResolver) Rarity() string    { return strings.ToUpper(c.m.Rarity) }
func (c *CardDefinitionResolver) Description() string { return c.m.Description }

// HeroDeck
type HeroDeckResolver struct{ m *HeroDeckModel }

func (h *HeroDeckResolver) HeroId() graphql.ID { return graphql.ID(h.m.HeroID) }
func (h *HeroDeckResolver) Cards() []*CardResolver {
	out := make([]*CardResolver, len(h.m.Cards))
	for i := range h.m.Cards {
		out[i] = &CardResolver{&h.m.Cards[i]}
	}
	return out
}

// Relic
type RelicResolver struct{ m *RelicModel }

func (r *RelicResolver) ID() graphql.ID    { return graphql.ID(r.m.ID) }
func (r *RelicResolver) DefId() graphql.ID { return graphql.ID(r.m.DefID) }
func (r *RelicResolver) Name() string      { return r.m.Name }
func (r *RelicResolver) Rarity() string    { return r.m.Rarity }
func (r *RelicResolver) Description() string { return r.m.Description }

// EncounterState
type EncounterStateResolver struct{ m *EncounterStateModel }

func (e *EncounterStateResolver) EncounterId() graphql.ID { return graphql.ID(e.m.EncounterID) }
func (e *EncounterStateResolver) HeroState() *HeroCombatStateResolver {
	return &HeroCombatStateResolver{&e.m.HeroState}
}
func (e *EncounterStateResolver) Monsters() []*EncounterMonsterResolver {
	out := make([]*EncounterMonsterResolver, len(e.m.Monsters))
	for i := range e.m.Monsters {
		out[i] = &EncounterMonsterResolver{&e.m.Monsters[i]}
	}
	return out
}
func (e *EncounterStateResolver) TurnNumber() int32 { return int32(e.m.TurnNumber) }
func (e *EncounterStateResolver) Status() string    { return e.m.Status }
func (e *EncounterStateResolver) Message() string   { return e.m.Message }

// HeroCombatState
type HeroCombatStateResolver struct{ m *HeroCombatStateModel }

func (h *HeroCombatStateResolver) Hp() int32        { return int32(h.m.HP) }
func (h *HeroCombatStateResolver) MaxHp() int32     { return int32(h.m.MaxHP) }
func (h *HeroCombatStateResolver) Block() int32     { return int32(h.m.Block) }
func (h *HeroCombatStateResolver) Energy() int32    { return int32(h.m.Energy) }
func (h *HeroCombatStateResolver) MaxEnergy() int32 { return int32(h.m.MaxEnergy) }
func (h *HeroCombatStateResolver) Hand() []*CardResolver {
	out := make([]*CardResolver, len(h.m.Hand))
	for i := range h.m.Hand {
		out[i] = &CardResolver{&h.m.Hand[i]}
	}
	return out
}
func (h *HeroCombatStateResolver) DrawPileCount() int32    { return int32(h.m.DrawPileCount) }
func (h *HeroCombatStateResolver) DiscardPileCount() int32 { return int32(h.m.DiscardPileCount) }
func (h *HeroCombatStateResolver) Statuses() []*StatusEffectResolver {
	out := make([]*StatusEffectResolver, 0, len(h.m.Statuses))
	for name, stacks := range h.m.Statuses {
		s := StatusEffectModel{Name: name, Stacks: stacks}
		out = append(out, &StatusEffectResolver{&s})
	}
	return out
}

// EncounterMonster
type EncounterMonsterResolver struct{ m *EncounterMonsterModel }

func (m *EncounterMonsterResolver) ID() graphql.ID  { return graphql.ID(m.m.ID) }
func (m *EncounterMonsterResolver) Name() string    { return m.m.Name }
func (m *EncounterMonsterResolver) Hp() int32       { return int32(m.m.HP) }
func (m *EncounterMonsterResolver) MaxHp() int32    { return int32(m.m.MaxHP) }
func (m *EncounterMonsterResolver) Block() int32    { return int32(m.m.Block) }
func (m *EncounterMonsterResolver) Status() string  { return m.m.Status }
func (m *EncounterMonsterResolver) Intents() []*MonsterIntentResolver {
	out := make([]*MonsterIntentResolver, len(m.m.Intents))
	for i := range m.m.Intents {
		out[i] = &MonsterIntentResolver{&m.m.Intents[i]}
	}
	return out
}

// MonsterIntent
type MonsterIntentResolver struct{ m *MonsterIntentModel }

func (i *MonsterIntentResolver) Type() string  { return i.m.Type }
func (i *MonsterIntentResolver) Value() int32  { return int32(i.m.Value) }

// StatusEffect
type StatusEffectResolver struct{ m *StatusEffectModel }

func (s *StatusEffectResolver) Name() string   { return s.m.Name }
func (s *StatusEffectResolver) Stacks() int32  { return int32(s.m.Stacks) }

// CardRewards
type CardRewardsResolver struct{ m *CardRewardsModel }

func (c *CardRewardsResolver) EncounterId() graphql.ID { return graphql.ID(c.m.EncounterID) }
func (c *CardRewardsResolver) Cards() []*CardDefinitionResolver {
	out := make([]*CardDefinitionResolver, len(c.m.Cards))
	for i := range c.m.Cards {
		out[i] = &CardDefinitionResolver{&c.m.Cards[i]}
	}
	return out
}

// ─── Shop Resolvers ───────────────────────────────────────────────────────────

type ShopInventoryResolver struct{ m *ShopInventoryModel }

func (s *ShopInventoryResolver) Items() []*ShopItemResolver {
	out := make([]*ShopItemResolver, len(s.m.Items))
	for i := range s.m.Items {
		out[i] = &ShopItemResolver{&s.m.Items[i]}
	}
	return out
}

type ShopItemResolver struct{ m *ShopItemModel }

func (s *ShopItemResolver) CardDef() *CardDefinitionResolver { return &CardDefinitionResolver{&s.m.CardDef} }
func (s *ShopItemResolver) Price() int32                     { return int32(s.m.Price) }

// ─── Inventory Resolvers ──────────────────────────────────────────────────────

func (r *Resolver) HeroInventory(ctx context.Context, args struct{ HeroId graphql.ID }) (*HeroInventoryResolver, error) {
	inv, err := encoreCall[HeroInventoryModel](ctx, http.MethodGet, fmt.Sprintf("/inventory/%s", args.HeroId), nil)
	if err != nil {
		return nil, err
	}
	return &HeroInventoryResolver{inv}, nil
}

func (r *Resolver) UseItem(ctx context.Context, args struct {
	HeroId graphql.ID
	ItemId graphql.ID
}) (bool, error) {
	type useResp struct{}
	_, err := encoreCall[useResp](ctx, http.MethodPost, "/inventory/use", map[string]any{
		"hero_id": string(args.HeroId),
		"item_id": string(args.ItemId),
	})
	return err == nil, err
}

type HeroInventoryResolver struct{ m *HeroInventoryModel }

func (h *HeroInventoryResolver) HeroId() graphql.ID { return graphql.ID(h.m.HeroID) }
func (h *HeroInventoryResolver) Items() []*InventoryItemResolver {
	out := make([]*InventoryItemResolver, len(h.m.Items))
	for i := range h.m.Items {
		out[i] = &InventoryItemResolver{&h.m.Items[i]}
	}
	return out
}

type InventoryItemResolver struct{ m *InventoryItemModel }

func (i *InventoryItemResolver) Id() graphql.ID       { return graphql.ID(i.m.ID) }
func (i *InventoryItemResolver) Name() string         { return i.m.Name }
func (i *InventoryItemResolver) Type() string         { return i.m.Type }
func (i *InventoryItemResolver) Value() int32         { return int32(i.m.Value) }
func (i *InventoryItemResolver) Rarity() string       { return i.m.Rarity }
func (i *InventoryItemResolver) Description() string  { return i.m.Description }

// ─── Event Resolvers ──────────────────────────────────────────────────────────

func (r *Resolver) RandomEvent(ctx context.Context) (*RandomEventResolver, error) {
	event, err := encoreCall[RandomEventModel](ctx, http.MethodGet, "/game/event", nil)
	if err != nil {
		return nil, err
	}
	return &RandomEventResolver{event}, nil
}

func (r *Resolver) ResolveEvent(ctx context.Context, args struct {
	HeroId   graphql.ID
	EventId  graphql.ID
	ChoiceId graphql.ID
}) (*EventOutcomeResolver, error) {
	outcome, err := encoreCall[EventOutcomeModel](ctx, http.MethodPost, "/game/event/resolve", map[string]any{
		"hero_id":   string(args.HeroId),
		"event_id":  string(args.EventId),
		"choice_id": string(args.ChoiceId),
	})
	if err != nil {
		return nil, err
	}
	return &EventOutcomeResolver{outcome}, nil
}

type RandomEventResolver struct{ m *RandomEventModel }

func (e *RandomEventResolver) Id() graphql.ID    { return graphql.ID(e.m.ID) }
func (e *RandomEventResolver) Title() string     { return e.m.Title }
func (e *RandomEventResolver) Description() string { return e.m.Description }
func (e *RandomEventResolver) Choices() []*EventChoiceResolver {
	out := make([]*EventChoiceResolver, len(e.m.Choices))
	for i := range e.m.Choices {
		out[i] = &EventChoiceResolver{&e.m.Choices[i]}
	}
	return out
}

type EventChoiceResolver struct{ m *EventChoiceModel }

func (c *EventChoiceResolver) Id() graphql.ID      { return graphql.ID(c.m.ID) }
func (c *EventChoiceResolver) Label() string       { return c.m.Label }
func (c *EventChoiceResolver) Description() string { return c.m.Description }

type EventOutcomeResolver struct{ m *EventOutcomeModel }

func (o *EventOutcomeResolver) ChoiceId() string    { return o.m.ChoiceID }
func (o *EventOutcomeResolver) Description() string { return o.m.Description }
func (o *EventOutcomeResolver) GoldDelta() int32    { return int32(o.m.GoldDelta) }
func (o *EventOutcomeResolver) HpDelta() int32      { return int32(o.m.HPDelta) }


