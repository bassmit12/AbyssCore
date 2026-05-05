"use client"
import { useState } from "react"
import { useMutation, useQuery } from "@apollo/client/react"
import {
  GET_HERO,
  START_RUN,
  TRAVEL_TO_NODE,
  START_ENCOUNTER,
  GET_FLOOR_GRAPH,
} from "@/lib/queries"
import MapView from "./MapView"
import CombatView from "./CombatView"
import ShopView from "./ShopView"
import VictoryScreen from "./VictoryScreen"
import EventView from "./EventView"
import InventoryPanel from "./InventoryPanel"
import HeroPanel from "./HeroPanel"
import CombatLog from "./CombatLog"
import type { FloorGraph, EncounterState } from "@/types/game"

type Screen = "lobby" | "map" | "combat" | "rest" | "shop" | "event" | "victory" | "dead"

interface Props {
  heroId: string
}

export default function GameView({ heroId }: Props) {
  const [screen, setScreen] = useState<Screen>("lobby")
  const [combatLog, setCombatLog] = useState<string[]>([])
  const [activeEncounter, setActiveEncounter] = useState<EncounterState | null>(null)
  const [activeNodeId, setActiveNodeId] = useState<string | null>(null)
  // Track floors cleared and monsters killed for the victory screen
  const [floorsCleared, setFloorsCleared] = useState(0)
  const [monstersKilled, setMonstersKilled] = useState(0)

  // ── Data ──────────────────────────────────────────────────────────────────

  const { data: heroData, refetch: refetchHero } = useQuery(GET_HERO, {
    variables: { id: heroId },
    pollInterval: screen === "map" ? 5000 : 0,
  })

  const { data: graphData, refetch: refetchGraph } = useQuery(GET_FLOOR_GRAPH, {
    variables: { heroId },
    skip: screen === "lobby",
    fetchPolicy: "network-only",
  })

  const hero = (heroData as { hero?: ReturnType<typeof Object> } | undefined)?.hero as {
    name: string; class: string; hp: number; maxHp: number; level: number; xp: number; gold: number; alive: boolean; runId?: string
  } | undefined
  const graph = (graphData as { floorGraph?: FloorGraph } | undefined)?.floorGraph

  // ── Mutations ─────────────────────────────────────────────────────────────

  const [startRun, { loading: startingRun }] = useMutation(START_RUN, {
    onCompleted: () => {
      addLog("A new run begins. Choose your path.")
      setFloorsCleared(0)
      setMonstersKilled(0)
      setScreen("map")
      refetchGraph()
    },
    onError: (err) => addLog(`Start run failed: ${err.message}`),
  })

  const [travelToNode, { loading: traveling }] = useMutation(TRAVEL_TO_NODE, {
    onCompleted: (raw) => {
      const data = raw as { travelToNode: FloorGraph }
      const g = data.travelToNode
      const node = g.nodes.find((n) => n.id === g.currentNodeId)
      if (!node) return
      addLog(`Arrived at ${node.type.toLowerCase()} node.`)
      setActiveNodeId(node.id)
      const t = node.type.toUpperCase()
      if (t === "COMBAT" || t === "ELITE" || t === "BOSS") {
        startEncounter({ variables: { heroId, nodeId: node.id } })
      } else if (t === "REST") {
        setScreen("rest")
      } else if (t === "SHOP") {
        setScreen("shop")
      } else if (t === "EVENT") {
        setScreen("event")
      } else {
        addLog("Nothing happens here yet. (treasure coming soon)")
        refetchGraph()
      }
    },
    onError: (err) => addLog(`Travel failed: ${err.message}`),
  })

  const [startEncounter, { loading: startingCombat }] = useMutation(START_ENCOUNTER, {
    onCompleted: (raw) => {
      const data = raw as { startEncounter: EncounterState }
      setActiveEncounter(data.startEncounter)
      addLog("Combat begins!")
      setScreen("combat")
    },
    onError: (err) => addLog(`Encounter failed: ${err.message}`),
  })

  // ── Helpers ───────────────────────────────────────────────────────────────

  const addLog = (msg: string) => {
    const t = new Date().toLocaleTimeString()
    setCombatLog((prev) => [`[${t}] ${msg}`, ...prev].slice(0, 50))
  }

  const handleVictory = () => {
    addLog("Victory! Returning to map...")
    // Track monsters killed (1 per combat win for now)
    setMonstersKilled((k) => k + 1)
    setActiveEncounter(null)

    // Check if the current node is a BOSS — if so, show victory screen
    if (graph) {
      const node = graph.nodes.find((n) => n.id === activeNodeId)
      if (node?.type.toUpperCase() === "BOSS") {
        setFloorsCleared((f) => f + 1)
        setScreen("victory")
        refetchHero()
        return
      }
    }

    setScreen("map")
    refetchHero()
    refetchGraph()
  }

  const handleDefeat = () => {
    addLog("💀 YOU HAVE DIED. Your run is over.")
    setActiveEncounter(null)
    setScreen("dead")
    refetchHero()
  }

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div className="flex flex-col h-screen bg-gray-950 text-gray-100 overflow-hidden">
      {/* Top bar — hidden on victory screen */}
      {screen !== "victory" && (
        <div className="flex items-center justify-between px-6 py-3 border-b border-gray-800 bg-gray-900 flex-shrink-0">
          <h1 className="text-xl font-bold text-purple-400">
            AbyssCore
            {process.env.NEXT_PUBLIC_APP_VERSION && (
              <span className="ml-2 text-xs font-mono text-gray-500">
                v{process.env.NEXT_PUBLIC_APP_VERSION.slice(0, 7)}
              </span>
            )}
          </h1>
          <div className="text-sm text-gray-400">
            {hero ? `${hero.name} · ${hero.class} · HP ${hero.hp}/${hero.maxHp} · 💰 ${hero.gold}` : "Loading..."}
          </div>
          <div className="text-xs text-gray-600 uppercase tracking-widest">
            {screen}
          </div>
        </div>
      )}

      {/* Main area */}
      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1 flex flex-col overflow-hidden">

          {/* LOBBY */}
          {screen === "lobby" && (
            <div className="flex-1 flex items-center justify-center">
              <button
                disabled={startingRun}
                onClick={() => startRun({ variables: { heroId } })}
                className="px-10 py-4 bg-purple-700 hover:bg-purple-600 disabled:opacity-40 rounded-xl text-lg font-bold transition"
              >
                {startingRun ? "Starting..." : "Begin Run"}
              </button>
            </div>
          )}

          {/* MAP */}
          {screen === "map" && graph && (
            <div className="flex-1 overflow-y-auto p-4">
              <MapView
                graph={graph}
                onTravel={(nodeId) => travelToNode({ variables: { heroId, nodeId } })}
                loading={traveling || startingCombat}
              />
              {(traveling || startingCombat) && (
                <div className="text-center text-sm text-gray-500 py-2">Traveling...</div>
              )}
            </div>
          )}

          {screen === "map" && !graph && (
            <div className="flex-1 flex items-center justify-center text-gray-500">
              Loading map...
            </div>
          )}

          {/* COMBAT */}
          {screen === "combat" && activeEncounter && (
            <CombatView
              encounterId={activeEncounter.encounterId}
              heroId={heroId}
              initial={activeEncounter}
              onVictory={handleVictory}
              onDefeat={handleDefeat}
            />
          )}

          {/* SHOP */}
          {screen === "shop" && (
            <ShopView
              heroId={heroId}
              nodeId={activeNodeId ?? ""}
              heroGold={hero?.gold ?? 0}
              onLeave={(updatedGold) => {
                if (updatedGold !== undefined) {
                  addLog(`Left shop. Gold: ${updatedGold}`)
                }
                setScreen("map")
                refetchHero()
                refetchGraph()
              }}
            />
          )}

          {/* EVENT */}
          {screen === "event" && (
            <EventView
              heroId={heroId}
              onDone={({ goldDelta, hpDelta }) => {
                if (goldDelta > 0) addLog(`You gain ${goldDelta} gold!`)
                if (goldDelta < 0) addLog(`You lose ${Math.abs(goldDelta)} gold.`)
                if (hpDelta > 0) addLog(`You recover ${hpDelta} HP.`)
                if (hpDelta < 0) addLog(`You lose ${Math.abs(hpDelta)} HP.`)
                setScreen("map")
                refetchHero()
                refetchGraph()
              }}
            />
          )}

          {/* REST */}
          {screen === "rest" && (
            <div className="flex-1 flex flex-col items-center justify-center gap-4">
              <div className="text-5xl">🔥</div>
              <h2 className="text-2xl font-bold text-green-400">Rest Site</h2>
              <p className="text-gray-400">The fire is warm. You feel your wounds closing.</p>
              <div className="flex gap-3">
                <button
                  onClick={() => {
                    addLog("You rest and recover 30% max HP.")
                    setScreen("map")
                    refetchHero()
                    refetchGraph()
                  }}
                  className="px-6 py-2 bg-green-800 hover:bg-green-700 rounded-lg font-semibold"
                >
                  Rest (heal 30%)
                </button>
                <button
                  onClick={() => {
                    addLog("You meditate, but nothing happens yet.")
                    setScreen("map")
                    refetchGraph()
                  }}
                  className="px-6 py-2 bg-gray-800 hover:bg-gray-700 rounded-lg"
                >
                  Smith (upgrade — coming soon)
                </button>
              </div>
            </div>
          )}

          {/* VICTORY */}
          {screen === "victory" && hero && (
            <VictoryScreen
              heroId={heroId}
              heroName={hero.name}
              heroClass={hero.class}
              floorsCleared={floorsCleared}
              monstersKilled={monstersKilled}
              onPlayAgain={() => {
                setScreen("lobby")
                refetchHero()
              }}
              onMainMenu={() => window.location.reload()}
            />
          )}

          {/* DEAD */}
          {screen === "dead" && (
            <div className="flex-1 flex flex-col items-center justify-center gap-4">
              <div className="text-6xl">💀</div>
              <h2 className="text-3xl font-bold text-red-500">YOU DIED</h2>
              <p className="text-gray-400">The abyss claimed another soul.</p>
              <button
                onClick={() => window.location.reload()}
                className="mt-4 px-8 py-3 bg-red-800 hover:bg-red-700 rounded-xl font-semibold"
              >
                Try Again
              </button>
            </div>
          )}
        </div>

        {/* Right sidebar: hero stats + inventory + log */}
        {(screen === "map" || screen === "lobby") && (
          <div className="w-72 border-l border-gray-800 flex flex-col flex-shrink-0">
            <HeroPanel hero={hero} />
            <div style={{ padding: '8px 12px', borderTop: '1px solid #1f2937' }}>
              <div style={{ color: '#9ca3af', fontSize: '11px', fontWeight: 700, marginBottom: '6px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                Inventory
              </div>
              <InventoryPanel heroId={heroId} onHeroUpdated={refetchHero} />
            </div>
            <CombatLog entries={combatLog} />
          </div>
        )}
      </div>
    </div>
  )
}
