"use client"
import { useMemo } from "react"
import type { FloorGraph, MapNode, MapEdge } from "@/types/game"

const NODE_CONFIG: Record<string, { icon: string; color: string; label: string }> = {
  combat:   { icon: "⚔️",  color: "bg-red-900 border-red-700",      label: "Combat"   },
  elite:    { icon: "💀",  color: "bg-orange-900 border-orange-600", label: "Elite"    },
  boss:     { icon: "👹",  color: "bg-red-950 border-red-500",       label: "Boss"     },
  rest:     { icon: "🔥",  color: "bg-green-900 border-green-700",   label: "Rest"     },
  shop:     { icon: "🛒",  color: "bg-yellow-900 border-yellow-700", label: "Shop"     },
  treasure: { icon: "💰",  color: "bg-yellow-800 border-yellow-500", label: "Treasure" },
  event:    { icon: "❓",  color: "bg-blue-900 border-blue-700",     label: "Event"    },
  COMBAT:   { icon: "⚔️",  color: "bg-red-900 border-red-700",      label: "Combat"   },
  ELITE:    { icon: "💀",  color: "bg-orange-900 border-orange-600", label: "Elite"    },
  BOSS:     { icon: "👹",  color: "bg-red-950 border-red-500",       label: "Boss"     },
  REST:     { icon: "🔥",  color: "bg-green-900 border-green-700",   label: "Rest"     },
  SHOP:     { icon: "🛒",  color: "bg-yellow-900 border-yellow-700", label: "Shop"     },
  EVENT:    { icon: "❓",  color: "bg-blue-900 border-blue-700",     label: "Event"    },
}

// Layout: top = boss floor, bottom = start floor
// X axis = position (row in backend), Y axis = floor (col in backend, inverted)
const NODE_W   = 64
const NODE_H   = 64
const COL_GAP  = 96   // vertical gap between floors (Y spacing)
const ROW_GAP  = 80   // horizontal gap between nodes on same floor (X spacing)
const MAX_ROWS = 3    // max nodes per floor
const PAD_X    = 60
const PAD_Y    = 40

interface Props {
  graph: FloorGraph
  onTravel: (nodeId: string) => void
  loading?: boolean
}

