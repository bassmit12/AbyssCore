package mapservice

import (
	"context"
	"fmt"
	"math/rand"

	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("map", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// ─── Types ────────────────────────────────────────────────────────────────────

type Node struct {
	ID      string `json:"id"`
	FloorID string `json:"floor_id"`
	Col     int    `json:"col"`
	Row     int    `json:"row"`
	Type    string `json:"type"` // combat | elite | event | shop | rest | boss
	Cleared bool   `json:"cleared"`
}

type Path struct {
	FromNodeID string `json:"from_node_id"`
	ToNodeID   string `json:"to_node_id"`
}

type Floor struct {
	ID            string `json:"id"`
	RunID         string `json:"run_id"`
	HeroID        string `json:"hero_id,omitempty"`
	CurrentNodeID string `json:"current_node_id,omitempty"`
	Level         int    `json:"level"`
	Nodes         []Node `json:"nodes"`
	Paths         []Path `json:"paths"`
	Edges         []Path `json:"edges"` // alias for paths, for gateway compatibility
}

type Run struct {
	ID           string `json:"id"`
	HeroID       string `json:"hero_id"`
	Status       string `json:"status"`
	CurrentFloor int    `json:"current_floor"`
}

type StartRunRequest struct {
	HeroID string `json:"hero_id"`
}

type TravelRequest struct {
	HeroID string `json:"hero_id"`
	NodeID string `json:"node_id"`
}

type TravelResponse struct {
	Node Node `json:"node"`
}

type GetFloorResponse struct {
	Floor Floor `json:"floor"`
}

type MarkClearedRequest struct {
	HeroID string `json:"hero_id"`
	NodeID string `json:"node_id"`
}

// ─── API ──────────────────────────────────────────────────────────────────────

// StartRun creates a new run and generates floor 1 for the hero.
//
//encore:api auth method=POST path=/map/runs
func StartRun(ctx context.Context, req *StartRunRequest) (*Floor, error) {
	// Create run
	var runID string
	err := db.QueryRow(ctx, `
		INSERT INTO map.runs (hero_id) VALUES ($1::uuid) RETURNING id
	`, req.HeroID).Scan(&runID)
	if err != nil {
		return nil, fmt.Errorf("create run: %w", err)
	}

	// Update hero run_id in game schema (cross-schema local dev)
	_, _ = db.Exec(ctx, `
		UPDATE game.heroes SET run_id = $1::uuid WHERE id = $2::uuid
	`, runID, req.HeroID)

	floor, err := generateFloor(ctx, runID, 1)
	if err != nil {
		return nil, err
	}
	floor.HeroID = req.HeroID

	// Create hero_positions row (node_id=NULL = not yet placed on map).
	// This row is required for Travel to work — without it the first move
	// always returns "hero position not found".
	_, err = db.Exec(ctx, `
		INSERT INTO map.hero_positions (hero_id, run_id, node_id, floor_id)
		VALUES ($1::uuid, $2::uuid, NULL, $3::uuid)
		ON CONFLICT (hero_id) DO UPDATE
		  SET run_id=$2::uuid, node_id=NULL, floor_id=$3::uuid, updated_at=now()
	`, req.HeroID, runID, floor.ID)
	if err != nil {
		return nil, fmt.Errorf("init hero position: %w", err)
	}

	return floor, nil
}

// GetFloor returns the node graph for a specific floor of a run.
//
//encore:api auth method=GET path=/map/runs/:runID/floor/:level
func GetFloor(ctx context.Context, runID string, level int) (*GetFloorResponse, error) {
	var floorID string
	err := db.QueryRow(ctx, `
		SELECT id FROM map.floors WHERE run_id = $1::uuid AND level = $2
	`, runID, level).Scan(&floorID)
	if err != nil {
		return nil, fmt.Errorf("floor not found: %w", err)
	}
	floor, err := loadFloor(ctx, floorID)
	if err != nil {
		return nil, err
	}
	return &GetFloorResponse{Floor: *floor}, nil
}

// Travel moves the hero to an adjacent node and returns the updated floor graph.
//
//encore:api auth method=POST path=/map/travel
func Travel(ctx context.Context, req *TravelRequest) (*Floor, error) {
	// Get hero's current position
	var currentNodeID *string
	var floorID, runID string
	err := db.QueryRow(ctx, `
		SELECT node_id, floor_id, run_id FROM map.hero_positions WHERE hero_id = $1::uuid
	`, req.HeroID).Scan(&currentNodeID, &floorID, &runID)
	if err != nil {
		return nil, fmt.Errorf("hero position not found: %w", err)
	}

	// Validate the path is available
	if currentNodeID != nil {
		var exists bool
		err = db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM map.paths
				WHERE from_node_id = $1::uuid AND to_node_id = $2::uuid
			)
		`, *currentNodeID, req.NodeID).Scan(&exists)
		if err != nil || !exists {
			return nil, fmt.Errorf("invalid travel: no path from current node to target")
		}
	} else {
		// First move: must be col=0
		var col int
		err = db.QueryRow(ctx, `SELECT col FROM map.nodes WHERE id = $1::uuid`, req.NodeID).Scan(&col)
		if err != nil || col != 0 {
			return nil, fmt.Errorf("first move must be to a starting node (col 0)")
		}
	}

	// Load the target node
	var node Node
	err = db.QueryRow(ctx, `
		SELECT id, floor_id, col, row, type, cleared FROM map.nodes WHERE id = $1::uuid
	`, req.NodeID).Scan(&node.ID, &node.FloorID, &node.Col, &node.Row, &node.Type, &node.Cleared)
	if err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}

	// Update position
	_, err = db.Exec(ctx, `
		INSERT INTO map.hero_positions (hero_id, run_id, node_id, floor_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid)
		ON CONFLICT (hero_id) DO UPDATE
		  SET node_id=$3::uuid, floor_id=$4::uuid, updated_at=now()
	`, req.HeroID, runID, req.NodeID, floorID)
	if err != nil {
		return nil, fmt.Errorf("update position: %w", err)
	}

	// Return the updated floor graph
	var level int
	if err := db.QueryRow(ctx, `SELECT current_floor FROM map.runs WHERE id = $1::uuid`, runID).Scan(&level); err != nil {
		return nil, fmt.Errorf("load current floor: %w", err)
	}
	floor, err := getFloorByRunAndLevel(ctx, runID, level)
	if err != nil {
		return nil, err
	}
	floor.CurrentNodeID = req.NodeID
	floor.HeroID = req.HeroID
	return floor, nil
}

// HeroGraph returns the current floor graph for a hero's active run.
//
//encore:api public method=GET path=/map/hero/:heroID/graph
func HeroGraph(ctx context.Context, heroID string) (*Floor, error) {
	// Find the hero's active run
	var runID string
	err := db.QueryRow(ctx, `
		SELECT id FROM map.runs WHERE hero_id = $1::uuid AND status = 'active' ORDER BY created_at DESC LIMIT 1
	`, heroID).Scan(&runID)
	if err != nil {
		return nil, fmt.Errorf("no active run for hero: %w", err)
	}
	// Get current floor
	var currentFloor int
	err = db.QueryRow(ctx, `
		SELECT current_floor FROM map.runs WHERE id = $1::uuid
	`, runID).Scan(&currentFloor)
	if err != nil {
		return nil, fmt.Errorf("run not found: %w", err)
	}
	floor, err := getFloorByRunAndLevel(ctx, runID, currentFloor)
	if err != nil {
		return nil, err
	}
	// Attach the hero's current node so the frontend can highlight it.
	var nodeID *string
	_ = db.QueryRow(ctx, `SELECT node_id FROM map.hero_positions WHERE hero_id = $1::uuid`, heroID).Scan(&nodeID)
	if nodeID != nil {
		floor.CurrentNodeID = *nodeID
	}
	floor.HeroID = heroID
	return floor, nil
}

// getFloorByRunAndLevel is a helper to load a persisted floor with nodes+paths.
func getFloorByRunAndLevel(ctx context.Context, runID string, level int) (*Floor, error) {
	var floorID string
	err := db.QueryRow(ctx, `
		SELECT id FROM map.floors WHERE run_id = $1::uuid AND level = $2
	`, runID, level).Scan(&floorID)
	if err != nil {
		return nil, fmt.Errorf("floor not found: %w", err)
	}
	rows, err := db.Query(ctx, `
		SELECT id, floor_id, col, row, type, cleared FROM map.nodes WHERE floor_id = $1::uuid ORDER BY col, row
	`, floorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.FloorID, &n.Col, &n.Row, &n.Type, &n.Cleared); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	pathRows, err := db.Query(ctx, `
		SELECT from_node_id, to_node_id FROM map.paths
		WHERE from_node_id IN (SELECT id FROM map.nodes WHERE floor_id = $1::uuid)
	`, floorID)
	if err != nil {
		return nil, err
	}
	defer pathRows.Close()
	var paths []Path
	for pathRows.Next() {
		var p Path
		if err := pathRows.Scan(&p.FromNodeID, &p.ToNodeID); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return &Floor{ID: floorID, RunID: runID, Level: level, Nodes: nodes, Paths: paths, Edges: paths}, nil
}

// MarkCleared marks a node as cleared after combat/event/etc.
//
//encore:api auth method=POST path=/map/clear
func MarkCleared(ctx context.Context, req *MarkClearedRequest) error {
	_, err := db.Exec(ctx, `
		UPDATE map.nodes SET cleared = TRUE WHERE id = $1::uuid
	`, req.NodeID)
	if err != nil {
		return fmt.Errorf("mark cleared: %w", err)
	}

	// Check if all non-boss nodes on floor are cleared → advance floor
	// (simplified: just mark for now, client calls StartRun nextFloor)
	return nil
}

// NextFloor generates the next floor for an active run.
//
//encore:api auth method=POST path=/map/runs/:runID/next-floor
func NextFloor(ctx context.Context, runID string) (*Floor, error) {
	var currentFloor int
	err := db.QueryRow(ctx, `
		SELECT current_floor FROM map.runs WHERE id = $1::uuid
	`, runID).Scan(&currentFloor)
	if err != nil {
		return nil, fmt.Errorf("run not found: %w", err)
	}

	next := currentFloor + 1
	_, err = db.Exec(ctx, `
		UPDATE map.runs SET current_floor=$1, updated_at=now() WHERE id=$2::uuid
	`, next, runID)
	if err != nil {
		return nil, err
	}

	return generateFloor(ctx, runID, next)
}

// ─── Floor Generator ──────────────────────────────────────────────────────────

// generateFloor builds a branching path floor and persists it.
// Layout: 7 columns. Col 0 = start (1 node), cols 1-5 = branches (2-3 nodes each),
// col 6 = boss (1 node).
func generateFloor(ctx context.Context, runID string, level int) (*Floor, error) {
	var floorID string
	err := db.QueryRow(ctx, `
		INSERT INTO map.floors (run_id, level, cols) VALUES ($1::uuid, $2, 7) RETURNING id
	`, runID, level).Scan(&floorID)
	if err != nil {
		return nil, fmt.Errorf("create floor: %w", err)
	}

	// Generate node grid: col -> []row indices
	colNodes := [][]int{
		{0},       // col 0: single start
		{0, 1},    // col 1
		{0, 1, 2}, // col 2
		{0, 1},    // col 3
		{0, 1, 2}, // col 4
		{0, 1},    // col 5
		{0},       // col 6: boss
	}

	nodeTypes := []string{"combat", "elite", "event", "shop", "rest"}
	bossEvery := 3

	// nodeGrid[col][row] = Node (for path linking)
	nodeGrid := make([][]*Node, len(colNodes))

	// Insert all nodes
	for col, rows := range colNodes {
		nodeGrid[col] = make([]*Node, len(rows))
		for _, row := range rows {
			var ntype string
			if col == 0 || col == len(colNodes)-1 {
				if col == len(colNodes)-1 {
					ntype = "boss"
					_ = bossEvery
				} else {
					ntype = "combat"
				}
			} else {
				// Force shop every 2 cols, rest every 2, otherwise random
				switch {
				case col%4 == 3:
					ntype = "shop"
				case col%4 == 1 && row == len(colNodes[col])-1:
					ntype = "rest"
				default:
					ntype = nodeTypes[rand.Intn(len(nodeTypes)-1)] // skip shop/rest from pure random
				}
				// Elites only on floors 2+
				if ntype == "elite" && level < 2 {
					ntype = "combat"
				}
			}

			var nodeID string
			err := db.QueryRow(ctx, `
				INSERT INTO map.nodes (floor_id, col, row, type)
				VALUES ($1::uuid, $2, $3, $4) RETURNING id
			`, floorID, col, row, ntype).Scan(&nodeID)
			if err != nil {
				return nil, fmt.Errorf("insert node col=%d row=%d: %w", col, row, err)
			}
			nodeGrid[col][row] = &Node{ID: nodeID, FloorID: floorID, Col: col, Row: row, Type: ntype}
		}
	}

	// Connect paths: each node in col N connects to 1-2 nodes in col N+1
	for col := 0; col < len(colNodes)-1; col++ {
		fromNodes := nodeGrid[col]
		toNodes := nodeGrid[col+1]

		// Ensure every to-node has at least one incoming edge
		connected := make(map[int]bool)
		for _, from := range fromNodes {
			// Connect to 1 or 2 to-nodes
			count := 1
			if len(toNodes) > 1 && rand.Float32() < 0.4 {
				count = 2
			}
			perm := rand.Perm(len(toNodes))
			for i := 0; i < count && i < len(perm); i++ {
				toRow := perm[i]
				to := toNodes[toRow]
				_, _ = db.Exec(ctx, `
					INSERT INTO map.paths (floor_id, from_node_id, to_node_id)
					VALUES ($1::uuid, $2::uuid, $3::uuid)
					ON CONFLICT DO NOTHING
				`, floorID, from.ID, to.ID)
				connected[toRow] = true
			}
		}

		// Any to-node with no incoming edge: connect from a random from-node
		for toRow, to := range toNodes {
			if !connected[toRow] {
				from := fromNodes[rand.Intn(len(fromNodes))]
				_, _ = db.Exec(ctx, `
					INSERT INTO map.paths (floor_id, from_node_id, to_node_id)
					VALUES ($1::uuid, $2::uuid, $3::uuid)
					ON CONFLICT DO NOTHING
				`, floorID, from.ID, to.ID)
			}
		}
	}

	return loadFloor(ctx, floorID)
}

func loadFloor(ctx context.Context, floorID string) (*Floor, error) {
	var f Floor
	err := db.QueryRow(ctx, `
		SELECT fl.id, fl.run_id, fl.level
		FROM map.floors fl WHERE fl.id = $1::uuid
	`, floorID).Scan(&f.ID, &f.RunID, &f.Level)
	if err != nil {
		return nil, fmt.Errorf("load floor: %w", err)
	}

	// Load nodes
	rows, err := db.Query(ctx, `
		SELECT id, floor_id, col, row, type, cleared
		FROM map.nodes WHERE floor_id = $1::uuid ORDER BY col, row
	`, floorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.FloorID, &n.Col, &n.Row, &n.Type, &n.Cleared); err != nil {
			return nil, err
		}
		f.Nodes = append(f.Nodes, n)
	}

	// Load paths
	prows, err := db.Query(ctx, `
		SELECT from_node_id, to_node_id FROM map.paths WHERE floor_id = $1::uuid
	`, floorID)
	if err != nil {
		return nil, err
	}
	defer prows.Close()
	for prows.Next() {
		var p Path
		if err := prows.Scan(&p.FromNodeID, &p.ToNodeID); err != nil {
			return nil, err
		}
		f.Paths = append(f.Paths, p)
	}

	return &f, nil
}
