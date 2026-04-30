"use client"
import { useState, useEffect, useCallback } from "react"
import { useMutation, useQuery, useSubscription } from "@apollo/client"
import {
  GET_HERO, START_DUNGEON, MOVE_HERO, ATTACK, GET_FLOOR,
  GET_INVENTORY, USE_ITEM, COMBAT_EVENTS
} from "@/lib/queries"
import DungeonMap from "./DungeonMap"
import HeroPanel from "./HeroPanel"
import CombatLog from "./CombatLog"
import InventoryPanel from "./InventoryPanel"

interface Props {
  heroId: string
}

export default function GameView({ heroId }: Props) {
  const [dungeonId, setDungeonId] = useState<string | null>(null)
  const [floorLevel, setFloorLevel] = useState(1)
  const [combatLog, setCombatLog] = useState<string[]>([])
  const [showInventory, setShowInventory] = useState(false)

  const { data: heroData, refetch: refetchHero } = useQuery(GET_HERO, {
    variables: { id: heroId },
    pollInterval: 2000,
  })

  const { data: floorData } = useQuery(GET_FLOOR, {
    variables: { dungeonId, level: floorLevel },
    skip: !dungeonId,
  })

  const [startDungeon] = useMutation(START_DUNGEON, {
    onCompleted: (data) => {
      setDungeonId(data.startDungeon.dungeonId)
      addLog("You descend into the abyss...")
    },
  })

  const [moveHero] = useMutation(MOVE_HERO, {
    onCompleted: (data) => {
      refetchHero()
      addLog(`You move. HP: ${data.moveHero.hp}/${data.moveHero.maxHp}`)
    },
  })

  const [attack] = useMutation(ATTACK, {
    onCompleted: (data) => {
      const r = data.attack
      addLog(r.message)
      if (r.monsterDied) addLog("You gain XP!")
      if (r.heroDied) addLog("💀 YOU HAVE DIED. Your run is over.")
      refetchHero()
    },
  })

  // Real-time combat events via GraphQL subscription
  useSubscription(COMBAT_EVENTS, {
    variables: { heroId },
    skip: !dungeonId,
    onData: ({ data }) => {
      const event = data.data?.combatEvent
      if (event) addLog(`[live] ${event.message}`)
    },
  })

  const addLog = (msg: string) => {
    const time = new Date().toLocaleTimeString()
    setCombatLog((prev) => [`[${time}] ${msg}`, ...prev].slice(0, 50))
  }

  // Keyboard movement
  const handleKey = useCallback((e: KeyboardEvent) => {
    if (!dungeonId) return
    const dirMap: Record<string, string> = {
      ArrowUp: "north", ArrowDown: "south", ArrowLeft: "west", ArrowRight: "east",
      w: "north", s: "south", a: "west", d: "east",
    }
    const dir = dirMap[e.key]
    if (dir) {
      e.preventDefault()
      moveHero({ variables: { heroId, direction: dir } })
    }
  }, [dungeonId, heroId, moveHero])

  useEffect(() => {
    window.addEventListener("keydown", handleKey)
    return () => window.removeEventListener("keydown", handleKey)
  }, [handleKey])

  const hero = heroData?.hero
  const floor = floorData?.dungeonFloor

  return (
    <div className="flex flex-col h-screen bg-gray-950 text-gray-100 overflow-hidden">
      {/* Top bar */}
      <div className="flex items-center justify-between px-6 py-3 border-b border-gray-800 bg-gray-900">
        <h1 className="text-xl font-bold text-purple-400">AbyssCore</h1>
        <div className="text-sm text-gray-400">
          {hero ? `${hero.name} the ${hero.class} · Floor ${floorLevel}` : "Loading..."}
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setShowInventory(!showInventory)}
            className="px-3 py-1 bg-gray-800 hover:bg-gray-700 rounded text-sm"
          >
            Inventory [I]
          </button>
        </div>
      </div>

      {/* Main game area */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left: dungeon map */}
        <div className="flex-1 flex flex-col items-center justify-center p-4 relative">
          {!dungeonId ? (
            <button
              onClick={() => startDungeon({ variables: { heroId } })}
              className="px-10 py-4 bg-purple-700 hover:bg-purple-600 rounded-xl text-lg font-bold transition"
            >
              Start Run
            </button>
          ) : (
            <>
              <DungeonMap floor={floor} hero={hero} onAttack={(monsterId) =>
                attack({ variables: { heroId, monsterId } })
              } />
              <div className="mt-3 text-xs text-gray-600">
                WASD / Arrow keys to move · Click monster to attack
              </div>
            </>
          )}
        </div>

        {/* Right: hero panel + combat log */}
        <div className="w-80 border-l border-gray-800 flex flex-col">
          <HeroPanel hero={hero} />
          <CombatLog entries={combatLog} />
        </div>
      </div>

      {/* Inventory overlay */}
      {showInventory && (
        <div className="absolute inset-0 bg-black/70 flex items-center justify-center z-50">
          <div className="bg-gray-900 border border-gray-700 rounded-xl p-6 w-[480px] max-h-[70vh] overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-bold text-purple-400">Inventory</h2>
              <button onClick={() => setShowInventory(false)} className="text-gray-500 hover:text-white">✕</button>
            </div>
            <InventoryPanel heroId={heroId} onUse={() => { setShowInventory(false); refetchHero() }} />
          </div>
        </div>
      )}
    </div>
  )
}
