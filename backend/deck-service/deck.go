package deck

import (
	"context"
	"fmt"
	"math/rand"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("deck", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// ─── Types ────────────────────────────────────────────────────────────────────

type CardDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Type        string `json:"type"`
	Cost        int    `json:"cost"`
	Effect      string `json:"effect"`       // raw JSON
	Rarity      string `json:"rarity"`
	Description string `json:"description"`
}

type Card struct {
	ID          string `json:"id"`          // hero_decks.id
	DefID       string `json:"def_id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Type        string `json:"type"`
	Cost        int    `json:"cost"`
	Effect      string `json:"effect"`
	Rarity      string `json:"rarity"`
	Description string `json:"description"`
	Upgraded    bool   `json:"upgraded"`
}

type RelicDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Rarity      string `json:"rarity"`
	Effect      string `json:"effect"`
	Description string `json:"description"`
}

type HeroRelic struct {
	ID     string   `json:"id"`
	HeroID string   `json:"hero_id"`
	Relic  RelicDef `json:"relic"`
}

type AddCardRequest struct {
	HeroID    string `json:"hero_id"`
	CardDefID string `json:"card_def_id"`
}

type RemoveCardRequest struct {
	HeroID string `json:"hero_id"`
	CardID string `json:"card_id"` // hero_decks.id
}

type UpgradeCardRequest struct {
	HeroID string `json:"hero_id"`
	CardID string `json:"card_id"`
}

type AddRelicRequest struct {
	HeroID     string `json:"hero_id"`
	RelicDefID string `json:"relic_def_id"`
}

type DeckResponse struct {
	HeroID string `json:"hero_id"`
	Cards  []Card `json:"cards"`
}

type RelicsResponse struct {
	HeroID string      `json:"hero_id"`
	Relics []HeroRelic `json:"relics"`
}

type CardRewardsResponse struct {
	Cards []CardDef `json:"cards"`
}

// ─── API ──────────────────────────────────────────────────────────────────────

// GetDeck returns all cards in a hero's current deck.
//
//encore:api public method=GET path=/deck/hero/:heroID
func GetDeck(ctx context.Context, heroID string) (*DeckResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT hd.id, cd.id, cd.name, cd.class, cd.type, cd.cost,
		       cd.effect::text, cd.rarity, cd.description, hd.upgraded
		FROM deck.hero_decks hd
		JOIN deck.card_definitions cd ON cd.id = hd.card_def_id
		WHERE hd.hero_id = $1::uuid
		ORDER BY hd.acquired_at
	`, heroID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resp := &DeckResponse{HeroID: heroID, Cards: []Card{}}
	for rows.Next() {
		var c Card
		if err := rows.Scan(&c.ID, &c.DefID, &c.Name, &c.Class, &c.Type,
			&c.Cost, &c.Effect, &c.Rarity, &c.Description, &c.Upgraded); err != nil {
			return nil, err
		}
		resp.Cards = append(resp.Cards, c)
	}
	return resp, rows.Err()
}

// AddCard adds a card definition to a hero's deck.
//
//encore:api auth method=POST path=/deck/add
func AddCard(ctx context.Context, req *AddCardRequest) (*Card, error) {
	c := &Card{}
	err := db.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO deck.hero_decks (hero_id, card_def_id)
			VALUES ($1::uuid, $2::uuid) RETURNING id, card_def_id, upgraded
		)
		SELECT ins.id, cd.id, cd.name, cd.class, cd.type, cd.cost,
		       cd.effect::text, cd.rarity, cd.description, ins.upgraded
		FROM ins JOIN deck.card_definitions cd ON cd.id = ins.card_def_id
	`, req.HeroID, req.CardDefID).Scan(
		&c.ID, &c.DefID, &c.Name, &c.Class, &c.Type,
		&c.Cost, &c.Effect, &c.Rarity, &c.Description, &c.Upgraded,
	)
	return c, err
}

// RemoveCard removes a card from a hero's deck.
//
//encore:api auth method=POST path=/deck/remove
func RemoveCard(ctx context.Context, req *RemoveCardRequest) error {
	res, err := db.Exec(ctx, `
		DELETE FROM deck.hero_decks WHERE id = $1::uuid AND hero_id = $2::uuid
	`, req.CardID, req.HeroID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("card not found in deck")
	}
	return nil
}

// UpgradeCard marks a card as upgraded.
//
//encore:api auth method=POST path=/deck/upgrade
func UpgradeCard(ctx context.Context, req *UpgradeCardRequest) (*Card, error) {
	c := &Card{}
	err := db.QueryRow(ctx, `
		WITH upd AS (
			UPDATE deck.hero_decks SET upgraded = TRUE
			WHERE id = $1::uuid AND hero_id = $2::uuid
			RETURNING id, card_def_id, upgraded
		)
		SELECT upd.id, cd.id, cd.name, cd.class, cd.type, cd.cost,
		       cd.effect::text, cd.rarity, cd.description, upd.upgraded
		FROM upd JOIN deck.card_definitions cd ON cd.id = upd.card_def_id
	`, req.CardID, req.HeroID).Scan(
		&c.ID, &c.DefID, &c.Name, &c.Class, &c.Type,
		&c.Cost, &c.Effect, &c.Rarity, &c.Description, &c.Upgraded,
	)
	return c, err
}

// GetRelics returns all relics held by a hero.
//
//encore:api public method=GET path=/deck/hero/:heroID/relics
func GetRelics(ctx context.Context, heroID string) (*RelicsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT hr.id, hr.hero_id, rd.id, rd.name, rd.rarity, rd.effect::text, rd.description
		FROM deck.hero_relics hr
		JOIN deck.relic_definitions rd ON rd.id = hr.relic_def_id
		WHERE hr.hero_id = $1::uuid
		ORDER BY hr.acquired_at
	`, heroID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resp := &RelicsResponse{HeroID: heroID, Relics: []HeroRelic{}}
	for rows.Next() {
		var hr HeroRelic
		if err := rows.Scan(&hr.ID, &hr.HeroID,
			&hr.Relic.ID, &hr.Relic.Name, &hr.Relic.Rarity,
			&hr.Relic.Effect, &hr.Relic.Description); err != nil {
			return nil, err
		}
		resp.Relics = append(resp.Relics, hr)
	}
	return resp, rows.Err()
}

