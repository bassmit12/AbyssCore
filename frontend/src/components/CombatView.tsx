"use client"
import { useState } from "react"
import { useMutation } from "@apollo/client/react"
import { PLAY_CARD, END_TURN } from "@/lib/queries"
import CardHand from "./CardHand"
import MonsterRow from "./MonsterRow"
import CardReward from "./CardReward"
import type { EncounterState } from "@/types/game"

interface Props {
  encounterId: string
  heroId: string
  initial: EncounterState
  onVictory: () => void
  onDefeat: () => void
}

export default function CombatView({ encounterId, heroId, initial, onVictory, onDefeat }: Props) {
  const [state, setState] = useState<EncounterState>(initial)
  const [selectedTarget, setSelectedTarget] = useState<string | null>(null)
  const [log, setLog] = useState<string[]>([initial.message])
  const [showReward, setShowReward] = useState(false)

  const addLog = (msg: string) => {
    if (!msg) return
    const t = new Date().toLocaleTimeString()
    setLog((prev) => [`[${t}] ${msg}`, ...prev].slice(0, 30))
  }

  const [playCard, { loading: playing }] = useMutation(PLAY_CARD, {
    onCompleted: (raw) => {
      const data = raw as { playCard: EncounterState }
      const next = data.playCard
      setState(next)
      addLog(next.message)
      if (next.status === "won") {
        addLog("Victory! Choose your reward.")
        setShowReward(true)
      } else if (next.status === "lost") {
        addLog("You have fallen...")
        setTimeout(onDefeat, 1500)
      }
    },
    onError: (err) => addLog(`Error: ${err.message}`),
  })

  const [endTurn, { loading: ending }] = useMutation(END_TURN, {
    onCompleted: (raw) => {
      const data = raw as { endTurn: EncounterState }
      const next = data.endTurn
      setState(next)
      addLog(next.message)
      if (next.status === "won") {
        setShowReward(true)
      } else if (next.status === "lost") {
        setTimeout(onDefeat, 1500)
      }
    },
    onError: (err) => addLog(`Error: ${err.message}`),
  })

  const handlePlayCard = (cardId: string) => {
    // If there are multiple monsters and none selected, auto-select first alive
    let target = selectedTarget
    const aliveMonsters = state.monsters.filter((m) => m.hp > 0)
    if (!target && aliveMonsters.length > 0) {
      target = aliveMonsters[0].id
    }
    playCard({
      variables: { encounterId, heroId, cardId, targetId: target },
    })
  }

  const busy = playing || ending
  const hero = state.heroState
  const hpPct = Math.max(0, (hero.hp / hero.maxHp) * 100)

  if (showReward) {
    return (
      <CardReward
        encounterId={encounterId}
        heroId={heroId}
        onDone={onVictory}
      />
    )
  }

  return (
    <div className="flex flex-col h-full gap-3 p-4 bg-gray-950">
      {/* Header: turn + status */}
      <div className="flex items-center justify-between text-sm text-gray-500">
        <span>Turn {state.turnNumber}</span>
        <span className={`font-semibold ${
          state.status === "active" ? "text-green-500" :
          state.status === "won" ? "text-yellow-400" : "text-red-400"
        }`}>{state.status.toUpperCase()}</span>
      </div>

      {/* Monsters */}
      <div className="flex-shrink-0 py-4">
        <MonsterRow
          monsters={state.monsters}
          selectedTarget={selectedTarget}
          onSelect={setSelectedTarget}
        />
      </div>

      {/* Divider */}
      <div className="border-t border-gray-800" />

      {/* Hero status bar */}
      <div className="flex gap-4 items-center px-2">
        <div className="flex-1">
          <div className="flex justify-between text-xs text-gray-400 mb-0.5">
            <span>HP</span>
            <span>{hero.hp}/{hero.maxHp}</span>
          </div>
          <div className="bg-gray-800 rounded-full h-3">
            <div
              className="h-3 rounded-full transition-all bg-green-600"
              style={{ width: `${hpPct}%` }}
            />
          </div>
        </div>
        <div className="text-sm text-blue-300 flex items-center gap-1">
          🛡️ <span>{hero.block}</span>
        </div>
        {/* Statuses */}
        <div className="flex gap-1">
          {hero.statuses.map((s) => (
            <span key={s.name} className="text-xs bg-gray-800 px-1.5 py-0.5 rounded-full text-gray-300"
              title={s.name}>
              {s.name} {s.stacks > 1 ? `×${s.stacks}` : ""}
            </span>
          ))}
        </div>
        {/* Pile counts */}
        <div className="text-[11px] text-gray-600 flex gap-2">
          <span title="Draw pile">🃏{hero.drawPileCount}</span>
          <span title="Discard">🗑️{hero.discardPileCount}</span>
        </div>
      </div>

      {/* Card hand */}
      <div className="flex-1 flex flex-col justify-end pb-2">
        <CardHand
          cards={hero.hand}
          energy={hero.energy}
          maxEnergy={hero.maxEnergy}
          onPlay={handlePlayCard}
          disabled={busy || state.status !== "active"}
        />
      </div>

      {/* End Turn */}
      <button
        disabled={busy || state.status !== "active"}
        onClick={() => endTurn({ variables: { encounterId, heroId } })}
        className="py-2 rounded-lg bg-yellow-700 hover:bg-yellow-600 disabled:opacity-40 font-semibold text-white transition"
      >
        {ending ? "..." : "End Turn"}
      </button>

      {/* Combat log */}
      <div className="bg-gray-900 rounded-lg border border-gray-800 p-2 max-h-24 overflow-y-auto text-xs text-gray-400">
        {log.map((entry, i) => (
          <div key={i}>{entry}</div>
        ))}
      </div>
    </div>
  )
}
