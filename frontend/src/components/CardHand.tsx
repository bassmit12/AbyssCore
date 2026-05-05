"use client"
import type { Card } from "@/types/game"

const TYPE_COLOR: Record<string, string> = {
  ATTACK: "bg-red-900 border-red-700 hover:border-red-500",
  SKILL:  "bg-blue-900 border-blue-700 hover:border-blue-500",
  POWER:  "bg-purple-900 border-purple-700 hover:border-purple-500",
  STATUS: "bg-gray-800 border-gray-600",
  CURSE:  "bg-gray-900 border-red-900",
}

// Cards that have artwork — keyed by lowercase name
const CARD_ART: Record<string, string> = {
  "ball lightning": "/cards/ball_lightning.png",
  "chaos theory":   "/cards/chaos_theory.png",
}

// Only these cards are playable for mage right now
const ENABLED_CARDS = new Set(["ball lightning", "chaos theory"])

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
          const key = card.name.toLowerCase()
          const isEnabled = ENABLED_CARDS.has(key)
          const artSrc = CARD_ART[key]
          const canPlay = !disabled && isEnabled && card.cost <= energy
          const colorClass = TYPE_COLOR[card.type.toUpperCase()] ?? TYPE_COLOR.SKILL

          return (
            <button
              key={card.id}
              disabled={!canPlay}
              onClick={() => onPlay(card.id)}
              title={isEnabled ? card.effect : "Not available"}
              className={`
                relative w-32 h-48 rounded-lg overflow-hidden transition-all
                ${canPlay ? "cursor-pointer hover:-translate-y-3 shadow-lg hover:shadow-2xl" : "opacity-40 cursor-not-allowed"}
              `}
            >
              {/* Artwork or placeholder */}
              {artSrc ? (
                /* eslint-disable-next-line @next/next/no-img-element */
                <img
                  src={artSrc}
                  alt={card.name}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className={`w-full h-full flex flex-col items-center justify-center gap-1 border-2 rounded-lg ${colorClass}`}>
                  <span className="text-3xl select-none">
                    {card.type.toUpperCase() === "ATTACK" ? "⚔️" : card.type.toUpperCase() === "SKILL" ? "✨" : "🌀"}
                  </span>
                  <span className="text-xs font-semibold text-white text-center px-1 leading-tight">{card.name}</span>
                  <span className="text-[9px] text-gray-400">{card.type}</span>
                  {/* Cost gem on placeholder */}
                  <div className="absolute top-1.5 left-1.5 bg-blue-700 rounded-full w-5 h-5 flex items-center justify-center text-[10px] font-bold shadow">
                    {card.cost}
                  </div>
                </div>
              )}
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