// AddRelic grants a relic to a hero.
//
//encore:api auth method=POST path=/deck/relics/add
func AddRelic(ctx context.Context, req *AddRelicRequest) (*HeroRelic, error) {
	hr := &HeroRelic{HeroID: req.HeroID}
	err := db.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO deck.hero_relics (hero_id, relic_def_id)
			VALUES ($1::uuid, $2::uuid)
			ON CONFLICT DO NOTHING
			RETURNING id, relic_def_id
		)
		SELECT ins.id, rd.id, rd.name, rd.rarity, rd.effect::text, rd.description
		FROM ins JOIN deck.relic_definitions rd ON rd.id = ins.relic_def_id
	`, req.HeroID, req.RelicDefID).Scan(
		&hr.ID,
		&hr.Relic.ID, &hr.Relic.Name, &hr.Relic.Rarity,
		&hr.Relic.Effect, &hr.Relic.Description,
	)
	return hr, err
}

// CardLibrary returns all card definitions.
//
//encore:api public method=GET path=/deck/library
func CardLibrary(ctx context.Context) (*CardRewardsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, class, type, cost, effect::text, rarity, description
		FROM deck.card_definitions ORDER BY class, rarity, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resp := &CardRewardsResponse{Cards: []CardDef{}}
	for rows.Next() {
		var c CardDef
		if err := rows.Scan(&c.ID, &c.Name, &c.Class, &c.Type,
			&c.Cost, &c.Effect, &c.Rarity, &c.Description); err != nil {
			return nil, err
		}
		resp.Cards = append(resp.Cards, c)
	}
	return resp, rows.Err()
}

// CardRewards returns 3 random card picks for a post-combat reward screen,
// weighted by rarity and filtered to the hero's class (+ 'any').
//
//encore:api public method=GET path=/deck/rewards/:heroClass
func CardRewards(ctx context.Context, heroClass string) (*CardRewardsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, class, type, cost, effect::text, rarity, description
		FROM deck.card_definitions
		WHERE class IN ($1, 'any')
	`, heroClass)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type weighted struct {
		card   CardDef
		weight int
	}
	rarityWeight := map[string]int{
		"common": 60, "uncommon": 30, "rare": 10,
	}
	var pool []weighted
	total := 0
	for rows.Next() {
		var c CardDef
		if err := rows.Scan(&c.ID, &c.Name, &c.Class, &c.Type,
			&c.Cost, &c.Effect, &c.Rarity, &c.Description); err != nil {
			return nil, err
		}
		w := rarityWeight[c.Rarity]
		if w == 0 {
			w = 10
		}
		pool = append(pool, weighted{c, w})
		total += w
	}

	// Pick 3 distinct cards
	picked := make([]CardDef, 0, 3)
	usedIdx := map[int]bool{}
	for len(picked) < 3 && len(picked) < len(pool) {
		roll := rand.Intn(total)
		cum := 0
		for i, w := range pool {
			cum += w.weight
			if roll < cum && !usedIdx[i] {
				picked = append(picked, w.card)
				usedIdx[i] = true
				break
			}
		}
	}
	return &CardRewardsResponse{Cards: picked}, nil
}

