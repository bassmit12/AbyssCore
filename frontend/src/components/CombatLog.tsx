"use client"
import { useEffect, useRef } from "react"

interface Props {
  entries: string[]
}

export default function CombatLog({ entries }: Props) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (ref.current) ref.current.scrollTop = 0
  }, [entries])

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="px-4 py-2 border-b border-gray-800 text-xs font-semibold text-gray-500 uppercase tracking-wide">
        Combat Log
      </div>
      <div ref={ref} className="flex-1 overflow-y-auto p-3 space-y-1 font-mono text-xs">
        {entries.length === 0 ? (
          <div className="text-gray-700">No events yet...</div>
        ) : (
          entries.map((entry, i) => (
            <div
              key={i}
              className={`leading-relaxed ${
                entry.includes("DIED") || entry.includes("💀")
                  ? "text-red-400"
                  : entry.includes("XP") || entry.includes("loot")
                  ? "text-yellow-400"
                  : entry.includes("[live]")
                  ? "text-cyan-400"
                  : "text-gray-400"
              }`}
            >
              {entry}
            </div>
          ))
        )}
      </div>
    </div>
  )
}
