"use client"

import { useMutation } from "@apollo/client/react"
import { SUBMIT_SCORE } from "@/lib/queries"
import { useEffect, useState } from "react"

interface Run {
  id: string
  heroName: string
  floorsCleared: number
  monstersKilled: number
  score: number
}

interface Props {
  heroId: string
  heroName: string
  heroClass: string
  floorsCleared: number
  monstersKilled: number
  onPlayAgain: () => void
  onMainMenu: () => void
}

export default function VictoryScreen({
  heroId,
  heroName,
  heroClass,
  floorsCleared,
  monstersKilled,
  onPlayAgain,
  onMainMenu,
}: Props) {
  const [submitScore] = useMutation(SUBMIT_SCORE)
  const [run, setRun] = useState<Run | null>(null)
  const [submitted, setSubmitted] = useState(false)

  useEffect(() => {
    if (submitted) return
    setSubmitted(true)
    submitScore({ variables: { heroId } })
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .then((res: any) => {
        if (res.data?.submitScore) setRun(res.data.submitScore)
      })
      .catch(() => {/* score submission is best-effort */})
  }, [heroId]) // eslint-disable-line

  return (
    <div className="min-h-screen bg-gray-950 text-white flex flex-col items-center justify-center px-6 py-12">
      {/* Title */}
      <div className="text-center mb-10">
        <div className="text-6xl mb-4">🏆</div>
        <h1 className="text-5xl font-black text-yellow-400 mb-2 tracking-tight">
          VICTORY
        </h1>
        <p className="text-gray-400 text-lg">
          {heroName} the {heroClass} escaped the abyss!
        </p>
      </div>

      {/* Stats card */}
      <div className="bg-gray-900 border border-yellow-500/30 rounded-2xl p-8 w-full max-w-md mb-8 shadow-xl">
        <h2 className="text-xl font-bold text-gray-200 mb-6 text-center">Run Summary</h2>
        <div className="space-y-4">
          <StatRow icon="🗺️" label="Floors Cleared" value={floorsCleared} />
          <StatRow icon="💀" label="Monsters Killed" value={monstersKilled} />
          {run && <StatRow icon="⭐" label="Score" value={run.score} highlight />}
        </div>
        {run && (
          <p className="text-center text-green-400 text-sm mt-6 font-semibold">
            Score submitted to leaderboard!
          </p>
        )}
      </div>

      {/* Actions */}
      <div className="flex gap-4">
        <button
          onClick={onPlayAgain}
          className="bg-yellow-500 hover:bg-yellow-400 text-black font-bold px-8 py-3 rounded-xl text-lg transition"
        >
          Play Again
        </button>
        <button
          onClick={onMainMenu}
          className="bg-gray-700 hover:bg-gray-600 text-white font-semibold px-8 py-3 rounded-xl text-lg transition"
        >
          Main Menu
        </button>
      </div>
    </div>
  )
}

function StatRow({
  icon,
  label,
  value,
  highlight,
}: {
  icon: string
  label: string
  value: number
  highlight?: boolean
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-gray-400 flex items-center gap-2">
        <span>{icon}</span>
        {label}
      </span>
      <span
        className={`font-bold text-lg ${highlight ? "text-yellow-400" : "text-white"}`}
      >
        {value}
      </span>
    </div>
  )
}