// ─── Inter-service endpoints (called by combat-service) ──────────────────────

type HeroDeckResponse struct {
	HeroID string         `json:"hero_id"`
	Cards  []HeroDeckCard `json:"cards"`
}

type HeroDeckCard struct {
	ID    string `json:"id"`
	DefID string `json:"def_id"`
	Name  string `json:"name"`
	Cost  int    `json:"cost"`
	Type  string `json:"type"`
}

type CardDetailResponse struct {
	ID     string `json:"id"`
	DefID  string `json:"def_id"`
	Name   string `json:"name"`
	Cost   int    `json:"cost"`
	Effect string `json:"effect"`
	Type   string `json:"type"`
}

// GetCard returns full details of a single hero-deck card by its hero_deck ID.
//
//encore:api public method=GET path=/deck/card/:cardID
func GetCard(ctx context.Context, cardID string) (*CardDetailResponse, error) {
	r := &CardDetailResponse{}
	err := db.QueryRow(ctx, `
		SELECT hd.id, cd.id, cd.name, cd.cost, cd.effect::text, cd.type
		FROM deck.hero_decks hd
		JOIN deck.card_definitions cd ON cd.id = hd.card_def_id
		WHERE hd.id = $1::uuid
	`, cardID).Scan(&r.ID, &r.DefID, &r.Name, &r.Cost, &r.Effect, &r.Type)
	return r, err
}

type RelicBonusResponse struct {
	Value int `json:"value"`
}

