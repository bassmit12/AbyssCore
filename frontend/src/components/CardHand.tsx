"use client"
import type { Card } from "@/types/game"

const TYPE_COLOR: Record<string, string> = {
  ATTACK: "bg-red-900 border-red-700 hover:border-red-500",
  SKILL:  "bg-blue-900 border-blue-700 hover:border-blue-500",
  POWER:  "bg-purple-900 border-purple-700 hover:border-purple-500",
  STATUS: "bg-gray-800 border-gray-600",
  CURSE:  "bg-gray-900 border-red-900",
}

interface Props {
  cards: Card[]
  energy: number
  maxEnergy: number
  onPlay: (cardId: string) => void
  disabled?: boolean
}

export default function CardHand({ cards, energy, maxEnergy, onPlay, disabled }: Props) {
  return (
    <div className="flex flex-col gap-2">
      {/* Energy pip row */}
      <div className="flex items-center gap-1 justify-center mb-1">
        {Array.from({ length: maxEnergy }).map((_, i) => (
          <div
            key={i}
            className={`w-5 h-5 rounded-full border-2 text-[10px] flex items-center justify-center font-bold transition-all
              ${i < energy ? "bg-yellow-400 border-yellow-300 text-gray-900" : "bg-gray-800 border-gray-600 text-gray-600"}`}
          >
            {i < energy ? "●" : "○"}
          </div>
        ))}
        <span className="ml-2 text-xs text-gray-400">{energy}/{maxEnergy} energy</span>
      </div>

      {/* Card row */}
      <div className="flex gap-2 flex-wrap justify-center">
        {cards.map((card) => {
          const canPlay = !disabled && card.cost <= energy
          const colorClass = TYPE_COLOR[card.type] ?? TYPE_COLOR.SKILL

          return (
            <button
              key={card.id}
              disabled={!canPlay}
              onClick={() => onPlay(card.id)}
              className={`
                flex flex-col items-center w-24 rounded-xl border-2 px-2 py-2 text-left transition-all
                ${colorClass}
                ${canPlay ? "cursor-pointer hover:-translate-y-2" : "opacity-40 cursor-not-allowed"}
              `}
              title={card.effect}
            >
              {/* Cost gem */}
              <div className="self-start bg-blue-700 rounded-full w-5 h-5 flex items-center justify-center text-[10px] font-bold mb-1">
                {card.cost}
              </div>
              <span className="text-xs font-semibold text-white text-center leading-tight">{card.name}</span>
              <span className="text-[9px] text-gray-400 mt-1 text-center">{card.type}</span>
            </button>
          )
        })}
        {cards.length === 0 && (
          <div className="text-gray-600 text-sm italic">No cards in hand</div>
        )}
      </div>
    </div>
  )
}
