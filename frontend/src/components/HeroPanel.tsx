"use client"

interface Hero {
  name: string
  class: string
  hp: number
  maxHp: number
  level: number
  xp: number
  alive: boolean
}

interface Props {
  hero?: Hero
}

export default function HeroPanel({ hero }: Props) {
  if (!hero) return <div className="p-4 text-gray-600 text-sm">Loading hero...</div>

  const hpPct = Math.round((hero.hp / hero.maxHp) * 100)
  const xpNeeded = hero.level * 100
  const xpPct = Math.round((hero.xp / xpNeeded) * 100)

  const classColor = {
    warrior: "text-blue-400",
    rogue: "text-green-400",
    mage: "text-purple-400",
  }[hero.class.toLowerCase()] ?? "text-gray-400"

  return (
    <div className="p-4 border-b border-gray-800">
      <div className="flex items-baseline gap-2 mb-3">
        <span className="font-bold text-white">{hero.name}</span>
        <span className={`text-sm capitalize ${classColor}`}>{hero.class}</span>
        <span className="ml-auto text-gray-500 text-sm">Lv {hero.level}</span>
      </div>

      {/* HP bar */}
      <div className="mb-2">
        <div className="flex justify-between text-xs text-gray-500 mb-1">
          <span>HP</span>
          <span>{hero.hp}/{hero.maxHp}</span>
        </div>
        <div className="h-2 bg-gray-800 rounded-full">
          <div
            className={`h-2 rounded-full transition-all ${hpPct > 50 ? "bg-green-500" : hpPct > 25 ? "bg-yellow-500" : "bg-red-500"}`}
            style={{ width: `${hpPct}%` }}
          />
        </div>
      </div>

      {/* XP bar */}
      <div>
        <div className="flex justify-between text-xs text-gray-500 mb-1">
          <span>XP</span>
          <span>{hero.xp}/{xpNeeded}</span>
        </div>
        <div className="h-1.5 bg-gray-800 rounded-full">
          <div
            className="h-1.5 bg-purple-500 rounded-full transition-all"
            style={{ width: `${xpPct}%` }}
          />
        </div>
      </div>

      {!hero.alive && (
        <div className="mt-3 text-center text-red-400 font-bold text-sm border border-red-800 rounded py-1">
          DEAD
        </div>
      )}
    </div>
  )
}