// GetRelicBonus returns the numeric value of a relic effect key for a hero.
//
//encore:api public method=GET path=/deck/relic-bonus/:heroID/:effectKey
func GetRelicBonus(ctx context.Context, heroID, effectKey string) (*RelicBonusResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT rd.effect->>$2 FROM deck.hero_relics hr
		JOIN deck.relic_definitions rd ON rd.id = hr.relic_def_id
		WHERE hr.hero_id = $1::uuid
	`, heroID, effectKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	total := 0
	for rows.Next() {
		var raw *string
		if err := rows.Scan(&raw); err != nil {
			continue
		}
		if raw != nil {
			var n int
			fmt.Sscanf(*raw, "%d", &n)
			total += n
		}
	}
	return &RelicBonusResponse{Value: total}, rows.Err()
}

type SeedStarterDeckRequest struct {
	HeroID string `json:"hero_id"`
	Class  string `json:"class"` // "warrior", "mage", "rogue"
}

type SeedStarterDeckResponse struct {
	Seeded bool   `json:"seeded"`
	Count  int    `json:"count"`
}

// ─── Shop ─────────────────────────────────────────────────────────────────────

type ShopItem struct {
	CardDef CardDef `json:"card_def"`
	Price   int     `json:"price"`
}

type ShopResponse struct {
	Items []ShopItem `json:"items"`
}

type BuyCardRequest struct {
	HeroID    string `json:"hero_id"`
	CardDefID string `json:"card_def_id"`
}

// ShopCards returns 5 cards for sale in the shop, filtered to the hero's class.
//
//encore:api public method=GET path=/deck/shop/:heroClass
func ShopCards(ctx context.Context, heroClass string) (*ShopResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, class, type, cost, effect::text, rarity, description
		FROM deck.card_definitions
		WHERE class IN ($1, 'any')
		ORDER BY random()
		LIMIT 5
	`, heroClass)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rarityPrice := map[string]int{"common": 50, "uncommon": 75, "rare": 150}
	resp := &ShopResponse{Items: []ShopItem{}}
	for rows.Next() {
		var c CardDef
		if err := rows.Scan(&c.ID, &c.Name, &c.Class, &c.Type,
			&c.Cost, &c.Effect, &c.Rarity, &c.Description); err != nil {
			return nil, err
		}
		price := rarityPrice[c.Rarity]
		if price == 0 {
			price = 50
		}
		resp.Items = append(resp.Items, ShopItem{CardDef: c, Price: price})
	}
	return resp, rows.Err()
}

// BuyCard adds a card to the hero's deck (gold is deducted by caller via game-service).
//
//encore:api public method=POST path=/deck/shop/buy
func BuyCard(ctx context.Context, req *BuyCardRequest) (*Card, error) {
	c := &Card{}
	err := db.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO deck.hero_decks (hero_id, card_def_id)
			VALUES ($1::uuid, $2::uuid) RETURNING id, card_def_id, upgraded
		)
		SELECT ins.id, cd.id, cd.name, cd.class, cd.type, cd.cost,
		       cd.effect::text, cd.rarity, cd.description, ins.upgraded
		FROM ins JOIN deck.card_definitions cd ON cd.id = ins.card_def_id
	`, req.HeroID, req.CardDefID).Scan(
		&c.ID, &c.DefID, &c.Name, &c.Class, &c.Type,
		&c.Cost, &c.Effect, &c.Rarity, &c.Description, &c.Upgraded,
	)
	return c, err
}

// SeedStarterDeck gives a hero their starter cards if they have no cards yet.
//
//encore:api public method=POST path=/deck/seed-starter
func SeedStarterDeck(ctx context.Context, req *SeedStarterDeckRequest) (*SeedStarterDeckResponse, error) {
	// Check if hero already has cards
	var count int
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM deck.hero_decks WHERE hero_id = $1::uuid`, req.HeroID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return &SeedStarterDeckResponse{Seeded: false, Count: count}, nil
	}

	// Get starter card definitions for the class: 5x basic attack, 4x basic skill
	// Use lowest-cost card of each type as the "starter" card
	starterCards := []struct {
		cardType string
		count    int
	}{
		{"attack", 5},
		{"skill", 4},
	}

	added := 0
	for _, sc := range starterCards {
		var defID string
		err := db.QueryRow(ctx, `
			SELECT id FROM deck.card_definitions
			WHERE class = $1 AND type = $2
			ORDER BY cost ASC, name ASC
			LIMIT 1
		`, req.Class, sc.cardType).Scan(&defID)
		if err != nil {
			continue // definition not found for this class/type, skip
		}
		for i := 0; i < sc.count; i++ {
			_, err := db.Exec(ctx, `
				INSERT INTO deck.hero_decks (hero_id, card_def_id)
				VALUES ($1::uuid, $2::uuid)
			`, req.HeroID, defID)
			if err != nil {
				return nil, fmt.Errorf("add starter card: %w", err)
			}
			added++
		}
	}

	return &SeedStarterDeckResponse{Seeded: true, Count: added}, nil
}