export default function MapView({ graph, onTravel, loading }: Props) {
  const enrichedNodes = useMemo(() => {
    const currentId = graph.currentNodeId
    const visitedIds = new Set(
      graph.nodes.filter(n => n.visited || n.id === currentId).map(n => n.id)
    )
    const reachable = new Set<string>()
    for (const e of graph.edges) {
      if (visitedIds.has(e.fromNodeId)) reachable.add(e.toNodeId)
    }
    const nothingVisited = visitedIds.size === 0

    return graph.nodes.map(n => ({
      ...n,
      visited:   n.visited || n.id === currentId,
      available: n.available ||
        (nothingVisited && n.floor === 0) ||
        (reachable.has(n.id) && n.id !== currentId && !n.visited),
    }))
  }, [graph])

  const { nodePos, canvasW, canvasH } = useMemo(() => {
    // Group by floor (backend "col") to compute X-centering per floor
    const byFloor = new Map<number, MapNode[]>()
    for (const n of enrichedNodes) {
      if (!byFloor.has(n.floor)) byFloor.set(n.floor, [])
      byFloor.get(n.floor)!.push(n)
    }

    const maxFloor = Math.max(0, ...enrichedNodes.map(n => n.floor))

    // Canvas size
    const cW = PAD_X * 2 + MAX_ROWS * NODE_W + (MAX_ROWS - 1) * ROW_GAP
    const cH = PAD_Y * 2 + (maxFloor + 1) * NODE_H + maxFloor * COL_GAP

    const pos = new Map<string, { x: number; y: number }>()
    for (const [floor, nodes] of byFloor.entries()) {
      const rowCount = nodes.length
      // Total width of this floor's nodes
      const rowW = rowCount * NODE_W + (rowCount - 1) * ROW_GAP
      const startX = (cW - rowW) / 2

      // Sort by position for consistent ordering
      const sorted = [...nodes].sort((a, b) => a.position - b.position)
      sorted.forEach((n, i) => {
        // Invert floor: floor 0 (start) is at bottom, maxFloor (boss) at top
        const floorY = maxFloor - floor
        const x = startX + i * (NODE_W + ROW_GAP)
        const y = PAD_Y + floorY * (NODE_H + COL_GAP)
        pos.set(n.id, { x, y })
      })
    }

    return { nodePos: pos, canvasW: cW, canvasH: cH }
  }, [enrichedNodes])

  return (
    <div className="relative w-full flex flex-col items-center">
      {/* "Boss" label at top, "Start" at bottom */}
      <div className="w-full flex justify-center mb-1">
        <span className="text-xs text-red-400 font-semibold tracking-widest uppercase opacity-70">
          ↑ Boss
        </span>
      </div>

      <div className="overflow-auto w-full flex justify-center">
        <div style={{ position: "relative", width: canvasW, height: canvasH, minWidth: canvasW }}>
          {/* SVG edges */}
          <svg
            style={{ position: "absolute", inset: 0, pointerEvents: "none", overflow: "visible", zIndex: 0 }}
            width={canvasW}
            height={canvasH}
          >
            {graph.edges.map((e: MapEdge, i: number) => {
              const from = nodePos.get(e.fromNodeId)
              const to   = nodePos.get(e.toNodeId)
              if (!from || !to) return null
              // from-node: top-center; to-node: bottom-center
              // (from has higher floor = lower Y; to has lower floor = higher Y)
              // Actually from = lower floor (start side), to = higher floor (boss side)
              // From is closer to bottom (higher Y), to is closer to top (lower Y)
              const x1 = from.x + NODE_W / 2
              const y1 = from.y  // top of from-node (which is below to-node visually)
              const x2 = to.x + NODE_W / 2
              const y2 = to.y + NODE_H  // bottom of to-node
              const cy = (y1 + y2) / 2
              return (
                <path
                  key={i}
                  d={`M${x1},${y1} C${x1},${cy} ${x2},${cy} ${x2},${y2}`}
                  stroke="#4b5563"
                  strokeWidth={2}
                  strokeDasharray="5 4"
                  fill="none"
                  opacity={0.7}
                />
              )
            })}
          </svg>

          {/* Nodes */}
          {enrichedNodes.map((node: MapNode) => {
            const pos = nodePos.get(node.id)
            if (!pos) return null
            const cfg       = NODE_CONFIG[node.type] ?? NODE_CONFIG.event
            const isCurrent = node.id === graph.currentNodeId
            const isStart   = graph.currentNodeId === null && node.floor === 0
            const isAvail   = node.available && !node.visited

            let ringClass    = ""
            let opacityClass = "opacity-40"
            let cursorClass  = ""

            if (isCurrent) {
              ringClass    = "ring-2 ring-purple-500 border-purple-400"
              opacityClass = "opacity-100"
            } else if (isStart) {
              ringClass    = "ring-2 ring-emerald-400 border-emerald-400"
              opacityClass = "opacity-100"
              cursorClass  = "cursor-pointer"
            } else if (node.visited) {
              opacityClass = "opacity-50"
            } else if (isAvail) {
              ringClass    = cfg.color.split(" ")[1]
              opacityClass = "opacity-100"
              cursorClass  = "cursor-pointer hover:scale-110 transition-transform"
            }

            const baseColor = isCurrent
              ? "bg-purple-950 border-purple-400"
              : isStart
              ? "bg-emerald-950 border-emerald-400"
              : cfg.color

            return (
              <div
                key={node.id}
                style={{
                  position: "absolute",
                  left: pos.x,
                  top:  pos.y,
                  width:  NODE_W,
                  height: NODE_H,
                  zIndex: 1,
                }}
                className={`flex flex-col items-center justify-center rounded-xl border-2 text-center select-none
                  ${baseColor} ${opacityClass} ${ringClass} ${cursorClass}`}
                title={cfg.label}
                onClick={() => {
                  if ((isAvail || isStart) && !loading) onTravel(node.id)
                }}
              >
                <span className="text-xl leading-none">{cfg.icon}</span>
                <span className="text-[9px] text-gray-300 mt-0.5 leading-none">{cfg.label}</span>
                {isCurrent && <span className="text-[8px] text-purple-300 leading-none mt-0.5">HERE</span>}
                {isStart && !graph.currentNodeId && (
                  <span className="text-[8px] text-emerald-300 leading-none mt-0.5">START</span>
                )}
              </div>
            )
          })}
        </div>
      </div>

      <div className="w-full flex justify-center mt-1">
        <span className="text-xs text-emerald-400 font-semibold tracking-widest uppercase opacity-70">
          Start ↓
        </span>
      </div>
    </div>
  )
}
