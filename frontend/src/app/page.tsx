"use client"
import { useSession, signIn, signOut } from "next-auth/react"
import { useState } from "react"
import { useMutation } from "@apollo/client/react"
import { CREATE_HERO } from "@/lib/queries"
import GameView from "@/components/GameView"

export default function Home() {
  const { data: session, status } = useSession()
  const [heroId, setHeroId] = useState<string | null>(null)
  const [heroName, setHeroName] = useState("")
  const [heroClass, setHeroClass] = useState("warrior")

  const [createHero, { loading }] = useMutation(CREATE_HERO, {
    onCompleted: (raw) => setHeroId((raw as { createHero: { id: string } }).createHero.id),
  })

  if (status === "loading") {
    return <div className="flex items-center justify-center h-screen text-gray-400">Loading...</div>
  }

  if (!session) {
    return (
      <div className="flex flex-col items-center justify-center h-screen gap-6">
        <h1 className="text-5xl font-bold tracking-tight text-purple-400">AbyssCore</h1>
        <p className="text-gray-400 text-lg">A dungeon crawler. You will die.</p>
        <button
          onClick={() => signIn("keycloak")}
          className="px-8 py-3 bg-purple-600 hover:bg-purple-500 rounded-lg font-semibold transition"
        >
          Enter the Abyss
        </button>
      </div>
    )
  }

  if (!heroId) {
    return (
      <div className="min-h-screen flex flex-col">
        <header className="flex justify-between items-center px-6 py-3 border-b border-gray-800">
          <span className="text-purple-400 font-semibold">
            AbyssCore
            {process.env.NEXT_PUBLIC_APP_VERSION && (
              <span className="ml-2 text-xs font-mono text-gray-500">
                v{process.env.NEXT_PUBLIC_APP_VERSION.slice(0, 7)}
              </span>
            )}
          </span>
          <div className="flex items-center gap-3 text-sm text-gray-400">
            <span>{session.user?.name ?? session.user?.email}</span>
            <button
              onClick={() => signOut({ callbackUrl: "/" })}
              className="px-3 py-1 bg-gray-800 hover:bg-gray-700 rounded text-gray-300 transition"
            >
              Logout
            </button>
          </div>
        </header>
      <div className="flex flex-col items-center justify-center flex-1 gap-6">
        <h1 className="text-4xl font-bold text-purple-400">Create Your Hero</h1>
        <div className="bg-gray-900 p-8 rounded-xl border border-gray-800 flex flex-col gap-4 w-96">
          <input
            className="bg-gray-800 border border-gray-700 rounded px-4 py-2 text-white focus:outline-none focus:border-purple-500"
            placeholder="Hero name"
            value={heroName}
            onChange={(e) => setHeroName(e.target.value)}
          />
          <div className="grid grid-cols-3 gap-2">
            {["warrior", "rogue", "mage"].map((cls) => (
              <button
                key={cls}
                onClick={() => setHeroClass(cls)}
                className={`py-2 rounded capitalize font-medium transition ${
                  heroClass === cls
                    ? "bg-purple-600 text-white"
                    : "bg-gray-800 text-gray-400 hover:bg-gray-700"
                }`}
              >
                {cls}
              </button>
            ))}
          </div>
          <div className="text-sm text-gray-500">
            {heroClass === "warrior" && "120 HP · 12 ATK · Tanky fighter"}
            {heroClass === "rogue" && "80 HP · 18 ATK · 20% dodge chance"}
            {heroClass === "mage" && "70 HP · 22 ATK · Glass cannon"}
          </div>
          <button
            disabled={!heroName || loading}
            onClick={() => createHero({ variables: { name: heroName, class: heroClass.toUpperCase() } })}
            className="py-3 bg-purple-600 hover:bg-purple-500 disabled:opacity-40 rounded-lg font-semibold transition"
          >
            {loading ? "Creating..." : "Descend"}
          </button>
        </div>
      </div>
      </div>
    )
  }

  return <GameView heroId={heroId} />
}
