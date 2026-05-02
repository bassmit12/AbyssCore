"use client"
import type { EncounterMonster } from "@/types/game"

interface Props {
  monsters: EncounterMonster[]
  selectedTarget: string | null
  onSelect: (id: string) => void
}

const INTENT_ICONS: Record<string, string> = {
  attack: "⚔️",
  defend: "🛡️",
  buff:   "✨",
  debuff: "💫",
  heal:   "💚",
}

export default function MonsterRow({ monsters, selectedTarget, onSelect }: Props) {
  return (
    <div className="flex gap-4 justify-center flex-wrap">
      {monsters.map((m) => {
        const isDead = m.hp <= 0
        const isSelected = m.id === selectedTarget
        const hpPct = Math.max(0, (m.hp / m.maxHp) * 100)

        return (
          <div
            key={m.id}
            onClick={() => !isDead && onSelect(m.id)}
            className={`
              flex flex-col items-center gap-1 p-3 rounded-xl border-2 transition-all w-36 cursor-pointer
              ${isDead ? "opacity-30 border-gray-700 grayscale" : ""}
              ${isSelected && !isDead ? "border-red-400 ring-2 ring-red-500 bg-red-950/40" : "border-gray-700 hover:border-gray-500 bg-gray-900"}
            `}
          >
            {/* Intent badges */}
            <div className="flex gap-1 mb-1 min-h-5">
              {m.intents.map((intent, i) => (
                <span key={i} title={`${intent.type} ${intent.value > 0 ? intent.value : ""}`} className="text-sm">
                  {INTENT_ICONS[intent.type] ?? "❓"}
                  {intent.value > 0 && (
                    <span className="text-[9px] text-red-300 ml-0.5">{intent.value}</span>
                  )}
                </span>
              ))}
            </div>

            {/* Monster avatar */}
            <div className="text-4xl select-none">👾</div>

            {/* Name */}
            <div className="text-xs font-semibold text-gray-200 text-center leading-tight">{m.name}</div>

            {/* HP bar */}
            <div className="w-full bg-gray-800 rounded-full h-2 mt-1">
              <div
                className="h-2 rounded-full transition-all bg-red-500"
                style={{ width: `${hpPct}%` }}
              />
            </div>
            <div className="text-[10px] text-gray-400">{m.hp}/{m.maxHp}</div>

            {/* Block */}
            {m.block > 0 && (
              <div className="text-[10px] text-blue-300">🛡️ {m.block}</div>
            )}
          </div>
        )
      })}
    </div>
  )
}
